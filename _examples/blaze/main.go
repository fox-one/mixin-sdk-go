package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/fox-one/mixin-sdk-go"
	"github.com/gofrs/uuid"
)

var (
	config = flag.String("config", "", "keystore file path")
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

	h := func(ctx context.Context, msg *mixin.MessageView, userID string) error {
		if userID, _ := uuid.FromString(msg.UserID); userID == uuid.Nil {
			return nil
		}

		id, _ := uuid.FromString(msg.MessageID)

		reply := &mixin.MessageRequest{
			ConversationID: msg.ConversationID,
			RecipientID:    msg.UserID,
			MessageID:      uuid.NewV5(id, "reply").String(),
			Category:       msg.Category,
			Data:           msg.Data,
		}

		return client.SendMessage(ctx, reply)
	}

	ctx := context.Background()

	for {
		if err := client.LoopBlaze(ctx, mixin.BlazeListenFunc(h)); err != nil {
			log.Printf("LoopBlaze: %v", err)
		}

		time.Sleep(time.Second)
	}
}
