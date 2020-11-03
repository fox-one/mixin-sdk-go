package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/shopspring/decimal"
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

	var (
		addr      = mixin.NewMixinnetAddress(rand.Reader)
		tx        *mixin.Transaction
		privGhost *mixin.Key
	)

	{
		snapshot, err := client.Transaction(ctx, &mixin.TransferInput{
			AssetID:     "965e5c6e-434c-3fa9-b780-c50f43cd955c", // cnb
			Amount:      decimal.NewFromFloat(1),
			TraceID:     mixin.RandomTraceID(),
			Memo:        "send to mixin net address",
			OpponentKey: addr.String(),
		}, *pin)

		if err != nil {
			log.Printf("Transaction: %v", err)
			return
		}

		h, err := mixin.HashFromString(snapshot.TransactionHash)
		if err != nil {
			log.Printf("HashFromString (%s): %v", snapshot.TransactionHash, err)
			return
		}

		for {
			if tx, err = mixin.GetTransaction(ctx, h); err != nil || !tx.Asset.HasValue() {
				log.Printf("GetTransaction %v failed: %v", h, err)
				time.Sleep(time.Second)
				continue
			}
			break
		}

		// verify output
		if key := mixin.ViewGhostOutputKey(&tx.Outputs[0].Keys[0], &addr.PrivateViewKey, &tx.Outputs[0].Mask, 0); key.String() != addr.PublicSpendKey.String() {
			log.Printf("ViewGhostOutputKey check failed: %v != %v", key, addr.PublicSpendKey)
			return
		}

		privGhost = mixin.DeriveGhostPrivateKey(&tx.Outputs[0].Mask, &addr.PrivateViewKey, &addr.PrivateSpendKey, 0)
		if privGhost.Public().String() != tx.Outputs[0].Keys[0].String() {
			log.Printf("DeriveGhostPrivateKey check failed: expect %v; got priv ghost %v, public %v", tx.Outputs[0].Keys[0], privGhost, privGhost.Public())
			return
		}

		if ok, err := mixin.VerifyTransaction(ctx, addr, h); ok || err != nil {
			log.Printf("VerifyTransaction failed: %v; expect false bug got %v", err, ok)
			return
		}
	}

	{
		tx, err = client.MakeMultisigTransaction(ctx, &mixin.TransactionInput{
			Memo: "transaction test",
			Inputs: []*mixin.MultisigUTXO{
				{
					AssetID:         "965e5c6e-434c-3fa9-b780-c50f43cd955c",
					TransactionHash: *tx.Hash,
					OutputIndex:     0,
					Amount:          decimal.NewFromFloat(1),
					Members:         []string{client.ClientID},
					Threshold:       1,
				},
			},
		})

		if err != nil {
			log.Printf("MakeMultisigTransaction: %v", err)
			return
		}

		{
			raw, err := tx.DumpTransactionPayload()
			if err != nil {
				log.Printf("DumpTransactionPayload: %v", err)
				return
			}

			bts, _ := hex.DecodeString(raw)
			tx.Signatures = [][]mixin.Signature{
				{
					privGhost.Sign(bts),
				},
			}
		}

		raw, err := tx.DumpTransaction()
		if err != nil {
			log.Printf("DumpTransaction: %v", err)
			return
		}

		for {
			if tx, err = mixin.SendRawTransaction(ctx, raw); err == nil {
				break
			}
			log.Printf("SendRawTransaction: %v", err)
			time.Sleep(time.Second)
		}

		if ok, err := mixin.VerifyTransaction(ctx, addr, *tx.Hash); !ok || err != nil {
			log.Printf("VerifyTransaction failed: %v; expect true bug got %v", err, ok)
			return
		}
	}
}
