package mixinnet

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"

	"filippo.io/edwards25519"
)

type (
	Key [32]byte
)

func GenerateEd25519Key() ed25519.PrivateKey {
	_, private, _ := ed25519.GenerateKey(nil)
	return private
}

func GenerateKey(randReader io.Reader) Key {
	var (
		seed = make([]byte, 32)
		s    = 0
	)

	for s < len(seed) {
		n, _ := randReader.Read(seed[s:])
		s += n
	}
	k, err := keyFromSeed(seed)
	if err != nil {
		panic(err)
	}
	return k
}

func keyFromSeed(seed []byte) (Key, error) {
	var key Key
	h := sha512.Sum512(seed[:32])
	x := h[:32]
	var wideBytes [64]byte
	copy(wideBytes[:], x[:])
	wideBytes[0] &= 248
	wideBytes[31] &= 63
	wideBytes[31] |= 64
	s, err := edwards25519.NewScalar().SetUniformBytes(wideBytes[:])
	if err != nil {
		return key, err
	}
	copy(key[:], s.Bytes())
	return key, nil
}

func KeyFromSeed(seed string) (Key, error) {
	var key Key

	b, err := hex.DecodeString(seed)
	if err != nil {
		return key, err
	}

	return keyFromSeed(b)
}

func KeyFromBytes(bts []byte) Key {
	var key [32]byte
	s, err := edwards25519.NewScalar().SetUniformBytes(bts)
	if err != nil {
		panic(err)
	}
	copy(key[:], s.Bytes())
	return key
}

func KeyFromString(s string) (Key, error) {
	var key Key
	b, err := hex.DecodeString(s)
	if err != nil {
		return key, err
	}

	switch len(b) {
	case 32:
		copy(key[:], b)
		return key, nil
	case 64:
		return keyFromSeed(b[:32])
	default:
		return key, fmt.Errorf("invalid key size %d", len(b))
	}
}

func ParseKeyWithPub(s, pub string) (Key, error) {
	if key, err := KeyFromString(s); err == nil && key.Public().String() == pub {
		return key, nil
	}

	if key, err := KeyFromSeed(s); err == nil && key.Public().String() == pub {
		return key, nil
	}

	return Key{}, fmt.Errorf("invalid key")
}

func (k Key) CheckKey() bool {
	_, err := edwards25519.NewIdentityPoint().SetBytes(k[:])
	return err == nil
}

func (k Key) Public() Key {
	x, err := edwards25519.NewScalar().SetCanonicalBytes(k[:])
	if err != nil {
		return Key{}
	}
	v := edwards25519.NewIdentityPoint().ScalarBaseMult(x)
	var tmp Key
	copy(tmp[:], v.Bytes())
	return tmp
}

func (k Key) ToScalar() (*edwards25519.Scalar, error) {
	return edwards25519.NewScalar().SetCanonicalBytes(k[:])
}

func (k Key) ToPoint() (*edwards25519.Point, error) {
	return edwards25519.NewIdentityPoint().SetBytes(k[:])
}

func (k Key) HasValue() bool {
	zero := Key{}
	return !bytes.Equal(k[:], zero[:])
}

func (k Key) DeterministicHashDerive() Key {
	seed := NewHash(k[:])
	return KeyFromBytes(append(seed[:], seed[:]...))
}

func KeyMultPubPriv(pub, priv *Key) *edwards25519.Point {
	q, err := edwards25519.NewIdentityPoint().SetBytes(pub[:])
	if err != nil {
		panic(pub.String())
	}
	x, err := edwards25519.NewScalar().SetCanonicalBytes(priv[:])
	if err != nil {
		panic(priv.String())
	}

	v := edwards25519.NewIdentityPoint().ScalarMult(x, q)
	return v
}

func HashScalar(txVer uint8, k *edwards25519.Point, outputIndex uint8) *edwards25519.Scalar {
	tmp := make([]byte, 12)
	length := binary.PutUvarint(tmp, uint64(outputIndex))
	tmp = tmp[:length]

	var src [64]byte
	var buf bytes.Buffer
	buf.Write(k.Bytes())
	buf.Write(tmp)

	hashFunc := NewBlake3Hash
	if txVer < TxVersionHashSignature {
		hashFunc = NewHash
	}

	hash := hashFunc(buf.Bytes())
	copy(src[:32], hash[:])
	hash = hashFunc(hash[:])
	copy(src[32:], hash[:])
	s, err := edwards25519.NewScalar().SetUniformBytes(src[:])
	if err != nil {
		panic(err)
	}

	hash = hashFunc(s.Bytes())
	copy(src[:32], hash[:])
	hash = hashFunc(hash[:])
	copy(src[32:], hash[:])
	x, err := edwards25519.NewScalar().SetUniformBytes(src[:])
	if err != nil {
		panic(err)
	}
	return x
}

func DeriveGhostPublicKey(txVer uint8, r, A, B *Key, outputIndex uint8) *Key {
	x := HashScalar(txVer, KeyMultPubPriv(A, r), outputIndex)
	p1, err := edwards25519.NewIdentityPoint().SetBytes(B[:])
	if err != nil {
		panic(B.String())
	}
	p2 := edwards25519.NewIdentityPoint().ScalarBaseMult(x)
	p4 := edwards25519.NewIdentityPoint().Add(p1, p2)
	var key Key
	copy(key[:], p4.Bytes())
	return &key
}

func DeriveGhostPrivateKey(txVer uint8, R, a, b *Key, outputIndex uint8) *Key {
	x := HashScalar(txVer, KeyMultPubPriv(R, a), outputIndex)
	y, err := edwards25519.NewScalar().SetCanonicalBytes(b[:])
	if err != nil {
		panic(b.String())
	}
	t := edwards25519.NewScalar().Add(x, y)
	var key Key
	copy(key[:], t.Bytes())
	return &key
}

func ViewGhostOutputKey(txVer uint8, P, a, R *Key, outputIndex uint8) *Key {
	x := HashScalar(txVer, KeyMultPubPriv(R, a), outputIndex)
	p1, err := edwards25519.NewIdentityPoint().SetBytes(P[:])
	if err != nil {
		panic(P.String())
	}
	p2 := edwards25519.NewIdentityPoint().ScalarBaseMult(x)
	p4 := edwards25519.NewIdentityPoint().Subtract(p1, p2)
	var key Key
	copy(key[:], p4.Bytes())
	return &key
}

func (k Key) String() string {
	return hex.EncodeToString(k[:])
}

func (k Key) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(k.String())), nil
}

func (k *Key) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	key, err := KeyFromString(unquoted)
	if err != nil {
		return err
	}
	*k = key
	return nil
}
