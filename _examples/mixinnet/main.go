package main

import (
	"bytes"
	"context"
	"crypto/rand"
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

		ctx = mixin.WithMixinNetHost(context.Background(), mixin.RandomMixinNetHost())
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
			log.Printf("Transaction: %v\n", err)
			return
		}
		log.Printf("send to address: %v; hash: %v\n", addr, snapshot.TransactionHash)

		h, err := mixin.HashFromString(snapshot.TransactionHash)
		if err != nil {
			log.Printf("HashFromString (%s): %v\n", snapshot.TransactionHash, err)
			return
		}

		for {
			if tx, err = client.GetRawTransaction(ctx, h); err != nil || !tx.Asset.HasValue() {
				log.Printf("GetRawTransaction %v failed: %v\n", h, err)
				time.Sleep(time.Second)
				continue
			}
			break
		}

		{
			tx.Hash = nil
			hash1, err := tx.TransactionHash()
			if err != nil {
				log.Printf("tx.TransactionHash: %v\n", err)
				return
			}
			if !bytes.Equal(h[:], hash1[:]) {
				log.Println("transaction hash verify failed", h, hash1)
			}
		}

		// verify output
		if key := mixin.ViewGhostOutputKey(&tx.Outputs[0].Keys[0], &addr.PrivateViewKey, &tx.Outputs[0].Mask, 0); key.String() != addr.PublicSpendKey.String() {
			log.Printf("ViewGhostOutputKey check failed: %v != %v\n", key, addr.PublicSpendKey)
			return
		}
		log.Println("ViewGhostOutputKey passed")

		privGhost = mixin.DeriveGhostPrivateKey(&tx.Outputs[0].Mask, &addr.PrivateViewKey, &addr.PrivateSpendKey, 0)
		if privGhost.Public().String() != tx.Outputs[0].Keys[0].String() {
			log.Printf("DeriveGhostPrivateKey check failed: expect %v; got priv ghost %v, public %v\n", tx.Outputs[0].Keys[0], privGhost, privGhost.Public())
			return
		}

		{
			raw, err := tx.DumpTransaction()
			if err != nil {
				log.Printf("DumpTransaction failed: %v\n", err)
				return
			}

			tx1, err := mixin.TransactionFromRaw(raw)
			if err != nil {
				log.Printf("TransactionFromRaw failed: %v\n", err)
				return
			}

			hash, err := tx1.TransactionHash()
			if err != nil {
				log.Printf("TransactionHash failed: %v\n", err)
				return
			}

			if !bytes.Equal(h[:], hash[:]) {
				log.Println("Marshal & Unmarshal failed, hash not matched")
				return
			}
			log.Println("Marshal & Unmarshal passed")
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
			log.Printf("MakeMultisigTransaction: %v\n", err)
			return
		}

		{
			raw, err := tx.DumpTransactionPayload()
			if err != nil {
				log.Printf("DumpTransactionPayload: %v\n", err)
				return
			}

			sig := privGhost.Sign(raw)
			tx.Signatures = []map[uint16]*mixin.Signature{
				{
					0: &sig,
				},
			}
		}

		raw, err := tx.DumpTransaction()
		if err != nil {
			log.Printf("DumpTransaction: %v\n", err)
			return
		}

		for {
			_, err := client.SendRawTransaction(ctx, raw)
			if err == nil || mixin.IsErrorCodes(err, mixin.InvalidOutputKey) {
				break
			}
			log.Printf("SendRawTransaction: %v\n", err)
			time.Sleep(time.Second)
		}

		h, err := tx.TransactionHash()
		if err != nil {
			log.Printf("TransactionHash: %v\n", err)
			return
		}
		log.Println("Transaction sent,", h)

		for i := 0; i < 5; i++ {
			if tx, err = client.GetRawTransaction(ctx, h); err != nil || !tx.Asset.HasValue() {
				log.Printf("GetRawTransaction %v failed: %v\n", h, err)
				time.Sleep(time.Second)
				continue
			}
			break
		}

		if ok, err := mixin.VerifyTransaction(ctx, addr, *tx.Hash); !ok || err != nil {
			log.Printf("VerifyTransaction %v failed: %v; expect true but got %v", tx.Hash, err, ok)
			return
		}
		log.Println("VerifyTransaction passed")
	}

	log.Println("all passed")
}
