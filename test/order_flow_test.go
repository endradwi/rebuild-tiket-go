package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"tiket/lib"
)

func TestOrderFlowIntegration(t *testing.T) {
	router := setupRouter()

	// 1. Generate a mock token for user ID 1
	token, err := lib.GenerateTokenJwt(map[string]interface{}{"userId": float64(1)})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	var orderId int

	t.Run("Step 1: Get Available Seats", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/orders/seats?showtime_id=1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Step 2: Create Order (Checkout now)", func(t *testing.T) {
		orderReq := lib.OrderCreateRequest{
			ShowtimeId: 1,
			SeatIds:    []int{1, 2},
		}
		body, _ := json.Marshal(orderReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp lib.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		result := resp.Result.(map[string]interface{})
		orderId = int(result["id"].(float64))
	})

	t.Run("Step 2.5: Get Order Details (Summary)", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/orders/%d", orderId), nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Step 3: Process Payment (Pay now)", func(t *testing.T) {
		paymentReq := lib.PaymentRequest{
			FullName:      "John Doe",
			Email:         "john@example.com",
			PhoneNumber:   "08123456789",
			PaymentMethod: "GOPAY",
		}
		body, _ := json.Marshal(paymentReq)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("/orders/%d/payment", orderId), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}
	})
}
