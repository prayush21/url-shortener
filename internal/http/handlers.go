package http

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"github.com/prayushdave/url-shortener/internal/id"
	"github.com/prayushdave/url-shortener/internal/storage"
)

// URLRequest represents the request body for URL shortening
type URLRequest struct {
	URL string `json:"url" binding:"required"`
}

// URLResponse represents the response for URL shortening
type URLResponse struct {
	ShortKey string `json:"short_key"`
	URL      string `json:"url"`
}

// Handler handles HTTP requests for the URL shortener
type Handler struct {
	store     storage.Store
	generator *id.Generator
	baseURL   string
}

// NewHandler creates a new Handler instance
func NewHandler(store storage.Store, generator *id.Generator, baseURL string) *Handler {
	return &Handler{
		store:     store,
		generator: generator,
		baseURL:   baseURL,
	}
}

// SetupRoutes configures the routes for the handler
func (h *Handler) SetupRoutes(r *gin.Engine) {
	v1 := r.Group("/api/v1")
	{
		v1.POST("/urls", h.CreateURL)
		v1.DELETE("/urls/:key", h.DeleteURL)
	}

	// Add redirect route at root level
	r.GET("/:key", h.RedirectURL)
}

// CreateURL handles the URL shortening request
func (h *Handler) CreateURL(c *gin.Context) {
	var req URLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil || (!parsedURL.IsAbs() || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https")) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL. Must be absolute with http(s) scheme"})
		return
	}

	// Generate a unique key
	var key string
	for attempts := 0; attempts < 3; attempts++ {
		key, err = h.generator.Generate()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate key"})
			return
		}

		// Try to store the URL
		err = h.store.Set(c.Request.Context(), key, req.URL)
		if err == nil {
			break
		}

		// If we got an error other than collision, return error
		if err != storage.ErrKeyExists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store URL"})
			return
		}

		// On collision, try again with a new key
		continue
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate unique key after multiple attempts"})
		return
	}

	response := URLResponse{
		ShortKey: key,
		URL:      req.URL,
	}

	c.JSON(http.StatusCreated, response)
}

// RedirectURL handles the URL redirection
func (h *Handler) RedirectURL(c *gin.Context) {
	key := c.Param("key")

	// Validate key format
	if !h.generator.ValidateKey(key) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid URL key format"})
		return
	}

	// Get the original URL from storage
	url, err := h.store.Get(c.Request.Context(), key)
	if err == storage.ErrNotFound {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve URL"})
		return
	}

	// Redirect to the original URL
	c.Redirect(http.StatusFound, url)
}

// DeleteURL handles the URL deletion request
func (h *Handler) DeleteURL(c *gin.Context) {
	key := c.Param("key")

	// Validate key format
	if !h.generator.ValidateKey(key) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL key format"})
		return
	}

	// Delete the URL mapping
	err := h.store.Delete(c.Request.Context(), key)
	if err == storage.ErrNotFound {
		c.Status(http.StatusNoContent)
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete URL"})
		return
	}

	c.Status(http.StatusOK)
}
