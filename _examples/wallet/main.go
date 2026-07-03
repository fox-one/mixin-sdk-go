package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"log"
	"os"

	"github.com/fox-one/mixin-sdk-go/v3"
	"github.com/fox-one/mixin-sdk-go/v3/mixinnet"
	"github.com/shopspring/decimal"
)

// SpenderKeystore is the keystore of a Safe-enabled bot. On top of the regular
// keystore it carries the Safe spend key that is used to sign UTXO based
// transactions.
type SpenderKeystore struct {
	mixin.Keystore
	SpendKey string `json:"spend_key"`
}

var (
	config   = flag.String("config", "", "keystore file path")
	pin      = flag.String("pin", "", "pin or tip key")
	asset    = flag.String("asset", "965e5c6e-434c-3fa9-b780-c50f43cd955c", "asset id, default CNB")
	receiver = flag.String("receiver", "", "receiver user id")
	amount   = flag.String("amount", "0.0001", "transfer amount")

	ctx = context.Background()
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

	me, err := client.UserMe(ctx)
	if err != nil {
		log.Panicf("UserMe: %v", err)
	}

	if !me.HasSafe {
		log.Panicln("the bot is not migrated to safe network yet, run SafeMigrate first")
	}

	// bind the spend key with its registered public key
	spendKey, err := mixinnet.ParseKeyWithPub(store.SpendKey, me.SpendPublicKey)
	if err != nil {
		log.Panicf("parse spend key: %v", err)
	}

	if *pin != "" {
		if err := client.VerifyPin(ctx, *pin); err != nil {
			log.Panicf("VerifyPin: %v", err)
		}
	}

	if *receiver == "" {
		log.Println("no receiver specified, skip transfer")
		return
	}

	snapshotHash, err := safeTransfer(
		ctx,
		client,
		spendKey,
		*asset,
		*receiver,
		decimal.RequireFromString(*amount),
		"mixin-sdk-go safe transfer example",
	)
	if err != nil {
		log.Panicf("safeTransfer: %v", err)
	}

	log.Println("transfer done, snapshot hash:", snapshotHash)
}

// safeTransfer sends `amount` of `assetID` to a single opponent user via the
// Safe (UTXO) network and returns the resulting transaction hash.
func safeTransfer(
	ctx context.Context,
	client *mixin.Client,
	spendKey mixinnet.Key,
	assetID, opponentID string,
	amount decimal.Decimal,
	memo string,
) (string, error) {
	// 1. list unspent outputs of the given asset
	utxos, err := client.SafeListUtxos(ctx, mixin.SafeListUtxoOption{
		Members: []string{client.ClientID},
		Asset:   assetID,
		State:   mixin.SafeUtxoStateUnspent,
		Limit:   256,
	})
	if err != nil {
		return "", err
	}

	// 2. select enough outputs to cover the amount
	var (
		inputs []*mixin.SafeUtxo
		total  decimal.Decimal
	)
	for _, utxo := range utxos {
		inputs = append(inputs, utxo)
		total = total.Add(utxo.Amount)
		if total.GreaterThanOrEqual(amount) {
			break
		}
	}

	if total.LessThan(amount) {
		return "", errors.New("insufficient balance")
	}

	// 3. build the transaction, change is appended automatically
	b := mixin.NewSafeTransactionBuilder(inputs)
	b.Memo = memo

	tx, err := client.MakeTransaction(ctx, b, []*mixin.TransactionOutput{
		{
			Address: mixin.RequireNewMixAddress([]string{opponentID}, 1),
			Amount:  amount,
		},
	})
	if err != nil {
		return "", err
	}

	raw, err := tx.Dump()
	if err != nil {
		return "", err
	}

	// 4. create the transaction request to fetch the ghost keys (views)
	request, err := client.SafeCreateTransactionRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      mixin.BuildSnapshotID(inputs[0].TransactionHash.String(), inputs[0].OutputIndex, opponentID),
		RawTransaction: raw,
	})
	if err != nil {
		return "", err
	}

	// 5. sign the transaction locally with the spend key
	if err := mixin.SafeSignTransaction(tx, spendKey, request.Views, 0); err != nil {
		return "", err
	}

	signedRaw, err := tx.Dump()
	if err != nil {
		return "", err
	}

	// 6. submit the signed transaction
	submitted, err := client.SafeSubmitTransactionRequest(ctx, &mixin.SafeTransactionRequestInput{
		RequestID:      request.RequestID,
		RawTransaction: signedRaw,
	})
	if err != nil {
		return "", err
	}

	return submitted.TransactionHash, nil
}
