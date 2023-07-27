package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type (
	ExchangePayload struct {
		Base *string `validate:"oneof=crypto fiat ''"`
	}

	ExchangeHistoricPayload struct {
		BaseCurrency   *string `validate:"oneof=USD SGD EUR BTC DOGE ETH"`
		TargetCurrency *string `validate:"oneof=USD SGD EUR BTC DOGE ETH"`
		Start          *string `validate:"required"`
		End            *string
	}

	ExchangeResponse struct {
		Data struct {
			Currency string `json:"currency"`
			Rates    struct {
				USD  json.Number `json:"USD"`
				SGD  json.Number `json:"SGD"`
				EUR  json.Number `json:"EUR"`
				BTC  json.Number `json:"BTC"`
				DOGE json.Number `json:"DOGE"`
				ETH  json.Number `json:"ETH"`
			} `json:"rates"`
		} `json:"data"`
	}

	RatesCurrencyBlock struct {
		USD string
		SGD string
		EUR string
	}

	RatesCryptoBlock struct {
		BTC  string
		DOGE string
		ETH  string
	}

	RatesCryptoResponse struct {
		BTC  RatesCurrencyBlock
		DOGE RatesCurrencyBlock
		ETH  RatesCurrencyBlock
	}

	RatesCurrencyResponse struct {
		USD RatesCryptoBlock
		SGD RatesCryptoBlock
		EUR RatesCryptoBlock
	}

	TimeSeries struct {
		Timestamp int
		Value     string
	}

	HistoricResponse struct {
		Results []TimeSeries
	}

	PayloadErrors struct {
		Field string
		Value interface{}
	}
)

const EXCHANGE_URL string = "https://api.coinbase.com/v2"

var myClient = &http.Client{Timeout: 10 * time.Second}
var validate = validator.New()

func getJson(url string, target interface{}) error {
	r, err := myClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}

func parsePayloadErrors(errors []PayloadErrors) *string {
	if len(errors) == 0 {
		return nil
	}

	errMsgs := make([]string, 0)

	for _, err := range errors {
		errMsgs = append(errMsgs, fmt.Sprintf(
			"[%s]: '%v' value error",
			strings.ToLower(err.Field),
			err.Value,
		))
	}

	strMsg := strings.Join(errMsgs, " and ")
	return &strMsg
}

func validateExchangePayload(payload interface{}) *string {
	validationErrors := []PayloadErrors{}
	errs := validate.Struct(payload)
	if errs != nil {
		for _, err := range errs.(validator.ValidationErrors) {
			var elem PayloadErrors
			elem.Field = err.Field()
			elem.Value = err.Value()
			validationErrors = append(validationErrors, elem)
		}
	}
	return parsePayloadErrors(validationErrors)
}

func parseCryptoResponse(data *ExchangeResponse) RatesCryptoResponse {
	usd, _ := data.Data.Rates.USD.Float64()
	sgd, _ := data.Data.Rates.SGD.Float64()
	eur, _ := data.Data.Rates.EUR.Float64()
	doge, _ := data.Data.Rates.DOGE.Float64()
	eth, _ := data.Data.Rates.ETH.Float64()

	btcBlock := RatesCurrencyBlock{
		USD: fmt.Sprintf("%f", usd),
		SGD: fmt.Sprintf("%f", sgd),
		EUR: fmt.Sprintf("%f", eur),
	}

	dogeBlock := RatesCurrencyBlock{
		USD: fmt.Sprintf("%f", usd/doge),
		SGD: fmt.Sprintf("%f", sgd/doge),
		EUR: fmt.Sprintf("%f", eur/doge),
	}

	ethBlock := RatesCurrencyBlock{
		USD: fmt.Sprintf("%f", usd/eth),
		SGD: fmt.Sprintf("%f", sgd/eth),
		EUR: fmt.Sprintf("%f", eur/eth),
	}

	resp := RatesCryptoResponse{
		BTC:  btcBlock,
		DOGE: dogeBlock,
		ETH:  ethBlock,
	}

	return resp
}

