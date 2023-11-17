package mixin

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type (
	// Only use in Legacy Network
	TransactionOutput struct {
		Receivers []string
		Threshold uint8
		Amount    decimal.Decimal
	}

	// Only use in Legacy Network
	TransactionInput struct {
		Memo       string
		Inputs     []*MultisigUTXO
		Outputs    []TransactionOutput
		References []Hash
		Hint       string
	}
)

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

	if len(i.Memo) > ExtraSizeGeneralLimit {
		return errors.New("invalid memo, extra too long")
	}

	if len(i.Inputs) > SliceCountLimit || len(i.Outputs) > SliceCountLimit || len(i.References) > SliceCountLimit {
		return fmt.Errorf("invalid tx inputs or outputs %d %d %d", len(i.Inputs), len(i.Outputs), len(i.References))
	}

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
		Version:    TxVersion,
		Asset:      input.Asset(),
		Extra:      []byte(input.Memo),
		References: input.References,
	}
	if len(tx.Extra) > tx.getExtraLimit() {
		return nil, errors.New("memo too long")
	}
	// add inputs
	for _, input := range input.Inputs {
		tx.Inputs = append(tx.Inputs, &Input{
			Hash:  &input.TransactionHash,
			Index: uint64(input.OutputIndex),
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
