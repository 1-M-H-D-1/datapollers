package quotePollersFactories

import (
	"DataPoller/internal/common/application/services/pollers"
	"DataPoller/internal/common/application/services/pollers/cryptocurrencyexchanges"
	"DataPoller/internal/common/domain/consts"
	"DataPoller/internal/common/domain/repositories"
	"DataPoller/internal/common/infrastructure/repositories/postgres"
	"DataPoller/internal/common/infrastructure/repositories/quest"
)

func BuildBinanceQuotePoller() *pollers.QuotePoller {
	pgDataSourceRepository := postgresrepositories.PostgresDataSourcesRepository{}
	var datasourceRepository repositories.DataSourcesRepository = pgDataSourceRepository
	questCryptoQuotesWriter := questrepositories.QuestCryptoQuotesWriter{}
	var cryptoQuotesWriter repositories.CryptoQuotesWriter = questCryptoQuotesWriter

	dataSource, err := datasourceRepository.FindById(consts.Binance)
	if err != nil {
		panic(err)
	}

	p := cryptocurrencyexchanges.NewBinancePoller(*dataSource, cryptoQuotesWriter)

	return &p
}