func parseFiatResponse(data *ExchangeResponse) RatesCurrencyResponse {
	btc, _ := data.Data.Rates.BTC.Float64()
	doge, _ := data.Data.Rates.DOGE.Float64()
	eth, _ := data.Data.Rates.ETH.Float64()
	sgd, _ := data.Data.Rates.SGD.Float64()
	eur, _ := data.Data.Rates.EUR.Float64()

	usdBlock := RatesCryptoBlock{
		BTC:  fmt.Sprintf("%f", btc),
		DOGE: fmt.Sprintf("%f", doge),
		ETH:  fmt.Sprintf("%f", eth),
	}

	sgdBlock := RatesCryptoBlock{
		BTC:  fmt.Sprintf("%f", btc/sgd),
		DOGE: fmt.Sprintf("%f", doge/sgd),
		ETH:  fmt.Sprintf("%f", eth/sgd),
	}

	eurBlock := RatesCryptoBlock{
		BTC:  fmt.Sprintf("%f", btc/eur),
		DOGE: fmt.Sprintf("%f", doge/eur),
		ETH:  fmt.Sprintf("%f", eth/eur),
	}

	resp := RatesCurrencyResponse{
		USD: usdBlock,
		SGD: sgdBlock,
		EUR: eurBlock,
	}

	return resp
}

func getExchange(base *string) *ExchangeResponse {
	exchange_response := new(ExchangeResponse)
	var url = EXCHANGE_URL
	if base != nil {
		currency := "USD"
		if *base == "crypto" {
			currency = "BTC"
		}
		url = fmt.Sprint(EXCHANGE_URL, "/exchange-rates", "?currency=", currency)
	}
	getJson(url, exchange_response)
	return exchange_response
}

func getExchangeDB(base *string, db *sql.DB) *ExchangeResponse {
	fetchSql := `
	select * from rates
	order by id DESC 
	limit 1;
	`
	rows, err := db.Query(fetchSql)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()
	var resp ExchangeResponse
	for rows.Next() {
		var id int
		var sgd float64
		var eur float64
		var btc float64
		var doge float64
		var eth float64
		err = rows.Scan(
			&id,
			&sgd,
			&eur,
			&btc,
			&doge,
			&eth,
		)
		if err != nil {
			fmt.Println(err)
			return nil
		}

		resp = ExchangeResponse{
			Data: struct {
				Currency string "json:\"currency\""
				Rates    struct {
					USD  json.Number "json:\"USD\""
					SGD  json.Number "json:\"SGD\""
					EUR  json.Number "json:\"EUR\""
					BTC  json.Number "json:\"BTC\""
					DOGE json.Number "json:\"DOGE\""
					ETH  json.Number "json:\"ETH\""
				} "json:\"rates\""
			}{
				Currency: "USD",
				Rates: struct {
					USD  json.Number "json:\"USD\""
					SGD  json.Number "json:\"SGD\""
					EUR  json.Number "json:\"EUR\""
					BTC  json.Number "json:\"BTC\""
					DOGE json.Number "json:\"DOGE\""
					ETH  json.Number "json:\"ETH\""
				}{
					USD:  "1",
					SGD:  json.Number(fmt.Sprintf("%f", sgd)),
					EUR:  json.Number(fmt.Sprintf("%f", eur)),
					BTC:  json.Number(fmt.Sprintf("%f", btc)),
					DOGE: json.Number(fmt.Sprintf("%f", doge)),
					ETH:  json.Number(fmt.Sprintf("%f", eth)),
				},
			},
		}

		return &resp
	}

	return nil
}

func getExchangeHistoric(payload ExchangeHistoricPayload, db *sql.DB) *HistoricResponse {
	start, err := strconv.ParseInt(*payload.Start, 10, 64)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	var end int64
	if *payload.End == string("") {
		end = time.Now().UTC().Unix()
	} else {
		end, err = strconv.ParseInt(*payload.End, 10, 64)
		if err != nil {
			end = time.Now().UTC().Unix()
		}
	}

	base := *payload.BaseCurrency
	target := *payload.TargetCurrency

	fetchSql := fmt.Sprintf(`
	select %s, %s, id from rates
	where id >= %d and id <= %d;
	`, base, target, start, end)

	rows, err := db.Query(fetchSql)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	defer rows.Close()
	res := new(HistoricResponse)
	timeseries := make([]TimeSeries, 0)
	for rows.Next() {
		var timestamp int
		var target float64
		var base float64
		err := rows.Scan(
			&base,
			&target,
			&timestamp,
		)
		if err != nil {
			fmt.Println(err)
			continue
		}
		dataPoint := TimeSeries{
			Timestamp: timestamp,
			Value:     fmt.Sprintf("%f", target/base),
		}
		timeseries = append(timeseries, dataPoint)
	}

	res.Results = timeseries
	return res
}
