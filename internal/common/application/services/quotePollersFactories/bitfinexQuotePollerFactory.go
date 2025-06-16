package quotePollersFactories

import (
	"DataPoller/internal/common/application/services/pollers"
	"DataPoller/internal/common/application/services/pollers/cryptocurrencyexchanges"
	"DataPoller/internal/common/domain/consts"
	"DataPoller/internal/common/domain/repositories"
	"DataPoller/internal/common/infrastructure/repositories/postgres"
	"DataPoller/internal/common/infrastructure/repositories/quest"
	"fmt"
)

func BuildBitfinexQuotePoller() *pollers.QuotePoller {
	pgDataSourceRepository := postgresrepositories.PostgresDataSourcesRepository{}
	var datasourceRepository repositories.DataSourcesRepository = pgDataSourceRepository
	questCryptoQuotesWriter := questrepositories.QuestCryptoQuotesWriter{}
	var cryptoQuotesWriter repositories.CryptoQuotesWriter = questCryptoQuotesWriter

	dataSource, err := datasourceRepository.FindById(consts.Bitfinex)
	fmt.Printf("dataSource: %+v\n", dataSource)

	if err != nil {
		panic(err)
	}

	p := cryptocurrencyexchanges.NewBitfinexPoller(*dataSource, cryptoQuotesWriter)

	return &p
}
