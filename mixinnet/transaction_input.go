package mixinnet

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type (
	InputUTXO struct {
		Input

		Asset  Hash            `json:"asset"`
		Amount decimal.Decimal `json:"amount,omitempty"`
	}

	TransactionInput struct {
		TxVersion  uint8
		Memo       string
		Inputs     []*InputUTXO
		Outputs    []*Output
		References []Hash
		Hint       string
	}
)

func (input *TransactionInput) Asset() Hash {
	if len(input.Inputs) == 0 {
		return Hash{}
	}
	return input.Inputs[0].Asset
}

func (input *TransactionInput) TotalInputAmount() decimal.Decimal {
	var total decimal.Decimal
	for _, input := range input.Inputs {
		total = total.Add(input.Amount)
	}
	return total
}

func (input *TransactionInput) Validate() error {
	if len(input.Inputs) == 0 {
		return errors.New("no input utxo")
	}

	var (
		total = input.TotalInputAmount()
		asset = input.Asset()
	)

	if len(input.Memo) > ExtraSizeGeneralLimit {
		return errors.New("invalid memo, extra too long")
	}

	if len(input.Inputs) > SliceCountLimit || len(input.Outputs) > SliceCountLimit || len(input.References) > SliceCountLimit {
		return fmt.Errorf("invalid tx inputs or outputs %d %d %d", len(input.Inputs), len(input.Outputs), len(input.References))
	}

	for _, input := range input.Inputs {
		if asset != input.Asset {
			return errors.New("invalid input utxo, asset not matched")
		}
	}

	for _, output := range input.Outputs {
		if total = total.Sub(decimal.RequireFromString(output.Amount.String())); total.IsNegative() {
			return errors.New("invalid output: amount exceed")
		}
	}

	if !total.IsZero() {
		return errors.New("invalid output: amount not matched")
	}
	return nil
}

func (input *TransactionInput) Build() (*Transaction, error) {
	if err := input.Validate(); err != nil {
		return nil, err
	}

	var tx = Transaction{
		Version:    input.TxVersion,
		Asset:      input.Asset(),
		Extra:      []byte(input.Memo),
		References: input.References,
		Outputs:    input.Outputs,
	}
	if len(tx.Extra) > tx.ExtraLimit() {
		return nil, errors.New("memo too long")
	}
	// add inputs
	for _, input := range input.Inputs {
		tx.Inputs = append(tx.Inputs, &input.Input)
	}

	return &tx, nil
}
