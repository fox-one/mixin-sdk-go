package mixin

import (
	"bytes"
	"encoding/hex"

	"github.com/fox-one/msgpack"
)

type (
	TransactionV1 struct {
		Transaction
		Signatures [][]*Signature `json:"signatures,omitempty" msgpack:",omitempty"`
	}
)

func (t *TransactionV1) DumpTransaction() (string, error) {
	return t.Transaction.DumpTransaction()
}

func TransactionFromData(data []byte) (*Transaction, error) {
	if !checkTxVersion(data) {
		return transactionV1FromRaw(data)
	}

	return transactionV2FromRaw(data)
}

func TransactionFromRaw(raw string) (*Transaction, error) {
	bts, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	return TransactionFromData(bts)
}

func transactionV1FromRaw(bts []byte) (*Transaction, error) {
	var tx TransactionV1
	if err := msgpack.Unmarshal(bts, &tx); err != nil {
		return nil, err
	}
	if len(tx.Signatures) > 0 {
		tx.Transaction.Signatures = make([]map[uint16]*Signature, len(tx.Signatures))
		for i, sigs := range tx.Signatures {
			tx.Transaction.Signatures[i] = make(map[uint16]*Signature, len(sigs))
			for k, sig := range sigs {
				tx.Transaction.Signatures[i][uint16(k)] = sig
			}
		}
	}
	return &tx.Transaction, nil
}

func transactionV2FromRaw(bts []byte) (*Transaction, error) {
	return NewDecoder(bts).DecodeTransaction()
}

func checkTxVersion(val []byte) bool {
	if len(val) < 4 {
		return false
	}
	v := append(magic, 0, TxVersion)
	return bytes.Equal(v, val[:4])
}
