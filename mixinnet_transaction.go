package mixin

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
	"github.com/vmihailenco/msgpack"
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

	TransactionInput struct {
		Memo    string
		Inputs  []*MultisigUTXO
		Outputs []struct {
			Receivers []string
			Threshold uint8
			Amount    decimal.Decimal
		}
	}
)

func (t *Transaction) DumpTransaction() (string, error) {
	bts, err := msgpack.Marshal(t)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bts), nil
}

func (t *Transaction) DumpTransactionPayload() (string, error) {
	sigs := t.Signatures
	t.Signatures = nil
	bts, err := msgpack.Marshal(t)
	t.Signatures = sigs
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(bts), nil
}

func (t *Transaction) TransactionHash() (Hash, error) {
	if t.Hash == nil {
		bts, err := msgpack.Marshal(t)
		if err != nil {
			return Hash{}, err
		}

		h := NewHash(bts)
		t.Hash = &h
	}
	return *t.Hash, nil
}

func (i *TransactionInput) AppendUTXO(utxo *MultisigUTXO) {
	i.Inputs = append(i.Inputs, utxo)
}

func (i *TransactionInput) AppendOutput(receivers []string, threshold uint8, amount decimal.Decimal) {
	i.Outputs = append(i.Outputs, struct {
		Receivers []string
		Threshold uint8
		Amount    decimal.Decimal
	}{
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

	// add outputs
	for i, output := range input.Outputs {
		ghosts, err := c.ReadGhostKeys(ctx, output.Receivers, i)
		if err != nil {
			return nil, err
		}

		tx.Outputs = append(tx.Outputs, ghosts.DumpOutput(uint8(output.Threshold), output.Amount))
	}

	// refund the change
	{
		change := input.TotalInputAmount()
		for _, output := range input.Outputs {
			change = change.Sub(output.Amount)
		}
		if change.IsPositive() {
			ghosts, err := c.ReadGhostKeys(ctx, input.Inputs[0].Members, len(input.Outputs))
			if err != nil {
				return nil, err
			}

			tx.Outputs = append(tx.Outputs, ghosts.DumpOutput(uint8(input.Inputs[0].Threshold), change))
		}
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
