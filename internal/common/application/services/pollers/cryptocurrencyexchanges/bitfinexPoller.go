package cryptocurrencyexchanges

import (
	"DataPoller/internal/common/application/services/pollers"
	"DataPoller/internal/common/domain/entities"
	"DataPoller/internal/common/domain/repositories"
	questrepositories "DataPoller/internal/common/infrastructure/repositories/quest"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type BitfinexPoller struct {
	dataSource         entities.DataSource
	cryptoQuotesWriter repositories.CryptoQuotesWriter
}

type BitfinexSubscribeMessage struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Symbol  string `json:"symbol"`
}

/*{
   event: "subscribed",
   channel: "ticker",
   chanId: CHANNEL_ID,
   symbol: SYMBOL,
   pair: PAIR
}*/

type BitfinexSubscriptionResponse struct {
	Event     string `json:"event"`
	Channel   string `json:"channel"`
	ChannelId int    `json:"chanId"`
	Symbol    string `json:"symbol"`
	Pair      string `json:"pair"`
}

/*
Ticker Data:
[

	443378, //CHANNEL_ID
	[
	  7616.5, //BID
	  31.89055171, //BID_SIZE
	  7617.5, //ASK
	  43.358118629999986, //ASK_SIZE
	  -550.8, //DAILY_CHANGE
	  -0.0674, //DAILY_CHANGE_RELATIVE
	  7617.1, //LAST_PRICE
	  8314.71200815, //VOLUME
	  8257.8, //HIGH
	  7500 //LOW
	] //UPDATE

]
*/
type BitfinexTickerMessage struct {
	ChannelId int
	Ticker    BitfinexTickerData
}

type BitfinexTickerData struct {
	Bid            string
	BidSize        string
	Ask            string
	AskSize        string
	DailyChange    string
	DailyChangeRel string
	LastPrice      string
	Volume         string
	High           string
	Low            string
}

func NewBitfinexPoller(dataSource entities.DataSource,
	cryptoQuotesWriter repositories.CryptoQuotesWriter) pollers.QuotePoller {
	return &BitfinexPoller{dataSource: dataSource, cryptoQuotesWriter: cryptoQuotesWriter}
}

func (bitfinexPoller *BitfinexPoller) Poll() {
	symbolPairs := bitfinexPoller.dataSource.SymbolPairs
	rateLimit := bitfinexPoller.dataSource.RateLimit

	if rateLimit == 0 {
		go bitfinexPoller.pollSymbolChunk(symbolPairs)
		select {}
	}

	totalPairs := len(symbolPairs)
	for i := 0; i < totalPairs; i += rateLimit {
		end := i + rateLimit
		if end > totalPairs {
			end = totalPairs
		}

		go bitfinexPoller.pollSymbolChunk(symbolPairs[i:end])
	}

	select {}
}

