package mixinnet

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/zeebo/blake3"
	"golang.org/x/crypto/sha3"
)

type (
	Hash [32]byte
)

// Hash

func NewHash(data []byte) Hash {
	return sha3.Sum256(data)
}

func NewBlake3Hash(data []byte) Hash {
	return Hash(blake3.Sum256(data))
}

func HashFromString(src string) (Hash, error) {
	var hash Hash
	data, err := hex.DecodeString(src)
	if err != nil {
		return hash, err
	}
	if len(data) != len(hash) {
		return hash, fmt.Errorf("invalid hash length %d", len(data))
	}
	copy(hash[:], data)
	return hash, nil
}

func (h Hash) HasValue() bool {
	zero := Hash{}
	return !bytes.Equal(h[:], zero[:])
}

func (h Hash) String() string {
	return hex.EncodeToString(h[:])
}

func (h Hash) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(h.String())), nil
}

func (h *Hash) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	data, err := hex.DecodeString(unquoted)
	if err != nil {
		return err
	}
	if len(data) != len(h) {
		return fmt.Errorf("invalid hash length %d", len(data))
	}
	copy(h[:], data)
	return nil
}
