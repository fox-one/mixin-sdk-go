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

	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/fox-one/mixin-sdk-go/v2/mixinnet"
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
		addr      = mixinnet.GenerateAddress(rand.Reader)
		tx        *mixinnet.Transaction
		privGhost *mixinnet.Key
		mnClient  = mixinnet.NewClient(mixinnet.DefaultLegacyConfig)

		ctx = mnClient.WithHost(context.Background(), mnClient.RandomHost())
	)

	log.Println("addr.private_spend", addr.PrivateSpendKey, "addr.private_view", addr.PrivateViewKey)

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

		h, err := mixinnet.HashFromString(snapshot.TransactionHash)
		if err != nil {
			log.Printf("HashFromString (%s): %v\n", snapshot.TransactionHash, err)
			return
		}

		for {
			if tx, err = mnClient.GetTransaction(ctx, h); err != nil || !tx.Asset.HasValue() {
				log.Printf("GetTransaction %v failed: %v\n", h, err)
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
		if key := mixinnet.ViewGhostOutputKey(tx.Version, &tx.Outputs[0].Keys[0], &addr.PrivateViewKey, &tx.Outputs[0].Mask, 0); key.String() != addr.PublicSpendKey.String() {
			log.Printf("ViewGhostOutputKey check failed: %v != %v\n", key, addr.PublicSpendKey)
			return
		}
		log.Println("ViewGhostOutputKey passed")

		privGhost = mixinnet.DeriveGhostPrivateKey(tx.Version, &tx.Outputs[0].Mask, &addr.PrivateViewKey, &addr.PrivateSpendKey, 0)
		if privGhost.Public().String() != tx.Outputs[0].Keys[0].String() {
			log.Printf("DeriveGhostPrivateKey check failed: expect %v; got priv ghost %v, public %v\n", tx.Outputs[0].Keys[0], privGhost, privGhost.Public())
			return
		}

		{
			raw, err := tx.Dump()
			if err != nil {
				log.Printf("Dump failed: %v\n", err)
				return
			}

			tx1, err := mixinnet.TransactionFromRaw(raw)
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
		builder := mixin.NewLegacyTransactionBuilder([]*mixin.MultisigUTXO{
			{
				AssetID:         "965e5c6e-434c-3fa9-b780-c50f43cd955c",
				TransactionHash: *tx.Hash,
				OutputIndex:     0,
				Amount:          decimal.RequireFromString(tx.Outputs[0].Amount.String()),
				Members:         []string{client.ClientID},
				Threshold:       1,
			},
		})
		builder.Memo = "transaction test to mixinnet address"
		tx, err = client.MakeTransaction(ctx, builder, []*mixin.TransactionOutput{
			{
				Address: mixin.RequireNewMainnetMixAddress([]string{addr.String()}, 1),
				Amount:  decimal.New(1, -8),
			},
		})
		if err != nil {
			log.Printf("MakeTransaction failed: %v\n", err)
			return
		}

		if pub := mixinnet.ViewGhostOutputKey(tx.Version, &tx.Outputs[0].Keys[0], &addr.PrivateViewKey, &tx.Outputs[0].Mask, 0); pub.String() != addr.PublicSpendKey.String() {
			log.Printf("ViewGhostOutputKey check failed: %v != %v\n", pub, addr.PublicSpendKey)
			return
		}
		log.Println("ViewGhostOutputKey passed")

		{
			raw, err := tx.DumpPayload()
			if err != nil {
				log.Printf("DumpPayload: %v\n", err)
				return
			}

			sig := privGhost.Sign(raw)
			tx.Signatures = []map[uint16]*mixinnet.Signature{
				{
					0: &sig,
				},
			}
		}

		raw, err := tx.Dump()
		log.Println("tx.Dump", raw, err)
		if err != nil {
			return
		}

		for {
			_, err := mnClient.SendRawTransaction(ctx, raw)
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
			if tx, err = mnClient.GetTransaction(ctx, h); err != nil || !tx.Asset.HasValue() {
				log.Printf("GetTransaction %v failed: %v\n", h, err)
				time.Sleep(time.Second)
				continue
			}
			break
		}

		if ok, err := mnClient.VerifyTransaction(ctx, addr, *tx.Hash); !ok || err != nil {
			log.Printf("VerifyTransaction %v failed: %v; expect true but got %v", tx.Hash, err, ok)
			return
		}
		log.Println("VerifyTransaction passed")

		privGhost = mixinnet.DeriveGhostPrivateKey(tx.Version, &tx.Outputs[0].Mask, &addr.PrivateViewKey, &addr.PrivateSpendKey, 0)
		if privGhost.Public().String() != tx.Outputs[0].Keys[0].String() {
			log.Printf("DeriveGhostPrivateKey check failed: expect %v; got priv ghost %v, public %v\n", tx.Outputs[0].Keys[0], privGhost, privGhost.Public())
			return
		}
		log.Println("DeriveGhostPrivateKey passed")
	}

	{
		builder := mixin.NewLegacyTransactionBuilder([]*mixin.MultisigUTXO{
			{
				AssetID:         "965e5c6e-434c-3fa9-b780-c50f43cd955c",
				TransactionHash: *tx.Hash,
				OutputIndex:     0,
				Amount:          decimal.RequireFromString(tx.Outputs[0].Amount.String()),
				Members:         []string{client.ClientID},
				Threshold:       1,
			},
		})
		builder.Memo = "transaction test to mixinnet address"
		tx, err = client.MakeTransaction(ctx, builder, []*mixin.TransactionOutput{
			{
				Address: mixin.RequireNewMixAddress([]string{client.ClientID}, 1),
				Amount:  decimal.RequireFromString(tx.Outputs[0].Amount.String()),
			},
		})
		if err != nil {
			log.Printf("MakeTransaction failed: %v\n", err)
			return
		}

		{
			raw, err := tx.DumpPayload()
			if err != nil {
				log.Printf("DumpPayload: %v\n", err)
				return
			}

			sig := privGhost.Sign(raw)
			tx.Signatures = []map[uint16]*mixinnet.Signature{
				{
					0: &sig,
				},
			}
		}

		raw, err := tx.Dump()
		log.Println("tx.Dump:", raw, err)
		if err != nil {
			return
		}

		for {
			_, err := mnClient.SendRawTransaction(ctx, raw)
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
			if tx, err = mnClient.GetTransaction(ctx, h); err != nil || !tx.Asset.HasValue() {
				log.Printf("GetTransaction %v failed: %v\n", h, err)
				time.Sleep(time.Second)
				continue
			}
			break
		}

		if ok, err := mnClient.VerifyTransaction(ctx, addr, *tx.Hash); !ok || err != nil {
			log.Printf("VerifyTransaction %v failed: %v; expect true but got %v", tx.Hash, err, ok)
			return
		}
		log.Println("VerifyTransaction passed")
	}

	log.Println("all passed")
}
