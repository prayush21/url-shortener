package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/prayushdave/url-shortener/internal/id"
	"github.com/prayushdave/url-shortener/internal/storage"
)

func setupTestServer(t *testing.T) (*gin.Engine, *storage.RedisStore) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Initialize Redis store
	store := storage.NewRedisStore("localhost:6379", "", 0)

	// Clear test database
	err := store.FlushDB(context.Background())
	require.NoError(t, err)

	// Initialize ID generator
	generator := id.NewGenerator()

	// Create handler
	handler := NewHandler(store, generator, "http://localhost:8080")

	// Setup router
	router := gin.New()
	handler.SetupRoutes(router)

	return router, store
}

func TestCreateURL_Integration(t *testing.T) {
	router, store := setupTestServer(t)
	defer store.Close()

	tests := []struct {
		name             string
		requestBody      map[string]interface{}
		rawBody          string // For malformed JSON tests
		expectedStatus   int
		validateResponse func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "Valid URL",
			requestBody: map[string]interface{}{
				"url": "https://example.com",
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response URLResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)

				// Validate response format
				assert.NotEmpty(t, response.ShortKey)
				assert.Equal(t, "https://example.com", response.URL)

				// Verify URL was stored in Redis
				url, err := store.Get(context.Background(), response.ShortKey)
				assert.NoError(t, err)
				assert.Equal(t, "https://example.com", url)

				// Validate key format
				assert.Len(t, response.ShortKey, id.KeyLength)
				assert.Regexp(t, "^[0-9A-Za-z]+$", response.ShortKey)
			},
		},
		{
			name: "Very Long URL",
			requestBody: map[string]interface{}{
				"url": "https://example.com/" + strings.Repeat("very-long-path/", 100),
			},
			expectedStatus: http.StatusCreated,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response URLResponse
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.NotEmpty(t, response.ShortKey)
			},
		},
		{
			name:           "Malformed JSON",
			rawBody:        `{"url": "https://example.com"`, // Missing closing brace
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request body")
			},
		},
		{
			name: "Invalid URL Format",
			requestBody: map[string]interface{}{
				"url": "not-a-url",
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL")
			},
		},
		{
			name: "Missing URL Field",
			requestBody: map[string]interface{}{
				"wrong_field": "https://example.com",
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request body")
			},
		},
		{
			name: "Empty URL",
			requestBody: map[string]interface{}{
				"url": "",
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid request body")
			},
		},
		{
			name: "URL Without Scheme",
			requestBody: map[string]interface{}{
				"url": "example.com",
			},
			expectedStatus: http.StatusBadRequest,
			validateResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var body []byte
			var err error

			if tt.rawBody != "" {
				body = []byte(tt.rawBody)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResponse(t, w)
		})
	}
}

func TestCreateURL_Concurrent(t *testing.T) {
	router, store := setupTestServer(t)
	defer store.Close()

	// Number of concurrent requests
	n := 50
	var wg sync.WaitGroup
	wg.Add(n)

	// Channel to collect errors
	errCh := make(chan error, n)
	successCh := make(chan string, n) // Channel to collect generated keys

	// Run concurrent requests
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()

			body := map[string]interface{}{
				"url": fmt.Sprintf("https://example.com/concurrent/%d", i),
			}
			jsonBody, err := json.Marshal(body)
			if err != nil {
				errCh <- fmt.Errorf("failed to marshal request %d: %v", i, err)
				return
			}

			req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != http.StatusCreated {
				errCh <- fmt.Errorf("request %d: expected status %d, got %d", i, http.StatusCreated, w.Code)
				return
			}

			var response URLResponse
			if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
				errCh <- fmt.Errorf("request %d: failed to decode response: %v", i, err)
				return
			}

			successCh <- response.ShortKey
		}(i)
	}

	// Wait for all requests to complete
	wg.Wait()
	close(errCh)
	close(successCh)

	// Check for any errors
	for err := range errCh {
		t.Error(err)
	}

	// Verify all generated keys are unique
	keys := make(map[string]bool)
	for key := range successCh {
		if keys[key] {
			t.Errorf("Duplicate key generated: %s", key)
		}
		keys[key] = true
	}
}

func TestRedirectURL_Integration(t *testing.T) {
	router, store := setupTestServer(t)
	defer store.Close()

	// Create a test URL first
	testURL := "https://example.com/test"
	createResp := createTestURL(t, router, testURL)

	tests := []struct {
		name           string
		key            string
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Valid Key - Successful Redirect",
			key:            createResp.ShortKey,
			expectedStatus: http.StatusFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				assert.Equal(t, testURL, w.Header().Get("Location"))
			},
		},
		{
			name:           "Invalid Key Format - Special Characters",
			key:            "invalid!@#",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Invalid Key Format - Too Short",
			key:            "abc123",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Invalid Key Format - Too Long",
			key:            "abc123def456ghi789",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Invalid Key Format - Contains Spaces",
			key:            "abc%20123d", // URL-encoded space
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Invalid Key Format - Contains Underscores",
			key:            "abc_123d",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Invalid Key Format - Contains Hyphens",
			key:            "abc-123d",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Non-existent key",
			key:            "abcd1234", // Valid format (8 chars, base62) but doesn't exist
			expectedStatus: http.StatusNoContent,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "URL not found")
			},
		},
		{
			name:           "Non-existent Key - Another Valid Format",
			key:            "XYZ98765",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "URL not found")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/"+tt.key, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResp(t, w)
		})
	}
}

