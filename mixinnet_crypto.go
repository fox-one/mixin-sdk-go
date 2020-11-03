package mixin

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"

	"golang.org/x/crypto/sha3"
)

const (
	Operator0   = 0x00
	Operator64  = 0x40
	OperatorSum = 0xfe
	OperatorCmp = 0xff
)

type (
	Script []uint8
	Hash   [32]byte

	TransactionExtra []byte
)

// Script

func NewThresholdScript(threshold uint8) Script {
	return Script{OperatorCmp, OperatorSum, threshold}
}

func (s Script) VerifyFormat() error {
	if len(s) != 3 {
		return fmt.Errorf("invalid script %d", len(s))
	}
	if s[0] != OperatorCmp || s[1] != OperatorSum {
		return fmt.Errorf("invalid script %d %d", s[0], s[1])
	}
	return nil
}

func (s Script) Validate(sum int) error {
	err := s.VerifyFormat()
	if err != nil {
		return err
	}
	if sum < int(s[2]) {
		return fmt.Errorf("invalid signature keys %d %d", sum, s[2])
	}
	return nil
}

func (s Script) String() string {
	return hex.EncodeToString(s[:])
}

func (s Script) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(s.String())), nil
}

func (s *Script) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	data, err := hex.DecodeString(string(unquoted))
	if err != nil {
		return err
	}
	*s = data
	return nil
}

// Hash

func NewHash(data []byte) Hash {
	return Hash(sha3.Sum256(data))
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
	return bytes.Compare(h[:], zero[:]) != 0
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
	data, err := hex.DecodeString(string(unquoted))
	if err != nil {
		return err
	}
	if len(data) != len(h) {
		return fmt.Errorf("invalid hash length %d", len(data))
	}
	copy(h[:], data)
	return nil
}

// Transaction Extra

func (e TransactionExtra) String() string {
	return base64.StdEncoding.EncodeToString(e[:])
}

func (e TransactionExtra) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote(e.String())), nil
}

func (e *TransactionExtra) UnmarshalJSON(b []byte) error {
	unquoted, err := strconv.Unquote(string(b))
	if err != nil {
		return err
	}
	data, err := hex.DecodeString(string(unquoted))
	if err != nil {
		if data, err = base64.StdEncoding.DecodeString(string(unquoted)); err != nil {
			return err
		}
	}

	*e = data
	return nil
}
