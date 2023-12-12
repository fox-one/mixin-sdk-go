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

	token, scope, err := mixin.AuthorizeToken(ctx, *clientID, *clientSecret, *code, "")
	if err != nil {
		log.Printf("AuthorizeToken: %v", err)
		return
	}

	log.Println("scope", scope)

	user, err := mixin.UserMe(ctx, token)
	if err != nil {
		log.Printf("UserMe: %v", err)
		return
	}

	log.Println("user", user.UserID)
}
