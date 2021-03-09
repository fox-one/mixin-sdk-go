package mixin

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/fox-one/msgpack"
)

type (
	SignedTransactionV1 struct {
		Transaction
		Signatures [][]*Signature `json:"signatures,omitempty" msgpack:",omitempty"`
	}

	SignedTransactionV2 struct {
		Transaction
		Signatures []map[uint16]*Signature `json:"signatures,omitempty" msgpack:",omitempty"`
	}
)

func (t *SignedTransactionV1) DumpTransaction() (string, error) {
	t.Transaction.Signatures = t.Signatures
	return t.Transaction.DumpTransaction()
}

func (t *SignedTransactionV2) DumpTransaction() (string, error) {
	t.Transaction.Signatures = t.Signatures
	return t.Transaction.DumpTransaction()
}

func TransactionFromRaw(raw string) (*Transaction, error) {
	bts, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	if len(bts) > 4 {
		switch bts[3] {
		case 0, 1:
			var tx SignedTransactionV1
			if err := msgpack.Unmarshal(bts, &tx); err != nil {
				return nil, err
			}
			tx.Transaction.Signatures = tx.Signatures
			return &tx.Transaction, nil

		case 2:
			tx, err := NewDecoder(bts).DecodeTransaction()
			if err != nil {
				return nil, err
			}
			tx.Transaction.Signatures = tx.Signatures
			return &tx.Transaction, nil
		}
	}
	return nil, errors.New("invalid transaction data")
}

func checkTxVersion(val []byte) bool {
	if len(val) < 4 {
		return false
	}
	v := append(magic, 0, TxVersion)
	return bytes.Equal(v, val[:4])
}
