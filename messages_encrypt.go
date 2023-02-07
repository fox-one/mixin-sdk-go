package mixin

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/gofrs/uuid"
	"golang.org/x/crypto/curve25519"
)

const (
	encryptedCategoryLabel = "ENCRYPTED_"
	plainCategoryLabel     = "PLAIN_"
)

type MessageLocker interface {
	Lock(data []byte, sessions []*Session) ([]byte, error)
	Unlock(data []byte) ([]byte, error)
}

func IsEncryptedMessageCategory(category string) bool {
	return strings.Contains(category, encryptedCategoryLabel)
}

func IsPlainMessageCategory(category string) bool {
	return strings.Contains(category, plainCategoryLabel)
}

func EncryptMessageCategory(category string) string {
	return strings.ReplaceAll(category, plainCategoryLabel, encryptedCategoryLabel)
}

func DecryptMessageCategory(category string) string {
	return strings.ReplaceAll(category, encryptedCategoryLabel, plainCategoryLabel)
}

func dumpRawData(req *MessageRequest) ([]byte, error) {
	if req.DataBase64 != "" {
		return base64.RawURLEncoding.DecodeString(req.DataBase64)
	}

	return base64.StdEncoding.DecodeString(req.Data)
}

func (c *Client) EncryptMessageRequest(req *MessageRequest, sessions []*Session) error {
	if plain := IsPlainMessageCategory(req.Category); !plain {
		return nil
	}

	if ok := IsEncryptedMessageSupported(sessions); !ok {
		return nil
	}

	data, err := dumpRawData(req)
	if err != nil {
		return err
	}

	encryptedData, err := c.Lock(data, sessions)
	if err != nil {
		return err
	}

	req.Data = ""
	req.DataBase64 = base64.RawURLEncoding.EncodeToString(encryptedData)
	req.Category = EncryptMessageCategory(req.Category)

	req.Checksum = GenerateSessionChecksum(sessions)
	for _, s := range sessions {
		req.RecipientSessions = append(req.RecipientSessions, RecipientSession{
			SessionID: s.SessionID,
		})
	}

	return nil
}

const (
	EncryptedMessageReceiptStateSuccess = "SUCCESS"
	EncryptedMessageReceiptStateFailed  = "FAILED"
)

type EncryptedMessageReceipt struct {
	MessageID   string     `json:"message_id"`
	RecipientID string     `json:"recipient_id"`
	State       string     `json:"state"`
	Sessions    []*Session `json:"sessions"`
}

func (c *Client) SendEncryptedMessages(ctx context.Context, messages []*MessageRequest) ([]*EncryptedMessageReceipt, error) {
	var receipts []*EncryptedMessageReceipt
	if err := c.Post(ctx, "/encrypted_messages", messages, &receipts); err != nil {
		return nil, err
	}

	for _, receipt := range receipts {
		for _, s := range receipt.Sessions {
			s.UserID = receipt.RecipientID
		}
	}

	return receipts, nil
}

func EncryptMessageData(data []byte, sessions []*Session, private ed25519.PrivateKey) ([]byte, error) {
	key := make([]byte, 16)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}

	nonce := make([]byte, 12)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext := aesgcm.Seal(nil, nonce, data, nil)

	var sessionLen [2]byte
	binary.LittleEndian.PutUint16(sessionLen[:], uint16(len(sessions)))
	pub, _ := publicKeyToCurve25519(ed25519.PublicKey(private[32:]))

	var sessionsBytes []byte
	for _, s := range sessions {
		clientPublic, err := base64.RawURLEncoding.DecodeString(s.PublicKey)
		if err != nil {
			return nil, err
		}
		var priv, clientPub [32]byte
		copy(clientPub[:], clientPublic[:])
		privateKeyToCurve25519(&priv, private)
		dst, err := curve25519.X25519(priv[:], clientPub[:])
		if err != nil {
			return nil, err
		}

		block, err := aes.NewCipher(dst[:])
		if err != nil {
			return nil, err
		}
		padding := aes.BlockSize - len(key)%aes.BlockSize
		padtext := bytes.Repeat([]byte{byte(padding)}, padding)
		shared := make([]byte, len(key))
		copy(shared[:], key[:])
		shared = append(shared, padtext...)
		ciphertext := make([]byte, aes.BlockSize+len(shared))
		iv := ciphertext[:aes.BlockSize]
		_, err = rand.Read(iv)
		if err != nil {
			return nil, err
		}
		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext[aes.BlockSize:], shared)
		id, err := uuid.FromString(s.SessionID)
		if err != nil {
			return nil, err
		}
		sessionsBytes = append(sessionsBytes, id.Bytes()...)
		sessionsBytes = append(sessionsBytes, ciphertext...)
	}

	result := []byte{1}
	result = append(result, sessionLen[:]...)
	result = append(result, pub[:]...)
	result = append(result, sessionsBytes...)
	result = append(result, nonce[:]...)
	result = append(result, ciphertext...)

	return result, nil
}

func DecryptMessageData(data []byte, sessionID string, private ed25519.PrivateKey) ([]byte, error) {
	size := 16 + 48 // session id data + encrypted key data size
	total := len(data)
	if total < 1+2+32+size+12 {
		return nil, fmt.Errorf("invalid data size")
	}
	sessionLen := int(binary.LittleEndian.Uint16(data[1:3]))
	prefixSize := 35 + sessionLen*size
	var key []byte
	for i := 35; i < prefixSize; i += size {
		if uid, _ := uuid.FromBytes(data[i : i+16]); uid.String() == sessionID {
			var priv, pub [32]byte
			copy(pub[:], data[3:35])
			privateKeyToCurve25519(&priv, private)
			dst, err := curve25519.X25519(priv[:], pub[:])
			if err != nil {
				return nil, err
			}

			block, err := aes.NewCipher(dst[:])
			if err != nil {
				return nil, err
			}
			iv := data[i+16 : i+16+aes.BlockSize]
			key = data[i+16+aes.BlockSize : i+size]
			mode := cipher.NewCBCDecrypter(block, iv)
			mode.CryptBlocks(key, key)
			key = key[:16]
			break
		}
	}
	if len(key) != 16 {
		return nil, fmt.Errorf("session id not found")
	}
	nonce := data[prefixSize : prefixSize+12]
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes cipher: %w", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create aes gcm: %w", err)
	}

	raw, err := aesgcm.Open(nil, nonce, data[prefixSize+12:], nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data: %w", err)
	}

	return raw, nil
}

type messageLockNotSupported struct{}

func (m *messageLockNotSupported) Error() string {
	return "message encrypt & decrypt not supported"
}

func (m *messageLockNotSupported) Lock(data []byte, sessions []*Session) ([]byte, error) {
	return nil, m
}

func (m *messageLockNotSupported) Unlock(data []byte) ([]byte, error) {
	return nil, m
}

type ed25519MessageLocker struct {
	sessionID string
	key       ed25519.PrivateKey
}

func (m *ed25519MessageLocker) Lock(data []byte, sessions []*Session) ([]byte, error) {
	return EncryptMessageData(data, sessions, m.key)
}

func (m *ed25519MessageLocker) Unlock(data []byte) ([]byte, error) {
	return DecryptMessageData(data, m.sessionID, m.key)
}
