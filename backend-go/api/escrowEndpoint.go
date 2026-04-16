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

type createEscrowPaymentRequest struct {
	Token string `json:"token" validate:"required"`
}

type escrowActionRequest struct {
	PaymentID string `json:"paymentId" validate:"required"`
}

func InitEscrowEndpoint(group fiber.Router) {
	// POST /api/escrow/create - Create escrow payment
	group.Post("/escrow/create", func(ctx *fiber.Ctx) error {
		var request createEscrowPaymentRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("ESCROW PAYMENT CREATION REQUEST RECEIVED")
		log.Println("=================================================================")

		claims, err := jwe.ParseAndValidateJWE(request.Token)
		if err != nil {
			log.Printf("[ERROR] Invalid token: %v\n", err)
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid token: "+err.Error())
		}

		log.Printf("[INFO] Creating escrow payment for user ID: %s\n", claims.UserID)

		paymentResponse, err := createEscrowPayment(claims.UserID)
		if err != nil {
			log.Printf("[ERROR] Failed to create escrow payment: %v\n", err)
			return fiber.NewError(fiber.StatusInternalServerError, "Failed to create escrow payment: "+err.Error())
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

		log.Println("[SUCCESS] Returning escrow payment response to frontend")
		return ctx.JSON(response)
	})

	// POST /api/escrow/merchant-accept - Merchant accepts escrow payment
	group.Post("/escrow/merchant-accept", func(ctx *fiber.Ctx) error {
		var request escrowActionRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("MERCHANT ACCEPT REQUEST RECEIVED")
		log.Println("=================================================================")
		log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)

		if request.PaymentID == "" {
			log.Println("[ERROR] Payment ID is required")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": "Payment ID is required",
			})
		}

		merchantAcceptRequest := alipay.MerchantAcceptRequest{
			PaymentID: request.PaymentID,
		}

		merchantAcceptResponse, err := alipay.Interface.MerchantAccept(merchantAcceptRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to merchant accept: %v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": err.Error(),
			})
		}

		response := fiber.Map{
			"resultStatus":  merchantAcceptResponse.Result.ResultStatus,
			"resultCode":    merchantAcceptResponse.Result.ResultCode,
			"resultMessage": merchantAcceptResponse.Result.ResultMessage,
		}

		if merchantAcceptResponse.Result.ResultStatus == "S" {
			response["success"] = true
			response["paymentId"] = merchantAcceptResponse.PaymentID
			log.Println("[SUCCESS] Merchant accept successful")
		} else {
			response["success"] = false
			log.Printf("[ERROR] Merchant accept failed: %s\n", merchantAcceptResponse.Result.ResultMessage)
		}

		log.Println("=================================================================")
		return ctx.JSON(response)
	})

	// POST /api/escrow/confirm - Confirm escrow order (customer confirms)
	group.Post("/escrow/confirm", func(ctx *fiber.Ctx) error {
		var request escrowActionRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("CONFIRM ORDER REQUEST RECEIVED")
		log.Println("=================================================================")
		log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)

		if request.PaymentID == "" {
			log.Println("[ERROR] Payment ID is required")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": "Payment ID is required",
			})
		}

		// Generate unique confirm request ID
		confirmRequestID := fmt.Sprintf("CONFIRM-%s-%d", uuid.New().String(), time.Now().Unix())
		log.Printf("[INFO] Generated Confirm Request ID: %s\n", confirmRequestID)

		confirmOrderRequest := alipay.ConfirmOrderRequest{
			PaymentID:        request.PaymentID,
			ConfirmRequestID: confirmRequestID,
		}

		confirmOrderResponse, err := alipay.Interface.ConfirmOrder(confirmOrderRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to confirm order: %v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": err.Error(),
			})
		}

		response := fiber.Map{
			"resultStatus":  confirmOrderResponse.Result.ResultStatus,
			"resultCode":    confirmOrderResponse.Result.ResultCode,
			"resultMessage": confirmOrderResponse.Result.ResultMessage,
		}

		if confirmOrderResponse.Result.ResultStatus == "S" {
			response["success"] = true
			response["confirmId"] = confirmOrderResponse.ConfirmID
			response["confirmTime"] = confirmOrderResponse.ConfirmTime
			log.Println("[SUCCESS] Confirm order successful")
			log.Printf("[INFO] Confirm ID: %s\n", confirmOrderResponse.ConfirmID)
			log.Printf("[INFO] Confirm Time: %s\n", confirmOrderResponse.ConfirmTime)
		} else {
			response["success"] = false
			log.Printf("[ERROR] Confirm order failed: %s\n", confirmOrderResponse.Result.ResultMessage)
		}

		log.Println("=================================================================")
		return ctx.JSON(response)
	})

	// POST /api/escrow/cancel - Cancel escrow payment
	group.Post("/escrow/cancel", func(ctx *fiber.Ctx) error {
		var request escrowActionRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("CANCEL PAYMENT REQUEST RECEIVED")
		log.Println("=================================================================")
		log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)

		if request.PaymentID == "" {
			log.Println("[ERROR] Payment ID is required")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": "Payment ID is required",
			})
		}

		cancelPaymentRequest := alipay.CancelPaymentRequest{
			PaymentID: request.PaymentID,
		}

		cancelPaymentResponse, err := alipay.Interface.CancelPayment(cancelPaymentRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to cancel payment: %v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": err.Error(),
			})
		}

		response := fiber.Map{
			"resultStatus":  cancelPaymentResponse.Result.ResultStatus,
			"resultCode":    cancelPaymentResponse.Result.ResultCode,
			"resultMessage": cancelPaymentResponse.Result.ResultMessage,
		}

		if cancelPaymentResponse.Result.ResultStatus == "S" {
			response["success"] = true
			response["paymentId"] = cancelPaymentResponse.PaymentID
			log.Println("[SUCCESS] Cancel payment successful")
		} else {
			response["success"] = false
			log.Printf("[ERROR] Cancel payment failed: %s\n", cancelPaymentResponse.Result.ResultMessage)
		}

		log.Println("=================================================================")
		return ctx.JSON(response)
	})

	// POST /api/escrow/void - Void escrow payment
	group.Post("/escrow/void", func(ctx *fiber.Ctx) error {
		var request escrowActionRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("VOID PAYMENT REQUEST RECEIVED")
		log.Println("=================================================================")
		log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)

		if request.PaymentID == "" {
			log.Println("[ERROR] Payment ID is required")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": "Payment ID is required",
			})
		}

		// Generate unique void request ID
		voidRequestID := fmt.Sprintf("VOID-%s-%d", uuid.New().String(), time.Now().Unix())
		log.Printf("[INFO] Generated Void Request ID: %s\n", voidRequestID)

		voidRequest := alipay.VoidRequest{
			PaymentID:     request.PaymentID,
			VoidRequestID: voidRequestID,
		}

		voidResponse, err := alipay.Interface.Void(voidRequest)
		if err != nil {
			log.Printf("[ERROR] Failed to void payment: %v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": err.Error(),
			})
		}

		response := fiber.Map{
			"resultStatus":  voidResponse.Result.ResultStatus,
			"resultCode":    voidResponse.Result.ResultCode,
			"resultMessage": voidResponse.Result.ResultMessage,
		}

		if voidResponse.Result.ResultStatus == "S" {
			response["success"] = true
			response["voidId"] = voidResponse.VoidID
			response["voidTime"] = voidResponse.VoidTime
			log.Println("[SUCCESS] Void payment successful")
			log.Printf("[INFO] Void ID: %s\n", voidResponse.VoidID)
			log.Printf("[INFO] Void Time: %s\n", voidResponse.VoidTime)
		} else {
			response["success"] = false
			log.Printf("[ERROR] Void payment failed: %s\n", voidResponse.Result.ResultMessage)
		}

		log.Println("=================================================================")
		return ctx.JSON(response)
	})
}

