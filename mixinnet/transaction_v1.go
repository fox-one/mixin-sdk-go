package mixinnet

import (
	"github.com/fox-one/msgpack"
)

type (
	TransactionV1 struct {
		Transaction
		Signatures [][]*Signature `json:"signatures,omitempty" msgpack:",omitempty"`
	}
)

func (t *TransactionV1) Dump() (string, error) {
	return t.Transaction.Dump()
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
