package mixin

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func gzipCompress(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	if _, err := w.Write(data); err != nil {
		t.Fatalf("gzip write: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("gzip close: %v", err)
	}
	return buf.Bytes()
}

func gzipJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return gzipCompress(t, data)
}

type testSigner struct{}

func (testSigner) SignToken(signature, requestID string, exp time.Duration) string {
	return "test-token"
}

func (testSigner) EncryptPin(pin string) string { return pin }

type testMessageLocker struct {
	raw       []byte
	unlockErr error
}

func (l *testMessageLocker) Lock(data []byte, _ []*Session) ([]byte, error) {
	return data, nil
}

func (l *testMessageLocker) Unlock(data []byte) ([]byte, error) {
	if l.unlockErr != nil {
		return nil, l.unlockErr
	}
	return l.raw, nil
}

type testBlazeListener struct {
	onMessage    func(ctx context.Context, msg *MessageView, userID string) error
	onAckReceipt func(ctx context.Context, msg *MessageView, userID string) error
}

func (l *testBlazeListener) OnMessage(ctx context.Context, msg *MessageView, userID string) error {
	if l.onMessage != nil {
		return l.onMessage(ctx, msg, userID)
	}
	return nil
}

func (l *testBlazeListener) OnAckReceipt(ctx context.Context, msg *MessageView, userID string) error {
	if l.onAckReceipt != nil {
		return l.onAckReceipt(ctx, msg, userID)
	}
	return nil
}

func TestParseBlazeMessage(t *testing.T) {
	t.Run("resets reused message fields between frames", func(t *testing.T) {
		var msg BlazeMessage

		first := gzipJSON(t, map[string]any{
			"id":     "frame-1",
			"action": "CREATE_MESSAGE",
			"data":   map[string]any{"message_id": "m1"},
			"error": map[string]any{
				"status":      500,
				"code":        500,
				"description": "boom",
			},
		})
		if err := parseBlazeMessage(bytes.NewReader(first), &msg); err != nil {
			t.Fatalf("parse first: %v", err)
		}
		if msg.Error == nil || msg.Data == nil {
			t.Fatalf("first frame should keep error and data: %+v", msg)
		}

		second := gzipJSON(t, map[string]any{
			"id":     "frame-2",
			"action": "LIST_PENDING_MESSAGES",
		})
		if err := parseBlazeMessage(bytes.NewReader(second), &msg); err != nil {
			t.Fatalf("parse second: %v", err)
		}
		if msg.Id != "frame-2" || msg.Action != "LIST_PENDING_MESSAGES" {
			t.Fatalf("unexpected second frame identity: %+v", msg)
		}
		if msg.Error != nil {
			t.Fatalf("second frame Error should be nil, got %#v", msg.Error)
		}
		if msg.Data != nil {
			t.Fatalf("second frame Data should be nil, got %s", msg.Data)
		}
	})

	t.Run("rejects multiple JSON values", func(t *testing.T) {
		payload := append(
			[]byte(`{"id":"one","action":"CREATE_MESSAGE"}`),
			[]byte(`{"id":"two","action":"CREATE_MESSAGE"}`)...,
		)
		err := parseBlazeMessage(bytes.NewReader(gzipCompress(t, payload)), &BlazeMessage{})
		if err == nil || err.Error() != "blaze message contains multiple JSON values" {
			t.Fatalf("expected multiple JSON values error, got %v", err)
		}
	})

	t.Run("validates gzip trailer checksum and truncation", func(t *testing.T) {
		valid := gzipJSON(t, map[string]any{
			"id":     "crc",
			"action": "CREATE_MESSAGE",
		})

		corrupted := append([]byte(nil), valid...)
		crcOffset := len(corrupted) - 8
		corrupted[crcOffset] ^= 0xff
		err := parseBlazeMessage(bytes.NewReader(corrupted), &BlazeMessage{})
		if !errors.Is(err, gzip.ErrChecksum) {
			t.Fatalf("expected gzip.ErrChecksum, got %v", err)
		}

		truncated := valid[:len(valid)-4]
		err = parseBlazeMessage(bytes.NewReader(truncated), &BlazeMessage{})
		if err == nil {
			t.Fatal("expected truncated trailer error")
		}
	})
}

