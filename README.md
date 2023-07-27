# backend-test

Sample backend written in Golang & Fiber, pulling data from Coinbase exchange API. Database is using SQLite3.

## Installation

Ensure that air is installed for Fiber hot reload for development.
Ensure that Golang version +1.17.8 is installed locally.

Run the following `go run .` and inspect responses with postman / browser to verify endpoints.

```
backend-test> go run .

 ┌───────────────────────────────────────────────────┐
 │                   Fiber v2.48.0                   │
 │               http://127.0.0.1:3000               │
 │       (bound on host 0.0.0.0 and port 3000)       │
 │                                                   │
 │ Handlers ............. 5  Processes ........... 1 │
 │ Prefork ....... Disabled  PID ............. 31080 │
 └───────────────────────────────────────────────────┘
```

## Scheduler limitations

The cron scheduler is limited to jobs where 1 minute is the smallest interval to queue them. So for this case, its not possible to do the data refresh in smaller timescales then seconds.

## Implementation Details

Golang & SQLite3 were used since the language natively supports async operations, and SQLite3 is robust local database which is suited for this task.

## Expected endpoints

`http://localhost:3000/rates?base=fiat`

```json
{
  "USD": {
    "BTC": "0.000034",
    "DOGE": "12.774655",
    "ETH": "0.000536"
  },
  "SGD": {
    "BTC": "0.000026",
    "DOGE": "9.611833",
    "ETH": "0.000403"
  },
  "EUR": {
    "BTC": "0.000037",
    "DOGE": "14.053804",
    "ETH": "0.000590"
  }
}
```

`http://localhost:3000/rates-v1?base=fiat`

```json
{
  "USD": {
    "BTC": "0.000034",
    "DOGE": "12.774655",
    "ETH": "0.000536"
  },
  "SGD": {
    "BTC": "0.000026",
    "DOGE": "9.611834",
    "ETH": "0.000403"
  },
  "EUR": {
    "BTC": "0.000038",
    "DOGE": "14.053808",
    "ETH": "0.000590"
  }
}
```

`http://localhost:3000/historical-rates?baseCurrency=SGD&targetCurrency=USD&start=1690478340`

```json
{
  "Results": [
    {
      "Timestamp": 1690478340,
      "Value": "0.752310"
    },
    {
      "Timestamp": 1690478400,
      "Value": "0.752312"
    },
    {
      "Timestamp": 1690478700,
      "Value": "0.752191"
    }
  ]
}
```

## Testing

Was validated for only rates-v1 and rates, running `go test`
