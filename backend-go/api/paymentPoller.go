package api

import (
	"log"
	"superQiMiniAppBackend/alipay"
	"time"
)

const (
	pollingInterval  = 5 * time.Second   // Poll every 5 seconds
	maxPollingTime   = 2 * time.Minute   // Poll for max 2 minutes
	cleanupDelay     = 10 * time.Minute  // Keep in cache for 10 minutes after completion
)

// StartPaymentPolling starts a background goroutine to poll payment status
func StartPaymentPolling(paymentID, paymentRequestID string) {
	log.Printf("[PaymentPoller] Starting polling for payment: %s", paymentID)

	// Initialize payment status as PENDING
	paymentStore.Set(paymentID, &PaymentStatusInfo{
		PaymentID:        paymentID,
		PaymentRequestID: paymentRequestID,
		Status:           "PENDING",
		LastChecked:      time.Now(),
		Completed:        false,
		Message:          "Payment initiated, waiting for completion",
	})

	// Start polling in background
	go pollPaymentStatus(paymentID, paymentRequestID)
}

// pollPaymentStatus is the background polling worker
func pollPaymentStatus(paymentID, paymentRequestID string) {
	startTime := time.Now()
	attemptCount := 0
	maxAttempts := int(maxPollingTime / pollingInterval) // 24 attempts (2 min / 5 sec)

	log.Printf("[PaymentPoller] Started polling for payment %s (max %d attempts)", paymentID, maxAttempts)

	ticker := time.NewTicker(pollingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			attemptCount++
			elapsed := time.Since(startTime)

			log.Printf("[PaymentPoller] Attempt %d/%d for payment %s (elapsed: %.0fs)",
				attemptCount, maxAttempts, paymentID, elapsed.Seconds())

			// Check payment status
			status := checkPaymentStatus(paymentID, paymentRequestID)

			// Update store
			status.LastChecked = time.Now()
			paymentStore.Set(paymentID, status)

			// Check if we should stop polling
			if status.Completed {
				log.Printf("[PaymentPoller] Payment %s is complete (status: %s). Stopping poll.", paymentID, status.Status)

				// Schedule cleanup after delay
				go func() {
					time.Sleep(cleanupDelay)
					paymentStore.Delete(paymentID)
					log.Printf("[PaymentPoller] Cleaned up payment %s from cache", paymentID)
				}()

				return
			}

			// Stop if max attempts reached
			if attemptCount >= maxAttempts {
				log.Printf("[PaymentPoller] Max polling time reached for payment %s. Stopping.", paymentID)

				// Mark as timeout
				status.Status = "TIMEOUT"
				status.Message = "Payment status check timed out after 2 minutes. Please check manually."
				status.Completed = true
				paymentStore.Set(paymentID, status)

				// Schedule cleanup
				go func() {
					time.Sleep(cleanupDelay)
					paymentStore.Delete(paymentID)
				}()

				return
			}
		}
	}
}

// checkPaymentStatus queries Alipay and returns current status
func checkPaymentStatus(paymentID, paymentRequestID string) *PaymentStatusInfo {
	inquiryRequest := alipay.InquiryPaymentRequest{
		PaymentID:        paymentID,
		PaymentRequestID: paymentRequestID,
	}

	inquiryResponse, err := alipay.Interface.InquiryPayment(inquiryRequest)
	if err != nil {
		log.Printf("[PaymentPoller] Error querying payment %s: %v", paymentID, err)
		return &PaymentStatusInfo{
			PaymentID:        paymentID,
			PaymentRequestID: paymentRequestID,
			Status:           "ERROR",
			Completed:        false,
			Message:          "Error checking payment status: " + err.Error(),
		}
	}

	// Parse response
	status := &PaymentStatusInfo{
		PaymentID:        paymentID,
		PaymentRequestID: paymentRequestID,
		PaymentStatus:    inquiryResponse.PaymentStatus,
	}

	// Handle based on result status
	switch inquiryResponse.Result.ResultStatus {
	case "S":
		// Success - inquiry was successful, check payment status
		switch inquiryResponse.PaymentStatus {
		case "SUCCESS":
			status.Status = "SUCCESS"
			status.Message = "Payment completed successfully"
			status.Completed = true
			log.Printf("[PaymentPoller] Payment %s SUCCESS", paymentID)

		case "PROCESSING":
			status.Status = "PROCESSING"
			status.Message = "Payment is still processing"
			status.Completed = false
			log.Printf("[PaymentPoller] Payment %s still PROCESSING", paymentID)

		case "AUTH_SUCCESS":
			status.Status = "AUTH_SUCCESS"
			status.Message = "Payment authorized but not finished"
			status.Completed = false
			log.Printf("[PaymentPoller] Payment %s AUTH_SUCCESS (not finished)", paymentID)

		case "FAIL":
			status.Status = "FAIL"
			status.Message = "Payment failed"
			status.Completed = true
			log.Printf("[PaymentPoller] Payment %s FAILED", paymentID)

		default:
			status.Status = "UNKNOWN"
			status.Message = "Unknown payment status: " + inquiryResponse.PaymentStatus
			status.Completed = false
			log.Printf("[PaymentPoller] Payment %s has unknown status: %s", paymentID, inquiryResponse.PaymentStatus)
		}

	case "F":
		// Failed - payment doesn't exist or inquiry failed
		if inquiryResponse.Result.ResultCode == "ORDER_NOT_EXIST" {
			status.Status = "NOT_FOUND"
			status.Message = "Payment not found. May not be accepted yet."
			status.Completed = false
			log.Printf("[PaymentPoller] Payment %s not found yet", paymentID)
		} else {
			status.Status = "FAIL"
			status.Message = inquiryResponse.Result.ResultMessage
			status.Completed = true
			log.Printf("[PaymentPoller] Payment %s inquiry failed: %s", paymentID, inquiryResponse.Result.ResultMessage)
		}

	case "U":
		// Unknown - retry
		status.Status = "UNKNOWN"
		status.Message = "Unknown exception, retrying..."
		status.Completed = false
		log.Printf("[PaymentPoller] Payment %s unknown exception", paymentID)

	default:
		status.Status = "UNKNOWN"
		status.Message = "Unexpected result status: " + inquiryResponse.Result.ResultStatus
		status.Completed = false
	}

	return status
}
