package mixin

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
