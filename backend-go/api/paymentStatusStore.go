package api

import (
	"log"
	"sync"
	"time"
)

// PaymentStatusInfo stores the current status of a payment
type PaymentStatusInfo struct {
	PaymentID        string    `json:"paymentId"`
	PaymentRequestID string    `json:"paymentRequestId"`
	Status           string    `json:"status"` // PENDING, SUCCESS, PROCESSING, FAIL, UNKNOWN
	PaymentStatus    string    `json:"paymentStatus,omitempty"`
	LastChecked      time.Time `json:"lastChecked"`
	Completed        bool      `json:"completed"` // Whether polling should stop
	Message          string    `json:"message,omitempty"`
}

// PaymentStatusStore is an in-memory store for tracking payment statuses
type PaymentStatusStore struct {
	mu       sync.RWMutex
	payments map[string]*PaymentStatusInfo
}

// Global payment status store
var paymentStore = &PaymentStatusStore{
	payments: make(map[string]*PaymentStatusInfo),
}

// Set updates or creates a payment status
func (s *PaymentStatusStore) Set(paymentID string, info *PaymentStatusInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[paymentID] = info
	log.Printf("[PaymentStore] Updated payment %s: Status=%s, Completed=%v", paymentID, info.Status, info.Completed)
}

// Get retrieves a payment status
func (s *PaymentStatusStore) Get(paymentID string) (*PaymentStatusInfo, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	info, exists := s.payments[paymentID]
	return info, exists
}

// Delete removes a payment from store (for cleanup)
func (s *PaymentStatusStore) Delete(paymentID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.payments, paymentID)
	log.Printf("[PaymentStore] Deleted payment %s from store", paymentID)
}

// GetAll returns all payment statuses (for debugging)
func (s *PaymentStatusStore) GetAll() map[string]*PaymentStatusInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid race conditions
	copy := make(map[string]*PaymentStatusInfo)
	for k, v := range s.payments {
		copy[k] = v
	}
	return copy
}