func createEscrowPayment(userID string) (alipay.PaymentResponse, error) {
	log.Println("=================================================================")
	log.Printf("CREATING ESCROW PAYMENT FOR USER: %s\n", userID)
	log.Println("=================================================================")

	paymentRequestID := fmt.Sprintf("ESCROW-PAY-%s-%d", uuid.New().String(), time.Now().Unix())

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
		ProductCode:      alipay.ESCROW_PAYMENT,
		PaymentRequestID: paymentRequestID,
		PaymentAmount: alipay.PaymentAmount{
			Currency: "IQD",
			Value:    "1000",
		},
		Order: alipay.Order{
			OrderDescription: "Escrow Test Order - Payment held until merchant accepts",
			Buyer: alipay.OrderBuyer{
				ReferenceBuyerID: userID,
			},
		},
		PaymentExpiryTime:  expiryTime,
		PaymentRedirectURL: frontendURL + "/payment-success.html",
	}

	log.Println("[INFO] Escrow payment request details:")
	requestJSON, _ := json.MarshalIndent(paymentRequest, "", "  ")
	log.Printf("%s\n\n", string(requestJSON))

	log.Println("[INFO] Calling payment API with ESCROW_PAYMENT product code...")
	paymentResponse, err := alipay.Interface.Pay(paymentRequest)
	if err != nil {
		log.Printf("[ERROR] Payment API error: %v\n", err)
		return alipay.PaymentResponse{}, err
	}

	responseJSON, _ := json.MarshalIndent(paymentResponse, "", "  ")
	log.Printf("[SUCCESS] Payment API response received:\n%s\n\n", string(responseJSON))

	redirectURL := paymentResponse.GetRedirectURL()

	if paymentResponse.Result.ResultStatus == "A" {
		log.Println("[SUCCESS] Escrow payment accepted")
		if redirectURL != "" {
			log.Printf("[INFO] Redirection URL: %s\n", redirectURL)
			log.Println("[INFO] Frontend should call my.tradePay() with this URL")
		} else {
			log.Println("[WARNING] Redirection URL is empty in response")
		}
		log.Printf("[INFO] Payment ID: %s\n", paymentResponse.PaymentID)
		log.Printf("[INFO] Payment Request ID: %s\n", paymentResponse.PaymentRequestID)
		log.Println("[INFO] Payment will be held in escrow until merchant accepts")
	} else if paymentResponse.Result.ResultStatus == "S" {
		log.Println("[SUCCESS] Payment completed immediately")
	} else if paymentResponse.Result.ResultStatus == "U" {
		log.Println("[WARNING] Unknown payment status - need to query later")
	} else {
		log.Printf("[ERROR] Payment failed: %s\n", paymentResponse.Result.ResultMessage)
	}

	log.Println("=================================================================")
	log.Println("ESCROW PAYMENT CREATION COMPLETED")
	log.Println("=================================================================")

	return paymentResponse, nil
}
