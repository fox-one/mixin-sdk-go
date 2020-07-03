package mixin

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"io"
	"time"

	"github.com/dgrijalva/jwt-go"
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
	binary.LittleEndian.PutUint64(iteratorBytes, uint64(time.Now().UnixNano()))
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
