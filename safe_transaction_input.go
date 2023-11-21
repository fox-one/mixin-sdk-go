package mixin

import (
	"context"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type (
	SafeTransactionOutput struct {
		Receivers []string
		Threshold uint8
		Amount    decimal.Decimal
	}

	SafeTransactionInput struct {
		Memo       string
		Inputs     []*SafeUtxo
		Outputs    []SafeTransactionOutput
		References []Hash
		Hint       string
	}
)

func (i *SafeTransactionInput) AppendUTXO(utxo *SafeUtxo) {
	i.Inputs = append(i.Inputs, utxo)
}

func (i *SafeTransactionInput) AppendOutput(receivers []string, threshold uint8, amount decimal.Decimal) {
	i.Outputs = append(i.Outputs, SafeTransactionOutput{
		Receivers: receivers,
		Threshold: threshold,
		Amount:    amount,
	})
}

func (i *SafeTransactionInput) Asset() Hash {
	if len(i.Inputs) == 0 {
		return Hash{}
	}
	return i.Inputs[0].Asset
}

func (i *SafeTransactionInput) TotalInputAmount() decimal.Decimal {
	var total decimal.Decimal
	for _, input := range i.Inputs {
		total = total.Add(input.Amount)
	}
	return total
}

func (i *SafeTransactionInput) Validate() error {
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
		if asset.String() != input.Asset.String() {
			return errors.New("invalid input utxo, asset not matched")
		}

		if len(members) == 0 {
			for _, u := range input.Receivers {
				members[u] = true
			}
			continue
		}
		if len(members) != len(input.Receivers) {
			return errors.New("invalid input utxo, member not matched")
		}
		for _, m := range input.Receivers {
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

func (c *Client) SafeBuildTransaction(ctx context.Context, input *SafeTransactionInput) (*Transaction, error) {
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
			outputs = append(outputs, SafeTransactionOutput{
				Receivers: input.Inputs[0].Receivers,
				Threshold: input.Inputs[0].ReceiversThreshold,
				Amount:    change,
			})
		}
	}

	ghostInputs := make([]*GhostKeyInput, 0, len(outputs))
	for idx, output := range outputs {
		ghostInputs = append(ghostInputs, &GhostKeyInput{
			Receivers: output.Receivers,
			Index:     idx,
			Hint:      uuidHash(append([]byte(input.Hint), byte(idx))),
		})
	}

	ghosts, err := c.SafeCreateGhostKeys(ctx, ghostInputs)
	if err != nil {
		return nil, err
	}

	for idx, output := range outputs {
		ghost := ghosts[idx]
		tx.Outputs = append(tx.Outputs, ghost.DumpOutput(output.Threshold, output.Amount))
	}

	return &tx, nil
}
