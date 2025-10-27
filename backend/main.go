package main

import (
	"log"
	"os"
	"superQiMiniAppBackend/alipay"
	"superQiMiniAppBackend/api"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	recover2 "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	err := alipay.InitAlipayClient()
	if err := err; err != nil {
		log.Fatal(err)
	}

	app := initWebServer()

	apiGroup := app.Group("/api")
	api.InitAuthEndpoint(apiGroup)
	api.InitPaymentEndpoint(apiGroup)
	api.InitRefundEndpoint(apiGroup)
	api.InitAgreementEndpoint(apiGroup)

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "1999"
	}

	log.Printf("Server starting on port %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Printf("Server error: %v", err)
	}
}

func initWebServer() *fiber.App {
	app := fiber.New()

	app.Use(cors.New())
	app.Use(recover2.New())

	return app
}
