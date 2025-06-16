package postgresrepositories

import (
	entities2 "DataPoller/internal/common/domain/entities"
	"DataPoller/internal/common/infrastructure"
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
)

type PostgresDataSourcesRepository struct {
}

func (repo PostgresDataSourcesRepository) FindAll() ([]*entities2.DataSource, error) {
	var config infrastructure.Configuration
	err := config.LoadFromFile()
	if err != nil {
		return nil, err
	}

	query := "SELECT " +
		"ds.id AS dataSourceId, " +
		"ds.name AS dataSourceName, " +
		"ds.connection_string AS dataSourceConnectionString, " +
		"ds.login AS dataSourceLogin, " +
		"ds.password AS dataSourcePassword, " +
		"ds.rate_limit AS dataSourceRateLimit, " +
		"sp.Id AS symbolPairId, " +
		"m.id as marketId, " +
		"m.name as marketName, " +
		"bs.id as baseSymbolId, " +
		"bs.name as baseSymbolName, " +
		"qs.id as quoteSymbolId, " +
		"qs.name as quoteSymbolName " +
		"FROM tds.data_sources ds " +
		"INNER JOIN tds.data_source_symbol_pairs dssp ON ds.id = dssp.data_source_id " +
		"INNER JOIN tds.symbol_pairs sp ON dssp.symbol_pair_id = sp.id " +
		"INNER JOIN tds.markets m ON m.id = sp.market_id " +
		"INNER JOIN tds.symbols bs ON bs.id = sp.base_symbol_id " +
		"INNER JOIN tds.symbols qs ON qs.id = sp.quoted_symbol_id "

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.MainDatabase.Username,
		config.MainDatabase.Password,
		config.MainDatabase.Host,
		config.MainDatabase.Port,
		config.MainDatabase.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		return nil, err
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataSources []*entities2.DataSource
	dataSourceMap := make(map[int]*entities2.DataSource)

	for rows.Next() {
		var dataSource entities2.DataSource = entities2.DataSource{}
		var symbolPair entities2.SymbolPair
		var market entities2.Market
		var baseSymbol entities2.Symbol
		var quoteSymbol entities2.Symbol

		if err := rows.Scan(&dataSource.Id,
			&dataSource.Name,
			&dataSource.ConnectionString,
			&dataSource.Login,
			&dataSource.Password,
			&dataSource.RateLimit,
			&symbolPair.Id,
			&market.Id,
			&market.Name,
			&baseSymbol.Id,
			&baseSymbol.Name,
			&quoteSymbol.Id,
			&quoteSymbol.Name); err != nil {
			return nil, err
		}

		if existingDataSource, found := dataSourceMap[dataSource.Id]; found {
			symbolPair.Market = market
			symbolPair.BaseSymbol = baseSymbol
			symbolPair.QuoteSymbol = quoteSymbol

			existingDataSource.SymbolPairs = append(existingDataSource.SymbolPairs, symbolPair)
		} else {
			symbolPair.Market = market
			symbolPair.BaseSymbol = baseSymbol
			symbolPair.QuoteSymbol = quoteSymbol

			dataSource.SymbolPairs = append(dataSource.SymbolPairs, symbolPair)

			dataSourceMap[dataSource.Id] = &dataSource
			dataSources = append(dataSources, &dataSource)
		}

	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dataSources, nil

}

func (repo PostgresDataSourcesRepository) FindById(id int) (*entities2.DataSource, error) {
	var config infrastructure.Configuration
	err := config.LoadFromFile()
	if err != nil {
		return nil, err
	}

	query := "SELECT " +
		"ds.id AS dataSourceId, " +
		"ds.name AS dataSourceName, " +
		"ds.connection_string AS dataSourceConnectionString, " +
		"ds.login AS dataSourceLogin, " +
		"ds.password AS dataSourcePassword, " +
		"ds.rate_limit AS dataSourceRateLimit, " +
		"sp.Id AS symbolPairId, " +
		"m.id as marketId, " +
		"m.name as marketName, " +
		"bs.id as baseSymbolId, " +
		"bs.name as baseSymbolName, " +
		"qs.id as quoteSymbolId, " +
		"qs.name as quoteSymbolName " +
		"FROM tds.data_sources ds " +
		"INNER JOIN tds.data_source_symbol_pairs dssp ON ds.id = dssp.data_source_id " +
		"INNER JOIN tds.symbol_pairs sp ON dssp.symbol_pair_id = sp.id " +
		"INNER JOIN tds.markets m ON m.id = sp.market_id " +
		"INNER JOIN tds.symbols bs ON bs.id = sp.base_symbol_id " +
		"INNER JOIN tds.symbols qs ON qs.id = sp.quoted_symbol_id " +
		"WHERE ds.id = $1 "

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		config.MainDatabase.Username,
		config.MainDatabase.Password,
		config.MainDatabase.Host,
		config.MainDatabase.Port,
		config.MainDatabase.Database)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return nil, err
	}

	rows, err := db.Query(query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dataSource *entities2.DataSource
	dataSourceMap := make(map[int]*entities2.DataSource)

	for rows.Next() {
		var tempDataSource entities2.DataSource
		var symbolPair entities2.SymbolPair
		var market entities2.Market
		var baseSymbol entities2.Symbol
		var quoteSymbol entities2.Symbol

		if err := rows.Scan(&tempDataSource.Id,
			&tempDataSource.Name,
			&tempDataSource.ConnectionString,
			&tempDataSource.Login,
			&tempDataSource.Password,
			&tempDataSource.RateLimit,
			&symbolPair.Id,
			&market.Id,
			&market.Name,
			&baseSymbol.Id,
			&baseSymbol.Name,
			&quoteSymbol.Id,
			&quoteSymbol.Name); err != nil {
			return nil, err
		}

		if existingDataSource, found := dataSourceMap[tempDataSource.Id]; found {
			symbolPair.Market = market
			symbolPair.BaseSymbol = baseSymbol
			symbolPair.QuoteSymbol = quoteSymbol
			existingDataSource.SymbolPairs = append(existingDataSource.SymbolPairs, symbolPair)
		} else {
			symbolPair.Market = market
			symbolPair.BaseSymbol = baseSymbol
			symbolPair.QuoteSymbol = quoteSymbol
			tempDataSource.SymbolPairs = append(tempDataSource.SymbolPairs, symbolPair)
			dataSourceMap[tempDataSource.Id] = &tempDataSource
			dataSource = &tempDataSource
		}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return dataSource, nil
}
