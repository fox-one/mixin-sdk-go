package mixin

import (
	"context"
	"fmt"
	"sort"

	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

func (c *Client) createGhostKeys(ctx context.Context, txVer uint8, inputs []*GhostInput, senders []string) ([]*GhostKeys, error) {
	// sort receivers
	for _, input := range inputs {
		sort.Strings(input.Receivers)
	}

	if txVer < mixinnet.TxVersionHashSignature {
		return c.BatchReadGhostKeys(ctx, inputs)
	}

	return c.SafeCreateGhostKeys(ctx, inputs, senders...)
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
			Asset:  utxo.KernelAssetID,
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

	if err := c.AppendOutputsToInput(ctx, b, outputs); err != nil {
		return nil, err
	}

	return b.Build()
}

func (c *Client) AppendOutputsToInput(ctx context.Context, b *TransactionBuilder, outputs []*TransactionOutput) error {
	var (
		ghostInputs  []*GhostInput
		ghostOutputs []*mixinnet.Output
	)

	for _, output := range outputs {
		txOutput := &mixinnet.Output{
			Type:   mixinnet.OutputTypeScript,
			Amount: mixinnet.IntegerFromDecimal(output.Amount),
			Script: mixinnet.NewThresholdScript(output.Address.Threshold),
		}

		index := uint8(len(b.Outputs))
		if len(output.Address.xinMembers) > 0 {
			key := SafeCreateXinAddressGhostKeys(b.TxVersion, output.Address.xinMembers, index)
			txOutput.Mask = key.Mask
			txOutput.Keys = key.Keys
		} else if len(output.Address.uuidMembers) > 0 {
			ghostInputs = append(ghostInputs, &GhostInput{
				Receivers: output.Address.Members(),
				Index:     index,
				Hint:      uuidHash([]byte(fmt.Sprintf("hint:%s;index:%d", b.Hint, index))),
			})

			ghostOutputs = append(ghostOutputs, txOutput)
		}

		b.Outputs = append(b.Outputs, txOutput)
	}

	if len(ghostInputs) > 0 {
		keys, err := c.createGhostKeys(ctx, b.TxVersion, ghostInputs, b.addr.Members())
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
