package repositories

import (
	"DataPoller/internal/common/domain/entities"
)

type CryptoQuotesWriter interface {
	Write(quotes []entities.CryptoQuote) error
}
