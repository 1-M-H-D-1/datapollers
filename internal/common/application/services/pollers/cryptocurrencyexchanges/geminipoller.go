package cryptocurrencyexchanges

import (
	"DataPoller/internal/common/application/services/pollers"
	"DataPoller/internal/common/domain/entities"
)

type GenimiPoller struct {
	dataSource entities.DataSource
}

func NewGenimiPoller(dataSource entities.DataSource) pollers.QuotePoller {
	return &GenimiPoller{dataSource: dataSource}
}

func (geminiPoller *GenimiPoller) Poll() {}
