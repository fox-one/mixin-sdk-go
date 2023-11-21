package mixin

import (
	"filippo.io/edwards25519"
)

func SafeHashScalar(k *edwards25519.Point, outputIndex uint64) *edwards25519.Scalar {
	return HashScalar(k, outputIndex, NewBlake3Hash)
}

func SafeDeriveGhostPublicKey(r, A, B *Key, outputIndex uint64) *Key {
	return DeriveGhostPublicKey(r, A, B, outputIndex, NewBlake3Hash)
}

func SafeDeriveGhostPrivateKey(R, a, b *Key, outputIndex uint64) *Key {
	return DeriveGhostPrivateKey(R, a, b, outputIndex, NewBlake3Hash)
}

func SafeViewGhostOutputKey(P, a, R *Key, outputIndex uint64) *Key {
	return ViewGhostOutputKey(P, a, R, outputIndex, NewBlake3Hash)
}