func TestBlazeHandlerHandleMessage(t *testing.T) {
	ctx := context.Background()

	t.Run("isolates message views across create frames", func(t *testing.T) {
		b := &blazeHandler{Client: &Client{ClientID: "bot"}}
		var seen []*MessageView
		listener := &testBlazeListener{
			onMessage: func(_ context.Context, msg *MessageView, _ string) error {
				seen = append(seen, msg)
				msg.Ack()
				return nil
			},
		}

		firstData, _ := json.Marshal(map[string]any{
			"message_id": "m1",
			"source":     "LIST_PENDING_MESSAGES",
			"status":     "SENT",
		})
		secondData, _ := json.Marshal(map[string]any{
			"message_id": "m2",
		})

		if err := b.handleMessage(ctx, listener, &BlazeMessage{
			Id: "e1", Action: CreateMessageAction, Data: firstData,
		}); err != nil {
			t.Fatalf("first handle: %v", err)
		}
		if err := b.handleMessage(ctx, listener, &BlazeMessage{
			Id: "e2", Action: CreateMessageAction, Data: secondData,
		}); err != nil {
			t.Fatalf("second handle: %v", err)
		}

		if len(seen) != 2 {
			t.Fatalf("expected 2 messages, got %d", len(seen))
		}
		if seen[0] == seen[1] {
			t.Fatal("message view pointers must differ")
		}
		if seen[0].Source != "LIST_PENDING_MESSAGES" || seen[0].Status != "SENT" {
			t.Fatalf("first message mutated: %+v", seen[0])
		}
		if seen[1].Source != "" || seen[1].Status != "" {
			t.Fatalf("second message should omit source/status: %+v", seen[1])
		}
	})

	t.Run("create acknowledgement and listener errors", func(t *testing.T) {
		sentinel := errors.New("listener failed")

		t.Run("queues original message id", func(t *testing.T) {
			b := &blazeHandler{Client: &Client{ClientID: "bot"}}
			listener := &testBlazeListener{
				onMessage: func(_ context.Context, msg *MessageView, _ string) error {
					msg.MessageID = "mutated"
					return nil
				},
			}
			data, _ := json.Marshal(map[string]any{"message_id": "original-id"})
			if err := b.handleMessage(ctx, listener, &BlazeMessage{
				Id: "e1", Action: CreateMessageAction, Data: data,
			}); err != nil {
				t.Fatalf("handle: %v", err)
			}
			got := b.queue.pull(10)
			if len(got) != 1 || got[0].MessageID != "original-id" || got[0].Status != MessageStatusRead {
				t.Fatalf("unexpected queue: %+v", got)
			}
		})

		t.Run("skips queue when Ack called", func(t *testing.T) {
			b := &blazeHandler{Client: &Client{ClientID: "bot"}}
			listener := &testBlazeListener{
				onMessage: func(_ context.Context, msg *MessageView, _ string) error {
					msg.Ack()
					return nil
				},
			}
			data, _ := json.Marshal(map[string]any{"message_id": "acked-id"})
			if err := b.handleMessage(ctx, listener, &BlazeMessage{
				Id: "e1", Action: CreateMessageAction, Data: data,
			}); err != nil {
				t.Fatalf("handle: %v", err)
			}
			if got := b.queue.pull(10); len(got) != 0 {
				t.Fatalf("expected empty queue, got %+v", got)
			}
		})

		t.Run("returns listener error without queueing", func(t *testing.T) {
			b := &blazeHandler{Client: &Client{ClientID: "bot"}}
			listener := &testBlazeListener{
				onMessage: func(context.Context, *MessageView, string) error {
					return sentinel
				},
			}
			data, _ := json.Marshal(map[string]any{"message_id": "err-id"})
			err := b.handleMessage(ctx, listener, &BlazeMessage{
				Id: "e1", Action: CreateMessageAction, Data: data,
			})
			if !errors.Is(err, sentinel) {
				t.Fatalf("expected sentinel, got %v", err)
			}
			if got := b.queue.pull(10); len(got) != 0 {
				t.Fatalf("expected empty queue, got %+v", got)
			}
		})
	})

	t.Run("rejects bad create and ignores unknown actions", func(t *testing.T) {
		called := false
		listener := &testBlazeListener{
			onMessage: func(context.Context, *MessageView, string) error {
				called = true
				return nil
			},
			onAckReceipt: func(context.Context, *MessageView, string) error {
				called = true
				return nil
			},
		}

		b := &blazeHandler{Client: &Client{ClientID: "bot"}}
		data, _ := json.Marshal(map[string]any{"category": "PLAIN_TEXT"})
		err := b.handleMessage(ctx, listener, &BlazeMessage{
			Id: "env-missing", Action: CreateMessageAction, Data: data,
		})
		want := `blaze action "CREATE_MESSAGE" id "env-missing" missing message_id`
		if err == nil || err.Error() != want {
			t.Fatalf("missing message_id: got %v", err)
		}
		if called {
			t.Fatal("listener should not be called for missing message_id")
		}
		if got := b.queue.pull(10); len(got) != 0 {
			t.Fatalf("queue should stay empty, got %+v", got)
		}

		called = false
		b = &blazeHandler{Client: &Client{ClientID: "bot"}}
		err = b.handleMessage(ctx, listener, &BlazeMessage{
			Id: "env-bad", Action: CreateMessageAction, Data: json.RawMessage(`{`),
		})
		if err == nil || !strings.Contains(err.Error(), `decode blaze action "CREATE_MESSAGE" id "env-bad" data:`) {
			t.Fatalf("malformed data error = %v", err)
		}
		if called {
			t.Fatal("listener should not be called for malformed data")
		}

		called = false
		b = &blazeHandler{Client: &Client{ClientID: "bot"}}
		if err := b.handleMessage(ctx, listener, &BlazeMessage{
			Id: "env-unknown", Action: "FUTURE_ACTION", Data: nil,
		}); err != nil {
			t.Fatalf("unknown action should be ignored: %v", err)
		}
		if called {
			t.Fatal("listener should not be called for unknown action")
		}
	})

	t.Run("receipt and encrypted create paths", func(t *testing.T) {
		t.Run("acknowledge receipt only", func(t *testing.T) {
			b := &blazeHandler{Client: &Client{ClientID: "bot"}}
			var receiptCalled, messageCalled bool
			listener := &testBlazeListener{
				onMessage: func(context.Context, *MessageView, string) error {
					messageCalled = true
					return nil
				},
				onAckReceipt: func(_ context.Context, msg *MessageView, userID string) error {
					receiptCalled = true
					if msg.MessageID != "r1" || userID != "bot" {
						t.Fatalf("unexpected receipt: msg=%+v userID=%s", msg, userID)
					}
					return nil
				},
			}
			data, _ := json.Marshal(map[string]any{"message_id": "r1", "status": "READ"})
			if err := b.handleMessage(ctx, listener, &BlazeMessage{
				Id: "e1", Action: AcknowledgeReceiptAction, Data: data,
			}); err != nil {
				t.Fatalf("handle: %v", err)
			}
			if !receiptCalled || messageCalled {
				t.Fatalf("receipt=%v message=%v", receiptCalled, messageCalled)
			}
			if got := b.queue.pull(10); len(got) != 0 {
				t.Fatalf("receipt should not enqueue ack: %+v", got)
			}
		})

		t.Run("decrypts encrypted create", func(t *testing.T) {
			raw := []byte("hello-plain")
			b := &blazeHandler{Client: &Client{
				ClientID:      "bot",
				MessageLocker: &testMessageLocker{raw: raw},
			}}
			var got *MessageView
			listener := &testBlazeListener{
				onMessage: func(_ context.Context, msg *MessageView, _ string) error {
					got = msg
					msg.Ack()
					return nil
				},
			}
			data, _ := json.Marshal(map[string]any{
				"message_id":  "enc-1",
				"category":    "ENCRYPTED_TEXT",
				"data_base64": base64.RawURLEncoding.EncodeToString([]byte("cipher")),
			})
			if err := b.handleMessage(ctx, listener, &BlazeMessage{
				Id: "e1", Action: CreateMessageAction, Data: data,
			}); err != nil {
				t.Fatalf("handle: %v", err)
			}
			if got == nil {
				t.Fatal("expected message")
			}
			if got.Category != "PLAIN_TEXT" {
				t.Fatalf("category=%q", got.Category)
			}
			if got.DataBase64 != base64.RawURLEncoding.EncodeToString(raw) {
				t.Fatalf("data_base64=%q", got.DataBase64)
			}
			if got.Data != base64.StdEncoding.EncodeToString(raw) {
				t.Fatalf("data=%q", got.Data)
			}
		})

		t.Run("wraps invalid encrypted base64", func(t *testing.T) {
			b := &blazeHandler{Client: &Client{
				ClientID:      "bot",
				MessageLocker: &testMessageLocker{raw: []byte("x")},
			}}
			data, _ := json.Marshal(map[string]any{
				"message_id":  "enc-bad",
				"category":    "ENCRYPTED_TEXT",
				"data_base64": "!!!",
			})
			err := b.handleMessage(ctx, &testBlazeListener{}, &BlazeMessage{
				Id: "e1", Action: CreateMessageAction, Data: data,
			})
			if err == nil || !strings.HasPrefix(err.Error(), `decode encrypted blaze message "enc-bad":`) {
				t.Fatalf("error = %v", err)
			}
		})

		t.Run("wraps unlock errors", func(t *testing.T) {
			unlockErr := errors.New("unlock boom")
			b := &blazeHandler{Client: &Client{
				ClientID:      "bot",
				MessageLocker: &testMessageLocker{unlockErr: unlockErr},
			}}
			data, _ := json.Marshal(map[string]any{
				"message_id":  "enc-unlock",
				"category":    "ENCRYPTED_TEXT",
				"data_base64": base64.RawURLEncoding.EncodeToString([]byte("cipher")),
			})
			err := b.handleMessage(ctx, &testBlazeListener{}, &BlazeMessage{
				Id: "e1", Action: CreateMessageAction, Data: data,
			})
			if !errors.Is(err, unlockErr) {
				t.Fatalf("expected unlock sentinel, got %v", err)
			}
			if !strings.HasPrefix(err.Error(), `decrypt blaze message "enc-unlock":`) {
				t.Fatalf("error = %v", err)
			}
		})
	})
}

