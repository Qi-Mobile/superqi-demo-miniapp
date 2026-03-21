package api

import (
	"encoding/json"
	"log"
	"superQiMiniAppBackend/alipay"

	"github.com/gofiber/fiber/v2"
)

type applyTokenRequest struct {
	AuthCode string `json:"auth_code" validate:"required"`
}

type inquiryCardsRequest struct {
	AccessToken string `json:"accessToken" validate:"required"`
}

func InitInquiryEndpoint(group fiber.Router) {
	// Endpoint to exchange auth code for access token (specifically for card inquiry)
	group.Post("/users/inquiry-cards/apply-token", func(ctx *fiber.Ctx) error {
		var request applyTokenRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("STARTING TOKEN EXCHANGE FOR CARD INQUIRY")
		log.Println("=================================================================")
		log.Println("[INFO] Auth code received from frontend")

		tokenResponse, err := alipay.Interface.ApplyToken(request.AuthCode)
		if err != nil {
			log.Printf("[ERROR] Token exchange failed: %v\n", err)
			log.Println("=================================================================\n")
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		tokenResponseJson, _ := json.MarshalIndent(tokenResponse, "", "  ")
		log.Printf("[SUCCESS] Token response received:\n%s\n\n", string(tokenResponseJson))

		if tokenResponse.Result.ResultCode != "SUCCESS" {
			log.Printf("[ERROR] Invalid token response: %s\n", tokenResponse.Result.ResultMessage)
			log.Println("=================================================================\n")
			return fiber.NewError(fiber.StatusBadRequest, "Invalid token response: "+tokenResponse.Result.ResultMessage)
		}

		log.Println("[INFO] Token exchange successful")
		log.Printf("[INFO] Customer ID: %s\n", tokenResponse.CustomerID)
		log.Printf("[INFO] Access token obtained (valid until: %s)\n", tokenResponse.AccessTokenExpiryTime)
		log.Println("[SUCCESS] Returning access token to frontend")
		log.Println("=================================================================\n")

		response := fiber.Map{
			"accessToken":            tokenResponse.AccessToken,
			"customerId":             tokenResponse.CustomerID,
			"accessTokenExpiryTime":  tokenResponse.AccessTokenExpiryTime,
		}

		return ctx.JSON(response)
	})

	// Endpoint to get user card list using access token
	group.Post("/users/inquiry-cards", func(ctx *fiber.Ctx) error {
		var request inquiryCardsRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("STARTING USER CARD LIST INQUIRY")
		log.Println("=================================================================")
		log.Println("[INFO] Access token received from frontend")
		log.Println("[INFO] Calling Alipay+ inquiryUserCardList API...")

		cardListResponse, err := alipay.Interface.InquiryUserCardList(request.AccessToken)
		if err != nil {
			log.Printf("[ERROR] Card inquiry failed: %v\n", err)
			log.Println("=================================================================\n")
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		cardListResponseJson, _ := json.MarshalIndent(cardListResponse, "", "  ")
		log.Printf("[SUCCESS] Card list response received:\n%s\n\n", string(cardListResponseJson))

		if cardListResponse.Result.ResultStatus == "S" {
			cardCount := len(cardListResponse.CardList)
			log.Printf("[SUCCESS] Card inquiry successful - %d card(s) found\n", cardCount)

			if cardCount > 0 {
				log.Println("[INFO] Card details:")
				for i, card := range cardListResponse.CardList {
					log.Printf("  Card %d:\n", i+1)
					log.Printf("    Masked Card No: %s\n", card.MaskedCardNo)
					log.Printf("    Account Number: %s\n", card.AccountNumber)
				}
			} else {
				log.Println("[INFO] User has no cards bound to their account")
			}
		} else if cardListResponse.Result.ResultStatus == "F" {
			log.Printf("[ERROR] Card inquiry failed: %s\n", cardListResponse.Result.ResultMessage)
			log.Printf("[ERROR] Result code: %s\n", cardListResponse.Result.ResultCode)
		} else if cardListResponse.Result.ResultStatus == "U" {
			log.Printf("[WARNING] Card inquiry status unknown: %s\n", cardListResponse.Result.ResultMessage)
		}

		log.Println("[SUCCESS] Returning card list to frontend")
		log.Println("=================================================================\n")

		return ctx.JSON(cardListResponse)
	})
}
