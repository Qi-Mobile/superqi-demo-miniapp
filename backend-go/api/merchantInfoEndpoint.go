package api

import (
	"encoding/json"
	"log"
	"superQiMiniAppBackend/alipay"
	"superQiMiniAppBackend/jwe"

	"github.com/gofiber/fiber/v2"
)

func InitMerchantInfoEndpoint(group fiber.Router) {
	group.Post("/merchant/info", func(ctx *fiber.Ctx) error {
		var request map[string]string
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		token := request["token"]
		if token == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Token is required")
		}

		log.Println("=================================================================")
		log.Println("INQUIRING MERCHANT INFO")
		log.Println("=================================================================")

		// Decrypt the JWE token to get the access token
		claims, err := jwe.ParseAndValidateJWE(token)
		if err != nil {
			log.Printf("[ERROR] Failed to decrypt JWE token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}

		log.Printf("[INFO] Customer ID from token: %s\n", claims.UserID)
		log.Printf("[INFO] Calling InquiryMerchantInfo API...\n")

		merchantInfo, err := alipay.Interface.InquiryMerchantInfo(claims.AccessToken)
		if err != nil {
			log.Printf("[ERROR] Merchant info inquiry failed: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		merchantInfoJson, _ := json.MarshalIndent(merchantInfo, "", "  ")
		log.Printf("[RESPONSE] Merchant info retrieved:\n%s\n\n", string(merchantInfoJson))

		if merchantInfo.Result.ResultCode != "SUCCESS" {
			log.Printf("[ERROR] Merchant info inquiry returned error: %s\n", merchantInfo.Result.ResultMessage)
			return fiber.NewError(fiber.StatusBadRequest, "Failed to retrieve merchant info: "+merchantInfo.Result.ResultMessage)
		}

		log.Println("[SUCCESS] Merchant info retrieved successfully")
		return ctx.JSON(merchantInfo)
	})
}
