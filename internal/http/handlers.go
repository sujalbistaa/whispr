package http

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"
	// "time" // This import is not needed here

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"gorm.io/gorm"

	"github.com/sujalbistaa/whispr/internal/models"
	"github.com/sujalbistaa/whispr/internal/ws"
)

// --- Configuration Constants ---

const (
	// Max content length for a post (in characters)
	maxPostLength = 1000
	// Rate limit: requests per second
	rateLimitRPS = 1.0 / 3.0 // 1 request every 3 seconds
	// Rate limit: burst size
	rateLimitBurst = 1
)

// --- Structs for request binding ---

// CreatePostInput defines the expected JSON body for creating a post.
type CreatePostInput struct {
	Content string `json:"content" binding:"required,min=1,max=1000"`
}

// VoteInput defines the expected JSON body for voting.
type VoteInput struct {
	Value int `json:"value" binding:"required,oneof=-1 1"` // Must be 1 or -1
}

// --- WebSocket Payloads ---

// WsMessage is the envelope for all our real-time messages.
type WsMessage struct {
	Type    string      `json:"type"`    // "new_post", "vote_update", "delete_post"
	Payload interface{} `json:"payload"` // The data (e.g., a Post object)
}

// --- Rate Limiter ---

// IPRateLimiter holds a map of IP addresses to rate limiters
type IPRateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit // requests per second
	burst    int        // max burst size
}

// NewIPRateLimiter creates a new rate limiter
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		visitors: make(map[string]*rate.Limiter),
		mu:       sync.RWMutex{},
		rps:      r,
		burst:    b,
	}
}

// GetLimiter returns the rate limiter for a given IP
func (rl *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rps, rl.burst)
		rl.visitors[ip] = limiter
	}
	return limiter
}

// RateLimitMiddleware is the Gin middleware for our rate limiter
func RateLimitMiddleware(limiter *IPRateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please wait."})
			return
		}
		c.Next()
	}
}

// --- Handlers ---

// Env holds dependencies for our handlers (DB, Hub)
type Env struct {
	DB  *gorm.DB
	Hub *ws.Hub
}

// GetPosts retrieves all non-hidden posts, ordered by creation date
func (e *Env) GetPosts(c *gin.Context) {
	var posts []models.Post
	if err := e.DB.Order("created_at desc").Where("hidden = ?", false).Find(&posts).Error; err != nil {
		log.Printf("Error fetching posts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// GetTrendingPosts retrieves top posts.
func (e *Env) GetTrendingPosts(c *gin.Context) {
	var posts []models.Post
	if err := e.DB.Order("score desc, created_at desc").Where("hidden = ?", false).Limit(20).Find(&posts).Error; err != nil {
		log.Printf("Error fetching trending posts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch posts"})
		return
	}
	c.JSON(http.StatusOK, posts)
}

// CreatePost creates a new post, saves it, and broadcasts it
func (e *Env) CreatePost(c *gin.Context) {
	var input CreatePostInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	post := models.Post{
		Content: input.Content,
		Score:   1, 
	}

	if err := e.DB.Create(&post).Error; err != nil {
		log.Printf("Error creating post: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create post"})
		return
	}

	msg := WsMessage{Type: "new_post", Payload: post}
	e.broadcastMessage(msg)

	c.JSON(http.StatusCreated, post)
}

// VoteOnPost applies a +1 or -1 vote to a post
func (e *Env) VoteOnPost(c *gin.Context) {
	var input VoteInput

	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	var post models.Post
	var newScore int

	err = e.DB.Transaction(func(tx *gorm.DB) error {
		// Find the post, lock it, and check that it's not hidden
		if err := tx.Set("gorm:query_option", "FOR UPDATE").Where("hidden = ?", false).First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("post not found")
			}
			return err
		}

		vote := models.Vote{
			PostID: uint(postID),
			Value:  input.Value,
		}
		if err := tx.Create(&vote).Error; err != nil {
			return errors.New("failed to record vote")
		}

		newScore = post.Score + input.Value
		if err := tx.Model(&post).Update("score", newScore).Error; err != nil {
			return errors.New("failed to update post score")
		}

		return nil
	})

	if err != nil {
		log.Printf("Error in vote transaction: %v", err)
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process vote"})
		}
		return
	}

	payload := gin.H{"id": post.ID, "newScore": newScore}
	msg := WsMessage{Type: "vote_update", Payload: payload}
	e.broadcastMessage(msg)

	c.JSON(http.StatusOK, payload)
}

// DeletePost performs a "soft delete" on a post, marking it as hidden.
func (e *Env) DeletePost(c *gin.Context) {
	postID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid post ID"})
		return
	}

	var post models.Post

	err = e.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&post, postID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("post not found")
			}
			return err
		}

		if err := tx.Model(&post).Update("hidden", true).Error; err != nil {
			return errors.New("failed to hide post")
		}

		return nil
	})

	if err != nil {
		log.Printf("Error in delete transaction: %v", err)
		if err.Error() == "post not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "Post not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete post"})
		}
		return
	}

	payload := gin.H{"id": post.ID}
	msg := WsMessage{Type: "delete_post", Payload: payload}
	e.broadcastMessage(msg)

	c.JSON(http.StatusOK, gin.H{"message": "Post hidden successfully"})
}

// broadcastMessage is a helper to marshal and send a WsMessage
func (e *Env) broadcastMessage(msg WsMessage) {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		log.Printf("Error marshalling WS message: %v", err)
		return
	}
	e.Hub.Broadcast <- jsonMsg
}