func (bitfinexPoller *BitfinexPoller) pollSymbolChunk(pairs []entities.SymbolPair) {
	conn, _, err := websocket.DefaultDialer.Dial(bitfinexPoller.dataSource.ConnectionString, nil)

	if err != nil {
		log.Fatal("Error connecting to BitfinexPoller WebSocket:", err)
		return
	}

	defer conn.Close()

	log.Printf("Started BitfinexPoller conn for pairs: %+v\n", pairs)

	chanIdToPair := make(map[int]entities.SymbolPair)

	var params []string

	for _, pair := range pairs {
		symbolParam := fmt.Sprintf("t%s%s",
			strings.ToUpper(pair.BaseSymbol.Name),
			strings.ToUpper(pair.QuoteSymbol.Name))

		subMsg := map[string]interface{}{
			"event":   "subscribe",
			"channel": "ticker",
			"symbol":  symbolParam,
		}

		msgJSON, err := json.Marshal(subMsg)

		if err != nil {
			log.Printf("Error marshaling BitfinexPoller subscription JSON:", err)
			continue
		}

		//log.Println(subMsg)
		//log.Println(msgJSON)

		if err = conn.WriteMessage(websocket.TextMessage, msgJSON); err != nil {
			log.Printf("Error sending BitfinexPoller subscription message:", err)
			continue
		}

		subscribed := false
		errorReceived := false

		for !subscribed && !errorReceived {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Error reading Bitfinex subscription response: %v", err)
				break
			}

			var rawMsg map[string]interface{}
			if err := json.Unmarshal(msg, &rawMsg); err == nil {
				eventType, ok := rawMsg["event"].(string)

				if !ok {
					log.Printf("Missing event field in", string(msg))
					break
				}

				switch eventType {
				case "info":
					log.Println("Info message:", rawMsg)

				case "error":
					code := int(rawMsg["code"].(float64))
					//msg := rawMsg["msg"].(string)
					//log.Printf("Bitfinex subscription error [%d]: %s\n", code, msg)

					switch code {
					case 10300:
						log.Println("Generic subscription failed for the pair", pair)
					case 10301:
						log.Println("Already subscribed to the pair:", pair)
					case 10302:
						log.Println("Unknown channel", subMsg["channel"].(string), "for pair", pair)
					default:
						log.Println("Unhandled error code", code, "for pair", pair)
					}

					errorReceived = true
				case "subscribed":
					var subResp BitfinexSubscriptionResponse
					if err := json.Unmarshal(msg, &subResp); err != nil {
						log.Println("Failed to unmarshal subscribe response from Bitfinex:", err)
						break
					}
					chanIdToPair[subResp.ChannelId] = pair
					log.Printf("Subscribed to %s, chanId: %d", subResp.Pair, subResp.ChannelId)
					params = append(params, symbolParam)
					subscribed = true

				default:
					log.Println("Unknown event type received from Bifinex:", eventType)
				}
			} else {
				//log.Fatal("Received unexpected message: ", string(msg), "in subscription stage")
				break
			}
		}
	}

	log.Println("Subscribed to Binance WebSocket streams:", params)

	for {
		_, message, err := conn.ReadMessage()

		if err != nil {
			log.Println("Error reading Bitfinex message:", err)
			return
		}

		var rawMsg interface{}
		if err := json.Unmarshal(message, &rawMsg); err != nil {
			log.Println("Error unmarshaling Bitfinex message:", err)
			continue
		}

		log.Println(rawMsg)

		switch msg := rawMsg.(type) {
		case []interface{}:
			if len(msg) != 2 {
				log.Println("Invalid Bitfinex response")
				continue
			}

			if hbMsg, ok := msg[1].(string); ok && hbMsg == "hb" {
				log.Println("Heartbeat received")
				continue
			}

			chanId, ok := msg[0].(float64)
			update, ok := msg[1].([]interface{})

			if !ok {
				log.Println("Error in processing Bitfinex response", msg)
				continue
			}
			if update == nil {
				log.Printf("Channel %v returned null data", chanId)
				continue
			}

			//Sprintf or formatFloat?
			tickerData := BitfinexTickerData{
				Bid:            strconv.FormatFloat(update[0].(float64), 'f', -1, 64),
				BidSize:        strconv.FormatFloat(update[1].(float64), 'f', -1, 64),
				Ask:            strconv.FormatFloat(update[2].(float64), 'f', -1, 64),
				AskSize:        strconv.FormatFloat(update[3].(float64), 'f', -1, 64),
				DailyChange:    strconv.FormatFloat(update[4].(float64), 'f', -1, 64),
				DailyChangeRel: strconv.FormatFloat(update[5].(float64), 'f', -1, 64),
				LastPrice:      strconv.FormatFloat(update[6].(float64), 'f', -1, 64),
				Volume:         strconv.FormatFloat(update[7].(float64), 'f', -1, 64),
				High:           strconv.FormatFloat(update[8].(float64), 'f', -1, 64),
				Low:            strconv.FormatFloat(update[9].(float64), 'f', -1, 64),
			}

			log.Println(tickerData)
			log.Println(chanId)
			log.Println(chanIdToPair[int(chanId)])

			quote, err := bitfinexPoller.tickerToCryptoQuote(tickerData, chanIdToPair[int(chanId)])

			if err != nil {
				log.Println("Error converting Binance ticker to quote:", err)
				continue
			}

			err = bitfinexPoller.cryptoQuotesWriter.Write([]entities.CryptoQuote{quote})

			if err != nil {
				log.Println("Error writing Binance quote:", err)
				continue
			}

		case map[string]interface{}:
			if event, ok := msg["event"].(string); ok {
				switch event {
				case "info":
					log.Println("System info message:", msg)
				case "subscribed":
					log.Println("Subscribed confirmation:", msg)
				case "error":
					code := int(msg["code"].(float64))
					//log.Printf("Bitfinex subscription error [%d]: %s\n", code, msg)

					switch code {
					case 10300:
						log.Println("Generic subscription failed: ", msg)
					case 10301:
						log.Println("Already subscribed to the pair: ", msg)
					case 10302:
						log.Println("Unknown channel: ", msg["channel"].(string))
					default:
						log.Println("Unhandled error code: ", code)
					}
				default:
					log.Println("Unhandled system event: ", event, msg)
				}
			} else {
				log.Println("Unknown map message without 'event' field:", msg)
			}
		default:
			log.Println("Unknown type", msg)
		}

	}
}

