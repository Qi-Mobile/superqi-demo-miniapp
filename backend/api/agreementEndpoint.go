package api

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"superQiMiniAppBackend/alipay"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// =========================================================================
// REQUEST/RESPONSE TYPES
// =========================================================================

type prepareContractRequest struct {
	ContractDescription string `json:"contractDescription" validate:"required"`
}

type applyAccessTokenRequest struct {
	AuthCode string `json:"authCode" validate:"required"`
}

type executeAgreementPaymentRequest struct {
	AccessToken      string `json:"accessToken" validate:"required"`
	Amount           int64  `json:"amount" validate:"required"`
	Currency         string `json:"currency"`
	OrderDescription string `json:"orderDescription"`
}

// =========================================================================
// ENDPOINT INITIALIZATION
// =========================================================================

func InitAgreementEndpoint(group fiber.Router) {
	log.Println("=================================================================")
	log.Println("[Backend] Initializing Agreement Payment Endpoints")
	log.Println("=================================================================")
	log.Println("[Backend] Registering route: POST /api/agreement/prepare")
	log.Println("[Backend] Registering route: POST /api/agreement/apply-token")
	log.Println("[Backend] Registering route: POST /api/agreement/pay")
	log.Println("=================================================================")

	agreementGroup := group.Group("/agreement")

	agreementGroup.Post("/prepare", handlePrepareContract)

	agreementGroup.Post("/apply-token", handleApplyAccessToken)

	agreementGroup.Post("/pay", handleExecuteAgreementPayment)
}

// =========================================================================
// STEP 1: PREPARE CONTRACT
// =========================================================================

func handlePrepareContract(ctx *fiber.Ctx) error {
	log.Println("\n=================================================================")
	log.Println("[Backend] INCOMING REQUEST: POST /api/agreement/prepare")
	log.Println("=================================================================")

	log.Println("[Backend] Request Headers:")
	ctx.Request().Header.VisitAll(func(key, value []byte) {
		log.Printf("[Backend]   %s: %s\n", string(key), string(value))
	})

	rawBody := string(ctx.Body())
	log.Printf("[Backend] Raw Request Body: %s\n", rawBody)

	var request prepareContractRequest
	if err := ctx.BodyParser(&request); err != nil {
		log.Printf("[Backend] ERROR: Failed to parse request body: %v\n", err)
		log.Println("=================================================================")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	log.Printf("[Backend] SUCCESS: Request parsed successfully\n")
	log.Printf("[Backend] Contract description: %s\n", request.ContractDescription)
	log.Println("[Backend] -----------------------------------------------------------")
	log.Println("[Backend] Calling Alipay+ PrepareAuthorization API...")

	prepareResponse, err := alipay.Interface.PrepareAuthorization(request.ContractDescription)
	if err != nil {
		log.Printf("[Backend] ERROR: Alipay+ API call failed: %v\n", err)
		log.Println("[Backend] This could mean:")
		log.Println("[Backend]   1. Alipay+ gateway is unreachable")
		log.Println("[Backend]   2. Invalid credentials in .env file")
		log.Println("[Backend]   3. Network/firewall issue")
		log.Println("[Backend]   4. PrepareAuthorization function not implemented")
		log.Println("=================================================================")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to prepare contract: "+err.Error())
	}

	log.Println("[Backend] SUCCESS: Alipay+ API call successful")
	responseJSON, _ := json.MarshalIndent(prepareResponse, "", "  ")
	log.Printf("[Backend] Alipay+ Response:\n%s\n", string(responseJSON))
	log.Println("[Backend] -----------------------------------------------------------")

	log.Printf("[Backend] Checking result status: %s\n", prepareResponse.Result.ResultStatus)
	log.Printf("[Backend] Result code: %s\n", prepareResponse.Result.ResultCode)
	log.Printf("[Backend] Result message: %s\n", prepareResponse.Result.ResultMessage)

	if prepareResponse.Result.ResultStatus != "S" {
		log.Printf("[Backend] ERROR: Contract preparation failed\n")
		log.Printf("[Backend] Status: %s, Code: %s, Message: %s\n",
			prepareResponse.Result.ResultStatus,
			prepareResponse.Result.ResultCode,
			prepareResponse.Result.ResultMessage)
		log.Println("=================================================================")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":       false,
			"resultStatus":  prepareResponse.Result.ResultStatus,
			"resultCode":    prepareResponse.Result.ResultCode,
			"resultMessage": prepareResponse.Result.ResultMessage,
		})
	}

	if prepareResponse.AuthURL == "" {
		log.Printf("[Backend] WARNING: authUrl is EMPTY in response!\n")
		log.Printf("[Backend] This should not happen if resultStatus is 'S'\n")
		log.Printf("[Backend] Full response: %+v\n", prepareResponse)
		log.Println("=================================================================")
	} else {
		log.Printf("[Backend] SUCCESS: Authorization URL received: %s\n", prepareResponse.AuthURL)
	}

	response := fiber.Map{
		"success":       true,
		"authUrl":       prepareResponse.AuthURL,
		"resultStatus":  prepareResponse.Result.ResultStatus,
		"resultCode":    prepareResponse.Result.ResultCode,
		"resultMessage": prepareResponse.Result.ResultMessage,
	}

	log.Println("[Backend] -----------------------------------------------------------")
	log.Println("[Backend] Building response to frontend:")
	responseToFrontendJSON, _ := json.MarshalIndent(response, "", "  ")
	log.Printf("%s\n", string(responseToFrontendJSON))
	log.Println("[Backend] SUCCESS: Sending response to frontend")
	log.Println("=================================================================")

	return ctx.JSON(response)
}

