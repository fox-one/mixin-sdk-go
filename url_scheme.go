package mixin

import (
	"net/url"
)

const Scheme = "mixin"

var URL urlScheme

type urlScheme struct{}

func (urlScheme) Users(userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "users",
		Path:   userID,
	}

	return u.String()
}

func (urlScheme) Transfer(userID string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "transfer",
		Path:   userID,
	}

	return u.String()
}

func (urlScheme) Pay(input *TransferInput) string {
	q := url.Values{}
	q.Set("asset", input.AssetID)
	q.Set("trace", input.TraceID)
	q.Set("amount", input.Amount.String())
	q.Set("recipient", input.OpponentID)
	q.Set("memo", input.Memo)

	u := url.URL{
		Scheme:   Scheme,
		Host:     "pay",
		RawQuery: q.Encode(),
	}

	return u.String()
}

func (urlScheme) Codes(code string) string {
	u := url.URL{
		Scheme: Scheme,
		Host:   "codes",
		Path:   code,
	}

	return u.String()
}
