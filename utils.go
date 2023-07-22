package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
)

type (
	ExchangePayload struct {
		Base *string `validate:"oneof=crypto fiat ''"`
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

	PayloadErrors struct {
		Field string
		Value interface{}
	}
)

const EXCHANGE_URL string = "https://api.coinbase.com/v2/exchange-rates"

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

func validateExchangePayload(payload *ExchangePayload) *string {
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
		url = fmt.Sprint(EXCHANGE_URL, "?currency=", currency)
	}
	getJson(url, exchange_response)
	return exchange_response
}
