package api

import (
	"fmt"
	"log"
	"strings"
	"superQiMiniAppBackend/alipay"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

const defaultShortLinkDescription = "Miniapp short link"

type generateShortLinkRequest struct {
	AppID       string `json:"appId"`
	PagePath    string `json:"pagePath"`
	QueryParams string `json:"queryParams"`
	Description string `json:"description"`
}

func InitShortLinkEndpoint(group fiber.Router) {
	// Endpoint to generate a short link (QR code link) that opens a miniapp page
	group.Post("/miniapps/short-link", func(ctx *fiber.Ctx) error {
		var request generateShortLinkRequest
		if err := ctx.BodyParser(&request); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if request.AppID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "appId is required")
		}
		// pagePath and queryParams are optional: without a page the link opens the miniapp's default page
		if request.PagePath != "" && !strings.HasPrefix(request.PagePath, "/") {
			return fiber.NewError(fiber.StatusBadRequest, "pagePath must start with /")
		}
		// The gateway rejects requests without a description
		if request.Description == "" {
			request.Description = defaultShortLinkDescription
		}

		log.Println("=================================================================")
		log.Println("STARTING SHORT LINK GENERATION")
		log.Println("=================================================================")

		requestID := fmt.Sprintf("SHORTLINK-%s-%d", uuid.New().String(), time.Now().Unix())
		log.Printf("[INFO] Request ID: %s\n", requestID)
		log.Printf("[INFO] App ID: %s, page: %s, params: %s\n", request.AppID, request.PagePath, request.QueryParams)

		response, err := alipay.Interface.GenerateShortLink(requestID, request.AppID, request.PagePath, request.QueryParams, request.Description)
		if err != nil {
			log.Printf("[ERROR] Failed to generate short link: %v\n", err)
			log.Println("=================================================================")
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		if response.Result.ResultStatus == "S" {
			log.Printf("[SUCCESS] Short link generated: %s\n", response.AppQrCode)
		} else {
			log.Printf("[ERROR] Short link generation failed: %s (%s)\n", response.Result.ResultMessage, response.Result.ResultCode)
		}
		log.Println("=================================================================")

		return ctx.JSON(response)
	})
}
