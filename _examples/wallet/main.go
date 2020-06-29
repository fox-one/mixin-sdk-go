package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"

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

	me, err := client.UserMe(ctx)
	if err != nil {
		log.Printf("UserMe: %v", err)
		return
	}

	if err := client.VerifyPin(ctx, *pin); err != nil {
		log.Printf("VerifyPin: %v", err)
		return
	}

	if app := me.App; app != nil {
		input := &mixin.TransferInput{
			AssetID:    "965e5c6e-434c-3fa9-b780-c50f43cd955c", // cnb
			OpponentID: app.CreatorID,
			Amount:     decimal.NewFromInt(100),
			TraceID:    mixin.RandomTraceID(),
			Memo:       "test",
		}

		snapshot, err := client.Transfer(ctx, input, *pin)
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

		if _, err := client.ReadSnapshot(ctx, snapshot.SnapshotID); err != nil {
			log.Printf("read snapshot: %v", err)
			return
		}
	}
}
