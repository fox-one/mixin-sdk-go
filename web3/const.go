package web3

import "time"

const (
	MixinRouteApiPrefix = "https://api.route.mixin.one"

	assetInfoCacheDuration      = 30 * time.Second
	assetInfoCacheDirtyDuration = assetInfoCacheDuration * 2
)

type HistoryPriceType uint8

const (
	Type1D HistoryPriceType = iota
	Type1W
	Type1M
	TypeYTD
	TypeALL
)

func (h HistoryPriceType) String() string {
	return []string{"1D", "1W", "1M", "YTD", "ALL"}[h]
}

func InvalidPriceHistoryType(typ uint8) bool {
	return typ > uint8(TypeALL)
}
