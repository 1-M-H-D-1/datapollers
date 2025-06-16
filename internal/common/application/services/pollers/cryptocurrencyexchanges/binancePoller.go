package cryptocurrencyexchanges

import (
	"DataPoller/internal/common/application/services/pollers"
	"DataPoller/internal/common/domain/entities"
	"DataPoller/internal/common/domain/repositories"
	questrepositories "DataPoller/internal/common/infrastructure/repositories/quest"
	"time"

	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

type BinancePoller struct {
	dataSource         entities.DataSource
	cryptoQuotesWriter repositories.CryptoQuotesWriter
}

type BinanceSubscribeMessage struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	Id     int      `json:"id"`
}

type BinanceSubscriptionResponse struct {
	Result interface{} `json:"result"`
	Id     int         `json:"id"`
}

type BinanceTickerMessage struct {
	EventType           string `json:"e"`
	EventTime           int64  `json:"E"`
	Symbol              string `json:"s"`
	PriceChange         string `json:"p"`
	PriceChangePercent  string `json:"P"`
	WeightedAvgPrice    string `json:"w"`
	FirstTradePrice     string `json:"x"`
	LastPrice           string `json:"c"`
	LastQuantity        string `json:"Q"`
	BestBidPrice        string `json:"b"`
	BestBidQuantity     string `json:"B"`
	BestAskPrice        string `json:"a"`
	BestAskQuantity     string `json:"A"`
	OpenPrice           string `json:"o"`
	HighPrice           string `json:"h"`
	LowPrice            string `json:"l"`
	Volume              string `json:"v"`
	QuoteVolume         string `json:"q"`
	StatisticsOpenTime  int64  `json:"O"`
	StatisticsCloseTime int64  `json:"C"`
	FirstTradeID        int64  `json:"F"`
	LastTradeID         int64  `json:"L"`
	TotalNumberOfTrades int64  `json:"n"`
}

func NewBinancePoller(dataSource entities.DataSource,
	cryptoQuotesWriter repositories.CryptoQuotesWriter) pollers.QuotePoller {
	return &BinancePoller{dataSource: dataSource, cryptoQuotesWriter: cryptoQuotesWriter}
}

func (binancePoller *BinancePoller) Poll() {
	symbolPairs := binancePoller.dataSource.SymbolPairs
	rateLimit := binancePoller.dataSource.RateLimit

	if rateLimit == 0 {
		go binancePoller.pollSymbolChunk(symbolPairs)
		select {}
	}

	totalPairs := len(symbolPairs)
	for i := 0; i < totalPairs; i += rateLimit {
		end := i + rateLimit
		if end > totalPairs {
			end = totalPairs
		}

		go binancePoller.pollSymbolChunk(symbolPairs[i:end])
	}

	select {}
}

func (binancePoller *BinancePoller) pollSymbolChunk(pairs []entities.SymbolPair) {
	conn, _, err := websocket.DefaultDialer.Dial(binancePoller.dataSource.ConnectionString, nil)
	if err != nil {
		log.Fatal("Error connecting to Binance WebSocket:", err)
		return
	}
	defer conn.Close()
	log.Printf("Started Binance conn for pairs: %+v\n", pairs)

	var params []string
	for _, pair := range pairs {
		symbolParam := fmt.Sprintf("%s%s@ticker",
			strings.ToLower(pair.BaseSymbol.Name),
			strings.ToLower(pair.QuoteSymbol.Name))
		params = append(params, symbolParam)
	}

	subMsg := BinanceSubscribeMessage{
		Method: "SUBSCRIBE",
		Params: params,
		Id:     1,
	}

	msgJSON, err := json.Marshal(subMsg)
	if err != nil {
		log.Fatal("Error marshaling Binance subscription JSON:", err)
		return
	}

	if err = conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
		log.Fatal("Error sending Binance subscription message:", err)
		return
	}

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading Binance message:", err)
			return
		}

		var subResp BinanceSubscriptionResponse
		if err := json.Unmarshal(message, &subResp); err == nil && subResp.Id == 1 {
			//fmt.Printf("Binance subscription response: %+v\n", subResp)
			log.Println("Subscribed to Binance WebSocket streams:", params)
			continue
		}

		var tickerMsg BinanceTickerMessage
		err = json.Unmarshal(message, &tickerMsg)
		if err != nil {
			log.Println("Error unmarshaling Binance message:", err)
			continue
		}

		quote, err := binancePoller.tickerToCryptoQuote(tickerMsg, binancePoller.dataSource)
		if err != nil {
			log.Println("Error converting Binance ticker to quote:", err)
			continue
		}

		err = binancePoller.cryptoQuotesWriter.Write([]entities.CryptoQuote{quote})
		if err != nil {
			log.Println("Error writing Binance quote:", err)
			continue
		}

		//fmt.Printf("Binance ticker update: %+v\n", tickerMsg)
	}
}

func (binancePoller *BinancePoller) tickerToCryptoQuote(ticker BinanceTickerMessage, dataSource entities.DataSource) (entities.CryptoQuote, error) {
	var quote entities.CryptoQuote

	pair, err := binancePoller.findSymbolPair(ticker.Symbol, dataSource.SymbolPairs)
	if err != nil {
		return quote, err
	}

	rate, _ := questrepositories.ToDatabaseRate(ticker.LastPrice)
	openRate, _ := questrepositories.ToDatabaseRate(ticker.OpenPrice)
	highRate, _ := questrepositories.ToDatabaseRate(ticker.HighPrice)
	lowRate, _ := questrepositories.ToDatabaseRate(ticker.LowPrice)
	closeRate := rate
	volume, _ := questrepositories.ToDatabaseRate(ticker.Volume)

	quote = entities.CryptoQuote{
		SymbolPair: pair,
		Market:     pair.Market,
		TimeStamp:  time.UnixMilli(ticker.EventTime),
		Rate:       rate,
		OpenRate:   openRate,
		HighRate:   highRate,
		LowRate:    lowRate,
		CloseRate:  closeRate,
		Volume:     volume,
	}

	return quote, nil
}

func (binancePoller *BinancePoller) findSymbolPair(symbol string, pairs []entities.SymbolPair) (entities.SymbolPair, error) {
	for _, pair := range pairs {
		combined := strings.ToUpper(pair.BaseSymbol.Name + pair.QuoteSymbol.Name)
		if combined == symbol {
			return pair, nil
		}
	}
	return entities.SymbolPair{}, fmt.Errorf("symbol pair not found for %s", symbol)
}
