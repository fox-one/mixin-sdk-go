package web3

import (
	"time"
)

// 定义价格历史策略接口
type PriceHistoryStrategy interface {
	GetCacheDuration() time.Duration
	GetCacheDirtyDuration() time.Duration
	GetType() string
}

// 具体策略实现
type priceHistoryBase struct {
	cacheDuration      time.Duration
	cacheDirtyDuration time.Duration
	typeStr            string
}

func (p *priceHistoryBase) GetCacheDuration() time.Duration {
	return p.cacheDuration
}

func (p *priceHistoryBase) GetCacheDirtyDuration() time.Duration {
	return p.cacheDirtyDuration
}

func (p *priceHistoryBase) GetType() string {
	return p.typeStr
}

// 具体策略实例
var (
	OneDayStrategy = &priceHistoryBase{
		cacheDuration:      5 * time.Minute,
		cacheDirtyDuration: 10 * time.Minute,
		typeStr:            "1D",
	}

	OneWeekStrategy = &priceHistoryBase{
		cacheDuration:      time.Hour,
		cacheDirtyDuration: 2 * time.Hour,
		typeStr:            "1W",
	}
	OneMonthStrategy = &priceHistoryBase{
		cacheDuration:      time.Hour,
		cacheDirtyDuration: 2 * time.Hour,
		typeStr:            "1M",
	}
	YTDStrategy = &priceHistoryBase{
		cacheDuration:      24 * time.Hour,
		cacheDirtyDuration: 48 * time.Hour,
		typeStr:            "YTD",
	}
	AllTimeStrategy = &priceHistoryBase{
		cacheDuration:      24 * time.Hour,
		cacheDirtyDuration: 48 * time.Hour,
		typeStr:            "ALL",
	}
)

type PriceHistoryFactory struct {
	strategies map[HistoryPriceType]PriceHistoryStrategy
}

func NewPriceHistoryFactory() *PriceHistoryFactory {
	return &PriceHistoryFactory{
		strategies: map[HistoryPriceType]PriceHistoryStrategy{
			Type1D:  OneDayStrategy,
			Type1W:  OneWeekStrategy,
			Type1M:  OneMonthStrategy,
			TypeYTD: YTDStrategy,
			TypeALL: AllTimeStrategy,
		},
	}
}

func (f *PriceHistoryFactory) GetStrategy(t HistoryPriceType) PriceHistoryStrategy {
	return f.strategies[t]
}
