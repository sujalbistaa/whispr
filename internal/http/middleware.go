package http

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// AdminAuthMiddleware checks for a secret X-Admin-Token header.
func AdminAuthMiddleware() gin.HandlerFunc {
	// Get the secret token from the environment
	// We read this once when the middleware is initialized
	requiredToken := os.Getenv("X_ADMIN_TOKEN")

	// If no token is set in the environment, we must fail closed.
	// We log a fatal error because this is a critical misconfiguration.
	if requiredToken == "" {
		panic("CRITICAL: X_ADMIN_TOKEN environment variable not set.")
	}

	return func(c *gin.Context) {
		// Get the token from the request header
		suppliedToken := c.GetHeader("X-Admin-Token")

		// Check if the token is missing or incorrect
		if suppliedToken == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Admin token required"})
			return
		}

		if suppliedToken != requiredToken {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden: Invalid admin token"})
			return
		}

		// Token is valid, proceed to the next handler
		c.Next()
	}
}

// SecurityHeadersMiddleware adds basic, sensible security headers.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevents clickjacking
		c.Header("X-Frame-Options", "DENY")
		// Prevents MIME-type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// --- THIS IS THE UPDATED SECTION ---
		// We must allow 'unsafe-inline' for Alpine.js attributes and styles.
		// We must allow the CDNs for Tailwind and Alpine.
		csp := "default-src 'self';"
		csp += " script-src 'self' 'unsafe-inline' cdn.jsdelivr.net;"
		csp += " style-src 'self' 'unsafe-inline' cdn.tailwindcss.com;"
		c.Header("Content-Security-Policy", csp)
		// --- END UPDATE ---

		c.Next()
	}
}