package controllers_test

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestUpdateProfile tests editing a user's details without auth validation
func TestUpdateProfile(t *testing.T) {
	router := setupRouter()

	t.Run("Update Profile Missing Auth", func(t *testing.T) {
		w := httptest.NewRecorder()
		
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("first_name", "Test")
		_ = writer.WriteField("last_name", "User")
		writer.Close()

		req, _ := http.NewRequest(http.MethodPatch, "/users/profile", &body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d, but got %d", http.StatusUnauthorized, w.Code)
		}
	})
	
	t.Run("Fetching Logged In User Missing Auth", func(t *testing.T) {
	    w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/users/profile", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected status %d for unauthenticated fetch, but got %d", http.StatusUnauthorized, w.Code)
		}
	})
}
