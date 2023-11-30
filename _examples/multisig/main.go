package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/gofrs/uuid"
	"github.com/shopspring/decimal"
)

var (
	config = flag.String("config", "", "keystore file path")
	pin    = flag.String("pin", "", "pin")
)

func newUUID() string {
	u, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}

	return u.String()
}

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

	mnClient := mixinnet.DefaultClient(false)
	ctx := mnClient.WithHost(context.Background(), mnClient.RandomHost())

	me, err := client.UserMe(ctx)
	if err != nil {
		log.Panicln(err)
	}

	// create sub wallet
	privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	sub, subStore, err := client.CreateUser(ctx, privateKey, "sub user")
	if err != nil {
		log.Printf("CreateUser: %v", err)
		return
	}

	log.Println("create sub user", sub.UserID)

	subClient, _ := mixin.NewFromKeystore(subStore)
	if err := subClient.ModifyPin(ctx, "", *pin); err != nil {
		log.Printf("CreateUser: %v", err)
		return
	}

	members := []string{subClient.ClientID, client.ClientID, me.App.CreatorID}
	var threshold uint8 = 2

	h, err := client.Transaction(ctx, &mixin.TransferInput{
		AssetID: "965e5c6e-434c-3fa9-b780-c50f43cd955c",
		Amount:  decimal.New(1, -4),
		TraceID: newUUID(),
		Memo:    "send to multisig",
		OpponentMultisig: struct {
			Receivers []string `json:"receivers,omitempty"`
			Threshold uint8    `json:"threshold,omitempty"`
		}{
			Receivers: members,
			Threshold: threshold,
		},
	}, *pin)
	if err != nil {
		log.Panicln(err)
	}
	time.Sleep(time.Second * 15)

	var (
		utxo   *mixin.MultisigUTXO
		offset time.Time
	)
	const limit = 10
	for utxo == nil {
		outputs, err := client.ReadMultisigOutputs(ctx, members, threshold, offset, limit)
		if err != nil {
			log.Panicf("ReadMultisigOutputs: %v", err)
		}

		for _, output := range outputs {
			offset = output.UpdatedAt
			if hex.EncodeToString(output.TransactionHash[:]) == h.TransactionHash {
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

	input := mixinnet.TransactionInput{
		TxVersion: mixinnet.TxVersionLegacy,
		Memo:      "multisig test",
		Inputs: []*mixinnet.InputUTXO{
			{
				Input: mixinnet.Input{
					Hash:  &utxo.TransactionHash,
					Index: uint64(utxo.OutputIndex),
				},
				Asset:  utxo.Asset(),
				Amount: utxo.Amount,
			},
		},
		Hint: newUUID(),
	}
	mixAddr, err := mixin.NewMixAddress([]string{client.ClientID}, 1)
	if err != nil {
		log.Panicf("NewMixAddress: %v", err)
	}

	if err := client.AppendOutputsToInput(ctx, &input, []*mixin.TransactionOutputInput{
		{
			Address: *mixAddr,
			Amount:  utxo.Amount,
		},
	}); err != nil {
		log.Panicf("AppendOutputsToInput: %v", err)
	}

	tx, err := input.Build()
	if err != nil {
		log.Panicf("TransactionInput.Build: %v", err)
	}

	raw, err := tx.Dump()
	if err != nil {
		log.Panicf("DumpTransaction: %v", err)
	}

	log.Println("raw transaction", raw)

	{
		req, err := client.CreateMultisig(ctx, mixin.MultisigActionSign, raw)
		if err != nil {
			log.Panicf("CreateMultisig: sign %v", err)
		}

		_, err = client.SignMultisig(ctx, req.RequestID, *pin)
		if err != nil {
			log.Panicf("CreateMultisig: %v", err)
		}

		req, err = subClient.CreateMultisig(ctx, mixin.MultisigActionUnlock, raw)
		if err != nil {
			log.Panicf("CreateMultisig: sign %v", err)
		}

		err = subClient.CancelMultisig(ctx, req.RequestID)
		if err != nil {
			log.Panicf("CancelMultisig: %v", err)
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

		if len(req.Signers) < int(req.Threshold) {
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

		_, err = client.SignMultisig(ctx, req.RequestID, *pin)
		if err != nil {
			log.Panicf("CreateMultisig: %v", err)
		}

		req, err = subClient.CreateMultisig(ctx, mixin.MultisigActionSign, raw)
		if err != nil {
			log.Panicf("CreateMultisig: sign %v", err)
		}

		req, err = subClient.SignMultisig(ctx, req.RequestID, *pin)
		if err != nil {
			log.Panicf("CreateMultisig: %v", err)
		}

		txHash, err := mnClient.SendRawTransaction(ctx, req.RawTransaction)
		if err != nil {
			log.Panicf("SendRawTransaction: %v\n", err)
		}
		log.Printf("submit transaction: %v", txHash)
	}
}
