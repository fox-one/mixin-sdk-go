package mixinnet

import (
	"github.com/shopspring/decimal"
)

type (
	UTXO struct {
		Type   uint8           `json:"type"`
		Amount decimal.Decimal `json:"amount"`
		Hash   Hash            `json:"hash"`
		Index  uint8           `json:"index,omitempty"`
		Lock   *Hash           `json:"lock,omitempty"`
	}
)
