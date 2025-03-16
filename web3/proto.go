package web3

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type (
	MarketAssetInfo struct {
		CoinID                       string          `json:"coin_id"`
		Name                         string          `json:"name"`
		Symbol                       string          `json:"symbol"`
		IconURL                      string          `json:"icon_url"`
		CurrentPrice                 decimal.Decimal `json:"current_price"`
		MarketCap                    decimal.Decimal `json:"market_cap"`                       // 市值
		MarketCapRank                string          `json:"market_cap_rank"`                  // 市值排名
		TotalVolume                  decimal.Decimal `json:"total_volume"`                     // 24小时交易量
		High24H                      decimal.Decimal `json:"high_24h"`                         // 24小时最高价
		Low24H                       decimal.Decimal `json:"low_24h"`                          // 24小时最低价
		PriceChange24H               decimal.Decimal `json:"price_change_24h"`                 // 24小时价格变动
		PriceChangePercentage1H      decimal.Decimal `json:"price_change_percentage_1h"`       // 1小时价格变动百分比
		PriceChangePercentage24H     decimal.Decimal `json:"price_change_percentage_24h"`      // 24小时价格变动百分比
		PriceChangePercentage7D      decimal.Decimal `json:"price_change_percentage_7d"`       // 7天价格变动百分比
		PriceChangePercentage30D     decimal.Decimal `json:"price_change_percentage_30d"`      // 30天价格变动百分比
		MarketCapChange24H           decimal.Decimal `json:"market_cap_change_24h"`            // 市值变动
		MarketCapChangePercentage24H decimal.Decimal `json:"market_cap_change_percentage_24h"` // 市值变动百分比
		CirculatingSupply            decimal.Decimal `json:"circulating_supply"`               // 流通供应量
		TotalSupply                  decimal.Decimal `json:"total_supply"`                     // 总供应量
		MaxSupply                    decimal.Decimal `json:"max_supply"`                       // 最大供应量
		Ath                          decimal.Decimal `json:"ath"`                              // 历史最高价
		AthChangePercentage          decimal.Decimal `json:"ath_change_percentage"`            // 历史最高价变动百分比
		AthDate                      time.Time       `json:"ath_date"`                         // 历史最高价日期
		Atl                          decimal.Decimal `json:"atl"`                              // 历史最低价
		AtlChangePercentage          decimal.Decimal `json:"atl_change_percentage"`            // 历史最低价变动百分比
		AtlDate                      time.Time       `json:"atl_date"`                         // 历史最低价日期
		AssetIDS                     []string        `json:"asset_ids"`
		SparklineIn7D                string          `json:"sparkline_in_7d"`
		SparklineIn24H               string          `json:"sparkline_in_24h"`
		UpdatedAt                    time.Time       `json:"updated_at"`
		Key                          string          `json:"key"`
	}

	HistoricalPrice struct {
		CoinID    string                 `json:"coin_id"`
		Type      string                 `json:"type"` // 1D, 1W, 1M, YTD, ALL
		Data      []HistoricalPriceDatum `json:"data"`
		UpdatedAt time.Time              `json:"updated_at"`
	}

	HistoricalPriceDatum struct {
		Price string `json:"price"`
		Unix  int64  `json:"unix"`
	}

	TokenView struct {
		AssetId string     `json:"assetId"`
		Name    string     `json:"name"`
		Symbol  string     `json:"symbol"`
		Icon    string     `json:"icon"`
		Chain   TokenChain `json:"chain"`
	}

	TokenChain struct {
		ChainId  string `json:"chainId"`
		Symbol   string `json:"symbol"`
		Name     string `json:"name"`
		Icon     string `json:"icon"`
		Decimals int    `json:"decimals"`
	}

	QuoteRequest struct {
		InputMint  string `json:"inputMint"`
		OutputMint string `json:"outputMint"`
		Amount     string `json:"amount"`
	}

	QuoteResponseView struct {
		InputMint  string `json:"inputMint"`
		InAmount   string `json:"inAmount"`
		OutputMint string `json:"outputMint"`
		OutAmount  string `json:"outAmount"`
		Payload    string `json:"payload"`
	}

	SwapRequest struct {
		Payer       string `json:"payer"`       // mixin user id
		InputMint   string `json:"inputMint"`   // mixin asset id
		InputAmount string `json:"inputAmount"` // mixin amount
		OutputMint  string `json:"outputMint"`  // mixin asset id
		Payload     string `json:"payload"`     // QuoteResponseView.Payload
		Referral    string `json:"referral"`    // optional
	}

	SwapResponseView struct {
		Tx    string            `json:"tx"` // mixin://mixin.one/pay/...
		Quote QuoteResponseView `json:"quote"`
	}

	SwapTx struct {
		Trace   string `json:"trace"`
		Payee   string `json:"payee"` // 收款人
		Asset   string `json:"asset"`
		Amount  string `json:"amount"`
		Memo    string `json:"memo"`
		OrderId string `json:"orderId"`
	}

	SwapOrder struct {
		OrderId        string          `json:"order_id" gorm:"type:varchar(36);not null;primary_key"`
		UserId         string          `json:"user_id" gorm:"type:varchar(36);not null"`
		AssetId        string          `json:"asset_id" gorm:"type:varchar(36);not null"`
		ReceiveAssetId string          `json:"receive_asset_id" gorm:"type:varchar(36);not null"`
		Amount         decimal.Decimal `json:"amount" gorm:"type:decimal(64,8);not null"`
		ReceiveAmount  decimal.Decimal `json:"receive_amount" gorm:"type:decimal(64,8);not null"`
		PaymentTraceId string          `json:"payment_trace_id" gorm:"type:varchar(36);not null"`
		ReceiveTraceId string          `json:"receive_trace_id" gorm:"type:varchar(36);not null"`
		State          SwapOrderState  `json:"state" gorm:"type:varchar(10);not null"`
		CreatedAt      time.Time       `json:"created_at" gorm:"type:timestamp;not null"`
	}

	ErrorResponse struct {
		Error struct {
			Status      int    `json:"status"`
			Code        int    `json:"code"`
			Description string `json:"description"`
		} `json:"error"`
	}

	MixinOracleAPIError struct {
		StatusCode  int
		Code        int
		Description string
		RawBody     string
	}

	SwapOrderState string
)

