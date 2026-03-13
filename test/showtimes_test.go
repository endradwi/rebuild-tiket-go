package controllers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"tiket/lib"
)

func TestGetMovieShowtimes(t *testing.T) {
	router := setupRouter()

	t.Run("Get Showtimes for Spider-Man", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Movie ID 1 is Spider-Man in the seed data
		req, _ := http.NewRequest(http.MethodGet, "/movies/1/showtimes", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, w.Code)
		}

		var response lib.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		results := response.Result.([]interface{})
		if len(results) == 0 {
			t.Errorf("Expected showtimes for movie ID 1, but got none")
		}
	})

	t.Run("Filter Showtimes by Location", func(t *testing.T) {
		w := httptest.NewRecorder()
		// Filter by Location ID 1
		req, _ := http.NewRequest(http.MethodGet, "/movies/1/showtimes?location_id=1", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, but got %d", http.StatusOK, w.Code)
		}

		var response lib.Response
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to decode JSON response: %v", err)
		}

		// Result should only contain showtimes for cinemas in location ID 1
		// In seed data: Cinema 1 is in Location 1
	})
}
