package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robfig/cron/v3"
)

func main() {
	app := SetupApp()
	db := SetupDb()
	SetupCron(db)
	defer db.Close()
	log.Fatal(app.Listen(":3000"))
	for {
		select {}
	}
}

func SetupApp() *fiber.App {
	app := fiber.New()
	app.Use(recover.New())

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Backend is very healthy!")
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
		data := getExchange(payload.Base)
		if *payload.Base == "crypto" {
			res := parseCryptoResponse(data)
			return c.JSON(res)
		} else {
			res := parseFiatResponse(data)
			return c.JSON(res)
		}
	})

	return app
}

func SetupDb() *sql.DB {
	db, err := sql.Open("sqlite3", "./main.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func SetupCron(db *sql.DB) {
	s := cron.New()
	fmt.Println(db.Stats().MaxOpenConnections)
	val, err := s.AddFunc("*/1 * * * *", func() {
		fmt.Println("hi there!")
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(s.Entries(), val)
	s.Start()
}
