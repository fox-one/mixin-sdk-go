package mixin

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/curve25519"
)

type Keystore struct {
	ClientID   string `json:"client_id"`
	SessionID  string `json:"session_id"`
	PrivateKey string `json:"private_key"`
	PinToken   string `json:"pin_token"`
	Scope      string `json:"scope"`
}

type KeystoreAuth struct {
	*Keystore
	signMethod jwt.SigningMethod
	signKey    interface{}
	pinCipher  cipher.Block

	// seq is increasing number
	iter uint64
	mux  sync.Mutex
}

// AuthFromKeystore produces a signer using both ed25519 & RSA keystore.
func AuthFromKeystore(store *Keystore) (*KeystoreAuth, error) {
	auth := &KeystoreAuth{Keystore: store}

	var decodePinToken func([]byte) ([]byte, error)

	if b, err := ed25519Encoding.DecodeString(store.PrivateKey); err == nil && len(b) == ed25519.PrivateKeySize {
		auth.signMethod = Ed25519SigningMethod
		auth.signKey = ed25519.PrivateKey(b)

		decodePinToken = func(token []byte) ([]byte, error) {
			var curve [32]byte
			privateKeyToCurve25519(&curve, b)
			return curve25519.X25519(curve[:], token)
		}
	} else if key, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(store.PrivateKey)); err == nil {
		auth.signMethod = jwt.SigningMethodRS512
		auth.signKey = key

		decodePinToken = func(token []byte) ([]byte, error) {
			return rsa.DecryptOAEP(sha256.New(), rand.Reader, key, token, []byte(store.SessionID))
		}
	} else {
		return nil, fmt.Errorf("parse private key failed")
	}

	if store.PinToken != "" {
		token, err := ed25519Encoding.DecodeString(store.PinToken)
		if err != nil {
			return nil, fmt.Errorf("decode pin token failed: %w", err)
		}

		keyBytes, err := decodePinToken(token)
		if err != nil {
			return nil, err
		}

		pinCipher, err := aes.NewCipher(keyBytes)
		if err != nil {
			return nil, err
		}

		auth.pinCipher = pinCipher
	}

	return auth, nil
}

// AuthEd25519FromKeystore produces a signer using an ed25519 keystore.
// Deprecated: use AuthFromKeystore instead.
func AuthEd25519FromKeystore(store *Keystore) (*KeystoreAuth, error) {
	return AuthFromKeystore(store)
}

func (k *KeystoreAuth) SignTokenAt(signature, requestID string, at time.Time, exp time.Duration) string {
	jwtMap := jwt.MapClaims{
		"uid": k.ClientID,
		"sid": k.SessionID,
		"iat": at.Unix(),
		"exp": at.Add(exp).Unix(),
		"jti": requestID,
		"sig": signature,
		"scp": ScopeFull,
	}

	if k.Scope != "" {
		jwtMap["scp"] = k.Scope
	}

	token, err := jwt.NewWithClaims(k.signMethod, jwtMap).SignedString(k.signKey)
	if err != nil {
		panic(err)
	}

	return token
}

func (k *KeystoreAuth) SignToken(signature, requestID string, exp time.Duration) string {
	return k.SignTokenAt(signature, requestID, time.Now(), exp)
}

func (k *KeystoreAuth) sequence() uint64 {
	k.mux.Lock()
	defer k.mux.Unlock()

	if iter := uint64(time.Now().UnixNano()); iter > k.iter {
		k.iter = iter
	} else {
		k.iter += 1
	}

	return k.iter
}

func (k *KeystoreAuth) EncryptPin(pin string) string {
	if k.pinCipher == nil {
		panic(errors.New("keystore: pin_token required"))
	}

	if err := ValidatePinPattern(pin); err != nil {
		panic(err)
	}

	pinByte := []byte(pin)
	timeBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(timeBytes, uint64(time.Now().Unix()))
	pinByte = append(pinByte, timeBytes...)
	iteratorBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(iteratorBytes, k.sequence())
	pinByte = append(pinByte, iteratorBytes...)
	padding := aes.BlockSize - len(pinByte)%aes.BlockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	pinByte = append(pinByte, padText...)
	cipherText := make([]byte, aes.BlockSize+len(pinByte))
	iv := cipherText[:aes.BlockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(k.pinCipher, iv)
	mode.CryptBlocks(cipherText[aes.BlockSize:], pinByte)
	return base64.StdEncoding.EncodeToString(cipherText)
}
