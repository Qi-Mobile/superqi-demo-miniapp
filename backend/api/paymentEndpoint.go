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
	Token      string                 `json:"token" validate:"required"`
	ProductID  string                 `json:"productId,omitempty"`
	Quantity   int                    `json:"quantity,omitempty"`
	OrderID    string                 `json:"orderId,omitempty"`
	CustomData map[string]interface{} `json:"customData,omitempty"`
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

		paymentResponse, err := createTestPayment(claims.UserID, request)
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
		} else {
			log.Println("[WARNING] No payment URL in response")
			response["success"] = false
			response["error"] = "No redirect URL received from payment API"
		}

		log.Println("[SUCCESS] Returning payment response to frontend")
		return ctx.JSON(response)
	})

	group.Post("/payment/inquiry", func(ctx *fiber.Ctx) error {
		var request struct {
			PaymentID        string `json:"paymentId,omitempty"`
			PaymentRequestID string `json:"paymentRequestId,omitempty"`
		}

		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		if request.PaymentID == "" && request.PaymentRequestID == "" {
			return fiber.NewError(fiber.StatusBadRequest, "Either paymentId or paymentRequestId is required")
		}

		log.Println("=================================================================")
		log.Println("PAYMENT INQUIRY REQUEST RECEIVED")
		log.Println("=================================================================")
		log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)
		log.Printf("[INFO] Payment Request ID: %s\n", request.PaymentRequestID)

		inquiryRequest := alipay.InquiryPaymentRequest{
			PaymentID:        request.PaymentID,
			PaymentRequestID: request.PaymentRequestID,
		}

		inquiryResponse, err := alipay.Interface.InquiryPayment(inquiryRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to inquiry payment: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to inquiry payment: "+err.Error())
		}

		log.Println("=================================================================")
		log.Println("PAYMENT INQUIRY RESPONSE")
		log.Println("=================================================================")
		responseJSON, _ := json.MarshalIndent(inquiryResponse, "", "  ")
		log.Printf("%s\n", string(responseJSON))

		if inquiryResponse.ExtendInfo != "" {
			log.Println("=================================================================")
			log.Println("EXTENDINFO DETAILS")
			log.Println("=================================================================")
			extendInfoData, err := alipay.ParseExtendInfoMap(inquiryResponse.ExtendInfo)
			if err != nil {
				log.Printf("[WARNING] Failed to parse extendInfo: %v\n", err)
				log.Printf("[INFO] Raw extendInfo: %s\n", inquiryResponse.ExtendInfo)
			} else {
				extendInfoJSON, _ := json.MarshalIndent(extendInfoData, "", "  ")
				log.Printf("Parsed ExtendInfo:\n%s\n", string(extendInfoJSON))

				if productID, ok := extendInfoData["productId"].(string); ok {
					log.Printf("[TRACKING] Product ID: %s\n", productID)
				}
				if quantity, ok := extendInfoData["quantity"].(float64); ok {
					log.Printf("[TRACKING] Quantity: %.0f\n", quantity)
				}
				if orderID, ok := extendInfoData["orderId"].(string); ok {
					log.Printf("[TRACKING] Order ID: %s\n", orderID)
				}
			}
			log.Println("=================================================================")
		} else {
			log.Println("[INFO] No extendInfo in inquiry response")
		}

		response := fiber.Map{
			"success":          true,
			"paymentId":        inquiryResponse.PaymentID,
			"paymentRequestId": inquiryResponse.PaymentRequestID,
			"paymentStatus":    inquiryResponse.PaymentStatus,
			"paymentTime":      inquiryResponse.PaymentTime,
			"paymentAmount":    inquiryResponse.PaymentAmount,
			"extendInfo":       inquiryResponse.ExtendInfo,
		}

		return ctx.JSON(response)
	})
}

func createTestPayment(userID string, request createPaymentRequest) (alipay.PaymentResponse, error) {
	log.Println("=================================================================")
	log.Printf("CREATING TEST PAYMENT FOR USER: %s\n", userID)
	log.Println("=================================================================")

	paymentRequestID := fmt.Sprintf("PAY-%s-%d", uuid.New().String(), time.Now().Unix())

	expiryTime := time.Now().Add(30 * time.Minute).Format("2006-01-02T15:04:05-07:00")

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1999"
	}

	extendInfoMap := map[string]interface{}{
		"paymentRequestId": paymentRequestID,
		"userId":           userID,
		"timestamp":        time.Now().Unix(),
	}

	if request.ProductID != "" {
		extendInfoMap["productId"] = request.ProductID
	}
	if request.Quantity > 0 {
		extendInfoMap["quantity"] = request.Quantity
	}
	if request.OrderID != "" {
		extendInfoMap["orderId"] = request.OrderID
	}

	if request.CustomData != nil {
		for key, value := range request.CustomData {
			extendInfoMap[key] = value
		}
	}

	extendInfoJSON, err := json.Marshal(extendInfoMap)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal extendInfo: %v\n", err)
		return alipay.PaymentResponse{}, err
	}

	log.Printf("[INFO] ExtendInfo: %s\n", string(extendInfoJSON))

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
		PaymentRedirectURL: baseURL + "/payment-success.html",
		ExtendInfo:         string(extendInfoJSON),
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
		if paymentResponse.ExtendInfo != "" {
			log.Printf("[INFO] ExtendInfo returned in response: %s\n", paymentResponse.ExtendInfo)
		} else {
			log.Println("[INFO] ExtendInfo not returned in response (this is expected)")
		}
	} else if paymentResponse.Result.ResultStatus == "S" {
		log.Println("[SUCCESS] Payment completed immediately")
		if paymentResponse.ExtendInfo != "" {
			log.Printf("[INFO] ExtendInfo returned: %s\n", paymentResponse.ExtendInfo)
		}
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
