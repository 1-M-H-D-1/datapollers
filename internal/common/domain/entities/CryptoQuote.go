package entities

import "time"

type CryptoQuote struct {
	SymbolPair SymbolPair
	Market     Market
	TimeStamp  time.Time
	Rate       uint64
	OpenRate   uint64
	HighRate   uint64
	LowRate    uint64
	CloseRate  uint64
	Volume     uint64
}
