package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/MixinNetwork/mixin/common"
	"github.com/MixinNetwork/mixin/crypto"
	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/sha3"
)

var (
	config = flag.String("config", "", "keystore file path")
	pin    = flag.String("pin", "", "pin")

	ctx = context.Background()
)

func main() {
	flag.Parse()

	f, err := os.Open(*config)
	if err != nil {
		log.Panicln(err)
	}

	var store mixin.Keystore
	if err := json.NewDecoder(f).Decode(&store); err != nil {
		log.Panicln(err)
	}

	client, err := mixin.NewFromKeystore(&store)
	if err != nil {
		log.Panicln(err)
	}

	go sumbitTransactionLoop(ctx, client)

	me, err := client.UserMe(ctx)
	if err != nil {
		log.Panicln(err)
	}

	{
		if _, err := client.Transaction(ctx, &mixin.TransferInput{
			AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c",
			Amount:  decimal.NewFromFloat(1),
			TraceID: uuid.Must(uuid.NewV4()).String(),
			Memo:    "send to multisig",
			OpponentMultisig: struct {
				Receivers []string
				Threshold int64
			}{
				Receivers: []string{client.ClientID, me.App.CreatorID},
				Threshold: 1,
			},
		}, *pin); err != nil {
			log.Panicln(err)
		}
		time.Sleep(time.Second * 5)
	}

	var (
		utxo   *mixin.MultisigUTXO
		offset time.Time
	)
	const limit = 10
	for utxo == nil {
		outputs, err := client.ReadMultisigOutputs(ctx, offset, limit)
		if err != nil {
			log.Panicf("ReadMultisigOutputs: %v", err)
		}

		for _, output := range outputs {
			offset = output.UpdatedAt
			if output.State == mixin.UTXOStateUnspent {
				utxo = output
				break
			}
		}
		if len(outputs) < limit {
			break
		}
	}

	if utxo == nil {
		log.Panicln("No Unspent UTXO")
	}

	tx, err := client.MakeMultisigTransaction(ctx, &mixin.TransactionInput{
		Memo:   "multisig test",
		Inputs: []*mixin.MultisigUTXO{utxo},
		Outputs: []struct {
			Receivers []string
			Threshold int
			Amount    decimal.Decimal
		}{
			{
				[]string{client.ClientID},
				1,
				utxo.Amount,
			},
		},
	})

	if err != nil {
		log.Panicf("MakeMultisigTransaction: %v", err)
	}

	raw := tx.DumpTransaction()

	{
		req, err := client.CreateMultisig(ctx, mixin.MultisigActionSign, raw)
		if err != nil {
			log.Panicf("CreateMultisig: sign %v", err)
		}

		if len(req.Signers) < req.Threshold {
			err = client.CancelMultisig(ctx, req.RequestID)
			if err != nil {
				log.Panicf("CancelMultisig: %v", err)
			}
		}
	}

	{
		req, err := client.CreateMultisig(ctx, mixin.MultisigActionSign, raw)
		if err != nil {
			log.Panicf("CreateMultisig: sign %v", err)
		}

		req, err = client.SignMultisig(ctx, req.RequestID, *pin)
		if err != nil {
			log.Panicf("CreateMultisig: %v", err)
		}

		if len(req.Signers) < req.Threshold {
			req, err = client.CreateMultisig(ctx, mixin.MultisigActionUnlock, raw)
			if err != nil {
				log.Panicf("CreateMultisig: unlock %v", err)
			}

			err = client.UnlockMultisig(ctx, req.RequestID, *pin)
			if err != nil {
				log.Panicf("UnlockMultisig: %v", err)
			}
		}
	}

	{
		req, err := client.CreateMultisig(ctx, mixin.MultisigActionSign, raw)
		if err != nil {
			log.Panicf("CreateMultisig: sign %v", err)
		}

		req, err = client.SignMultisig(ctx, req.RequestID, *pin)
		if err != nil {
			log.Panicf("CreateMultisig: %v", err)
		}
	}
	time.Sleep(time.Second * 10)
}

func sumbitTransactionLoop(ctx context.Context, client *mixin.Client) {
	const (
		limit = 10
	)

	var (
		offset   time.Time
		sleepDur = time.Second
	)

	for {
		select {
		case <-ctx.Done():
			return

		case <-time.After(sleepDur):
			outputs, err := client.ReadMultisigOutputs(ctx, offset, limit)
			if err != nil {
				log.Panicf("ReadMultisigOutputs: %v", err)
			}

			for _, output := range outputs {
				offset = output.UpdatedAt
				if output.State == mixin.UTXOStateSigned && output.SignedBy != "" {
					tx, err := mixin.SendRawTransaction(ctx, output.SignedTx)
					if err != nil {
						log.Printf("ReadMultisigOutputs: %v\n", err)
						continue
					}
					log.Printf("submit transaction: %v", tx.Hash)
				}
			}
			sleepDur = time.Second
		}
	}
}

func buildTransaction(utxo *mixin.MultisigUTXO, ghosts *mixin.GhostKeys) string {
	tx := common.NewTransaction(crypto.Hash(sha3.Sum256([]byte(utxo.AssetID))))
	{
		h, _ := crypto.HashFromString(utxo.TransactionHash)
		tx.AddInput(h, utxo.OutputIndex)

		mask, _ := crypto.KeyFromString(ghosts.Mask)
		var keys = make([]crypto.Key, len(ghosts.Keys))
		for i, k := range ghosts.Keys {
			key, _ := crypto.KeyFromString(k)
			keys[i] = key
		}
		tx.Outputs = append(tx.Outputs, &common.Output{
			Type:   common.TransactionTypeScript,
			Amount: common.NewIntegerFromString(utxo.Amount.String()),
			Keys:   keys,
			Script: common.NewThresholdScript(1),
			Mask:   mask,
		})
	}
	return hex.EncodeToString(tx.AsLatestVersion().Marshal())
}
