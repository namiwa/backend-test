package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	app := SetupApp()
	db := SetupDb()
	SetupCron(db)
	defer db.Close()
	log.Fatal(app.Listen(":3000"))
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
	s := gocron.NewScheduler(time.UTC)
	now := time.Now()
	fmt.Println(db.Stats().Idle)
	s.Every(1).Minute().StartAt(now).Do(func() {
		fmt.Println(db.Stats().Idle, db.Stats().InUse)
	})
}