// =========================================================================
// STEP 2: APPLY ACCESS TOKEN
// =========================================================================

func handleApplyAccessToken(ctx *fiber.Ctx) error {
	log.Println("\n=================================================================")
	log.Println("[Backend] INCOMING REQUEST: POST /api/agreement/apply-token")
	log.Println("=================================================================")

	rawBody := string(ctx.Body())
	log.Printf("[Backend] Raw Request Body: %s\n", rawBody)

	var request applyAccessTokenRequest
	if err := ctx.BodyParser(&request); err != nil {
		log.Printf("[Backend] ERROR: Failed to parse request body: %v\n", err)
		log.Println("=================================================================")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	log.Printf("[Backend] SUCCESS: Request parsed successfully\n")
	log.Printf("[Backend] Auth code received (first 20 chars): %s...\n", truncateString(request.AuthCode, 20))
	log.Printf("[Backend] Auth code length: %d\n", len(request.AuthCode))
	log.Println("[Backend] -----------------------------------------------------------")
	log.Println("[Backend] Calling Alipay+ ApplyToken API...")

	tokenResponse, err := alipay.Interface.ApplyToken(request.AuthCode)
	if err != nil {
		log.Printf("[Backend] ERROR: Token exchange failed: %v\n", err)
		log.Println("=================================================================")
		return fiber.NewError(fiber.StatusBadRequest, "Token exchange failed: "+err.Error())
	}

	log.Println("[Backend] SUCCESS: Alipay+ API call successful")
	responseJSON, _ := json.MarshalIndent(tokenResponse, "", "  ")
	log.Printf("[Backend] Alipay+ Response:\n%s\n", string(responseJSON))
	log.Println("[Backend] -----------------------------------------------------------")

	log.Printf("[Backend] Checking result status: %s\n", tokenResponse.Result.ResultStatus)
	log.Printf("[Backend] Result code: %s\n", tokenResponse.Result.ResultCode)

	if tokenResponse.Result.ResultStatus != "S" || tokenResponse.Result.ResultCode != "SUCCESS" {
		log.Printf("[Backend] ERROR: Token application failed: %s\n", tokenResponse.Result.ResultMessage)
		log.Println("=================================================================")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":       false,
			"resultStatus":  tokenResponse.Result.ResultStatus,
			"resultCode":    tokenResponse.Result.ResultCode,
			"resultMessage": tokenResponse.Result.ResultMessage,
		})
	}

	log.Printf("[Backend] SUCCESS: Access token obtained\n")
	log.Printf("[Backend] Access Token (first 20 chars): %s...\n", truncateString(tokenResponse.AccessToken, 20))
	log.Printf("[Backend] Refresh Token (first 20 chars): %s...\n", truncateString(tokenResponse.RefreshToken, 20))
	log.Printf("[Backend] Token Expiry: %s\n", tokenResponse.AccessTokenExpiryTime)
	log.Printf("[Backend] Customer ID: %s\n", tokenResponse.CustomerID)
	log.Printf("[Backend] IMPORTANT: Store this accessToken in database for future payments!\n")
	log.Println("[Backend] SUCCESS: Sending response to frontend")
	log.Println("=================================================================")

	return ctx.JSON(fiber.Map{
		"success":                true,
		"accessToken":            tokenResponse.AccessToken,
		"refreshToken":           tokenResponse.RefreshToken,
		"accessTokenExpiryTime":  tokenResponse.AccessTokenExpiryTime,
		"refreshTokenExpiryTime": tokenResponse.RefreshTokenExpiryTime,
		"customerId":             tokenResponse.CustomerID,
		"resultStatus":           tokenResponse.Result.ResultStatus,
		"resultCode":             tokenResponse.Result.ResultCode,
		"resultMessage":          tokenResponse.Result.ResultMessage,
	})
}