func TestClientLoopBlazeHandlesFramesIndependently(t *testing.T) {
	prev := blazeURL
	defer func() { blazeURL = prev }()

	sentinel := errors.New("stop after second frame")
	var serverErr error
	upgrader := websocket.Upgrader{
		Subprotocols: []string{"Mixin-Blaze-1"},
		CheckOrigin:  func(*http.Request) bool { return true },
	}

	frameBytes := [][]byte{
		mustJSON(t, map[string]any{
			"id":     "frame-1",
			"action": CreateMessageAction,
			"data": map[string]any{
				"message_id": "m1",
				"source":     "LIST_PENDING_MESSAGES",
				"status":     "SENT",
			},
		}),
		mustJSON(t, map[string]any{
			"id":     "frame-2",
			"action": CreateMessageAction,
			"data": map[string]any{
				"message_id": "m2",
			},
		}),
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			serverErr = err
			return
		}
		defer conn.Close()

		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		typ, reader, err := conn.NextReader()
		if err != nil {
			serverErr = err
			return
		}
		if typ != websocket.BinaryMessage {
			serverErr = fmt.Errorf("expected binary, got %d", typ)
			return
		}
		var pending BlazeMessage
		if err := parseBlazeMessage(reader, &pending); err != nil {
			serverErr = err
			return
		}
		if pending.Action != "LIST_PENDING_MESSAGES" {
			serverErr = fmt.Errorf("unexpected initial action %q", pending.Action)
			return
		}

		for _, frame := range frameBytes {
			if err := writeGzipToConn(conn, frame); err != nil {
				serverErr = err
				return
			}
		}

		// Keep connection open until client disconnects.
		_, _, _ = conn.NextReader()
	}))
	defer srv.Close()

	UseBlazeURL("ws" + strings.TrimPrefix(srv.URL, "http"))

	c := &Client{
		ClientID:      "bot",
		Signer:        testSigner{},
		MessageLocker: &messageLockNotSupported{},
	}

	var seen []*MessageView
	listener := &testBlazeListener{
		onMessage: func(_ context.Context, msg *MessageView, _ string) error {
			seen = append(seen, msg)
			msg.Ack()
			if len(seen) == 2 {
				return sentinel
			}
			return nil
		},
	}

	err := c.LoopBlaze(context.Background(), listener)
	if !errors.Is(err, sentinel) {
		t.Fatalf("LoopBlaze error = %v", err)
	}
	if serverErr != nil {
		t.Fatalf("server protocol error: %v", serverErr)
	}
	if len(seen) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(seen))
	}
	if seen[0] == seen[1] {
		t.Fatal("saved message pointers must differ")
	}
	if seen[0].MessageID != "m1" || seen[0].Source != "LIST_PENDING_MESSAGES" || seen[0].Status != "SENT" {
		t.Fatalf("first message: %+v", seen[0])
	}
	if seen[1].MessageID != "m2" || seen[1].Source != "" || seen[1].Status != "" {
		t.Fatalf("second message: %+v", seen[1])
	}
}

