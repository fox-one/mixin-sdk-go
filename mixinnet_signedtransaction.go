package mixin

import (
	"bytes"
	"encoding/hex"

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

	if !checkTxVersion(bts) {
		return transactionV1FromRaw(bts)
	}
	return transactionV2FromRaw(bts)
}

func transactionV1FromRaw(bts []byte) (*Transaction, error) {
	var tx SignedTransactionV1
	if err := msgpack.Unmarshal(bts, &tx); err != nil {
		return nil, err
	}
	if len(tx.Signatures) > 0 {
		tx.Transaction.Signatures = tx.Signatures
	}
	return &tx.Transaction, nil
}

func transactionV2FromRaw(bts []byte) (*Transaction, error) {
	tx, err := NewDecoder(bts).DecodeTransaction()
	if err != nil {
		return nil, err
	}
	if len(tx.Signatures) > 0 {
		tx.Transaction.Signatures = tx.Signatures
	}
	return &tx.Transaction, nil
}

func checkTxVersion(val []byte) bool {
	if len(val) < 4 {
		return false
	}
	v := append(magic, 0, TxVersion)
	return bytes.Equal(v, val[:4])
}
