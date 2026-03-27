package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"superQiMiniAppBackend/alipay"
	"superQiMiniAppBackend/jwe"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type createPaymentRequest struct {
	Token string `json:"token" validate:"required"`
}

func InitPaymentEndpoint(group fiber.Router) {
	group.Post("/payment/create", func(ctx *fiber.Ctx) error {
		var request createPaymentRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("PAYMENT CREATION REQUEST RECEIVED")
		log.Println("=================================================================")

		claims, err := jwe.ParseAndValidateJWE(request.Token)
		if err != nil {
			log.Printf("[ERROR] Invalid token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token: "+err.Error())
		}

		log.Printf("[INFO] Creating payment for user ID: %s\n", claims.UserID)

		paymentResponse, err := createTestPayment(claims.UserID)
		if err != nil {
			log.Printf("[ERROR] Failed to create payment: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create payment: "+err.Error())
		}

		response := fiber.Map{
			"success": true,
			"amount":  1,
		}

		if paymentResponse.GetRedirectURL() != "" {
			response["paymentUrl"] = paymentResponse.GetRedirectURL()
			response["paymentId"] = paymentResponse.PaymentID
			log.Printf("[INFO] Sending payment URL to frontend: %s\n", paymentResponse.GetRedirectURL())

			// Start background polling for payment status
			if paymentResponse.PaymentID != "" {
				log.Printf("[INFO] Starting background polling for payment: %s\n", paymentResponse.PaymentID)
				StartPaymentPolling(paymentResponse.PaymentID, paymentResponse.PaymentRequestID)
			}
		} else {
			log.Println("[WARNING] No payment URL in response")
			response["success"] = false
			response["error"] = "No redirect URL received from payment API"
		}

		log.Println("[SUCCESS] Returning payment response to frontend")
		return ctx.JSON(response)
	})

	// GET /api/payment/status/:paymentId - Check payment status from cache
	group.Get("/payment/status/:paymentId", func(ctx *fiber.Ctx) error {
		paymentId := ctx.Params("paymentId")

		log.Printf("[INFO] Status check request for payment: %s\n", paymentId)

		// Get status from store
		status, exists := paymentStore.Get(paymentId)
		if !exists {
			log.Printf("[WARNING] Payment %s not found in cache\n", paymentId)
			return ctx.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"success": false,
				"message": "Payment not found in cache. It may be too old or was never tracked.",
			})
		}

		log.Printf("[INFO] Payment %s status: %s (completed: %v)\n", paymentId, status.Status, status.Completed)

		return ctx.JSON(fiber.Map{
			"success":          true,
			"paymentId":        status.PaymentID,
			"paymentRequestId": status.PaymentRequestID,
			"status":           status.Status,
			"paymentStatus":    status.PaymentStatus,
			"completed":        status.Completed,
			"message":          status.Message,
			"lastChecked":      status.LastChecked,
		})
	})
}

func createTestPayment(userID string) (alipay.PaymentResponse, error) {
	log.Println("=================================================================")
	log.Printf("CREATING TEST PAYMENT FOR USER: %s\n", userID)
	log.Println("=================================================================")

	paymentRequestID := fmt.Sprintf("PAY-%s-%d", uuid.New().String(), time.Now().Unix())

	expiryTime := time.Now().Add(30 * time.Minute).Format("2006-01-02T15:04:05-07:00")

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1999"
	}

	// Frontend URL for redirect after payment
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://172.20.10.2:5173" // Default frontend URL
	}

	paymentRequest := alipay.PaymentRequest{
		ProductCode:      alipay.ONLINE_PURCHASE,
		PaymentRequestID: paymentRequestID,
		PaymentAmount: alipay.PaymentAmount{
			Currency: "IQD",
			Value:    "1000",
		},
		Order: alipay.Order{
			OrderDescription: "Test Order - Online Purchase",
			Buyer: alipay.OrderBuyer{
				ReferenceBuyerID: userID,
			},
		},
		PaymentExpiryTime:  expiryTime,
		PaymentRedirectURL: frontendURL + "/payment-success.html",
		// Public URL for payment notifications code should be set here
	}

	log.Println("[INFO] Payment request details:")
	requestJSON, _ := json.MarshalIndent(paymentRequest, "", "  ")
	log.Printf("%s\n\n", string(requestJSON))

	log.Println("[INFO] Calling payment API...")
	paymentResponse, err := alipay.Interface.Pay(paymentRequest)
	if err != nil {
		log.Printf("[ERROR] Payment API error: %v\n", err)
		return alipay.PaymentResponse{}, err
	}

	responseJSON, _ := json.MarshalIndent(paymentResponse, "", "  ")
	log.Printf("[SUCCESS] Payment API response received:\n%s\n\n", string(responseJSON))

	redirectURL := paymentResponse.GetRedirectURL()

	if paymentResponse.Result.ResultStatus == "A" {
		log.Println("[SUCCESS] Payment accepted")
		if redirectURL != "" {
			log.Printf("[INFO] Redirection URL: %s\n", redirectURL)
			log.Println("[INFO] Frontend should call my.tradePay() with this URL")
		} else {
			log.Println("[WARNING] Redirection URL is empty in response")
		}
		log.Printf("[INFO] Payment ID: %s\n", paymentResponse.PaymentID)
		log.Printf("[INFO] Payment Request ID: %s\n", paymentResponse.PaymentRequestID)
	} else if paymentResponse.Result.ResultStatus == "S" {
		log.Println("[SUCCESS] Payment completed immediately")
	} else if paymentResponse.Result.ResultStatus == "U" {
		log.Println("[WARNING] Unknown payment status - need to query later")
	} else {
		log.Printf("[ERROR] Payment failed: %s\n", paymentResponse.Result.ResultMessage)
	}

	log.Println("=================================================================")
	log.Println("TEST PAYMENT CREATION COMPLETED")
	log.Println("=================================================================")

	return paymentResponse, nil
}
