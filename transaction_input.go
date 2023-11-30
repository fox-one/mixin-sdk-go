package mixin

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

type (
	TransactionOutputInput struct {
		Address MixAddress      `json:"address,omitempty"`
		Amount  decimal.Decimal `json:"amount,omitempty"`
	}
)

func (c *Client) CreateGhostKeys(ctx context.Context, txVer uint8, inputs []*SafeGhostKeyInput) ([]*SafeGhostKeys, error) {
	var resp []*SafeGhostKeys
	if txVer < mixinnet.TxVersionHashSignature {
		// legacy
		if err := c.Post(ctx, "/outputs", inputs, &resp); err != nil {
			return nil, err
		}
	} else {
		// safe
		if err := c.Post(ctx, "/safe/keys", inputs, &resp); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func (c *Client) AppendOutputsToInput(ctx context.Context, input *mixinnet.TransactionInput, outputs []*TransactionOutputInput) error {
	if input == nil {
		return nil
	}
	if input.Hint == "" {
		return fmt.Errorf("empty hint: hint should be unique uuid string")
	}

	ghostInputs := make([]*SafeGhostKeyInput, 0, len(outputs))
	for i, output := range outputs {
		if len(output.Address.uuidMembers) > 0 {
			ghostInputs = append(ghostInputs, &SafeGhostKeyInput{
				Receivers: output.Address.Members(),
				Index:     len(input.Outputs) + i,
				Hint:      uuidHash([]byte(fmt.Sprintf("trace:%s;index:%d", input.Hint, len(input.Outputs)+i))),
			})
		}
	}

	var ghostKeys []*SafeGhostKeys
	if len(ghostInputs) > 0 {
		ghosts, err := c.CreateGhostKeys(ctx, input.TxVersion, ghostInputs)
		if err != nil {
			return err
		}
		ghostKeys = ghosts
	}

	if len(ghostKeys) != len(ghostInputs) {
		return fmt.Errorf("invalid ghost keys count: %d", len(ghostKeys))
	}

	ghostKeyOffset := 0
	for i, output := range outputs {
		if len(output.Address.xinMembers) > 0 {
			r := mixinnet.GenerateKey(rand.Reader)
			keys := make([]mixinnet.Key, len(output.Address.xinMembers))
			for i, addr := range output.Address.xinMembers {
				keys[i] = *mixinnet.DeriveGhostPublicKey(input.TxVersion, &r, &addr.PublicViewKey, &addr.PublicSpendKey, uint64(len(input.Outputs)))
			}
			input.Outputs = append(input.Outputs, &mixinnet.Output{
				Type:   mixinnet.OutputTypeScript,
				Amount: mixinnet.IntegerFromDecimal(outputs[i].Amount),
				Script: mixinnet.NewThresholdScript(outputs[i].Address.Threshold),
				Keys:   keys,
				Mask:   r.Public(),
			})
		} else {
			ghost := ghostKeys[ghostKeyOffset]
			ghostKeyOffset++
			input.Outputs = append(input.Outputs, &mixinnet.Output{
				Type:   mixinnet.OutputTypeScript,
				Amount: mixinnet.IntegerFromDecimal(outputs[i].Amount),
				Script: mixinnet.NewThresholdScript(outputs[i].Address.Threshold),
				Keys:   ghost.Keys,
				Mask:   ghost.Mask,
			})
		}
	}
	return nil
}

func (c *Client) MakeLegacyTransaction(ctx context.Context, hint string, utxos []*MultisigUTXO, outputs []*TransactionOutputInput, references []mixinnet.Hash, memo string) (*mixinnet.Transaction, error) {
	if len(utxos) == 0 {
		return nil, fmt.Errorf("empty utxos")
	}

	input := &mixinnet.TransactionInput{
		TxVersion:  mixinnet.TxVersionLegacy,
		Memo:       memo,
		Hint:       hint,
		References: references,
		Inputs:     make([]*mixinnet.InputUTXO, len(utxos)),
	}
	for i, utxo := range utxos {
		input.Inputs[i] = &mixinnet.InputUTXO{
			Input: mixinnet.Input{
				Hash:  &utxo.TransactionHash,
				Index: uint64(utxo.OutputIndex),
			},
			Asset:  utxo.Asset(),
			Amount: utxo.Amount,
		}
	}

	// refund the change
	{
		change := input.TotalInputAmount()
		for _, output := range outputs {
			change = change.Sub(output.Amount)
		}

		if change.IsPositive() {
			mixAddr, err := NewMixAddress(utxos[0].Members, utxos[0].Threshold)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &TransactionOutputInput{
				Address: *mixAddr,
				Amount:  change,
			})
		}
	}

	if err := c.AppendOutputsToInput(ctx, input, outputs); err != nil {
		return nil, err
	}

	return input.Build()
}

func (c *Client) MakeSafeTransaction(ctx context.Context, hint string, utxos []*SafeUtxo, outputs []*TransactionOutputInput, references []mixinnet.Hash, memo string) (*mixinnet.Transaction, error) {
	if len(utxos) == 0 {
		return nil, fmt.Errorf("empty utxos")
	}

	input := &mixinnet.TransactionInput{
		TxVersion:  mixinnet.TxVersion,
		Memo:       memo,
		Hint:       hint,
		References: references,
		Inputs:     make([]*mixinnet.InputUTXO, len(utxos)),
	}
	for i, utxo := range utxos {
		input.Inputs[i] = &mixinnet.InputUTXO{
			Input: mixinnet.Input{
				Hash:  &utxo.TransactionHash,
				Index: uint64(utxo.OutputIndex),
			},
			Asset:  utxo.Asset,
			Amount: utxo.Amount,
		}
	}

	// refund the change
	{
		change := input.TotalInputAmount()
		for _, output := range outputs {
			change = change.Sub(output.Amount)
		}

		if change.IsPositive() {
			mixAddr, err := NewMixAddress(utxos[0].Receivers, utxos[0].ReceiversThreshold)
			if err != nil {
				return nil, err
			}
			outputs = append(outputs, &TransactionOutputInput{
				Address: *mixAddr,
				Amount:  change,
			})
		}
	}

	if err := c.AppendOutputsToInput(ctx, input, outputs); err != nil {
		return nil, err
	}

	return input.Build()
}
