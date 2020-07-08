<img align="right" width="256px" src="https://raw.githubusercontent.com/fox-one/mixin-sdk-go/master/logo/logo.png">

# mixin-sdk-go
Golang sdk for Mixin Network & Mixin Messenger

## Install

`go get -u github.com/fox-one/mixin-sdk-go`

## Features

* **Comprehensive** most of the Mixin Network & Mixin Messenger api supported
* **Security** verify Response `X-Request-ID` & signature automatically
* **Flexible** initialize [Client](https://github.com/fox-one/mixin-sdk-go/blob/master/client.go) from `keystore`, `ed25519_oauth_token` or `access_token`

## Examples

See [_examples/](https://github.com/fox-one/mixin-sdk-go/blob/master/_examples/) for a variety of examples.

**Quick Start**

```go
package main

import (
	"context"
	"log"

	"github.com/fox-one/mixin-sdk-go"
)

func main() {
	ctx := context.Background()
	s := &mixin.Keystore{
		ClientID:   "",
		SessionID:  "",
		PrivateKey: "",
		PinToken: "",
	}

	client, err := mixin.NewFromKeystore(s)
	if err != nil {
		log.Panicln(err)
	}

	user, err := client.UserMe(ctx)
	if err != nil {
		log.Printf("UserMe: %v", err)
		return
	}

	log.Println("user id", user.UserID)
}
```

## Error handling?

check error code by `mixin.IsErrorCodes`

```go
if _, err := client.UserMe(ctx); err != nil {
    switch {
    case mixin.IsErrorCodes(err,mixin.Unauthorized,mixin.EndpointNotFound):
    	// handle unauthorized error
    case mixin.IsErrorCodes(err,mixin.InsufficientBalance):
        // handle insufficient balance error
    default:
    }
}
```

