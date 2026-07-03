package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go/v3"
	"github.com/fox-one/mixin-sdk-go/v3/mixinnet"
	"github.com/shopspring/decimal"
)

// SpenderKeystore is the keystore of a Safe-enabled bot together with its Safe
// spend key used to sign UTXO based transactions.
type SpenderKeystore struct {
	mixin.Keystore
	SpendKey string `json:"spend_key"`
}

var (
	config = flag.String("config", "", "keystore file path")
	asset  = flag.String("asset", "965e5c6e-434c-3fa9-b780-c50f43cd955c", "asset id, default CNB")
)

// This example demonstrates the low level mixin kernel primitives:
//   - transfer to a raw mixin kernel address via the Safe network;
//   - fetch the on-chain transaction from a kernel node;
//   - verify the ghost output key and derive its private key.
func main() {
	flag.Parse()

	f, err := os.Open(*config)
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()

	var store SpenderKeystore
	if err := json.NewDecoder(f).Decode(&store); err != nil {
		log.Panicln(err)
	}

	client, err := mixin.NewFromKeystore(&store.Keystore)
	if err != nil {
		log.Panicln(err)
	}

	mnClient := mixinnet.NewClient(mixinnet.DefaultSafeConfig)
	ctx := mnClient.WithHost(context.Background(), mnClient.RandomHost())

	me, err := client.UserMe(ctx)
	if err != nil {
		log.Panicf("UserMe: %v", err)
	}
	if !me.HasSafe {
		log.Panicln("the bot is not migrated to safe network yet, run SafeMigrate first")
	}

	spendKey, err := mixinnet.ParseKeyWithPub(store.SpendKey, me.SpendPublicKey)
	if err != nil {
		log.Panicf("parse spend key: %v", err)
	}

	// a fresh kernel address we fully control the keys of
	addr := mixinnet.GenerateAddress(rand.Reader)
	log.Println("kernel address", addr.String())

	// 1. transfer to the raw kernel address via the safe network
	txHash := transferToAddress(ctx, client, spendKey, *asset, addr)
	log.Println("transfer sent, transaction hash:", txHash)

	hash, err := mixinnet.HashFromString(txHash)
	if err != nil {
		log.Panicf("HashFromString: %v", err)
	}

	// 2. poll the kernel node until the transaction is confirmed
	var tx *mixinnet.Transaction
	for i := 0; i < 30; i++ {
		if tx, err = mnClient.GetTransaction(ctx, hash); err == nil && tx != nil && tx.Asset.HasValue() {
			break
		}
		log.Printf("waiting for transaction %v ...", hash)
		time.Sleep(time.Second * 2)
	}
	if tx == nil || !tx.Asset.HasValue() {
		log.Panicln("transaction not confirmed in time")
	}

	// 3. verify the ghost output key belongs to our address
	output := tx.Outputs[0]
	if key := mixinnet.ViewGhostOutputKey(tx.Version, &output.Keys[0], &addr.PrivateViewKey, &output.Mask, 0); key.String() != addr.PublicSpendKey.String() {
		log.Panicf("ViewGhostOutputKey mismatch: %v != %v", key, addr.PublicSpendKey)
	}
	log.Println("ViewGhostOutputKey passed")

	// 4. derive the ghost private key that can spend this output
	privGhost := mixinnet.DeriveGhostPrivateKey(tx.Version, &output.Mask, &addr.PrivateViewKey, &addr.PrivateSpendKey, 0)
	if privGhost.Public().String() != output.Keys[0].String() {
		log.Panicf("DeriveGhostPrivateKey mismatch: expect %v got %v", output.Keys[0], privGhost.Public())
	}
	log.Println("DeriveGhostPrivateKey passed, all checks ok")
}

func transferToAddress(
	ctx context.Context,
	client *mixin.Client,
	spendKey mixinnet.Key,
	assetID string,
	addr *mixinnet.Address,
) string {
	utxos, err := client.SafeListUtxos(ctx, mixin.SafeListUtxoOption{
		Members: []string{client.ClientID},
		Asset:   assetID,
		State:   mixin.SafeUtxoStateUnspent,
		Limit:   1,
	})
	if err != nil {
		log.Panicf("SafeListUtxos: %v", err)
	}
	if len(utxos) == 0 {
		log.Panicln("empty unspent utxo")
	}

	b := mixin.NewSafeTransactionBuilder(utxos)
	b.Memo = "send to mixin kernel address"

	tx, err := client.MakeTransaction(ctx, b, []*mixin.TransactionOutput{
		{
			Address: mixin.RequireNewMainnetMixAddress([]string{addr.String()}, 1),
			Amount:  decimal.New(1, -8),
		},
	})
	if err != nil {
		log.Panicf("MakeTransaction: %v", err)
	}

	raw, err := tx.Dump()
	if err != nil {
		log.Panicf("Dump: %v", err)
	}

	request, err := client.SafeCreateTransactionRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      mixin.BuildSnapshotID(utxos[0].TransactionHash.String(), utxos[0].OutputIndex, addr.String()),
		RawTransaction: raw,
	})
	if err != nil {
		log.Panicf("SafeCreateTransactionRequest: %v", err)
	}

	if err := mixin.SafeSignTransaction(tx, spendKey, request.Views, 0); err != nil {
		log.Panicf("SafeSignTransaction: %v", err)
	}

	signedRaw, err := tx.Dump()
	if err != nil {
		log.Panicf("Dump: %v", err)
	}

	submitted, err := client.SafeSubmitTransactionRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      request.RequestID,
		RawTransaction: signedRaw,
	})
	if err != nil {
		log.Panicf("SafeSubmitTransactionRequest: %v", err)
	}

	return submitted.TransactionHash
}
