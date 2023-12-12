package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"log"
	"os"

	"github.com/fox-one/mixin-sdk-go/v2"
)

var (
	// Specify the keystore file in the -config parameter
	config = flag.String("config", "", "keystore file path")
	text   = flag.String("text", "hello world", "text message")
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

	me, err := client.UserMe(ctx)
	if err != nil {
		log.Fatalln(err)
	}

	if me.App == nil {
		log.Fatalln("use a bot keystore instead")
	}

	receiptID := me.App.CreatorID

	sessions, err := client.FetchSessions(ctx, []string{receiptID})
	if err != nil {
		log.Fatalln(err)
	}

	_ = sessions

	req := &mixin.MessageRequest{
		ConversationID: mixin.UniqueConversationID(client.ClientID, receiptID),
		RecipientID:    receiptID,
		MessageID:      mixin.RandomTraceID(),
		Category:       mixin.MessageCategoryPlainText,
		Data:           base64.StdEncoding.EncodeToString([]byte(*text)),
	}

	if err := client.EncryptMessageRequest(req, sessions); err != nil {
		log.Fatalln(err)
	}

	receipts, err := client.SendEncryptedMessages(ctx, []*mixin.MessageRequest{req})
	if err != nil {
		log.Fatalln(err)
	}

	b, _ := json.Marshal(receipts)
	log.Println(string(b))
}
