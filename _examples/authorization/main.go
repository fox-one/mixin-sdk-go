package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/fox-one/mixin-sdk-go"
)

var (
	code   = flag.String("code", "", "oauth code")
	scope  = flag.String("scope", "PROFILE:READ", "oauth scope")
	config = flag.String("config", "", "keystore file path")
)

func main() {
	flag.Parse()
	ctx := context.Background()

	var cfg struct {
		mixin.Keystore
		Pin string `json:"pin"`
	}

	// Open the keystore file
	f, err := os.Open(*config)
	if err != nil {
		log.Panicln(err)
	}

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		log.Panicln(err)
	}

	client, err := mixin.NewFromKeystore(&cfg.Keystore)
	if err != nil {
		log.Panicln(err)
	}

	authCode, err := client.GetCode(ctx, *code)
	if err != nil {
		log.Panicln("get code detail failed", err)
	}

	auth := authCode.Authorization()
	if auth == nil {
		log.Panicln("invalid authorization")
	}

	scopes := strings.Fields(*scope)
	auth, err = client.Authorize(ctx, auth.AuthorizationID, scopes, cfg.Pin)
	if err != nil {
		log.Panicln("authorize failed", err)
	}

	if auth.AuthorizationCode == "" {
		log.Println("access denied")
		return
	}

	log.Println("auth code is", auth.AuthorizationCode)

	if callback, err := url.Parse(auth.App.RedirectURI); err == nil {
		q := url.Values{}
		q.Set("code", auth.AuthorizationCode)
		callback.RawQuery = q.Encode()
		log.Println("callback url is", callback.String())
	}
}
