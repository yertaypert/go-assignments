package utils

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getJWTSecret())

type rateLimitEntry struct {
	count       int
	windowStart time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	clients map[string]rateLimitEntry
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password),
		bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(hashedPassword, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword),
		[]byte(password)) == nil
}

func GenerateJWT(userID uuid.UUID, role string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, err := parseTokenClaims(c.GetHeader("Authorization"))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": err.Error()})
			return
		}

		userID, _ := claims["user_id"].(string)
		role, _ := claims["role"].(string)
		c.Set("userID", userID)
		c.Set("role", role)
		c.Next()
	}
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		clients: make(map[string]rateLimitEntry),
	}
}

func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		identifier := r.identifierForRequest(c)
		now := time.Now()

		r.mu.Lock()
		entry, exists := r.clients[identifier]
		if !exists || now.Sub(entry.windowStart) >= r.window {
			r.clients[identifier] = rateLimitEntry{
				count:       1,
				windowStart: now,
			}
			r.mu.Unlock()
			c.Next()
			return
		}

		if entry.count >= r.limit {
			r.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			return
		}

		entry.count++
		r.clients[identifier] = entry
		r.mu.Unlock()
		c.Next()
	}
}

func RoleMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleValue, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		role, ok := roleValue.(string)
		if !ok || strings.TrimSpace(role) != requiredRole {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}

func (r *RateLimiter) identifierForRequest(c *gin.Context) string {
	if claims, err := parseTokenClaims(c.GetHeader("Authorization")); err == nil {
		if userID, ok := claims["user_id"].(string); ok && strings.TrimSpace(userID) != "" {
			return "user:" + strings.TrimSpace(userID)
		}
	}

	return "ip:" + c.ClientIP()
}

func parseTokenClaims(authHeader string) (jwt.MapClaims, error) {
	tokenStr := strings.TrimSpace(authHeader)
	if tokenStr == "" {
		return nil, fmt.Errorf("token required")
	}

	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func getJWTSecret() string {
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		return secret
	}

	return "dev-secret"
}
