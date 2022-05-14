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

func AuthFromKeystore(store *Keystore) (*KeystoreAuth, error) {
	auth := &KeystoreAuth{
		Keystore: store,
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(store.PrivateKey))
	if err != nil {
		return nil, err
	}
	auth.signKey = signKey
	auth.signMethod = jwt.SigningMethodRS512

	if store.PinToken != "" {
		token, err := base64.StdEncoding.DecodeString(store.PinToken)
		if err != nil {
			return nil, err
		}

		keyBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, signKey, token, []byte(store.SessionID))
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

// AuthEd25519FromKeystore produces a signer using a ed25519 keystore.
func AuthEd25519FromKeystore(store *Keystore) (*KeystoreAuth, error) {
	auth := &KeystoreAuth{
		Keystore:   store,
		signMethod: Ed25519SigningMethod,
	}

	signKey, err := ed25519Encoding.DecodeString(store.PrivateKey)
	if err != nil {
		return nil, err
	}

	if len(signKey) != ed25519.PrivateKeySize {
		return nil, errors.New("invalid ed25519 private key")
	}

	auth.signKey = ed25519.PrivateKey(signKey)

	if store.PinToken != "" {
		token, err := ed25519Encoding.DecodeString(store.PinToken)
		if err != nil {
			return nil, err
		}

		var keyBytes, curve, pub [32]byte
		privateKeyToCurve25519(&curve, signKey)
		copy(pub[:], token[:])
		curve25519.ScalarMult(&keyBytes, &curve, &pub)

		pinCipher, err := aes.NewCipher(keyBytes[:])
		if err != nil {
			return nil, err
		}

		auth.pinCipher = pinCipher
	}
	return auth, nil
}

func (k *KeystoreAuth) SignToken(signature, requestID string, exp time.Duration) string {
	jwtMap := jwt.MapClaims{
		"uid": k.ClientID,
		"sid": k.SessionID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(exp).Unix(),
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
