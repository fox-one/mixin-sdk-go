package mixin

import (
	"context"

	"github.com/shopspring/decimal"
)

// Fiat is a struct for fiat currencies
type Fiat struct {
	Code string          `json:"code,omitempty"`
	Rate decimal.Decimal `json:"rate,omitempty"`
}

// ReadFiats returns the exchange rates between two currencies
func (c *Client) ReadFiats(ctx context.Context) ([]Fiat, error) {
	var rates []Fiat
	if err := c.Get(ctx, "/fiats", nil, &rates); err != nil {
		return nil, err
	}

	return rates, nil
}

// ExchangeRate represent the exchange rate between two currencies
// deprecated: use Fiat instead
type ExchangeRate Fiat

// ReadExchangeRates returns the exchange rates between two currencies
// deprecated: use ReadFiats instead
func (c *Client) ReadExchangeRates(ctx context.Context) ([]ExchangeRate, error) {
	fiats, err := c.ReadFiats(ctx)
	if err != nil {
		return nil, err
	}

	var rates []ExchangeRate
	for _, fiat := range fiats {
		rates = append(rates, ExchangeRate(fiat))
	}

	return rates, nil
}
