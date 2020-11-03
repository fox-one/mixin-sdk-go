package mixin

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/fox-one/mixin-sdk-go/edwards25519"
)

type (
	Key [32]byte
)

func NewKeyFromSeed(seed []byte) Key {
	var key [32]byte
	var src [64]byte
	if len(seed) != len(src) {
		panic(len(seed))
	}
	copy(src[:], seed)
	edwards25519.ScReduce(&key, &src)
	return key
}

func KeyFromString(s string) (Key, error) {
	var key Key
	b, err := hex.DecodeString(s)
	if err != nil {
		return key, err
	}
	if len(b) != len(key) {
		return key, fmt.Errorf("invalid key size %d", len(b))
	}
	copy(key[:], b)
	return key, nil
}

func (k Key) CheckKey() bool {
	var point edwards25519.ExtendedGroupElement
	tmp := [32]byte(k)
	return point.FromBytes(&tmp)
}

func (k Key) CheckScalar() bool {
	tmp := [32]byte(k)
	return edwards25519.ScValid(&tmp)
}

func (k Key) Public() Key {
	var point edwards25519.ExtendedGroupElement
	tmp := [32]byte(k)
	edwards25519.GeScalarMultBase(&point, &tmp)
	point.ToBytes(&tmp)
	return tmp
}

func (k Key) HasValue() bool {
	zero := Key{}
	return bytes.Compare(k[:], zero[:]) != 0
}

func (k Key) DeterministicHashDerive() Key {
	seed := NewHash(k[:])
	return NewKeyFromSeed(append(seed[:], seed[:]...))
}

func KeyMultPubPriv(pub, priv *Key) *Key {
	if !pub.CheckKey() {
		panic(pub.String())
	}
	if !priv.CheckScalar() {
		panic(priv.String())
	}

	var point edwards25519.ExtendedGroupElement
	var point2 edwards25519.ProjectiveGroupElement

	tmp := [32]byte(*pub)
	point.FromBytes(&tmp)
	tmp = [32]byte(*priv)
	edwards25519.GeScalarMult(&point2, &tmp, &point)

	point2.ToBytes(&tmp)
	key := Key(tmp)
	return &key
}

func (k *Key) MultScalar(outputIndex int) *Key {
	tmp := make([]byte, 12, 12)
	length := binary.PutUvarint(tmp, uint64(outputIndex))
	tmp = tmp[:length]

	var src [64]byte
	var buf bytes.Buffer
	buf.Write(k[:])
	buf.Write(tmp)
	hash := NewHash(buf.Bytes())
	copy(src[:32], hash[:])
	hash = NewHash(hash[:])
	copy(src[32:], hash[:])
	key := NewKeyFromSeed(src[:])
	return &key
}

func ViewGhostOutputKey(P, a, R *Key, outputIndex int) *Key {
	var point1, point2 edwards25519.ExtendedGroupElement
	var point3 edwards25519.CachedGroupElement
	var point4 edwards25519.CompletedGroupElement
	var point5 edwards25519.ProjectiveGroupElement

	tmp := [32]byte(*P)
	point1.FromBytes(&tmp)
	scalar := KeyMultPubPriv(R, a).MultScalar(outputIndex).HashScalar()
	edwards25519.GeScalarMultBase(&point2, scalar)
	point2.ToCached(&point3)
	edwards25519.GeSub(&point4, &point1, &point3)
	point4.ToProjective(&point5)
	point5.ToBytes(&tmp)
	key := Key(tmp)
	return &key
}

func (k Key) HashScalar() *[32]byte {
	var out [32]byte
	var src [64]byte
	hash := NewHash(k[:])
	copy(src[:32], hash[:])
	hash = NewHash(hash[:])
	copy(src[32:], hash[:])
	edwards25519.ScReduce(&out, &src)
	return &out
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
	data, err := hex.DecodeString(string(unquoted))
	if err != nil {
		return err
	}
	if len(data) != len(k) {
		return fmt.Errorf("invalid key length %d", len(data))
	}
	copy(k[:], data)
	return nil
}
