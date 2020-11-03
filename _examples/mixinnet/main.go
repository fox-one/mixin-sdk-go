package main

import (
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
		addr = mixin.NewMixinnetAddress(rand.Reader)
		tx   *mixin.Transaction
	)

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

	if ok, err := mixin.VerifyTransaction(ctx, addr, h); ok || err != nil {
		log.Printf("VerifyTransaction failed: %v; expect false bug got %v", err, ok)
		return
	}
}
