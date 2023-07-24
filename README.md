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
