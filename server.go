package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	app := Setup()
	log.Fatal(app.Listen(":3000"))
}

func Setup() *fiber.App {
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
