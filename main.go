package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
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
		Addr:     os.Getenv("REDISADDR"),
		Password: os.Getenv("REDISPASS"),
		DB:       0,
	})

	app := fiber.New()

	app.Use(cors.New())
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	app.Get("/:path", func(ctx *fiber.Ctx) error {
		reqPath := ctx.Params("path")

		//val = map[reqPath]value
		val, err := rdb.HGetAll(gctx, reqPath).Result()

		//Fails if there is no key or there is a redis error
		if err != nil {
			//If there is no key will look for it at the API, parse it,
			//cache it and send it to the client

			fmt.Println(err)
			//Any other redis error will be handled as 500
			return ctx.Status(fiber.StatusInternalServerError).JSON("Could not resolve")

		}
		//Values are cached as Redis Hashes. The redis client
		//retrieves them as a map. If there is nothing cached =>
		//is an empty map
		if len(val) == 0 {
			valuePath := DolarSiPaths[reqPath]
			//dolar => [buy, sell]
			dolar, err := DolarSiBuySell(valuePath)

			if err != nil {
				//The path isnt at the DolarSiPaths or the api has changed
				if err.Error() == "Non valid path" {
					return ctx.Status(fiber.StatusNotFound).JSON("Not found")
				}

				return ctx.Status(fiber.StatusInternalServerError).JSON("Could not resolve")
			}

			if rdb.HSet(gctx, reqPath, "buy", dolar[0], "sell", dolar[1]).Err() != nil {
				fmt.Println("Error at caching")
			}

			return ctx.Status(fiber.StatusOK).JSON(DolarResponse{
				Date:   ISOTimestamp(),
				Compra: dolar[0],
				Venta:  dolar[1],
			})
		}

		//If err == nil && len(map) != 0 => The value is cached and returned
		return ctx.Status(fiber.StatusOK).JSON(DolarResponse{
			Date:   ISOTimestamp(),
			Compra: val["buy"],
			Venta:  val["sell"],
		})

	})

	log.Fatal(app.Listen(":" + port))

}