// =========================================================================
// STEP 3: EXECUTE AGREEMENT PAYMENT
// =========================================================================

func handleExecuteAgreementPayment(ctx *fiber.Ctx) error {
	log.Println("\n=================================================================")
	log.Println("[Backend] INCOMING REQUEST: POST /api/agreement/pay")
	log.Println("=================================================================")

	rawBody := string(ctx.Body())
	log.Printf("[Backend] Raw Request Body: %s\n", rawBody)

	var request executeAgreementPaymentRequest
	if err := ctx.BodyParser(&request); err != nil {
		log.Printf("[Backend] ERROR: Failed to parse request body: %v\n", err)
		log.Println("=================================================================")
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if request.Currency == "" {
		request.Currency = "IQD"
	}
	if request.OrderDescription == "" {
		request.OrderDescription = "Agreement payment - Monthly subscription"
	}

	log.Printf("[Backend] SUCCESS: Request parsed successfully\n")
	log.Printf("[Backend] Access Token (first 20 chars): %s...\n", truncateString(request.AccessToken, 20))
	log.Printf("[Backend] Amount: %d %s\n", request.Amount, request.Currency)
	log.Printf("[Backend] Order Description: %s\n", request.OrderDescription)
	log.Println("[Backend] -----------------------------------------------------------")

	if request.Amount <= 0 {
		log.Println("[Backend] ERROR: Invalid payment amount")
		log.Println("=================================================================")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":       false,
			"resultStatus":  "F",
			"resultMessage": "Payment amount must be greater than 0",
		})
	}

	if request.AccessToken == "" {
		log.Println("[Backend] ERROR: Access token is required")
		log.Println("=================================================================")
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success":       false,
			"resultStatus":  "F",
			"resultMessage": "Access token is required",
		})
	}

	log.Println("[Backend] Executing agreement payment...")

	paymentResponse, err := executeAgreementPaymentInternal(request.AccessToken, request.Amount, request.Currency, request.OrderDescription)
	if err != nil {
		log.Printf("[Backend] ERROR: Failed to execute payment: %v\n", err)
		log.Println("=================================================================")
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success":       false,
			"resultStatus":  "F",
			"resultMessage": err.Error(),
		})
	}

	response := buildAgreementPaymentResponse(paymentResponse)

	log.Println("[Backend] SUCCESS: Sending payment response to frontend")
	log.Println("=================================================================")
	return ctx.JSON(response)
}

// =========================================================================
// INTERNAL HELPER FUNCTIONS
// =========================================================================

