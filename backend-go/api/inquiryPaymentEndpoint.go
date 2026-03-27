package api

import (
	"encoding/json"
	"log"
	"superQiMiniAppBackend/alipay"

	"github.com/gofiber/fiber/v2"
)

type inquiryPaymentRequest struct {
	PaymentID        string `json:"paymentId,omitempty"`
	PaymentRequestID string `json:"paymentRequestId,omitempty"`
}

func InitInquiryPaymentEndpoint(group fiber.Router) {
	group.Post("/payment/inquiry", handleInquiryPayment)
}

func handleInquiryPayment(ctx *fiber.Ctx) error {
	var request inquiryPaymentRequest
	if err := ctx.BodyParser(&request); err != nil {
		log.Printf("[ERROR] Invalid request body: %v\n", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	log.Println("=================================================================")
	log.Println("PAYMENT INQUIRY REQUEST RECEIVED")
	log.Println("=================================================================")

	// Validate that at least one identifier is provided
	if request.PaymentID == "" && request.PaymentRequestID == "" {
		log.Println("[ERROR] Either paymentId or paymentRequestId must be provided")
		return fiber.NewError(fiber.StatusBadRequest, "Either paymentId or paymentRequestId is required")
	}

	log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)
	log.Printf("[INFO] Payment Request ID: %s\n", request.PaymentRequestID)

	// Call Alipay InquiryPayment API
	inquiryRequest := alipay.InquiryPaymentRequest{
		PaymentID:        request.PaymentID,
		PaymentRequestID: request.PaymentRequestID,
	}

	inquiryResponse, err := alipay.Interface.InquiryPayment(inquiryRequest)
	if err != nil {
		log.Printf("[ERROR] Failed to inquiry payment: %v\n", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to inquiry payment: "+err.Error())
	}

	responseJSON, _ := json.MarshalIndent(inquiryResponse, "", "  ")
	log.Printf("[INFO] Inquiry response:\n%s\n", string(responseJSON))

	// Build response
	response := buildInquiryPaymentResponse(inquiryResponse)

	log.Println("[SUCCESS] Returning inquiry response to frontend")
	log.Println("=================================================================")
	return ctx.JSON(response)
}

func buildInquiryPaymentResponse(inquiryResponse alipay.InquiryPaymentResponse) fiber.Map {
	response := fiber.Map{
		"resultStatus":  inquiryResponse.Result.ResultStatus,
		"resultCode":    inquiryResponse.Result.ResultCode,
		"resultMessage": inquiryResponse.Result.ResultMessage,
	}

	// Handle result status
	switch inquiryResponse.Result.ResultStatus {
	case "S":
		// Success - inquiry was successful, check payment status
		response["success"] = true
		response["paymentStatus"] = inquiryResponse.PaymentStatus

		// Add payment details
		if inquiryResponse.PaymentID != "" {
			response["paymentId"] = inquiryResponse.PaymentID
		}
		if inquiryResponse.PaymentRequestID != "" {
			response["paymentRequestId"] = inquiryResponse.PaymentRequestID
		}
		if inquiryResponse.PaymentTime != "" {
			response["paymentTime"] = inquiryResponse.PaymentTime
		}
		if inquiryResponse.PaymentAmount.Value != "" {
			response["paymentAmount"] = fiber.Map{
				"value":    inquiryResponse.PaymentAmount.Value,
				"currency": inquiryResponse.PaymentAmount.Currency,
			}
		}
		if len(inquiryResponse.Transactions) > 0 {
			response["transactions"] = inquiryResponse.Transactions
		}
		if inquiryResponse.ExtendInfo != "" {
			response["extendInfo"] = inquiryResponse.ExtendInfo
		}

		// Provide user-friendly status message based on paymentStatus
		switch inquiryResponse.PaymentStatus {
		case "SUCCESS":
			response["statusMessage"] = "Payment completed successfully"
			response["paymentCompleted"] = true
		case "AUTH_SUCCESS":
			response["statusMessage"] = "Payment authorized but not finished"
			response["paymentCompleted"] = false
		case "PROCESSING":
			response["statusMessage"] = "Payment is still processing"
			response["paymentCompleted"] = false
		case "FAIL":
			response["statusMessage"] = "Payment failed"
			response["paymentCompleted"] = false
		default:
			response["statusMessage"] = "Unknown payment status: " + inquiryResponse.PaymentStatus
			response["paymentCompleted"] = false
		}

	case "U":
		// Unknown exception - retry
		response["success"] = false
		response["statusMessage"] = "Unknown exception occurred. Please retry."
		response["shouldRetry"] = true

	case "F":
		// Failed
		response["success"] = false
		response["shouldRetry"] = false

		// Check if order doesn't exist (special case)
		if inquiryResponse.Result.ResultCode == "ORDER_NOT_EXIST" {
			response["statusMessage"] = "Payment not found or not yet accepted. This can be treated as payment failure."
			response["paymentCompleted"] = false
		} else {
			response["statusMessage"] = inquiryResponse.Result.ResultMessage
		}

	default:
		response["success"] = false
		response["statusMessage"] = "Unexpected result status: " + inquiryResponse.Result.ResultStatus
	}

	return response
}
