package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
)

var (
	// Specify the keystore file in the -config parameter
	config = flag.String("config", "", "keystore file path")
	pin    = flag.String("pin", "", "pin")
	mint   = flag.Bool("mint", false, "mint new collectibles")
)

func main() {
	// Use flag package to parse the parameters
	flag.Parse()

	// Open the keystore file
	f, err := os.Open(*config)
	if err != nil {
		log.Panicln(err)
	}

	// Read the keystore file as json into mixin.Keystore, which is a go struct
	var store mixin.Keystore
	if err := json.NewDecoder(f).Decode(&store); err != nil {
		log.Panicln(err)
	}

	// Create a Mixin Client from the keystore, which is the instance to invoke Mixin APIs
	client, err := mixin.NewFromKeystore(&store)
	if err != nil {
		log.Panicln(err)
	}

	ctx := context.Background()

	if *mint {
		id, _ := uuid.NewV4()
		token := rand.Int63()
		tr := mixin.NewMintCollectibleTransferInput(id.String(), id.String(), token, mixin.MetaHash(id.Bytes()))
		payment, err := client.VerifyPayment(ctx, tr)
		if err != nil {
			log.Panicln(err)
		}

		fmt.Println("mint collectibles", id.String(), mixin.URL.Codes(payment.CodeID))
		return
	}

	mixin.GetRestyClient().Debug = true

	outputs, err := client.ReadCollectibleOutputs(ctx, []string{client.ClientID}, 1, "", time.Unix(0, 0), 100)
	if err != nil {
		log.Panicln(err)
	}

	for _, output := range outputs {
		switch output.State {
		case mixin.CollectibleOutputStateUnspent:
			token, err := client.ReadCollectiblesToken(ctx, output.TokenID)
			if err != nil {
				log.Panicln(err)
			}

			handleUnspentOutput(ctx, client, output, token)
		case mixin.CollectibleOutputStateSigned:
			handleSignedOutput(ctx, client, output)
		}
	}
}

func handleUnspentOutput(ctx context.Context, client *mixin.Client, output *mixin.CollectibleOutput, token *mixin.CollectibleToken) {
	log.Println("handle unspent output", output.OutputID, token.TokenID)

	receivers := []string{"8017d200-7870-4b82-b53f-74bae1d2dad7"}
	tx, err := client.MakeCollectibleTransaction(ctx, output, token, receivers, 1)
	if err != nil {
		log.Panicln(err)
	}

	signedTx, err := tx.DumpTransaction()
	if err != nil {
		log.Panicln(err)
	}

	// create sign request
	req, err := client.CreateCollectibleRequest(ctx, mixin.CollectibleRequestActionSign, signedTx)
	if err != nil {
		log.Panicln(err)
	}

	// sign
	req, err = client.SignCollectibleRequest(ctx, req.RequestID, *pin)
	if err != nil {
		log.Panicln(err)
	}
}

func handleSignedOutput(ctx context.Context, client *mixin.Client, output *mixin.CollectibleOutput) {
	log.Println("handle signed output", output.OutputID)

	tx, err := mixin.TransactionFromRaw(output.SignedTx)
	if err != nil {
		log.Panicln(err)
	}

	if tx.AggregatedSignature == nil {
		return
	}

	if _, err := client.SendRawTransaction(ctx, output.SignedTx); err != nil {
		log.Panicln(err)
	}
}
