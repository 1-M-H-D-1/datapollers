package binancepoller

import (
	"DataPoller/internal/common/application/services/quotePollersFactories"
)

func RunBinancePoller() {
	binancePoller := quotePollersFactories.BuildBinanceQuotePoller()
	(*binancePoller).Poll()
}
