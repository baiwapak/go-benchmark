package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func CORS(allowOrigins []string) gin.HandlerFunc {
	allowAll := false
	set := map[string]struct{}{}
	for _, o := range allowOrigins {
		if o == "*" {
			allowAll = true
		}
		set[o] = struct{}{}
	}
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := set[origin]; ok {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RateLimit limits requests per IP per window for a route group.
func RateLimit(perMin int) gin.HandlerFunc {
	if perMin < 1 {
		perMin = 10
	}
	type bucket struct {
		count int
		reset time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()
		mu.Lock()
		b, ok := buckets[ip]
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(time.Minute)}
			buckets[ip] = b
		}
		b.count++
		over := b.count > perMin
		mu.Unlock()
		if over {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}

func ClientWantsJSON(c *gin.Context) bool {
	accept := c.GetHeader("Accept")
	return strings.Contains(accept, "application/json") ||
		strings.HasPrefix(c.Request.URL.Path, "/api/")
}
