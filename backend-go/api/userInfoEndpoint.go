package api

import (
	"encoding/json"
	"log"
	"superQiMiniAppBackend/alipay"
	"superQiMiniAppBackend/jwe"

	"github.com/gofiber/fiber/v2"
)

func InitUserInfoEndpoint(group fiber.Router) {
	group.Post("/user/info", func(ctx *fiber.Ctx) error {
		var request map[string]string
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		token := request["token"]
		if token == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Token is required")
		}

		log.Println("=================================================================")
		log.Println("INQUIRING USER INFO")
		log.Println("=================================================================")

		// Decrypt the JWE token to get the access token
		claims, err := jwe.ParseAndValidateJWE(token)
		if err != nil {
			log.Printf("[ERROR] Failed to decrypt JWE token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
		}

		userInfo, err := alipay.Interface.InquiryUserInfo(claims.AccessToken)
		if err != nil {
			log.Printf("[ERROR] User info inquiry failed: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		userInfoJson, _ := json.MarshalIndent(userInfo, "", "  ")
		log.Printf("[RESPONSE] User info retrieved:\n%s\n\n", string(userInfoJson))

		if userInfo.Result.ResultCode != "SUCCESS" {
			log.Printf("[ERROR] User info inquiry returned error: %s\n", userInfo.Result.ResultMessage)
			return fiber.NewError(fiber.StatusBadRequest, "Failed to retrieve user info: "+userInfo.Result.ResultMessage)
		}

		log.Println("[SUCCESS] User info retrieved successfully")
		return ctx.JSON(userInfo)
	})
}
