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

	if _, err := client.UserMe(ctx); err != nil {
		log.Printf("UserMe: %v", err)
		return
	}

	if err := client.VerifyPin(ctx, *pin); err != nil {
		log.Printf("VerifyPin: %v", err)
		return
	}

	{
		createAndTestUser(ctx, client, mixin.NewKey(rand.Reader).String())
	}
	{
		_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
		createAndTestUser(ctx, client, hex.EncodeToString(privateKey))
	}

	{
		// create sub wallet
		// privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
		_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
		sub, subStore, err := client.CreateUser(ctx, privateKey, "sub user")
		if err != nil {
			log.Panicf("CreateUser: %v", err)
		}

		log.Println("create sub user", sub.UserID)

		// set pin
		newPin := mixin.RandomPin()
		subClient, _ := mixin.NewFromKeystore(subStore)
		log.Println("try ModifyPin", newPin)
		if err := subClient.ModifyPin(ctx, "", newPin); err != nil {
			log.Panicf("ModifyPin (%s) failed: %v", newPin, err)
		}

		tipPin := mixin.NewKey(rand.Reader)
		log.Println("try ModifyPin", tipPin, tipPin.Public())
		if err := subClient.ModifyPin(ctx, newPin, tipPin.Public().String()); err != nil {
			log.Panicf("ModifyPin (%s) failed: %v", tipPin, err)
		}

		if err := subClient.VerifyPin(ctx, tipPin.String()); err != nil {
			log.Panicf("sub user VerifyPin: %v", err)
		}
	}

	{
		// create sub wallet
		// privateKey, _ := rsa.GenerateKey(rand.Reader, 1024)
		_, privateKey, _ := ed25519.GenerateKey(rand.Reader)
		sub, subStore, err := client.CreateUser(ctx, privateKey, "sub user")
		if err != nil {
			log.Panicf("CreateUser: %v", err)
		}

		log.Println("create sub user", sub.UserID)

		// set pin
		subClient, _ := mixin.NewFromKeystore(subStore)
		anotherTipPin := hex.EncodeToString(mixin.GenerateEd25519Key())
		log.Println("try ModifyPin", anotherTipPin[:64], anotherTipPin[64:])
		if err := subClient.ModifyPin(ctx, "", anotherTipPin[64:]); err != nil {
			log.Panicf("ModifyPin (%s) failed: %v", anotherTipPin, err)
		}

		if err := subClient.VerifyPin(ctx, anotherTipPin); err != nil {
			log.Panicf("sub user VerifyPin: %v", err)
		}
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

	testTransfer(ctx, dapp, *pin, sub.UserID)

	// set pin
	newPin := mixin.RandomPin()
	subClient, _ := mixin.NewFromKeystore(subStore)
	log.Println("try ModifyPin", newPin)
	if err := subClient.ModifyPin(ctx, "", newPin); err != nil {
		log.Panicf("ModifyPin (%s) failed: %v", newPin, err)
	}

	tipPin, err := mixin.KeyFromString(userPin)
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

	testTransfer(ctx, subClient, userPin, dapp.ClientID)
}

func testTransfer(ctx context.Context, dapp *mixin.Client, pin, opponent string) {
	input := &mixin.TransferInput{
		AssetID:    "965e5c6e-434c-3fa9-b780-c50f43cd955c", // CNB
		OpponentID: opponent,
		Amount:     decimal.NewFromInt(100),
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
