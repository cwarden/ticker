package quote

import (
	"fmt"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
)

type ResponseQuote struct {
	ShortName                  string  `json:"shortName"`
	Symbol                     string  `json:"symbol"`
	MarketState                string  `json:"marketState"`
	Currency                   string  `json:"currency"`
	ExchangeName               string  `json:"fullExchangeName"`
	ExchangeDelay              float64 `json:"exchangeDataDelayedBy"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketPreviousClose float64 `json:"regularMarketPreviousClose"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	RegularMarketDayRange      string  `json:"regularMarketDayRange"`
	PostMarketChange           float64 `json:"postMarketChange"`
	PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
	PostMarketPrice            float64 `json:"postMarketPrice"`
	PreMarketChange            float64 `json:"preMarketChange"`
	PreMarketChangePercent     float64 `json:"preMarketChangePercent"`
	PreMarketPrice             float64 `json:"preMarketPrice"`
	PriceToBook                float64 `json:"priceToBook"`
	TrailingPE                 float64 `json:"trailingPE"`
	DividendDate               int64   `json:"dividendDate"`
	AnnualDividend             float64 `json:"trailingAnnualDividendRate"`
	DividendYield              float64 `json:"trailingAnnualDividendYield"`
}

type Quote struct {
	ResponseQuote
	Price                   float64
	Change                  float64
	ChangePercent           float64
	IsActive                bool
	IsRegularTradingSession bool
	DividendDate            time.Time
}

type Response struct {
	QuoteResponse struct {
		Quotes []ResponseQuote `json:"result"`
		Error  interface{}     `json:"error"`
	} `json:"quoteResponse"`
}

func transformResponseQuote(responseQuote ResponseQuote) Quote {
	q := Quote{
		ResponseQuote:           responseQuote,
		Price:                   responseQuote.RegularMarketPrice,
		Change:                  0.0,
		ChangePercent:           0.0,
		IsActive:                false,
		IsRegularTradingSession: false,
	}
	if responseQuote.DividendDate != 0 {
		q.DividendDate = time.Unix(responseQuote.DividendDate, 0)
	}

	if responseQuote.MarketState == "REGULAR" {
		q.Change = responseQuote.RegularMarketChange
		q.ChangePercent = responseQuote.RegularMarketChangePercent
		q.IsActive = true
		q.IsRegularTradingSession = true
	}

	if responseQuote.MarketState == "POST" {
		q.Price = responseQuote.PostMarketPrice
		q.Change = responseQuote.PostMarketChange + responseQuote.RegularMarketChange
		q.ChangePercent = responseQuote.PostMarketChangePercent + responseQuote.RegularMarketChangePercent
		q.IsActive = true
	}

	if responseQuote.MarketState == "PRE" {
		q.Price = responseQuote.PreMarketPrice
		q.Change = responseQuote.PreMarketChange
		q.ChangePercent = responseQuote.PreMarketChangePercent
		q.IsActive = true
	}
	return q
}

func transformResponseQuotes(responseQuotes []ResponseQuote) []Quote {

	quotes := make([]Quote, 0)
	for _, responseQuote := range responseQuotes {
		quotes = append(quotes, transformResponseQuote(responseQuote))
	}
	return quotes

}

func GetQuotes(client resty.Client, symbols []string) func() []Quote {
	return func() []Quote {
		symbolsString := strings.Join(symbols, ",")
		url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=%s", symbolsString)
		res, _ := client.R().
			SetResult(Response{}).
			Get(url)

		return transformResponseQuotes((res.Result().(*Response)).QuoteResponse.Quotes)
	}
}
