package main

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func Test_main(t *testing.T) {
	tests := []struct {
		description  string
		route        string
		expectedCode int
	}{
		{
			description:  "sanity health check for root path",
			route:        "/",
			expectedCode: 200,
		},
	}

	app := Setup()

	for _, test := range tests {
		// from the test case
		req, _ := http.NewRequest(
			"GET",
			test.route,
			nil,
		)

		// Perform the request plain with the app.
		// The -1 disables request latency.
		res, err := app.Test(req, -1)

		if err != nil {
			continue
		}

		// verify that no error occured, that is not expected
		// Verify if the status code is as expected
		assert.Equalf(t, test.expectedCode, res.StatusCode, test.description)
	}
}

func Test_rates(t *testing.T) {
	defer gock.Disable()
	tests := []struct {
		description  string
		route        string
		expectedCode int
		response     interface{}
	}{
		{
			description:  "check default rates",
			route:        "/rates",
			expectedCode: 200,
			response: RatesCurrencyResponse{
				USD: RatesCryptoBlock{
					BTC:  "1.000000",
					DOGE: "1.000000",
					ETH:  "1.000000",
				},
				SGD: RatesCryptoBlock{
					BTC:  "1.000000",
					DOGE: "1.000000",
					ETH:  "1.000000",
				},
				EUR: RatesCryptoBlock{
					BTC:  "1.000000",
					DOGE: "1.000000",
					ETH:  "1.000000",
				},
			},
		},
		{
			description:  "check fiat rates",
			route:        "/rates?base=fiat",
			expectedCode: 200,
			response: RatesCurrencyResponse{
				USD: RatesCryptoBlock{
					BTC:  "1.000000",
					DOGE: "1.000000",
					ETH:  "1.000000",
				},
				SGD: RatesCryptoBlock{
					BTC:  "1.000000",
					DOGE: "1.000000",
					ETH:  "1.000000",
				},
				EUR: RatesCryptoBlock{
					BTC:  "1.000000",
					DOGE: "1.000000",
					ETH:  "1.000000",
				},
			},
		},
		{
			description:  "check crypto rates",
			route:        "/rates?base=crypto",
			expectedCode: 200,
			response: RatesCryptoResponse{
				BTC: RatesCurrencyBlock{
					USD: "1.000000",
					SGD: "1.000000",
					EUR: "1.000000",
				},
				DOGE: RatesCurrencyBlock{
					USD: "1.000000",
					SGD: "1.000000",
					EUR: "1.000000",
				},
				ETH: RatesCurrencyBlock{
					USD: "1.000000",
					SGD: "1.000000",
					EUR: "1.000000",
				},
			},
		},
		{
			description:  "check incorrect base params",
			route:        "/rates?base=error",
			expectedCode: 422,
			response:     "[base]: 'error' value error",
		},
	}

	app := Setup()

	for _, test := range tests {
		mockRes := ExchangeResponse{
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
					USD:  "1.0",
					SGD:  "1.0",
					EUR:  "1.0",
					BTC:  "1.0",
					DOGE: "1.0",
					ETH:  "1.0",
				},
			},
		}
		gock.New(EXCHANGE_URL).
			Get("/exchange-rates").
			Reply(200).
			JSON(mockRes)

		// from the test case
		req, err := http.NewRequest(
			"GET",
			test.route,
			nil,
		)

		if err != nil {
			continue
		}

		// Perform the request plain with the app.
		// The -1 disables request latency.
		res, err := app.Test(req, -1)

		if err != nil {
			continue
		}

		// verify that no error occured, that is not expected
		// Verify if the status code is as expected
		assert.Equalf(t, test.expectedCode, res.StatusCode, test.description)
		if test.route == "/rates?base=crypto" {
			sample := new(RatesCryptoResponse)
			defer res.Body.Close()
			json.NewDecoder(res.Body).Decode(sample)
			assert.Equalf(t, *sample, test.response, test.description)
		} else if test.expectedCode == 422 {
			assert.Equalf(t, "[base]: 'error' value error", test.response, test.description)
		} else {
			sample := new(RatesCurrencyResponse)
			defer res.Body.Close()
			json.NewDecoder(res.Body).Decode(sample)
			assert.Equalf(t, *sample, test.response, test.description)
		}
	}
}
