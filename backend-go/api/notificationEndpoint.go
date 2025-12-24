package api

import (
	"encoding/json"
	"fmt"
	"log"
	"superQiMiniAppBackend/alipay"
	"superQiMiniAppBackend/jwe"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type sendInboxRequest struct {
	Token   string `json:"token" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
	Url     string `json:"url,omitempty"`
}

type sendPushRequest struct {
	Token   string `json:"token" validate:"required"`
	Title   string `json:"title" validate:"required"`
	Content string `json:"content" validate:"required"`
	Url     string `json:"url,omitempty"`
}

func InitNotificationEndpoint(group fiber.Router) {
	// POST /api/notification/send-inbox
	group.Post("/notification/send-inbox", func(ctx *fiber.Ctx) error {
		var request sendInboxRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("SEND INBOX NOTIFICATION REQUEST RECEIVED")
		log.Println("=================================================================")

		claims, err := jwe.ParseAndValidateJWE(request.Token)
		if err != nil {
			log.Printf("[ERROR] Invalid token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token: "+err.Error())
		}

		log.Printf("[INFO] Sending notification for user ID: %s\n", claims.UserID)
		log.Printf("[INFO] Title: %s\n", request.Title)
		log.Printf("[INFO] Content: %s\n", request.Content)

		notificationResponse, err := sendInboxNotification(claims.AccessToken, request.Title, request.Content, request.Url)
		if err != nil {
			log.Printf("[ERROR] Failed to send notification: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to send notification: "+err.Error())
		}

		response := buildNotificationResponse(notificationResponse)

		log.Println("[SUCCESS] Returning notification response to frontend")
		log.Println("=================================================================")
		return ctx.JSON(response)
	})

	// POST /api/notification/send-push
	group.Post("/notification/send-push", func(ctx *fiber.Ctx) error {
		var request sendPushRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("SEND PUSH NOTIFICATION REQUEST RECEIVED")
		log.Println("=================================================================")

		claims, err := jwe.ParseAndValidateJWE(request.Token)
		if err != nil {
			log.Printf("[ERROR] Invalid token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token: "+err.Error())
		}

		log.Printf("[INFO] Sending push notification for user ID: %s\n", claims.UserID)
		log.Printf("[INFO] Title: %s\n", request.Title)
		log.Printf("[INFO] Content: %s\n", request.Content)

		pushResponse, err := sendPushNotification(claims.AccessToken, request.Title, request.Content, request.Url)
		if err != nil {
			log.Printf("[ERROR] Failed to send push notification: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to send push notification: "+err.Error())
		}

		response := buildPushNotificationResponse(pushResponse)

		log.Println("[SUCCESS] Returning push notification response to frontend")
		log.Println("=================================================================")
		return ctx.JSON(response)
	})
}

func sendInboxNotification(accessToken, title, content, url string) (alipay.SendInboxResponse, error) {
	log.Println("=================================================================")
	log.Println("PROCESSING INBOX NOTIFICATION")
	log.Println("=================================================================")

	requestID := generateNotificationRequestID()
	log.Printf("[INFO] Generated Request ID: %s\n", requestID)

	if url == "" {
		url = "mini://platformapi/startapp?_ariver_appid=888888"
	}

	templateParams := map[string]string{
		"Title":   title,
		"Content": content,
		"Url":     url,
	}

	notificationRequest := alipay.SendInboxRequest{
		AccessToken:  accessToken,
		RequestID:    requestID,
		TemplateCode: "MINI_APP_COMMON_INBOX",
		Templates: []alipay.InboxTemplate{
			{
				TemplateParameters: templateParams,
			},
		},
	}

	requestJSON, _ := json.MarshalIndent(notificationRequest, "", "  ")
	log.Printf("[INFO] Notification request details:\n%s\n\n", string(requestJSON))

	log.Println("[INFO] Calling Alipay SendInbox API...")
	notificationResponse, err := alipay.Interface.SendInbox(notificationRequest)
	if err != nil {
		log.Printf("[ERROR] SendInbox API call failed: %v\n", err)
		return alipay.SendInboxResponse{}, fmt.Errorf("SendInbox API call failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(notificationResponse, "", "  ")
	log.Printf("[SUCCESS] SendInbox API response:\n%s\n\n", string(responseJSON))

	switch notificationResponse.Result.ResultStatus {
	case "S":
		log.Println("[SUCCESS] Notification sent successfully")
		if notificationResponse.MessageID != "" {
			log.Printf("[INFO] Message ID: %s\n", notificationResponse.MessageID)
		}

	case "A":
		log.Println("[SUCCESS] Notification accepted by wallet")

	case "U":
		log.Println("[WARNING] Notification status unknown")

	case "F":
		log.Printf("[ERROR] Notification failed: %s\n", notificationResponse.Result.ResultMessage)
		log.Printf("[ERROR] Error Code: %s\n", notificationResponse.Result.ResultCode)
	}

	log.Println("=================================================================")
	log.Println("NOTIFICATION PROCESSING COMPLETED")
	log.Println("=================================================================")

	return notificationResponse, nil
}

func buildNotificationResponse(notificationResponse alipay.SendInboxResponse) fiber.Map {
	response := fiber.Map{
		"resultStatus":  notificationResponse.Result.ResultStatus,
		"resultCode":    notificationResponse.Result.ResultCode,
		"resultMessage": notificationResponse.Result.ResultMessage,
	}

	switch notificationResponse.Result.ResultStatus {
	case "S", "A":
		response["status"] = "SUCCESS"
		response["success"] = true
		if notificationResponse.MessageID != "" {
			response["messageId"] = notificationResponse.MessageID
		}
		if notificationResponse.ExtendInfo != "" {
			response["extendInfo"] = notificationResponse.ExtendInfo
		}

	case "U":
		response["status"] = "UNKNOWN"
		response["success"] = false
		response["message"] = "Notification status is unknown. It may still be processed."

	case "F":
		response["status"] = "FAILED"
		response["success"] = false
		response["message"] = notificationResponse.Result.ResultMessage
	}

	return response
}

func sendPushNotification(accessToken, title, content, url string) (alipay.SendPushResponse, error) {
	log.Println("=================================================================")
	log.Println("PROCESSING PUSH NOTIFICATION")
	log.Println("=================================================================")

	requestID := generateNotificationRequestID()
	log.Printf("[INFO] Generated Request ID: %s\n", requestID)

	if url == "" {
		url = "mini://platformapi/startapp?_ariver_appid=888888"
	}

	templateParams := map[string]string{
		"Title":   title,
		"Content": content,
		"Url":     url,
	}

	pushRequest := alipay.SendPushRequest{
		AccessToken:  accessToken,
		RequestID:    requestID,
		TemplateCode: "MINI_APP_COMMON_PUSH",
		Templates: []alipay.PushTemplate{
			{
				TemplateParameters: templateParams,
			},
		},
	}

	requestJSON, _ := json.MarshalIndent(pushRequest, "", "  ")
	log.Printf("[INFO] Push notification request details:\n%s\n\n", string(requestJSON))

	log.Println("[INFO] Calling Alipay SendPush API...")
	pushResponse, err := alipay.Interface.SendPush(pushRequest)
	if err != nil {
		log.Printf("[ERROR] SendPush API call failed: %v\n", err)
		return alipay.SendPushResponse{}, fmt.Errorf("SendPush API call failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(pushResponse, "", "  ")
	log.Printf("[SUCCESS] SendPush API response:\n%s\n\n", string(responseJSON))

	switch pushResponse.Result.ResultStatus {
	case "S":
		log.Println("[SUCCESS] Push notification sent successfully")
		if pushResponse.MessageID != "" {
			log.Printf("[INFO] Message ID: %s\n", pushResponse.MessageID)
		}

	case "A":
		log.Println("[SUCCESS] Push notification accepted by wallet")

	case "U":
		log.Println("[WARNING] Push notification status unknown")

	case "F":
		log.Printf("[ERROR] Push notification failed: %s\n", pushResponse.Result.ResultMessage)
		log.Printf("[ERROR] Error Code: %s\n", pushResponse.Result.ResultCode)
	}

	log.Println("=================================================================")
	log.Println("PUSH NOTIFICATION PROCESSING COMPLETED")
	log.Println("=================================================================")

	return pushResponse, nil
}

func buildPushNotificationResponse(pushResponse alipay.SendPushResponse) fiber.Map {
	response := fiber.Map{
		"resultStatus":  pushResponse.Result.ResultStatus,
		"resultCode":    pushResponse.Result.ResultCode,
		"resultMessage": pushResponse.Result.ResultMessage,
	}

	switch pushResponse.Result.ResultStatus {
	case "S", "A":
		response["status"] = "SUCCESS"
		response["success"] = true
		if pushResponse.MessageID != "" {
			response["messageId"] = pushResponse.MessageID
		}
		if pushResponse.ExtendInfo != "" {
			response["extendInfo"] = pushResponse.ExtendInfo
		}

	case "U":
		response["status"] = "UNKNOWN"
		response["success"] = false
		response["message"] = "Push notification status is unknown. It may still be processed."

	case "F":
		response["status"] = "FAILED"
		response["success"] = false
		response["message"] = pushResponse.Result.ResultMessage
	}

	return response
}

func generateNotificationRequestID() string {
	return fmt.Sprintf("NOTIF-%s-%d", uuid.New().String(), time.Now().Unix())
}