func executeAgreementPaymentInternal(accessToken string, amount int64, currency string, orderDescription string) (alipay.PaymentResponse, error) {
	log.Println("[Backend] Preparing agreement payment request...")

	paymentRequestID := fmt.Sprintf("AGREEMENT-PAY-%s-%d", uuid.New().String(), time.Now().Unix())
	log.Printf("[Backend] Generated Payment Request ID: %s\n", paymentRequestID)

	expiryTime := time.Now().Add(30 * time.Minute).Format("2006-01-02T15:04:05-07:00")

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:1999"
	}

	paymentRequest := alipay.PaymentRequest{
		ProductCode:      alipay.AGREEMENT_PAYMENT, // Important: Use agreement payment product code
		PaymentRequestID: paymentRequestID,
		PaymentAuthCode:  accessToken, // This is the key difference - we use accessToken here
		PaymentAmount: alipay.PaymentAmount{
			Currency: currency,
			Value:    fmt.Sprintf("%d", amount),
		},
		Order: alipay.Order{
			OrderDescription: orderDescription,
		},
		PaymentExpiryTime: expiryTime,
		PaymentNotifyURL:  baseURL + "/api/webhook/payment-notify",
	}

	requestJSON, _ := json.MarshalIndent(paymentRequest, "", "  ")
	log.Printf("[Backend] Agreement payment request:\n%s\n", string(requestJSON))

	log.Println("[Backend] Calling /v1/payments/pay API...")
	paymentResponse, err := alipay.Interface.Pay(paymentRequest)
	if err != nil {
		log.Printf("[Backend] ERROR: Payment API call failed: %v\n", err)
		return alipay.PaymentResponse{}, fmt.Errorf("payment API call failed: %v", err)
	}

	responseJSON, _ := json.MarshalIndent(paymentResponse, "", "  ")
	log.Printf("[Backend] Payment API response:\n%s\n", string(responseJSON))

	switch paymentResponse.Result.ResultStatus {
	case "S":
		log.Println("[Backend] SUCCESS: Payment completed immediately")
		log.Printf("[Backend] Payment ID: %s\n", paymentResponse.PaymentID)
		log.Printf("[Backend] Payment Time: %s\n", paymentResponse.PaymentTime)
		log.Println("[Backend] WARNING: Money deducted from user's wallet automatically!")

	case "U":
		log.Println("[Backend] WARNING: Payment status unknown")
		log.Println("[Backend] Backend should poll /v1/payments/inquiryPayment for status")

	case "F":
		log.Printf("[Backend] ERROR: Payment failed - %s\n", paymentResponse.Result.ResultMessage)
		log.Printf("[Backend] Error Code: %s\n", paymentResponse.Result.ResultCode)

	default:
		log.Printf("[Backend] WARNING: Unexpected status: %s\n", paymentResponse.Result.ResultStatus)
	}

	return paymentResponse, nil
}

func buildAgreementPaymentResponse(paymentResponse alipay.PaymentResponse) fiber.Map {
	response := fiber.Map{
		"resultStatus":  paymentResponse.Result.ResultStatus,
		"resultCode":    paymentResponse.Result.ResultCode,
		"resultMessage": paymentResponse.Result.ResultMessage,
	}

	switch paymentResponse.Result.ResultStatus {
	case "S":
		response["status"] = "SUCCESS"
		response["success"] = true
		if paymentResponse.PaymentID != "" {
			response["paymentId"] = paymentResponse.PaymentID
		}
		if paymentResponse.PaymentRequestID != "" {
			response["paymentRequestId"] = paymentResponse.PaymentRequestID
		}
		if paymentResponse.PaymentTime != "" {
			response["paymentTime"] = paymentResponse.PaymentTime
		}

	case "U":
		response["status"] = "PENDING"
		response["success"] = false
		response["message"] = "Payment status is unknown. Backend should poll for status."
		if paymentResponse.PaymentID != "" {
			response["paymentId"] = paymentResponse.PaymentID
		}

	case "F":
		response["status"] = "FAILED"
		response["success"] = false
		response["message"] = paymentResponse.Result.ResultMessage

	default:
		response["status"] = "UNKNOWN"
		response["success"] = false
		response["message"] = "Unexpected payment status: " + paymentResponse.Result.ResultStatus
	}

	return response
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
