package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"superQiMiniAppBackend/alipay"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type refundRequest struct {
	PaymentID string  `json:"paymentId" validate:"required"`
	Amount    float64 `json:"amount" validate:"required"`
}

func InitRefundEndpoint(group fiber.Router) {
	group.Post("/payment/refund", func(ctx *fiber.Ctx) error {
		var request refundRequest
		if err := ctx.BodyParser(&request); err != nil {
			log.Printf("[ERROR] Invalid refund request body: %v\n", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
		}

		log.Println("=================================================================")
		log.Println("REFUND REQUEST RECEIVED")
		log.Println("=================================================================")
		log.Printf("[INFO] Payment ID: %s\n", request.PaymentID)
		log.Printf("[INFO] Refund Amount (IQD): %.2f\n", request.Amount)

		// Validate inputs
		if request.PaymentID == "" {
			log.Println("[ERROR] Payment ID is required")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": "Payment ID is required",
			})
		}

		if request.Amount <= 0 {
			log.Println("[ERROR] Invalid refund amount")
			return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": "Refund amount must be greater than 0",
			})
		}

		refundResponse, err := processRefund(request.PaymentID, request.Amount)
		if err != nil {
			log.Printf("[ERROR] Failed to process refund: %v\n", err)
			return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success":       false,
				"resultStatus":  "F",
				"resultMessage": err.Error(),
			})
		}

		response := buildRefundResponse(refundResponse)

		log.Println("[SUCCESS] Returning refund response to frontend")
		log.Println("=================================================================")
		return ctx.JSON(response)
	})
}

