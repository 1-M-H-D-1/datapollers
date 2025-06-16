package bitfinexpoller

import (
	"DataPoller/internal/common/application/services/quotePollersFactories"
)

func RunBitfinexPoller() {
	bitfinexPoller := quotePollersFactories.BuildBitfinexQuotePoller()
	(*bitfinexPoller).Poll()
}
