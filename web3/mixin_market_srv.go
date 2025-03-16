package web3

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	bot "github.com/MixinNetwork/bot-api-go-client/v3"
	"github.com/fox-one/mixin-sdk-go/v2"
	"github.com/fox-one/mixin-sdk-go/v2/cacheflight"
	"github.com/go-resty/resty/v2"
)

type MixinMarketSrv struct {
	client     *resty.Client
	web3Client Web3Client

	assetInfoCache *cacheflight.Group

	priceHistoryFactory *PriceHistoryFactory
	historyPriceCache   map[HistoryPriceType]*cacheflight.Group
}

func NewMixinMarketSrv(keystore *mixin.Keystore, spendkey string) *MixinMarketSrv {
	historyPriceCache := make(map[HistoryPriceType]*cacheflight.Group)
	factory := NewPriceHistoryFactory()
	for t := range factory.strategies {
		strategy := factory.GetStrategy(t)
		historyPriceCache[t] = cacheflight.New(
			strategy.GetCacheDuration(),
			strategy.GetCacheDirtyDuration(),
		)
	}

	su := &bot.SafeUser{
		UserId:            keystore.ClientID,
		SessionId:         keystore.SessionID,
		SessionPrivateKey: keystore.SessionPrivateKey,
		ServerPublicKey:   keystore.ServerPublicKey,
		SpendPrivateKey:   spendkey,
	}
	logger := slog.Default()
	client := bot.NewDefaultClient(su, logger)

	return &MixinMarketSrv{
		web3Client: NewWeb3Client(client),
		client: resty.New().
			SetHeader("Content-Type", "application/json").
			SetBaseURL(MixinRouteApiPrefix).
			SetTimeout(10 * time.Second),
		assetInfoCache: cacheflight.New(
			assetInfoCacheDuration,
			assetInfoCacheDirtyDuration,
		),
		priceHistoryFactory: factory,
		historyPriceCache:   historyPriceCache,
	}
}

/*
GET /markets/:coin_idcoin_id: string, coin_id from GET /markets. OR mixin asset idresponse:
*/
func (m *MixinMarketSrv) GetAssetInfo(ctx context.Context, assetId string) (*MarketAssetInfo, error) {
	var response struct {
		Data MarketAssetInfo `json:"data"`
	}

	data, err := m.assetInfoCache.Do(assetId, func() (interface{}, error) {
		_, err := m.client.R().SetContext(ctx).SetPathParams(map[string]string{
			"coin_id": assetId,
		}).SetResult(&response).Get("/markets/{coin_id}")
		if err != nil {
			return nil, err
		}
		return &response.Data, nil
	})
	if err != nil {
		return nil, err
	}

	return data.(*MarketAssetInfo), nil
}

/*
GET /markets/:coin_id/price-history?type=${type}paramdescriptioncoin_idcoin_id from GET /markets, or mixin asset idtype1D, 1W, 1M, YTD, ALLresponse:
*/
func (m *MixinMarketSrv) GetPriceHistory(ctx context.Context, assetId string, t HistoryPriceType) (*HistoricalPrice, error) {
	strategy := m.priceHistoryFactory.GetStrategy(t)
	if strategy == nil {
		return nil, fmt.Errorf("unsupported price history type: %v", t)
	}

	data, err := m.historyPriceCache[t].Do(assetId, func() (interface{}, error) {
		var response struct {
			Data HistoricalPrice `json:"data"`
		}

		_, err := m.client.R().
			SetContext(ctx).
			SetPathParams(map[string]string{
				"coin_id": assetId,
			}).
			SetQueryParam("type", strategy.GetType()).
			SetResult(&response).
			Get("/markets/{coin_id}/price-history")

		if err != nil {
			return nil, err
		}
		return &response.Data, nil
	})

	if err != nil {
		return nil, err
	}

	return data.(*HistoricalPrice), nil
}
