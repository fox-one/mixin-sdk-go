package mixin

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/vmihailenco/msgpack/v4"
)

const (
	TxVersion = 0x01
)

type (
	MintData struct {
		Group  string  `json:"group"`
		Batch  uint64  `json:"batch"`
		Amount Integer `json:"amount"`
	}

	DepositData struct {
		Chain           Hash    `json:"chain"`
		AssetKey        string  `json:"asset"`
		TransactionHash string  `json:"transaction"`
		OutputIndex     uint64  `json:"index"`
		Amount          Integer `json:"amount"`
	}

	Input struct {
		Hash    *Hash        `json:"hash,omitempty"`
		Index   int          `json:"index,omitempty"`
		Genesis []byte       `json:"genesis,omitempty"`
		Deposit *DepositData `json:"deposit,omitempty"`
		Mint    *MintData    `json:"mint,omitempty"`
	}

	Output struct {
		Type   uint8   `json:"type"`
		Amount Integer `json:"amount"`
		Keys   []Key   `json:"keys,omitempty"`
		Script Script  `json:"script"`
		Mask   Key     `json:"mask,omitempty"`
	}

	Transaction struct {
		Hash       *Hash            `json:"hash,omitempty" msgpack:"-"`
		Version    uint8            `json:"version"`
		Asset      Hash             `json:"asset"`
		Inputs     []*Input         `json:"inputs"`
		Outputs    []*Output        `json:"outputs"`
		Extra      TransactionExtra `json:"extra,omitempty"`
		Signatures [][]Signature    `json:"signatures,omitempty" msgpack:",omitempty"`
		Snapshot   *Hash            `json:"snapshot,omitempty" msgpack:"-"`
	}

	TransactionOutput struct {
		Receivers []string
		Threshold uint8
		Amount    decimal.Decimal
	}

	TransactionInput struct {
		Memo    string
		Inputs  []*MultisigUTXO
		Outputs []TransactionOutput
		Hint    string
	}
)

func (t *Transaction) DumpTransaction() (string, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf).UseCompactEncoding(true)
	err := enc.Encode(t)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(buf.Bytes()), nil
}

func (t *Transaction) DumpTransactionPayload() (string, error) {
	sigs := t.Signatures
	t.Signatures = nil
	raw, err := t.DumpTransaction()
	t.Signatures = sigs
	return raw, err
}

func (t *Transaction) TransactionHash() (Hash, error) {
	if t.Hash == nil {
		raw, err := t.DumpTransactionPayload()
		if err != nil {
			return Hash{}, err
		}

		bts, _ := hex.DecodeString(raw)
		h := NewHash(bts)
		t.Hash = &h
	}
	return *t.Hash, nil
}

func (i *TransactionInput) AppendUTXO(utxo *MultisigUTXO) {
	i.Inputs = append(i.Inputs, utxo)
}

func (i *TransactionInput) AppendOutput(receivers []string, threshold uint8, amount decimal.Decimal) {
	i.Outputs = append(i.Outputs, TransactionOutput{
		Receivers: receivers,
		Threshold: threshold,
		Amount:    amount,
	})
}

func (i *TransactionInput) Asset() Hash {
	if len(i.Inputs) == 0 {
		return Hash{}
	}
	return i.Inputs[0].Asset()
}

func (i *TransactionInput) TotalInputAmount() decimal.Decimal {
	var total decimal.Decimal
	for _, input := range i.Inputs {
		total = total.Add(input.Amount)
	}
	return total
}

func (i *TransactionInput) Validate() error {
	if len(i.Inputs) == 0 {
		return errors.New("no input utxo")
	}

	var (
		members = map[string]bool{}
		total   = i.TotalInputAmount()
		asset   = i.Asset()
	)

	for _, input := range i.Inputs {
		if asset != input.Asset() {
			return errors.New("invalid input utxo, asset not matched")
		}

		if len(members) == 0 {
			for _, u := range input.Members {
				members[u] = true
			}
			continue
		}
		if len(members) != len(input.Members) {
			return errors.New("invalid input utxo, member not matched")
		}
		for _, m := range input.Members {
			if _, f := members[m]; !f {
				return errors.New("invalid input utxo, member not matched")
			}
		}
	}

	for _, output := range i.Outputs {
		if output.Threshold == 0 || int(output.Threshold) > len(output.Receivers) {
			return fmt.Errorf("invalid output threshold: %d", output.Threshold)
		}

		if !output.Amount.IsPositive() {
			return fmt.Errorf("invalid output amount: %v", output.Amount)
		}

		if total = total.Sub(output.Amount); total.IsNegative() {
			return errors.New("invalid output: amount exceed")
		}
	}

	return nil
}

func (c *Client) MakeMultisigTransaction(ctx context.Context, input *TransactionInput) (*Transaction, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var tx = Transaction{
		Version: TxVersion,
		Asset:   input.Asset(),
		Extra:   []byte(input.Memo),
	}
	// add inputs
	for _, input := range input.Inputs {
		tx.Inputs = append(tx.Inputs, &Input{
			Hash:  &input.TransactionHash,
			Index: input.OutputIndex,
		})
	}

	outputs := input.Outputs

	// refund the change
	{
		change := input.TotalInputAmount()
		for _, output := range input.Outputs {
			change = change.Sub(output.Amount)
		}

		if change.IsPositive() {
			outputs = append(outputs, TransactionOutput{
				Receivers: input.Inputs[0].Members,
				Threshold: input.Inputs[0].Threshold,
				Amount:    change,
			})
		}
	}

	ghostInputs := make([]*GhostInput, 0, len(outputs))
	for idx, output := range outputs {
		ghostInputs = append(ghostInputs, &GhostInput{
			Receivers: output.Receivers,
			Index:     idx,
			Hint:      input.Hint,
		})
	}

	ghosts, err := c.BatchReadGhostKeys(ctx, ghostInputs)
	if err != nil {
		return nil, err
	}

	for idx, output := range outputs {
		ghost := ghosts[idx]
		tx.Outputs = append(tx.Outputs, ghost.DumpOutput(output.Threshold, output.Amount))
	}

	return &tx, nil
}

func TransactionFromRaw(raw string) (*Transaction, error) {
	bts, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var tx Transaction
	if err := msgpack.Unmarshal(bts, &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}
