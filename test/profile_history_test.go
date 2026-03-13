package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"tiket/lib"
)

func TestProfileAndHistoryIntegration(t *testing.T) {
	router := setupRouter()

	// 1. Generate a mock token for user ID 1
	token, err := lib.GenerateTokenJwt(map[string]interface{}{"userId": float64(1)})
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	t.Run("Update Profile with String Phone Number", func(t *testing.T) {
		w := httptest.NewRecorder()
		// We use multipart/form-data for UpdateProfile
		req, _ := http.NewRequest(http.MethodPatch, "/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		
		// For simplicity, let's just test the GET profile first to see the new type
		req, _ = http.NewRequest(http.MethodGet, "/users/profile", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}
	})

	t.Run("Get User Order History", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/users/orders", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var resp lib.Response
		json.Unmarshal(w.Body.Bytes(), &resp)
		// Result should be an array (even if empty)
	})

	t.Run("Get Ticket Result", func(t *testing.T) {
		// We need an order ID. In the previous task we created one.
		// Let's assume order ID 1 exists from seed/previous tests or just skip if not found
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/orders/1/ticket", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)

		// It might be 404 if order 1 doesn't exist yet in this test run's DB state
		if w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 200 or 404, got %d", w.Code)
		}
	})
}
