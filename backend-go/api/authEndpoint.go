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

		info, err := alipay.Interface.InquiryUserInfo(tokenResponse.AccessToken)
		if err != nil {
			log.Printf("[ERROR] Failed to retrieve user info: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		infoJson, _ := json.MarshalIndent(info, "", "  ")
		log.Printf("[SUCCESS] User info retrieved:\n%s\n\n", string(infoJson))

		// Return a JWE to the MiniApp containing the required access token to be used in future calls to A+ backend
		jweToken, err := jwe.CreateJWE(jwe.TokenClaims{
			UserID:      info.UserInfo.UserID,
			AccessToken: tokenResponse.AccessToken,
		})

		if err != nil {
			log.Printf("[ERROR] Failed to create JWE token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		// Build response object - only return token, no payment info
		response := fiber.Map{
			"token": jweToken,
		}

		log.Println("[SUCCESS] Returning auth token to frontend")
		return ctx.JSON(response)
	})
}
