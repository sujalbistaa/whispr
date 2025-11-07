package http

import (
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/sujalbistaa/whispr/internal/ws"
)

// SetupRoutes configures all application routes and middleware.
func SetupRoutes(router *gin.Engine, db *gorm.DB, hub *ws.Hub) {

	// --- Dependencies ---
	env := &Env{DB: db, Hub: hub}

	// --- Middleware ---

	// Apply global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(SecurityHeadersMiddleware()) // Security headers
	
	// CORS Middleware
	corsOrigin := os.Getenv("CORS_ORIGIN")
	if corsOrigin == "" {
		corsOrigin = "*" // Default to allow all for local dev
	}
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{corsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Admin-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// --- Rate Limiter Setup ---
	limiter := NewIPRateLimiter(rate.Limit(rateLimitRPS), rateLimitBurst)
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			limiter.mu.Lock()
			for ip, v := range limiter.visitors {
				if !v.Allow() {
					// If the limiter is still full, keep it.
				} else {
					// If allowed, it means it's old, so remove it.
					delete(limiter.visitors, ip)
				}
			}
			limiter.mu.Unlock()
		}
	}()


	// --- API Routes ---

	api := router.Group("/api")
	{
		api.GET("/posts", env.GetPosts)
		api.GET("/trending", env.GetTrendingPosts)
		api.POST("/posts", RateLimitMiddleware(limiter), env.CreatePost)
		api.POST("/posts/:id/vote", env.VoteOnPost)
		api.DELETE("/posts/:id", AdminAuthMiddleware(), env.DeletePost)
	}

	// --- WebSocket Route ---

	router.GET("/ws", func(c *gin.Context) {
		ws.ServeWs(hub, c.Writer, c.Request)
	})

	// --- Serve Frontend ---
	// This MUST come AFTER your API routes.
	// We serve a single file at the root. This does not conflict with /api.
	router.StaticFile("/", "./public/index.html") // <-- THIS IS THE FIX
}