// Helper function to create a test URL and return the response
func createTestURL(t *testing.T, router *gin.Engine, url string) *URLResponse {
	body := map[string]interface{}{
		"url": url,
	}
	jsonBody, err := json.Marshal(body)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/urls", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	var response URLResponse
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	return &response
}

func TestRedirectURL_EdgeCases(t *testing.T) {
	router, store := setupTestServer(t)
	defer store.Close()

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		validateResp   func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name:           "Empty Key - Root Path",
			path:           "/",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				// For root path with empty key, Gin might return different response
				// This tests the actual behavior
				if w.Code == http.StatusNotFound {
					// Check if it's JSON error response or HTML 404
					contentType := w.Header().Get("Content-Type")
					if strings.Contains(contentType, "application/json") {
						var response map[string]string
						err := json.NewDecoder(w.Body).Decode(&response)
						if err == nil {
							assert.Contains(t, response["error"], "Invalid URL key format")
						}
					}
				}
			},
		},
		{
			name:           "URL Encoded Key - Invalid Characters",
			path:           "/abc%20123", // Space encoded as %20
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
		{
			name:           "Key With Invalid Dot Character",
			path:           "/abcd.234",
			expectedStatus: http.StatusNotFound,
			validateResp: func(t *testing.T, w *httptest.ResponseRecorder) {
				// This tests keys with dots, which are invalid in our Base62 character set
				var response map[string]string
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Contains(t, response["error"], "Invalid URL key format")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateResp(t, w)
		})
	}
}

func TestDeleteURL_Integration(t *testing.T) {
	router, store := setupTestServer(t)
	defer store.Close()

	tests := []struct {
		name           string
		setup          func(t *testing.T) string // returns key if needed
		key            string
		expectedStatus int
		validateState  func(t *testing.T, key string)
	}{
		{
			name: "Successful deletion",
			setup: func(t *testing.T) string {
				resp := createTestURL(t, router, "https://example.com")
				return resp.ShortKey
			},
			expectedStatus: http.StatusOK,
			validateState: func(t *testing.T, key string) {
				// Verify URL was deleted from Redis
				_, err := store.Get(context.Background(), key)
				assert.ErrorIs(t, err, storage.ErrNotFound)
			},
		},
		{
			name:           "Non-existent key",
			key:            "abcd1234", // Valid format (8 chars, base62) but doesn't exist
			expectedStatus: http.StatusNoContent,
			validateState: func(t *testing.T, key string) {
				// Verify key still doesn't exist
				_, err := store.Get(context.Background(), key)
				assert.ErrorIs(t, err, storage.ErrNotFound)
			},
		},
		{
			name:           "Invalid key format - too short",
			key:            "abc",
			expectedStatus: http.StatusBadRequest,
			validateState: func(t *testing.T, key string) {
				// No state change needed
			},
		},
		{
			name:           "Invalid key format - invalid characters",
			key:            "invalid@#$key",
			expectedStatus: http.StatusBadRequest,
			validateState: func(t *testing.T, key string) {
				// No state change needed
			},
		},
		{
			name: "Delete already deleted key",
			setup: func(t *testing.T) string {
				resp := createTestURL(t, router, "https://example.com")
				key := resp.ShortKey
				// Delete it first time
				req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/urls/%s", key), nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)
				assert.Equal(t, http.StatusOK, w.Code)
				return key
			},
			expectedStatus: http.StatusNoContent,
			validateState: func(t *testing.T, key string) {
				// Verify URL is still deleted
				_, err := store.Get(context.Background(), key)
				assert.ErrorIs(t, err, storage.ErrNotFound)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var key string
			if tt.setup != nil {
				key = tt.setup(t)
			} else {
				key = tt.key
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/urls/%s", key), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.validateState(t, key)
		})
	}
}

func TestDeleteURL_Concurrent(t *testing.T) {
	router, store := setupTestServer(t)
	defer store.Close()

	// Create a URL to be deleted
	resp := createTestURL(t, router, "https://example.com")
	key := resp.ShortKey

	// Number of concurrent deletion attempts
	n := 50
	var wg sync.WaitGroup
	wg.Add(n)

	// Channels to collect results
	successCh := make(chan int, n) // Channel to collect successful status codes
	errCh := make(chan error, n)

	// Run concurrent deletion requests
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/urls/%s", key), nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
				errCh <- fmt.Errorf("unexpected status code: %d", w.Code)
				return
			}
			successCh <- w.Code
		}()
	}

	// Wait for all goroutines to complete
	wg.Wait()
	close(successCh)
	close(errCh)

	// Check for errors
	for err := range errCh {
		t.Errorf("concurrent deletion error: %v", err)
	}

	// Verify results
	okCount := 0
	noContentCount := 0
	for code := range successCh {
		switch code {
		case http.StatusOK:
			okCount++
		case http.StatusNoContent:
			noContentCount++
		}
	}

	// We should have exactly one OK (the first successful deletion)
	// and the rest should be NoContent (subsequent attempts)
	assert.Equal(t, 1, okCount, "Expected exactly one successful deletion")
	assert.Equal(t, n-1, noContentCount, "Expected all other attempts to return NoContent")

	// Verify the URL is actually deleted
	_, err := store.Get(context.Background(), key)
	assert.ErrorIs(t, err, storage.ErrNotFound, "URL should be deleted after concurrent deletion attempts")
}
