package web3

import (
	"context"
)

func (m *MixinMarketSrv) Web3Tokens(ctx context.Context) (tokens []TokenView, err error) {
	var result struct {
		Data []TokenView `json:"data"`
	}

	err = m.web3Client.Get(
		ctx,
		"/web3/tokens",
		"source=mixin",
		&result,
	)

	return result.Data, err
}

func (m *MixinMarketSrv) Web3Quote(ctx context.Context, req QuoteRequest) (resp QuoteResponseView, err error) {
	var result struct {
		Data QuoteResponseView `json:"data"`
	}

	err = m.web3Client.DoRequest(
		ctx,
		"GET",
		"/web3/quote",
		req.ToQuery(),
		nil,
		&result,
	)

	return result.Data, err
}

func (m *MixinMarketSrv) Web3Swap(ctx context.Context, req SwapRequest) (resp SwapResponseView, err error) {
	var result struct {
		Data SwapResponseView `json:"data"`
	}

	err = m.web3Client.Post(
		ctx,
		"/web3/swap",
		req,
		&result,
	)

	return result.Data, err
}

func (m *MixinMarketSrv) GetWeb3SwapOrder(ctx context.Context, orderId string) (order SwapOrder, err error) {
	var result struct {
		Data SwapOrder `json:"data"`
	}

	err = m.web3Client.Get(
		ctx,
		"/web3/swap/orders/"+orderId,
		"",
		&result,
	)

	return result.Data, err
}
