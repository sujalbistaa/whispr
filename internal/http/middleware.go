package http

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware checks for a secret X-Admin-Token header.
func AdminAuthMiddleware() gin.HandlerFunc {
	// Get the secret token from the environment
	requiredToken := os.Getenv("X_ADMIN_TOKEN")

	if requiredToken == "" {
		panic("CRITICAL: X_ADMIN_TOKEN environment variable not set.")
	}

	return func(c *gin.Context) {
		// Get the token from the request header
		suppliedToken := c.GetHeader("X-Admin-Token")

		if suppliedToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Admin token required"})
			return
		}

		if suppliedToken != requiredToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Invalid admin token"})
			return
		}
		c.Next()
	}
}

// SecurityHeadersMiddleware adds basic, sensible security headers.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-Content-Type-Options", "nosniff")

		// --- THIS IS THE FINAL CSP FIX ---
		// We MUST add 'unsafe-eval' for Alpine.js to work.
		csp := "default-src 'self';"
		csp += " script-src 'self' 'unsafe-inline' 'unsafe-eval' cdn.jsdelivr.net cdn.tailwindcss.com;" // <-- 'unsafe-eval' ADDED
		csp += " style-src 'self' 'unsafe-inline' cdn.tailwindcss.com;" // Allow tailwindcss CDN
		csp += " connect-src 'self' ws: wss:;" // Allow fetch, XHR, and WebSockets
		
		c.Header("Content-Security-Policy", csp)
		// --- END FIX ---

		c.Next()
	}
}