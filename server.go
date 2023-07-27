package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron/v3"
)

func main() {
	db := SetupDb()
	app := SetupApp(db)
	SetupCron(db)
	defer db.Close()
	log.Fatal(app.Listen(":3000"))
	for {
		select {}
	}
}

func GetDBFile() string {
	if flag.Lookup("test.v") == nil {
		return "./main.db"
	}

	return "./test.db"
}

func SetupApp(db *sql.DB) *fiber.App {
	app := fiber.New()
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Backend is very healthy!")
	})

	app.Get("/rates-v1", func(c *fiber.Ctx) error {
		base := c.Query("base")
		payload := &ExchangePayload{
			Base: &base,
		}
		errMsg := validateExchangePayload(payload)
		if errMsg != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, *errMsg)
		}
		data := getExchange(payload.Base)
		if *payload.Base == "crypto" {
			res := parseCryptoResponse(data)
			return c.JSON(res)
		} else {
			res := parseFiatResponse(data)
			return c.JSON(res)
		}
	})

	app.Get("/rates", func(c *fiber.Ctx) error {
		base := c.Query("base")
		payload := &ExchangePayload{
			Base: &base,
		}
		errMsg := validateExchangePayload(payload)
		if errMsg != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, *errMsg)
		}
		data := getExchangeDB(&base, db)
		if data == nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "error fetching from db")
		}
		if *payload.Base == "crypto" {
			res := parseCryptoResponse(data)
			return c.JSON(res)
		} else {
			res := parseFiatResponse(data)
			return c.JSON(res)
		}
	})

	app.Get("/historical-rates", func(c *fiber.Ctx) error {
		baseCurrency := c.Query("baseCurrency")
		targetCurrency := c.Query("targetCurrency")
		start := c.Query("start")
		end := c.Query("end")

		payload := ExchangeHistoricPayload{
			BaseCurrency:   &baseCurrency,
			TargetCurrency: &targetCurrency,
			Start:          &start,
			End:            &end,
		}

		errMsg := validateExchangePayload(payload)
		if errMsg != nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, *errMsg)
		}

		if *payload.Start == string("") {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "[start]: start invalid time ''")
		}

		data := getExchangeHistoric(payload, db)
		if data == nil {
			return fiber.NewError(fiber.StatusUnprocessableEntity, "unable to get data from db")
		}

		return c.JSON(*data)
	})

	return app
}

func SetupDb() *sql.DB {
	db_file := GetDBFile()
	db, err := sql.Open("sqlite3", db_file)
	if err != nil {
		log.Fatal(err)
	}
	createSql := `
	create table if not exists rates (
		id integer not null primary key,
		USD double default 1,
		SGD double,
		EUR double,
		BTC double,
		DOGE double,
		ETH double
	);
	`
	_, err = db.Exec(createSql)
	if err != nil {
		log.Fatal("error creating rates database", err)
	}

	return db
}

func SetupCron(db *sql.DB) {
	s := cron.New()
	fmt.Println("Starting cron scheduler")
	_, err := s.AddFunc("*/1 * * * *", func() {
		fmt.Println("fetching data from api")
		base := "USD"
		data := getExchange(&base)
		if data == nil {
			fmt.Println("Error encountered fetching data, trying again 1 minute")
			return
		}
		utcTimestamp := time.Now().UTC().Unix()

		sgd, _ := data.Data.Rates.SGD.Float64()
		eur, _ := data.Data.Rates.EUR.Float64()
		btc, _ := data.Data.Rates.BTC.Float64()
		doge, _ := data.Data.Rates.DOGE.Float64()
		eth, _ := data.Data.Rates.ETH.Float64()

		insertSql := fmt.Sprintf(`
		insert into 
			rates(id, SGD, EUR, BTC, DOGE, ETH)
		values(
			'%d',
			%f,
			%f,
			%f,
			%f,
			%f
		);`,
			utcTimestamp,
			sgd,
			eur,
			btc,
			doge,
			eth,
		)
		fmt.Println("writing statement", insertSql)
		_, err := db.Exec(insertSql)
		if err != nil {
			fmt.Println("error writing to db", err)
			return
		}
		fmt.Println("Data fetch & db store success")
	})
	if err != nil {
		log.Fatal(err)
	}
	s.Start()
}