func processRefund(paymentID string, amountIQD float64) (alipay.RefundResponse, error) {
	log.Println("=================================================================")
	log.Printf("PROCESSING REFUND FOR PAYMENT: %s\n", paymentID)
	log.Println("=================================================================")

	refundRequestID := generateRefundRequestID()
	log.Printf("[INFO] Generated Refund Request ID: %s\n", refundRequestID)

	amountInFils := int64(amountIQD * 1000)
	log.Printf("[INFO] Amount in fils: %d\n", amountInFils)

	refundRequest := alipay.RefundRequest{
		RefundRequestID: refundRequestID,
		PaymentID:       paymentID,
		RefundAmount: alipay.RefundAmount{
			Currency: "IQD",
			Value:    strconv.FormatInt(amountInFils, 10),
		},
		RefundReason: "Customer requested refund from mini app",
	}

	requestJSON, _ := json.MarshalIndent(refundRequest, "", "  ")
	log.Printf("[INFO] Refund request details:\n%s\n\n", string(requestJSON))

	log.Println("[INFO] Calling Alipay refund API...")
	refundResponse, err := alipay.Interface.Refund(refundRequest)
	if err != nil {
		log.Printf("[ERROR] Refund API call failed: %v\n", err)
		return alipay.RefundResponse{}, fmt.Errorf("refund API call failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(refundResponse, "", "  ")
	log.Printf("[SUCCESS] Refund API response:\n%s\n\n", string(responseJSON))

	switch refundResponse.Result.ResultStatus {
	case "S":
		log.Println("[SUCCESS]  Refund successful immediately")
		log.Printf("[INFO] Refund ID: %s\n", refundResponse.RefundID)
		log.Printf("[INFO] Refund Time: %s\n", refundResponse.RefundTime)

	case "U":
		log.Println("[WARNING] Refund status unknown - starting polling...")
		finalResponse := pollRefundStatus(refundRequestID)
		if finalResponse != nil {
			return *finalResponse, nil
		}
		log.Println("[WARNING] Polling completed but status still unknown")

	case "F":
		log.Printf("[ERROR] Refund failed: %s\n", refundResponse.Result.ResultMessage)
		log.Printf("[ERROR] Error Code: %s\n", refundResponse.Result.ResultCode)
	}

	log.Println("=================================================================")
	log.Println("REFUND PROCESSING COMPLETED")
	log.Println("=================================================================")

	return refundResponse, nil
}

func pollRefundStatus(refundRequestID string) *alipay.RefundResponse {
	const maxAttempts = 12
	const intervalSeconds = 5

	log.Println("=================================================================")
	log.Printf("STARTING REFUND STATUS POLLING FOR: %s\n", refundRequestID)
	log.Println("=================================================================")
	log.Printf("[INFO] Max attempts: %d, Interval: %d seconds\n", maxAttempts, intervalSeconds)

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		log.Printf("[INFO] Polling attempt %d/%d...\n", attempt, maxAttempts)

		inquiryRequest := alipay.InquiryRefundRequest{
			RefundRequestID: refundRequestID,
		}

		inquiryResponse, err := alipay.Interface.InquiryRefund(inquiryRequest)
		if err != nil {
			log.Printf("[ERROR] Inquiry attempt %d failed: %v\n", attempt, err)
			time.Sleep(intervalSeconds * time.Second)
			continue
		}

		inquiryJSON, _ := json.MarshalIndent(inquiryResponse, "", "  ")
		log.Printf("[INFO] Inquiry response:\n%s\n\n", string(inquiryJSON))

		if inquiryResponse.Result.ResultStatus == "S" {
			switch inquiryResponse.RefundStatus {
			case "SUCCESS":
				log.Println("[SUCCESS] Refund completed successfully!")
				log.Printf("[INFO] Refund ID: %s\n", inquiryResponse.RefundID)
				log.Printf("[INFO] Refund Time: %s\n", inquiryResponse.RefundTime)

				return &alipay.RefundResponse{
					Result: alipay.Result{
						ResultCode:    "SUCCESS",
						ResultStatus:  "S",
						ResultMessage: "Success",
					},
					RefundID:   inquiryResponse.RefundID,
					RefundTime: inquiryResponse.RefundTime,
				}

			case "FAIL":
				log.Printf("[ERROR] Refund failed: %s\n", inquiryResponse.RefundFailReason)

				return &alipay.RefundResponse{
					Result: alipay.Result{
						ResultCode:    "REFUND_FAILED",
						ResultStatus:  "F",
						ResultMessage: inquiryResponse.RefundFailReason,
					},
				}

			case "PROCESSING":
				log.Println("[INFO] Refund still processing...")

			default:
				log.Printf("[WARNING] Unknown refund status: %s\n", inquiryResponse.RefundStatus)
			}

		} else if inquiryResponse.Result.ResultStatus == "F" {
			log.Printf("[ERROR] Inquiry failed: %s\n", inquiryResponse.Result.ResultMessage)

			if inquiryResponse.Result.ResultCode == "REFUND_NOT_EXIST" {
				log.Println("[ERROR] Refund does not exist in wallet system")
				return &alipay.RefundResponse{
					Result: alipay.Result{
						ResultCode:    "REFUND_NOT_EXIST",
						ResultStatus:  "F",
						ResultMessage: "Refund not found in wallet system",
					},
				}
			}
		}

		if attempt < maxAttempts {
			log.Printf("[INFO] Waiting %d seconds before next attempt...\n", intervalSeconds)
			time.Sleep(intervalSeconds * time.Second)
		}
	}

	log.Println("=================================================================")
	log.Println("[WARNING] POLLING TIMEOUT - Refund status still unknown")
	log.Println("[WARNING] Manual intervention may be required")
	log.Println("=================================================================")

	return nil
}

func buildRefundResponse(refundResponse alipay.RefundResponse) fiber.Map {
	response := fiber.Map{
		"resultStatus":  refundResponse.Result.ResultStatus,
		"resultCode":    refundResponse.Result.ResultCode,
		"resultMessage": refundResponse.Result.ResultMessage,
	}

	switch refundResponse.Result.ResultStatus {
	case "S":
		response["status"] = "SUCCESS"
		response["success"] = true
		if refundResponse.RefundID != "" {
			response["refundId"] = refundResponse.RefundID
		}
		if refundResponse.RefundTime != "" {
			response["refundTime"] = refundResponse.RefundTime
		}

	case "U":
		response["status"] = "PENDING"
		response["success"] = false
		response["message"] = "Refund is being processed. Status is unknown."

	case "F":
		response["status"] = "FAILED"
		response["success"] = false
		response["message"] = refundResponse.Result.ResultMessage
	}

	return response
}

func generateRefundRequestID() string {
	return fmt.Sprintf("REFUND-%s-%d", uuid.New().String(), time.Now().Unix())
}
