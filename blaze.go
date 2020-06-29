package mixin

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 10 * time.Second
	pingPeriod = pongWait * 8 / 10

	ackBatch = 80

	CreateMessageAction      = "CREATE_MESSAGE"
	AcknowledgeReceiptAction = "ACKNOWLEDGE_MESSAGE_RECEIPT"
)

type BlazeMessage struct {
	Id     string                 `json:"id"`
	Action string                 `json:"action"`
	Params map[string]interface{} `json:"params,omitempty"`
	Data   json.RawMessage        `json:"data,omitempty"`
	Error  *Error                 `json:"error,omitempty"`
}

type MessageView struct {
	ConversationID   string    `json:"conversation_id"`
	UserID           string    `json:"user_id"`
	MessageID        string    `json:"message_id"`
	Category         string    `json:"category"`
	Data             string    `json:"data"`
	RepresentativeID string    `json:"representative_id"`
	QuoteMessageID   string    `json:"quote_message_id"`
	Status           string    `json:"status"`
	Source           string    `json:"source"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`

	// ack status
	ack bool
}

func (m *MessageView) reset() {
	m.ack = false
	m.RepresentativeID = ""
	m.QuoteMessageID = ""
}

// Ack mark messageView as acked
// otherwise sdk will ack this message
func (m *MessageView) Ack() {
	m.ack = true
}

type TransferView struct {
	Type          string    `json:"type"`
	SnapshotID    string    `json:"snapshot_id"`
	CounterUserID string    `json:"counter_user_id"`
	AssetID       string    `json:"asset_id"`
	Amount        string    `json:"amount"`
	TraceID       string    `json:"trace_id"`
	Memo          string    `json:"memo"`
	CreatedAt     time.Time `json:"created_at"`
}

type SystemConversationPayload struct {
	Action        string `json:"action"`
	ParticipantID string `json:"participant_id"`
	UserID        string `json:"user_id,omitempty"`
	Role          string `json:"role,omitempty"`
}

type BlazeListener interface {
	OnAckReceipt(ctx context.Context, msg *MessageView, userID string) error
	OnMessage(ctx context.Context, msg *MessageView, userID string) error
}

func (c *Client) LoopBlaze(ctx context.Context, listener BlazeListener) error {
	conn, err := connectMixinBlaze(c)
	if err != nil {
		return err
	}
	defer conn.Close()

	b := &blazeHandler{
		Client: c,
		conn:   conn,
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go tick(ctx, conn)

	ackBuffer := make(chan string)
	defer close(ackBuffer)

	go b.ack(ctx, ackBuffer)

	_ = b.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return b.SetReadDeadline(time.Now().Add(pongWait))
	})

	if err = writeMessage(conn, "LIST_PENDING_MESSAGES", nil); err != nil {
		return fmt.Errorf("write LIST_PENDING_MESSAGES failed: %w", err)
	}

	var (
		blazeMessage BlazeMessage
		message      MessageView
	)

	for {
		typ, r, err := conn.NextReader()
		if err != nil {
			return err
		}

		if typ != websocket.BinaryMessage {
			return fmt.Errorf("invalid message type %d", typ)
		}

		if err := parseBlazeMessage(r, &blazeMessage); err != nil {
			return err
		}

		if blazeMessage.Error != nil {
			return err
		}

		message.reset()
		if err := json.Unmarshal(blazeMessage.Data, &message); err != nil {
			continue
		}

		switch blazeMessage.Action {
		case CreateMessageAction:
			messageID := message.MessageID
			if err := listener.OnMessage(ctx, &message, b.ClientID); err != nil {
				return err
			}

			if !message.ack {
				ackBuffer <- messageID
			}
		case AcknowledgeReceiptAction:
			if err := listener.OnAckReceipt(ctx, &message, b.ClientID); err != nil {
				return err
			}
		}

		if time.Until(b.readDeadline) < time.Second {
			// 可能因为收到的消息过多或者消息处理太慢或者 ack 太慢
			// 导致没有及时处理 pong frame 而 read deadline 没有刷新
			// 这种情况下不应该读超时，在这里重置一下 read deadline
			_ = b.SetReadDeadline(time.Now().Add(pongWait))
		}
	}
}

func connectMixinBlaze(s Signer) (*websocket.Conn, error) {
	sig := SignRaw("GET", "/", nil)
	token := s.SignToken(sig, newRequestID(), time.Minute)
	header := make(http.Header)
	header.Add("Authorization", "Bearer "+token)
	u := url.URL{Scheme: "wss", Host: blazeHost, Path: "/"}
	dialer := &websocket.Dialer{
		Subprotocols:   []string{"Mixin-Blaze-1"},
		ReadBufferSize: 1024,
	}
	conn, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		return nil, err
	}

	// no limit
	conn.SetReadLimit(0)
	return conn, nil
}

func tick(ctx context.Context, conn *websocket.Conn) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		conn.Close()
		ticker.Stop()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}

func writeMessage(coon *websocket.Conn, action string, params map[string]interface{}) error {
	blazeMessage, err := json.Marshal(BlazeMessage{
		Id:     newUUID(),
		Action: action,
		Params: params,
	})
	if err != nil {
		return err
	}

	if err := writeGzipToConn(coon, blazeMessage); err != nil {
		return err
	}

	return nil
}

func writeGzipToConn(conn *websocket.Conn, msg []byte) error {
	if err := conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		return err
	}

	wsWriter, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return err
	}
	gzWriter, err := gzip.NewWriterLevel(wsWriter, 3)
	if err != nil {
		return err
	}
	if _, err := gzWriter.Write(msg); err != nil {
		return err
	}

	if err := gzWriter.Close(); err != nil {
		return err
	}
	return wsWriter.Close()
}

func parseBlazeMessage(r io.Reader, msg *BlazeMessage) error {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return err
	}

	err = json.NewDecoder(gzReader).Decode(msg)
	_ = gzReader.Close()
	return err
}

type BlazeListenFunc func(ctx context.Context, msg *MessageView, userID string) error

func (f BlazeListenFunc) OnAckReceipt(ctx context.Context, msg *MessageView, userID string) error {
	return nil
}

func (f BlazeListenFunc) OnMessage(ctx context.Context, msg *MessageView, userID string) error {
	return f(ctx, msg, userID)
}
