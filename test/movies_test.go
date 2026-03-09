package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"tiket/routers"

	"github.com/gin-gonic/gin"
)

// setupRouter initializes the Gin router with all application routes for testing
func setupRouter() *gin.Engine {
	// Set Gin to run in Test Mode so we don't get excessive logs during tests
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	routers.InitRouter(router)
	return router
}

// TestGetAllMovies Tests the pagination and filtering features of the Movie listing endpoint
func TestGetAllMovies(t *testing.T) {
	router := setupRouter()

	tests := []struct {
		name         string
		query        string
		expectedCode int
	}{
		{
			name:         "Default Pagination Request",
			query:        "",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Valid Pagination Query",
			query:        "?page=1&limit=5",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Search Query Filter",
			query:        "?search=the&page=1",
			expectedCode: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/movies"+tc.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != tc.expectedCode {
				t.Errorf("Expected status %d, but got %d", tc.expectedCode, w.Code)
			}

			// Parse response to ensure there's a JSON block
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to decode JSON response: %v", err)
			}
			
			// Verify response contains the 'page_info' key from pagination
			if _, exists := response["page_info"]; !exists {
				t.Errorf("Expected response to contain 'page_info' metadata block")
			}
		})
	}
}

// TestGetMovieById validates that retrieving a movie yields correct responses
func TestGetMovieById(t *testing.T) {
	router := setupRouter()

	t.Run("Invalid Non-Integer Movie ID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/movies/abc", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid ID, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("Valid Integer Movie ID Not Found", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Try accessing a significantly high ID that shouldn't exist
		req, _ := http.NewRequest(http.MethodGet, "/movies/999999", nil)
		router.ServeHTTP(w, req)

		// 404 or 500 depending on how pgx handles pgx.ErrNoRows in GetMovieById
		if w.Code != http.StatusNotFound && w.Code != http.StatusInternalServerError {
			t.Errorf("Expected Not Found / Internal error, got %d", w.Code)
		}
	})
}
