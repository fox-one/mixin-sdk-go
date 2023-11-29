package mixinnet

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"strconv"

	"filippo.io/edwards25519"
)

type (
	Signature [64]byte
)

func (privateKey *Key) Sign(message []byte) Signature {
	var digest1, messageDigest, hramDigest [64]byte

	// the hash costs almost nothing compared to elliptic curve ops
	h := sha512.New()
	h.Write(privateKey[:32])
	h.Sum(digest1[:0])
	h.Reset()
	h.Write(digest1[32:])
	h.Write(message)
	h.Sum(messageDigest[:0])

	z, err := edwards25519.NewScalar().SetUniformBytes(messageDigest[:])
	if err != nil {
		panic(err)
	}

	R := edwards25519.NewIdentityPoint().ScalarBaseMult(z)

	pub := privateKey.Public()
	h.Reset()
	h.Write(R.Bytes())
	h.Write(pub[:])
	h.Write(message[:])
	h.Sum(hramDigest[:0])
	x, err := edwards25519.NewScalar().SetUniformBytes(hramDigest[:])
	if err != nil {
		panic(err)
	}

	y, err := edwards25519.NewScalar().SetCanonicalBytes(privateKey[:])
	if err != nil {
		panic(privateKey.String())
	}
	s := edwards25519.NewScalar().MultiplyAdd(x, y, z)

	var signature Signature
	copy(signature[:], R.Bytes())
	copy(signature[32:], s.Bytes())

	return signature
}

func (privateKey *Key) SignHash(h Hash) Signature {
	return privateKey.Sign(h[:])
}

func (publicKey *Key) VerifyWithChallenge(sig Signature, a *edwards25519.Scalar) bool {
	p, err := edwards25519.NewIdentityPoint().SetBytes(publicKey[:])
	if err != nil {
		return false
	}
	A := edwards25519.NewIdentityPoint().Negate(p)

	b, err := edwards25519.NewScalar().SetCanonicalBytes(sig[32:])
	if err != nil {
		return false
	}
	R := edwards25519.NewIdentityPoint().VarTimeDoubleScalarBaseMult(a, A, b)
	return bytes.Equal(sig[:32], R.Bytes())
}

func (publicKey *Key) Verify(message []byte, sig Signature) bool {
	h := sha512.New()
	h.Write(sig[:32])
	h.Write(publicKey[:])
	h.Write(message[:])
	var digest [64]byte
	h.Sum(digest[:0])

	x, err := edwards25519.NewScalar().SetUniformBytes(digest[:])
	if err != nil {
		panic(err)
	}
	return publicKey.VerifyWithChallenge(sig, x)
}

func (publicKey *Key) VerifyHash(message Hash, sig Signature) bool {
	return publicKey.Verify(message[:], sig)
}

func (s Signature) String() string {
	return hex.EncodeToString(s[:])
}

func (s Signature) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.String())), nil
}

func (s *Signature) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	data, err := hex.DecodeString(unquoted)
	if err != nil {
		return err
	}
	if len(data) != len(s) {
		return fmt.Errorf("invalid signature length %d", len(data))
	}
	copy(s[:], data)
	return nil
}
