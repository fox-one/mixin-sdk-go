package mixinnet

import (
	"bytes"
	"encoding/hex"
	"errors"

	"github.com/fox-one/msgpack"
)

const (
	TxVersionCommonEncoding = 0x02
	TxVersionBlake3Hash     = 0x03
	TxVersionReferences     = 0x04
	TxVersionHashSignature  = 0x05

	TxVersionLegacy = TxVersionReferences
	TxVersion       = TxVersionHashSignature
)

const (
	OutputTypeScript              = 0x00
	OutputTypeWithdrawalSubmit    = 0xa1
	OutputTypeWithdrawalFuel      = 0xa2
	OutputTypeNodePledge          = 0xa3
	OutputTypeNodeAccept          = 0xa4
	outputTypeNodeResign          = 0xa5
	OutputTypeNodeRemove          = 0xa6
	OutputTypeDomainAccept        = 0xa7
	OutputTypeDomainRemove        = 0xa8
	OutputTypeWithdrawalClaim     = 0xa9
	OutputTypeNodeCancel          = 0xaa
	OutputTypeCustodianEvolution  = 0xb1
	OutputTypeCustodianMigration  = 0xb2
	OutputTypeCustodianDeposit    = 0xb3
	OutputTypeCustodianWithdrawal = 0xb4
)

const (
	TransactionTypeScript               = 0x00
	TransactionTypeMint                 = 0x01
	TransactionTypeDeposit              = 0x02
	TransactionTypeWithdrawalSubmit     = 0x03
	TransactionTypeWithdrawalFuel       = 0x04
	TransactionTypeWithdrawalClaim      = 0x05
	TransactionTypeNodePledge           = 0x06
	TransactionTypeNodeAccept           = 0x07
	transactionTypeNodeResign           = 0x08
	TransactionTypeNodeRemove           = 0x09
	TransactionTypeDomainAccept         = 0x10
	TransactionTypeDomainRemove         = 0x11
	TransactionTypeNodeCancel           = 0x12
	TransactionTypeCustodianUpdateNodes = 0x13
	TransactionTypeCustodianSlashNodes  = 0x14
	TransactionTypeUnknown              = 0xff
)

type (
	MintData struct {
		Group  string  `json:"group"`
		Batch  uint64  `json:"batch"`
		Amount Integer `json:"amount"`
	}

	DepositData struct {
		Chain       Hash    `json:"chain"`
		AssetKey    string  `json:"asset"`
		Transaction string  `json:"transaction"`
		Index       uint64  `json:"index"`
		Amount      Integer `json:"amount"`
	}

	WithdrawalData struct {
		Address string `json:"address"`
		Tag     string `json:"tag"`

		// DEPRECATED since safe, tx 5, TxVersionHashSignature
		Chain Hash `json:"chain"`
		// DEPRECATED since safe, tx 5, TxVersionHashSignature
		AssetKey string `json:"asset"`
	}

	Input struct {
		Hash    *Hash        `json:"hash,omitempty"`
		Index   uint8        `json:"index,omitempty"`
		Genesis []byte       `json:"genesis,omitempty"`
		Deposit *DepositData `json:"deposit,omitempty"`
		Mint    *MintData    `json:"mint,omitempty"`
	}

	Output struct {
		Type       uint8           `json:"type"`
		Amount     Integer         `json:"amount"`
		Keys       []Key           `json:"keys,omitempty"`
		Withdrawal *WithdrawalData `json:"withdrawal,omitempty" msgpack:",omitempty"`

		Script Script `json:"script"`
		Mask   Key    `json:"mask,omitempty"`
	}

	AggregatedSignature struct {
		Signers   []int      `json:"signers"`
		Signature *Signature `json:"signature"`
	}

	Transaction struct {
		Hash                *Hash                   `json:"hash,omitempty" msgpack:"-"`
		Snapshot            *Hash                   `json:"snapshot,omitempty" msgpack:"-"`
		Signatures          []map[uint16]*Signature `json:"signatures,omitempty" msgpack:"-"`
		AggregatedSignature *AggregatedSignature    `json:"aggregated_signature,omitempty" msgpack:"-"`

		Version    uint8            `json:"version"`
		Asset      Hash             `json:"asset"`
		Inputs     []*Input         `json:"inputs"`
		Outputs    []*Output        `json:"outputs"`
		References []Hash           `msgpack:"-"`
		Extra      TransactionExtra `json:"extra,omitempty"`
	}
)

func TransactionFromRaw(raw string) (*Transaction, error) {
	bts, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	return TransactionFromData(bts)
}

func TransactionFromData(data []byte) (*Transaction, error) {
	txVer := checkTxVersion(data)
	if txVer < TxVersionCommonEncoding {
		return transactionV1FromRaw(data)
	}

	return NewDecoder(data).DecodeTransaction()
}

func (t *Transaction) DumpData() ([]byte, error) {
	switch t.Version {
	case 0, 1:
		tx := TransactionV1{
			Transaction: *t,
		}
		if len(t.Signatures) > 0 {
			tx.Signatures = make([][]*Signature, len(t.Signatures))
			for i, sigs := range t.Signatures {
				tx.Signatures[i] = make([]*Signature, len(sigs))
				for k, sig := range sigs {
					tx.Signatures[i][k] = sig
				}
			}
		}

		var buf bytes.Buffer
		enc := msgpack.NewEncoder(&buf).UseCompactEncoding(true)
		err := enc.Encode(tx)
		if err != nil {
			return nil, err
		}
		return buf.Bytes(), nil

	case TxVersionCommonEncoding, TxVersionBlake3Hash, TxVersionReferences, TxVersionHashSignature:
		return NewEncoder().EncodeTransaction(t), nil

	default:
		return nil, errors.New("unknown tx version")
	}
}

func (t *Transaction) Dump() (string, error) {
	bts, err := t.DumpData()
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bts), nil
}

func (t Transaction) DumpPayload() ([]byte, error) {
	t.Signatures = nil
	t.AggregatedSignature = nil
	return t.DumpData()
}

func (t *Transaction) TransactionHash() (Hash, error) {
	if t.Hash == nil {
		raw, err := t.DumpPayload()
		if err != nil {
			return Hash{}, err
		}

		if t.Version >= TxVersionBlake3Hash {
			h := NewBlake3Hash(raw)
			t.Hash = &h
		} else {
			h := NewHash(raw)
			t.Hash = &h
		}
	}
	return *t.Hash, nil
}
