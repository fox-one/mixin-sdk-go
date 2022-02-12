package mixin

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
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
	ConversationID string `json:"conversation_id"`
	UserID         string `json:"user_id"`
	MessageID      string `json:"message_id"`
	Category       string `json:"category"`
	Data           string `json:"data"`
	// DataBase64 is same as Data but encoded by base64.RawURLEncoding
	DataBase64       string    `json:"data_base64"`
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

type BlazeOption func(dialer *websocket.Dialer)

func (c *Client) LoopBlaze(ctx context.Context, listener BlazeListener, opts ...BlazeOption) error {
	conn, err := connectMixinBlaze(c, opts...)
	if err != nil {
		return err
	}

	defer conn.Close()

	b := &blazeHandler{
		Client: c,
	}

	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(s string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	if err = writeMessage(conn, "LIST_PENDING_MESSAGES", nil); err != nil {
		return fmt.Errorf("write LIST_PENDING_MESSAGES failed: %w", err)
	}

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return tick(ctx, conn)
	})

	g.Go(func() error {
		return b.ack(ctx)
	})

	g.Go(func() error {
		var (
			blazeMessage BlazeMessage
			message      MessageView
		)

		for {
			_ = conn.SetReadDeadline(time.Now().Add(pongWait))
			typ, r, err := conn.NextReader()
			if err != nil {
				if ctxErr := ctx.Err(); ctxErr != nil {
					return ctxErr
				}

				return err
			}

			if typ != websocket.BinaryMessage {
				return fmt.Errorf("invalid message type %d", typ)
			}

			if err := parseBlazeMessage(r, &blazeMessage); err != nil {
				return err
			}

			if err := blazeMessage.Error; err != nil {
				return err
			}

			message.reset()
			if err := json.Unmarshal(blazeMessage.Data, &message); err != nil {
				continue
			}

			if IsEncryptedMessageCategory(message.Category) {
				data, err := base64.RawURLEncoding.DecodeString(message.DataBase64)
				if err != nil {
					return err
				}

				rawData, err := c.Unlock(data)
				if err != nil {
					return err
				}

				message.Category = DecryptMessageCategory(message.Category)
				message.DataBase64 = base64.RawURLEncoding.EncodeToString(rawData)
				message.Data = base64.StdEncoding.EncodeToString(rawData)
			}

			switch blazeMessage.Action {
			case CreateMessageAction:
				messageID := message.MessageID
				if err := listener.OnMessage(ctx, &message, b.ClientID); err != nil {
					return err
				}

				if !message.ack {
					b.queue.pushBack(&AcknowledgementRequest{
						MessageID: messageID,
						Status:    MessageStatusRead,
					})
				}
			case AcknowledgeReceiptAction:
				if err := listener.OnAckReceipt(ctx, &message, b.ClientID); err != nil {
					return err
				}
			}
		}
	})

	return g.Wait()
}

func connectMixinBlaze(s Signer, opts ...BlazeOption) (*websocket.Conn, error) {
	sig := SignRaw("GET", "/", nil)
	token := s.SignToken(sig, newRequestID(), time.Minute)
	header := make(http.Header)
	header.Add("Authorization", "Bearer "+token)

	dialer := &websocket.Dialer{
		Subprotocols:   []string{"Mixin-Blaze-1"},
		ReadBufferSize: 1024,
	}

	for _, opt := range opts {
		opt(dialer)
	}

	conn, _, err := dialer.Dial(blazeURL, header)
	if err != nil {
		return nil, err
	}

	// no limit
	conn.SetReadLimit(0)
	return conn, nil
}

func tick(ctx context.Context, conn *websocket.Conn) error {
	// Call conn.Close() to close the connection immediately when func return.
	// If not called, the Loop process may get stuck on conn.NextReader().
	defer conn.Close()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pingPeriod):
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return fmt.Errorf("send ping message failed: %w", err)
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
