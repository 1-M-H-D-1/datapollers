package questrepositories

import (
	"DataPoller/internal/common/domain/entities"
	"DataPoller/internal/common/infrastructure"
	"strconv"

	"context"
	"fmt"

	qdb "github.com/questdb/go-questdb-client"
)

type QuestCryptoQuotesWriter struct{}

func (repo QuestCryptoQuotesWriter) Write(quotes []entities.CryptoQuote) error {
	var config infrastructure.Configuration
	err := config.LoadFromFile()
	if err != nil {
		return err
	}

	connStr := fmt.Sprintf("%s:%s",
		config.TimeSeriesDatabase.Host,
		config.TimeSeriesDatabase.Port)

	ctx := context.TODO()

	client, err := qdb.NewLineSender(ctx, qdb.WithAddress(connStr))
	if err != nil {
		panic("Failed to create QuestDB client")
	}
	defer client.Close()

	for _, quote := range quotes {
		err := client.
			Table("crypto_quotes").
			Symbol("Base", quote.SymbolPair.BaseSymbol.Name).
			Symbol("Quote", quote.SymbolPair.QuoteSymbol.Name).
			Symbol("MarketName", quote.Market.Name).
			Symbol("BaseQuote", quote.SymbolPair.BaseSymbol.Name+quote.SymbolPair.QuoteSymbol.Name).
			Int64Column("BaseId", int64(quote.SymbolPair.BaseSymbol.Id)).
			Int64Column("QuoteId", int64(quote.SymbolPair.QuoteSymbol.Id)).
			TimestampColumn("TimeStamp", quote.TimeStamp.UnixMicro()).
			Int64Column("Rate", int64(quote.Rate)).
			Int64Column("MarketId", int64(quote.Market.Id)).
			Int64Column("OpenRate", int64(quote.OpenRate)).
			Int64Column("HighRate", int64(quote.HighRate)).
			Int64Column("LowRate", int64(quote.LowRate)).
			Int64Column("CloseRate", int64(quote.CloseRate)).
			Int64Column("Volume", int64(quote.Volume)).
			At(ctx, quote.TimeStamp.UnixMicro())
		if err != nil {
			return err
		}
	}

	if err := client.Flush(ctx); err != nil {
		return fmt.Errorf("failed to flush lines to QuestDB: %w", err)
	}

	return nil
}

func ToDatabaseRate(rate string) (uint64, error) {
	rateFloat, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return 0, fmt.Errorf("error converting rate to float: %v", err)
	}
	return uint64(rateFloat * 10000), nil
}
