package main

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
	"github.com/shopspring/decimal"
)

const (
	ASSET_CNB = "965e5c6e-434c-3fa9-b780-c50f43cd955c"
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

	if _, err := client.UserMe(ctx); err != nil {
		log.Printf("UserMe: %v", err)
		return
	}

	if err := client.VerifyPin(ctx, *pin); err != nil {
		log.Printf("VerifyPin: %v", err)
		return
	}

	{
		createAndTestUser(ctx, client, mixinnet.GenerateKey(rand.Reader).String())
	}
	{
		_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
		createAndTestUser(ctx, client, hex.EncodeToString(privateKey))
	}
}

func createAndTestUser(ctx context.Context, dapp *mixin.Client, userPin string) {
	// create sub wallet
	// privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
	_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
	sub, subStore, err := dapp.CreateUser(ctx, privateKey, "sub user")
	if err != nil {
		log.Panicf("CreateUser: %v", err)
	}
	log.Println("create sub user", sub.UserID)

	testTransfer(ctx, dapp, *pin, sub.UserID, decimal.NewFromInt(100))

	// set pin
	newPin := mixin.RandomPin()
	subClient, _ := mixin.NewFromKeystore(subStore)
	log.Println("try ModifyPin", newPin)
	if err := subClient.ModifyPin(ctx, "", newPin); err != nil {
		log.Panicf("ModifyPin (%s) failed: %v", newPin, err)
	}

	tipPin, err := mixinnet.KeyFromString(userPin)
	if err != nil {
		log.Panicf("KeyFromString(%s) failed: %v", userPin, err)
	}
	log.Println("try ModifyPin", userPin, tipPin, tipPin.Public())
	if err := subClient.ModifyPin(ctx, newPin, tipPin.Public().String()); err != nil {
		log.Panicf("ModifyPin (%s) failed: %v", tipPin, err)
	}

	if err := subClient.VerifyPin(ctx, userPin); err != nil {
		log.Panicf("sub user VerifyPin: %v", err)
	}

	testTransfer(ctx, subClient, userPin, dapp.ClientID, decimal.NewFromInt(99))
}

func testTransfer(ctx context.Context, dapp *mixin.Client, pin, opponent string, amount decimal.Decimal) {
	{
		input := &mixin.TransferInput{
			AssetID:    ASSET_CNB, // CNB
			OpponentID: opponent,
			Amount:     amount,
			// THIS IS AN EXAMPLE.
			// NEVER USE A RANDOM TRACE ID IN YOU REAL PROJECT.
			TraceID: mixin.RandomTraceID(),
			Memo:    "test",
		}

		snapshot, err := dapp.Transfer(ctx, input, pin)
		if err != nil {
			switch {
			case mixin.IsErrorCodes(err, mixin.InsufficientBalance):
				log.Println("insufficient balance")
			default:
				log.Printf("transfer: %v", err)
			}

			return
		}

		log.Println("transfer done", snapshot.SnapshotID, snapshot.Memo)
		log.Println("sleep 5 seconds")
		time.Sleep(5 * time.Second)

		transfer, err := dapp.ReadTransfer(ctx, snapshot.TraceID)
		if err != nil {
			log.Panicf("ReadTransfer: %v", err)
		}

		if transfer.SnapshotID != snapshot.SnapshotID {
			log.Panicf("expect %v but got %v", snapshot.SnapshotID, transfer.SnapshotID)
		}

		if _, err := dapp.ReadSnapshot(ctx, snapshot.SnapshotID); err != nil {
			log.Panicf("read snapshot: %v", err)
		}
	}

	{
		input := mixin.CreateAddressInput{
			AssetID:     ASSET_CNB,
			Destination: "0xe20FE5C04Fa6b044b720F8CA019Cd896881ED13B",
			Label:       "mixin-sdk-go wallet example test",
		}
		addr, err := dapp.CreateAddress(ctx, input, pin)
		if err != nil {
			log.Panicf("create address: %v", err)
		}

		winput := mixin.WithdrawInput{
			AddressID: addr.AddressID,
			Amount:    decimal.New(1, 0),
			TraceID:   mixin.RandomTraceID(),
			Memo:      "withdraw test",
		}
		if _, err := dapp.Withdraw(ctx, winput, pin); err != nil {
			log.Panicf("withdraw: %v", err)
		}
	}
}
