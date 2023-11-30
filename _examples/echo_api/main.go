package main

import (
	"context"
	"fmt"
	"log"

	"github.com/fox-one/mixin-sdk-go/v2"
)

func main() {
	ctx := context.Background()

	mixin.UseApiHost(mixin.EchoApiHost)
	client := &mixin.Client{}

	fiats, err := client.ReadExchangeRates(ctx)
	if err != nil {
		log.Printf("ReadExchangeRates: %v", err)
		return
	}

	for _, rate := range fiats {
		fmt.Println(rate.Code, rate.Rate)
	}
}
