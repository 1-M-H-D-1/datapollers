package main

import "DataPoller/internal/app/bitfinexpoller"

func main() {
	bitfinexpoller.RunBitfinexPoller()
}

/*func main() {
	url := "wss://api-pub.bitfinex.com/ws/2"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Dial error:", err)
	}
	defer conn.Close()

	// Подписка на тикер BTC/USD
	subMsg := map[string]interface{}{
		"event":   "subscribe",
		"channel": "ticker",
		"symbol":  "tBTCUSD",
	}
	subData, _ := json.Marshal(subMsg)
	err = conn.WriteMessage(websocket.TextMessage, subData)
	if err != nil {
		log.Fatal("Write error:", err)
	}

	// Чтение сообщений
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			break
		}

		fmt.Println("Message:", string(msg))
		time.Sleep(1 * time.Second) // чуть замедляем поток
	}
}*/
