package mixin

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

func (c *Client) CreateGhostKeys(ctx context.Context, txVer uint8, inputs []*SafeGhostKeyInput) ([]*SafeGhostKeys, error) {
	path := "/safe/keys"
	if txVer < mixinnet.TxVersionHashSignature {
		path = "/outputs"
	}

	var resp []*SafeGhostKeys
	if err := c.Post(ctx, path, inputs, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

type TransactionBuilder struct {
	*mixinnet.TransactionInput
	addr *MixAddress
}

type TransactionOutput struct {
	Address *MixAddress     `json:"address,omitempty"`
	Amount  decimal.Decimal `json:"amount,omitempty"`
}

func NewLegacyTransactionBuilder(utxos []*MultisigUTXO) *TransactionBuilder {
	b := &TransactionBuilder{
		TransactionInput: &mixinnet.TransactionInput{
			TxVersion: mixinnet.TxVersionLegacy,
			Hint:      newUUID(),
			Inputs:    make([]*mixinnet.InputUTXO, len(utxos)),
		},
	}

	for i, utxo := range utxos {
		b.Inputs[i] = &mixinnet.InputUTXO{
			Input: mixinnet.Input{
				Hash:  &utxo.TransactionHash,
				Index: uint8(utxo.OutputIndex),
			},
			Asset:  utxo.Asset(),
			Amount: utxo.Amount,
		}

		addr, err := NewMixAddress(utxo.Members, utxo.Threshold)
		if err != nil {
			panic(err)
		}

		if i == 0 {
			b.addr = addr
		} else if b.addr.String() != addr.String() {
			panic("invalid utxos")
		}
	}

	return b
}

func NewSafeTransactionBuilder(utxos []*SafeUtxo) *TransactionBuilder {
	b := &TransactionBuilder{
		TransactionInput: &mixinnet.TransactionInput{
			TxVersion: mixinnet.TxVersion,
			Hint:      newUUID(),
			Inputs:    make([]*mixinnet.InputUTXO, len(utxos)),
		},
	}

	for i, utxo := range utxos {
		b.Inputs[i] = &mixinnet.InputUTXO{
			Input: mixinnet.Input{
				Hash:  &utxo.TransactionHash,
				Index: utxo.OutputIndex,
			},
			Asset:  utxo.Asset,
			Amount: utxo.Amount,
		}

		addr, err := NewMixAddress(utxo.Receivers, utxo.ReceiversThreshold)
		if err != nil {
			panic(err)
		}

		if i == 0 {
			b.addr = addr
		} else if b.addr.String() != addr.String() {
			panic("invalid utxos")
		}
	}

	return b
}

func (c *Client) MakeTransaction(ctx context.Context, b *TransactionBuilder, outputs []*TransactionOutput) (*mixinnet.Transaction, error) {
	remain := b.TotalInputAmount()
	for _, output := range outputs {
		remain = remain.Sub(output.Amount)
	}

	if remain.IsPositive() {
		outputs = append(outputs, &TransactionOutput{
			Address: b.addr,
			Amount:  remain,
		})
	}

	if err := c.AppendOutputsToInput(ctx, b.TransactionInput, outputs); err != nil {
		return nil, err
	}

	return b.Build()
}

func (c *Client) AppendOutputsToInput(ctx context.Context, input *mixinnet.TransactionInput, outputs []*TransactionOutput) error {
	var (
		ghostInputs  []*SafeGhostKeyInput
		ghostOutputs []*mixinnet.Output
	)

	for _, output := range outputs {
		txOutput := &mixinnet.Output{
			Type:   mixinnet.OutputTypeScript,
			Amount: mixinnet.IntegerFromDecimal(output.Amount),
			Script: mixinnet.NewThresholdScript(output.Address.Threshold),
		}

		if len(output.Address.xinMembers) > 0 {
			r := mixinnet.GenerateKey(rand.Reader)
			for _, addr := range output.Address.xinMembers {
				key := mixinnet.DeriveGhostPublicKey(input.TxVersion, &r, &addr.PublicViewKey, &addr.PublicSpendKey, uint8(len(input.Outputs)))
				txOutput.Keys = append(txOutput.Keys, *key)
			}

			txOutput.Mask = r.Public()
		} else if len(output.Address.uuidMembers) > 0 {
			index := uint8(len(input.Outputs))
			ghostInputs = append(ghostInputs, &SafeGhostKeyInput{
				Receivers: output.Address.Members(),
				Index:     index,
				Hint:      uuidHash([]byte(fmt.Sprintf("hint:%s;index:%d", input.Hint, index))),
			})

			ghostOutputs = append(ghostOutputs, txOutput)
		}

		input.Outputs = append(input.Outputs, txOutput)
	}

	if len(ghostInputs) > 0 {
		keys, err := c.CreateGhostKeys(ctx, input.TxVersion, ghostInputs)
		if err != nil {
			return err
		}

		for i, key := range keys {
			output := ghostOutputs[i]
			output.Keys = key.Keys
			output.Mask = key.Mask
		}
	}

	return nil
}
