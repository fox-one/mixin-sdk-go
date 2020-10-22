package mixin

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/MixinNetwork/mixin/common"
	"github.com/MixinNetwork/mixin/crypto"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
)

type (
	Input struct {
		Hash  string `json:"hash"`
		Index int64  `json:"index"`
	}

	Output struct {
		Mask   string   `json:"mask"`
		Keys   []string `json:"keys"`
		Amount string   `json:"amount"`
		Script string   `json:"script"`
	}

	Transaction struct {
		Inputs  []*Input  `json:"inputs"`
		Outputs []*Output `json:"outputs"`
		Asset   string    `json:"asset"`
		Extra   string    `json:"extra"`
		Hash    string    `json:"hash"`
	}

	TransactionInput struct {
		Inputs  []*MultisigUTXO
		Outputs []struct {
			Receivers []string
			Threshold int
			Amount    decimal.Decimal
		}
	}
)

func (i *TransactionInput) AppendUTXO(utxo *MultisigUTXO) {
	i.Inputs = append(i.Inputs, utxo)
}

func (i *TransactionInput) AppendOutput(receivers []string, threshold int, amount decimal.Decimal) {
	i.Outputs = append(i.Outputs, struct {
		Receivers []string
		Threshold int
		Amount    decimal.Decimal
	}{
		Receivers: receivers,
		Threshold: threshold,
		Amount:    amount,
	})
}

func (i *TransactionInput) Asset() string {
	if len(i.Inputs) == 0 {
		return ""
	}
	s := sha3.Sum256([]byte(i.Inputs[0].AssetID))
	return hex.EncodeToString(s[:])
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
		s := sha3.Sum256([]byte(input.AssetID))
		if asset != hex.EncodeToString(s[:]) {
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
		if output.Threshold == 0 || output.Threshold > len(output.Receivers) {
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

	var tx Transaction
	tx.Asset = input.Asset()
	// add inputs
	for _, input := range input.Inputs {
		tx.Inputs = append(tx.Inputs, &Input{
			Hash:  input.TransactionHash,
			Index: int64(input.OutputIndex),
		})
	}

	// add outputs
	for i, output := range input.Outputs {
		ghosts, err := c.ReadGhostKeys(ctx, output.Receivers, i)
		if err != nil {
			return nil, err
		}

		tx.Outputs = append(tx.Outputs, ghosts.DumpOutput(output.Threshold, output.Amount))
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

			tx.Outputs = append(tx.Outputs, ghosts.DumpOutput(int(input.Inputs[0].Threshold), change))
		}
	}

	return &tx, nil
}

func TransactionFromRaw(raw string) (*Transaction, error) {
	bts, err := hex.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	ver, err := common.UnmarshalVersionedTransaction(bts)
	if err != nil {
		return nil, err
	}

	var tx = Transaction{
		Inputs:  make([]*Input, len(ver.Inputs)),
		Outputs: make([]*Output, len(ver.Outputs)),
		Asset:   ver.Asset.String(),
		Extra:   string(ver.Extra),
		Hash:    ver.PayloadHash().String(),
	}

	for i, input := range ver.Inputs {
		tx.Inputs[i] = &Input{
			Hash:  input.Hash.String(),
			Index: int64(input.Index),
		}
	}
	for i, output := range ver.Outputs {
		o := &Output{
			Mask:   output.Mask.String(),
			Keys:   make([]string, len(output.Keys)),
			Amount: output.Amount.String(),
			Script: output.Script.String(),
		}
		for i, k := range output.Keys {
			o.Keys[i] = k.String()
		}
		tx.Outputs[i] = o
	}

	return &tx, nil
}

func (t *Transaction) DumpTransaction() string {
	asset, _ := crypto.HashFromString(t.Asset)
	tx := common.NewTransaction(asset)
	for _, utxo := range t.Inputs {
		h, _ := crypto.HashFromString(utxo.Hash)
		tx.AddInput(h, int(utxo.Index))
	}

	for _, output := range t.Outputs {
		mask, _ := crypto.KeyFromString(output.Mask)
		var keys = make([]crypto.Key, len(output.Keys))
		for i, k := range output.Keys {
			key, _ := crypto.KeyFromString(k)
			keys[i] = key
		}
		tx.Outputs = append(tx.Outputs, &common.Output{
			Type:   common.TransactionTypeScript,
			Amount: common.NewIntegerFromString(output.Amount),
			Keys:   keys,
			Script: common.NewThresholdScript(1),
			Mask:   mask,
		})
	}
	return hex.EncodeToString(tx.AsLatestVersion().Marshal())
}
