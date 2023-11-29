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

func (i *TransactionInput) AppendUTXO(utxo *InputUTXO) {
	i.Inputs = append(i.Inputs, utxo)
}

func (i *TransactionInput) AppendOutput(output *Output) {
	i.Outputs = append(i.Outputs, output)
}

func (i *TransactionInput) Asset() Hash {
	if len(i.Inputs) == 0 {
		return Hash{}
	}
	return i.Inputs[0].Asset
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
		total = i.TotalInputAmount()
		asset = i.Asset()
	)

	if len(i.Memo) > ExtraSizeGeneralLimit {
		return errors.New("invalid memo, extra too long")
	}

	if len(i.Inputs) > SliceCountLimit || len(i.Outputs) > SliceCountLimit || len(i.References) > SliceCountLimit {
		return fmt.Errorf("invalid tx inputs or outputs %d %d %d", len(i.Inputs), len(i.Outputs), len(i.References))
	}

	for _, input := range i.Inputs {
		if asset != input.Asset {
			return errors.New("invalid input utxo, asset not matched")
		}
	}

	for _, output := range i.Outputs {
		if total = total.Sub(decimal.RequireFromString(output.Amount.String())); total.IsNegative() {
			return errors.New("invalid output: amount exceed")
		}
	}

	if !total.IsZero() {
		return errors.New("invalid output: amount not matched")
	}
	return nil
}

func (input *TransactionInput) BuildTransaction() (*Transaction, error) {
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
