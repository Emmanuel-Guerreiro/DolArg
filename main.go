package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

type DolarResponse struct {
	Date   string
	Compra string
	Venta  string
}

var gctx = context.Background()

func main() {

	var port string
	godotenv.Load(".env")

	if os.Getenv("MODE") == "DEVELOPMENT" {
		port = "3000"
	} else {
		port = os.Getenv("PORT")
		if port == "" {
			log.Fatal("$PORT must be set")
		}
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	app := fiber.New()

	app.Use(cors.New())

	app.Get("/:path", func(ctx *fiber.Ctx) error {
		reqPath := ctx.Params("path")

		val, err := rdb.Get(gctx, reqPath).Result()
		fmt.Println("val,", val)

		//Fails if there is no key or there is a redis error
		if err != nil {
			//If there is no key will look for it at the API, parse it,
			//cache it and send it to the client
			if err.Error() == "redis: nil" {
				valuePath := DolarSiPaths[reqPath]
				fmt.Println(valuePath)
				//dolar => [buy, sell]
				dolar, err := DolarSiBuySell(valuePath)

				if err != nil {
					//The path isnt at the DolarSiPaths or the api has changed
					if err.Error() == "Non valid path" {
						return ctx.Status(fiber.StatusNotFound).JSON("Not found")
					}

					return ctx.Status(fiber.StatusInternalServerError).JSON("Could not resolve")
				}

				return ctx.Status(fiber.StatusOK).JSON(DolarResponse{
					Date:   ISOTimestamp(),
					Compra: dolar[0],
					Venta:  dolar[1],
				})
			} else {
				//Any other redis error will be handled as 500
				return ctx.Status(fiber.StatusInternalServerError).JSON("Could not resolve")
			}
		}

		//If err == nil => The value is cached and returned
		return ctx.Status(fiber.StatusOK).JSON(DolarResponse{
			Date:   ISOTimestamp(),
			Compra: "10",
			Venta:  "10",
		})

	})

	log.Fatal(app.Listen(":" + port))

}
