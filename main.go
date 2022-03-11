package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

type DolarResponse struct {
	Date   string
	Compra string
	Venta  string
}

func main() {

	app := fiber.New()

	app.Use(cors.New())

	app.Get("/:path", func(ctx *fiber.Ctx) error {
		reqPath := ctx.Params("path")
		valuePath := DolarSiPaths[reqPath]

		//dolar => [buy, sell]
		dolar, err := DolarSiBuySell("cotiza." + valuePath)
		if err != nil {
			return ctx.Status(fiber.StatusInternalServerError).JSON("Could not resolve")
		}

		return ctx.Status(fiber.StatusOK).JSON(DolarResponse{
			Date:   ISOTimestamp(),
			Compra: dolar[0],
			Venta:  dolar[1],
		})
	})

	log.Fatal(app.Listen(":3000"))

}
