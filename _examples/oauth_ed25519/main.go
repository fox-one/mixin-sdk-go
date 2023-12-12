package main

import (
	"context"
	"flag"
	"log"

	"github.com/fox-one/mixin-sdk-go/v2"
)

var (
	clientID     = flag.String("client", "", "client id")
	clientSecret = flag.String("secret", "", "client secret")
	code         = flag.String("code", "", "oauth code")

	ctx = context.Background()
)

func main() {
	flag.Parse()

	key := mixin.GenerateEd25519Key()
	store, err := mixin.AuthorizeEd25519(ctx, *clientID, *clientSecret, *code, "", key)
	if err != nil {
		log.Printf("AuthorizeEd25519: %v", err)
		return
	}

	client, err := mixin.NewFromOauthKeystore(store)
	if err != nil {
		log.Panicln(err)
	}

	user, err := client.UserMe(ctx)
	if err != nil {
		log.Printf("UserMe: %v", err)
		return
	}

	log.Println("user", user.UserID)
}