func (bitfinexPoller *BitfinexPoller) tickerToCryptoQuote(ticker BitfinexTickerData, pair entities.SymbolPair) (entities.CryptoQuote, error) {
	var quote entities.CryptoQuote

	rate, _ := questrepositories.ToDatabaseRate(ticker.LastPrice)
	openRate, _ := questrepositories.ToDatabaseRate(ticker.Bid)
	highRate, _ := questrepositories.ToDatabaseRate(ticker.High)
	lowRate, _ := questrepositories.ToDatabaseRate(ticker.Low)
	closeRate := rate
	volume, _ := questrepositories.ToDatabaseRate(ticker.Volume)

	quote = entities.CryptoQuote{
		SymbolPair: pair,
		Market:     pair.Market,
		TimeStamp:  time.Now(),
		Rate:       rate,
		OpenRate:   openRate,
		HighRate:   highRate,
		LowRate:    lowRate,
		CloseRate:  closeRate,
		Volume:     volume,
	}

	return quote, nil
}

func (bitfinexPoller *BitfinexPoller) findSymbolPair(symbol string, pairs []entities.SymbolPair) (entities.SymbolPair, error) {

	for _, pair := range pairs {
		combined := strings.ToUpper(pair.BaseSymbol.Name + pair.QuoteSymbol.Name)
		if combined == symbol {
			return pair, nil
		}
	}
	return entities.SymbolPair{}, fmt.Errorf("symbol pair not found for %s", symbol)
}

func handleSystemEvent(RawMsg []byte) {
	var msg map[string]interface{}
	if err := json.Unmarshal(RawMsg, &msg); err != nil {
		log.Printf("Error unmarshaling system message: %v", err)
		return
	}

	event, ok := msg["event"].(string)
	if !ok {
		log.Println("Received system message without 'event':", msg)
		return
	}

	switch event {
	case "info":
		code, hasCode := msg["code"].(float64)
		if hasCode {
			switch int(code) {
			case 20051:
				log.Println("Info: Stop, Restart WebSocket Server. Reconnecting")
			case 20060:
				log.Println("Info: Entering maintenance mode. Pause activity for at most 120 seconds.")
				//waiting for 20061 -
			case 20061:
				log.Println("Info: Maintenance ended. Resubscribe to channels needed.")
			default:
				log.Printf("Info message with unhandled code (%d): %v\n", int(code), msg)
			}
		} else {
			log.Println("Info message without code:", msg)
		}

	case "subscribed":
		log.Printf("Sucessfully subscribed to channel: %v", msg)

	case "error":
		code := int(msg["code"].(float64))
		msgText := msg["msg"].(string)
		log.Printf("Error (%d): %s", code, msgText)

		switch code {
		case 10000:
			log.Println("Unknown event")
		case 10001:
			log.Println("Unknown pair")
		case 10300:
			log.Println("Generic subscription failed")
		case 10301:
			log.Println("Already subscribed")
		case 10302:
			log.Println("Unknown channel")
		case 10305:
			log.Println("Reached limit of open channels")
		default:
			log.Println("Unhandled error code")
		}

	default:
		log.Println("Unhandled system event type:", event, "with message: ", msg)
	}
}