func TestClientLoopBlazeRejectsNilListener(t *testing.T) {
	optionCalled := false
	err := (&Client{Signer: testSigner{}}).LoopBlaze(context.Background(), nil, func(*websocket.Dialer) {
		optionCalled = true
	})
	if err == nil || err.Error() != "mixin: nil BlazeListener" {
		t.Fatalf("error = %v", err)
	}
	if optionCalled {
		t.Fatal("option should not run before nil listener check")
	}
}

func TestAckQueuePushFrontPreservesOrder(t *testing.T) {
	var q AckQueue
	q.pushBack(&AcknowledgementRequest{MessageID: "tail"})
	q.pushFront(
		&AcknowledgementRequest{MessageID: "a"},
		&AcknowledgementRequest{MessageID: "b"},
		&AcknowledgementRequest{MessageID: "c"},
	)

	got := q.pull(10)
	want := []string{"a", "b", "c", "tail"}
	if len(got) != len(want) {
		t.Fatalf("len=%d got=%v", len(got), idsOf(got))
	}
	for i := range want {
		if got[i].MessageID != want[i] {
			t.Fatalf("order=%v want=%v", idsOf(got), want)
		}
	}
}

func mustJSON(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return data
}

func idsOf(reqs []*AcknowledgementRequest) []string {
	out := make([]string, len(reqs))
	for i, req := range reqs {
		out[i] = req.MessageID
	}
	return out
}

func TestClient_LoopBlaze(t *testing.T) {
	store := newKeystoreFromEnv(t)
	c, err := NewFromKeystore(&store.Keystore)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = c.LoopBlaze(ctx, BlazeListenFunc(func(ctx context.Context, msg *MessageView, userID string) error {
		t.Log(msg.Category, msg.Data)
		return nil
	}))

	t.Log(err)
}
