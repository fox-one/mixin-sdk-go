package main

import (
	"context"
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
	pin    = flag.String("pin", "", "pin or tip key")
	asset  = flag.String("asset", "965e5c6e-434c-3fa9-b780-c50f43cd955c", "asset id, default CNB")
	// the extra member of the 1-of-2 multisig group besides the bot itself
	partner = flag.String("partner", "6a00a4bc-229e-3c39-978a-91d2d6c382bf", "the other multisig member id")
)

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

	ctx := context.Background()

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

	members := []string{client.ClientID, *partner}
	var threshold uint8 = 1

	// Phase 1: fund the 1-of-2 multisig address from the bot's own wallet.
	utxo := fundMultisig(ctx, client, spendKey, *asset, members, threshold)
	log.Println("funded multisig utxo", utxo.OutputID)

	// Phase 2: spend from the multisig address back to the bot.
	spendFromMultisig(ctx, client, spendKey, utxo, members)
}

// fundMultisig transfers a small amount from the bot's wallet to the multisig
// address and returns the newly created multisig utxo (synthesized locally).
func fundMultisig(
	ctx context.Context,
	client *mixin.Client,
	spendKey mixinnet.Key,
	assetID string,
	members []string,
	threshold uint8,
) *mixin.SafeUtxo {
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
	b.Memo = "Transfer To Multisig"

	tx, err := client.MakeTransaction(ctx, b, []*mixin.TransactionOutput{
		{
			Address: mixin.RequireNewMixAddress(members, threshold),
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
		RequestID:      mixin.BuildSnapshotID(utxos[0].TransactionHash.String(), utxos[0].OutputIndex, "fund"),
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

	if _, err := client.SafeSubmitTransactionRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      request.RequestID,
		RawTransaction: signedRaw,
	}); err != nil {
		log.Panicf("SafeSubmitTransactionRequest: %v", err)
	}

	// wait for the network to confirm, then synthesize the multisig utxo
	time.Sleep(time.Second * 10)

	return &mixin.SafeUtxo{
		OutputID:           mixin.BuildSnapshotID(tx.Hash.String(), 0, "multisig"),
		KernelAssetID:      tx.Asset,
		TransactionHash:    *tx.Hash,
		OutputIndex:        0,
		Amount:             decimal.RequireFromString(tx.Outputs[0].Amount.String()),
		ReceiversThreshold: threshold,
		Receivers:          members,
	}
}

// spendFromMultisig spends the multisig utxo back to the bot using the safe
// multisig request flow: create request -> sign -> (repeat for each signer).
func spendFromMultisig(
	ctx context.Context,
	client *mixin.Client,
	spendKey mixinnet.Key,
	utxo *mixin.SafeUtxo,
	members []string,
) {
	// the signer index is the position of the bot within the (sorted) receivers
	var k uint16
	for i, member := range utxo.Receivers {
		if member == client.ClientID {
			k = uint16(i)
			break
		}
	}

	b := mixin.NewSafeTransactionBuilder([]*mixin.SafeUtxo{utxo})
	b.Memo = "Transfer From Multisig"

	tx, err := client.MakeTransaction(ctx, b, []*mixin.TransactionOutput{
		{
			Address: mixin.RequireNewMixAddress([]string{client.ClientID}, 1),
			Amount:  utxo.Amount,
		},
	})
	if err != nil {
		log.Panicf("MakeTransaction: %v", err)
	}

	raw, err := tx.Dump()
	if err != nil {
		log.Panicf("Dump: %v", err)
	}

	request, err := client.SafeCreateMultisigRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      mixin.BuildSnapshotID(utxo.TransactionHash.String(), utxo.OutputIndex, "spend"),
		RawTransaction: raw,
	})
	if err != nil {
		log.Panicf("SafeCreateMultisigRequest: %v", err)
	}

	if err := mixin.SafeSignTransaction(tx, spendKey, request.Views, k); err != nil {
		log.Panicf("SafeSignTransaction: %v", err)
	}

	signedRaw, err := tx.Dump()
	if err != nil {
		log.Panicf("Dump: %v", err)
	}

	request, err = client.SafeSignMultisigRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      request.RequestID,
		RawTransaction: signedRaw,
	})
	if err != nil {
		log.Panicf("SafeSignMultisigRequest: %v", err)
	}

	log.Println("multisig request signed", request.RequestID, "signers", request.Signers)
}
