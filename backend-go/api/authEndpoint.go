package api

import (
	"encoding/json"
	"log"
	"superQiMiniAppBackend/alipay"
	"superQiMiniAppBackend/jwe"

	"github.com/gofiber/fiber/v2"
)

type authRequest struct {
	AuthCode string `json:"auth_code" validate:"required"`
}

func InitAuthEndpoint(group fiber.Router) {
	group.Post("/auth/apply-token", func(ctx *fiber.Ctx) error {
		var request authRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("STARTING AUTH TOKEN EXCHANGE")
		log.Println("=================================================================")

		tokenResponse, err := alipay.Interface.ApplyToken(request.AuthCode)
		if err != nil {
			log.Printf("[ERROR] Token exchange failed: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		tokenResponseJson, _ := json.MarshalIndent(tokenResponse, "", "  ")
		log.Printf("[SUCCESS] Token response received:\n%s\n\n", string(tokenResponseJson))

		if tokenResponse.Result.ResultCode != "SUCCESS" {
			log.Printf("[ERROR] Invalid token response: %s\n", tokenResponse.Result.ResultMessage)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid token response: "+tokenResponse.Result.ResultMessage)
		}

		log.Println("[INFO] Token exchange successful")
		log.Printf("[INFO] Customer ID: %s\n", tokenResponse.CustomerID)
		log.Println("[INFO] User/Merchant detailed info can be retrieved via separate endpoints")

		// Return a JWE to the MiniApp containing the access token and customer ID
		// The customerID is returned from the token exchange and can be either userId or merchantId
		jweToken, err := jwe.CreateJWE(jwe.TokenClaims{
			UserID:      tokenResponse.CustomerID,
			AccessToken: tokenResponse.AccessToken,
		})

		if err != nil {
			log.Printf("[ERROR] Failed to create JWE token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		response := fiber.Map{
			"token": jweToken,
		}

		log.Println("[SUCCESS] Returning auth token to frontend")
		return ctx.JSON(response)
	})
}
