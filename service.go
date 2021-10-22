package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type exchangeData struct {
	Buy     string `json:"buy"`
	Sell    string `json:"sell"`
	BuyNow  bool   `json:"buyNow"`
	SellNow bool   `json:"sellNow"`
}

type coinbaseSchema struct {
	Data struct {
		Amount string `json:"amount"`
	} `json:"data"`
}

type krakenSchemaBTC struct {
	Result struct {
		Data struct {
			Buy  []string `json:"a"` // ask
			Sell []string `json:"b"` // bid
		} `json:"XXBTZUSD"`
	} `json:"result"`
}

type krakenSchemaETH struct {
	Result struct {
		Data struct {
			Buy  []string `json:"a"` // ask
			Sell []string `json:"b"` // bid
		} `json:"XETHZUSD"`
	} `json:"result"`
}

func coinbaseToData(transactions map[string]interface{}) *exchangeData {
	buy := transactions["buy"].(*coinbaseSchema).Data.Amount
	sell := transactions["sell"].(*coinbaseSchema).Data.Amount

	return &exchangeData{
		Buy:  buy,
		Sell: sell,
	}
}

func krakenToData(transactions map[string]interface{}, symbol string) *exchangeData {
	var buy, sell string

	switch symbol {
	case "eth":
		buy = transactions["buy-sell"].(*krakenSchemaETH).Result.Data.Buy[0]
		sell = transactions["buy-sell"].(*krakenSchemaETH).Result.Data.Sell[0]

	case "btc":
		buy = transactions["buy-sell"].(*krakenSchemaBTC).Result.Data.Buy[0]
		sell = transactions["buy-sell"].(*krakenSchemaBTC).Result.Data.Sell[0]
	}

	return &exchangeData{
		Buy:  buy,
		Sell: sell,
	}
}

func getRespJSON(url string, schema interface{}) ([]byte, error) {
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	json.NewDecoder(response.Body).Decode(schema)

	j, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	return j, nil
}

func main() {
	apis := map[string]map[string]map[string]struct {
		schema interface{}
		url    string
	}{
		"coinbase": {
			"btc": {
				"buy": {
					schema: &coinbaseSchema{},
					url:    "https://api.coinbase.com/v2/prices/BTC-USD/buy",
				},
				"sell": {
					schema: &coinbaseSchema{},
					url:    "https://api.coinbase.com/v2/prices/BTC-USD/sell",
				},
			},

			"eth": {
				"buy": {
					schema: &coinbaseSchema{},
					url:    "https://api.coinbase.com/v2/prices/ETH-USD/buy",
				},
				"sell": {
					schema: &coinbaseSchema{},
					url:    "https://api.coinbase.com/v2/prices/ETH-USD/sell",
				},
			},
		},

		"kraken": {
			"btc": {
				"buy-sell": {
					schema: &krakenSchemaBTC{},
					url:    "https://api.kraken.com/0/public/Ticker?pair=BTCUSD",
				},
			},
			"eth": {
				"buy-sell": {
					schema: &krakenSchemaETH{},
					url:    "https://api.kraken.com/0/public/Ticker?pair=ETHUSD",
				},
			},
		},
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := make(map[string]*exchangeData)
		symbol := r.URL.Path[1:]

		for exchange, coins := range apis {
			if _, ok := coins[symbol]; !ok {
				continue
			}

			transactions := make(map[string]interface{})

			for transaction, pair := range coins[symbol] {
				_, err := getRespJSON(pair.url, pair.schema)

				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				transactions[transaction] = pair.schema
			}

			switch exchange {
			case "coinbase":
				data[exchange] = coinbaseToData(transactions)
			case "kraken":
				data[exchange] = krakenToData(transactions, symbol)
			}
		}

		j, err := json.Marshal(data)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(j)
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