const (
	SwapOrderStateCreated SwapOrderState = "created"
	SwapOrderStatePending SwapOrderState = "pending"
	SwapOrderStateSuccess SwapOrderState = "success"
	SwapOrderStateFailed  SwapOrderState = "failed"
)

func (e *MixinOracleAPIError) Error() string {
	return fmt.Sprintf("API error: status=%d, code=%d, description=%s",
		e.StatusCode, e.Code, e.Description)
}

func (q QuoteRequest) ToQuery() string {
	return fmt.Sprintf("inputMint=%s&outputMint=%s&amount=%s&source=mixin", q.InputMint, q.OutputMint, q.Amount)
}

func (s SwapResponseView) DecodeTx() (*SwapTx, error) {
	// mixin://mixin.one/pay/${uid}?asset=965e5c6e-434c-3fa9-b780-c50f43cd955c&amount=0.1&memo=test&trace=74518d17-e3df-46e5-a615-07793af27d5d
	tx, err := url.Parse(s.Tx)
	if err != nil {
		return nil, err
	}

	query, err := url.ParseQuery(tx.RawQuery)
	if err != nil {
		return nil, err
	}

	// mixin://mixin.one/pay/${uid}
	uid := strings.TrimPrefix(tx.Path, "/pay/")
	if uid == "" {
		return nil, fmt.Errorf("invalid uid in path: %s", tx.Path)
	}

	return &SwapTx{
		Trace:   query.Get("trace"),
		Payee:   uid,
		Asset:   query.Get("asset"),
		Amount:  query.Get("amount"),
		Memo:    query.Get("memo"),
		OrderId: query.Get("memo"),
	}, nil
}
