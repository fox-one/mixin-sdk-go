package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/fox-one/mixin-sdk-go/v2"
)

var (
	clientID = flag.String("client_id", "", "client id")
	scope    = flag.String("scope", "PROFILE:READ", "oauth scope")
	config   = flag.String("config", "", "keystore file path")
	callback = flag.Bool("callback", false, "callback")
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

	scopes := strings.Fields(*scope)

	var verifier, challenge string

	if !*callback {
		verifier, challenge = mixin.RandomCodeChallenge()
	}

	auth, err := mixin.RequestAuthorization(ctx, *clientID, scopes, challenge)
	if err != nil {
		log.Panicln("request authorization failed", err)
	}

	log.Println("auth id is", auth.AuthorizationID)
	auth, err = client.Authorize(ctx, auth.AuthorizationID, scopes, cfg.Pin)
	if err != nil {
		log.Panicln("authorize failed", err)
	}

	if auth.AuthorizationCode == "" {
		log.Println("access denied")
		return
	}

	log.Println("auth code is", auth.AuthorizationCode)

	if showCallback := *callback; showCallback {
		if callbackURL, err := url.Parse(auth.App.RedirectURI); err == nil {
			q := url.Values{}
			q.Set("code", auth.AuthorizationCode)
			callbackURL.RawQuery = q.Encode()
			log.Println("callback url is", callbackURL.String())
		}
	} else {
		token, _, err := mixin.AuthorizeToken(ctx, *clientID, "", auth.AuthorizationCode, verifier)
		if err != nil {
			log.Panicln("authorize token failed", err)
		}

		log.Println("token is", token)
	}
}
