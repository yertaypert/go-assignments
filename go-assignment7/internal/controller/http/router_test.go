package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/yertaypert/go-assignment7/internal/entity"
	"github.com/yertaypert/go-assignment7/utils"
)

type fakeUserUseCase struct {
	users map[string]*entity.User
}

func (f *fakeUserUseCase) GetMe(userID string) (*entity.User, error) {
	return f.users[userID], nil
}

func (f *fakeUserUseCase) PromoteUser(userID string) (*entity.User, error) {
	user := f.users[userID]
	if user == nil {
		return nil, http.ErrMissingFile
	}
	user.Role = "admin"
	return user, nil
}

func (f *fakeUserUseCase) LoginUser(user *entity.LoginUserDTO) (string, error) {
	return "", nil
}

func (f *fakeUserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {
	return user, "", nil
}

func TestGetMeUsesJWTUserIdentity(t *testing.T) {
	userAID := uuid.New()
	userBID := uuid.New()

	router := NewRouter(&fakeUserUseCase{
		users: map[string]*entity.User{
			userAID.String(): {
				ID:       userAID,
				Username: "user-a",
				Email:    "usera@example.com",
				Role:     "user",
			},
			userBID.String(): {
				ID:       userBID,
				Username: "user-b",
				Email:    "userb@example.com",
				Role:     "user",
			},
		},
	})

	tokenA, err := utils.GenerateJWT(userAID, "user")
	if err != nil {
		t.Fatalf("generate token A: %v", err)
	}

	tokenB, err := utils.GenerateJWT(userBID, "user")
	if err != nil {
		t.Fatalf("generate token B: %v", err)
	}

	reqA := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	reqA.Header.Set("Authorization", "Bearer "+tokenA)
	recA := httptest.NewRecorder()
	router.ServeHTTP(recA, reqA)

	if recA.Code != http.StatusOK {
		t.Fatalf("user A status = %d, body = %s", recA.Code, recA.Body.String())
	}

	var bodyA struct {
		User entity.UserResponse `json:"user"`
	}
	if err := json.Unmarshal(recA.Body.Bytes(), &bodyA); err != nil {
		t.Fatalf("decode user A response: %v", err)
	}
	if bodyA.User.Email != "usera@example.com" {
		t.Fatalf("user A email = %q", bodyA.User.Email)
	}

	reqB := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	reqB.Header.Set("Authorization", "Bearer "+tokenB)
	recB := httptest.NewRecorder()
	router.ServeHTTP(recB, reqB)

	if recB.Code != http.StatusOK {
		t.Fatalf("user B status = %d, body = %s", recB.Code, recB.Body.String())
	}

	var bodyB struct {
		User entity.UserResponse `json:"user"`
	}
	if err := json.Unmarshal(recB.Body.Bytes(), &bodyB); err != nil {
		t.Fatalf("decode user B response: %v", err)
	}
	if bodyB.User.Email != "userb@example.com" {
		t.Fatalf("user B email = %q", bodyB.User.Email)
	}
}

func TestGetMeRequiresToken(t *testing.T) {
	router := NewRouter(&fakeUserUseCase{users: map[string]*entity.User{}})

	req := httptest.NewRequest(http.MethodGet, "/v1/users/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "{\"error\":\"token required\"}" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestPromoteUserRequiresAdminRole(t *testing.T) {
	targetID := uuid.New()
	router := NewRouter(&fakeUserUseCase{
		users: map[string]*entity.User{
			targetID.String(): {
				ID:       targetID,
				Username: "target",
				Email:    "target@example.com",
				Role:     "user",
			},
		},
	})

	userToken, err := utils.GenerateJWT(uuid.New(), "user")
	if err != nil {
		t.Fatalf("generate user token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/v1/users/promote/"+targetID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+userToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "{\"error\":\"forbidden\"}" {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestPromoteUserWithAdminRole(t *testing.T) {
	targetID := uuid.New()
	router := NewRouter(&fakeUserUseCase{
		users: map[string]*entity.User{
			targetID.String(): {
				ID:       targetID,
				Username: "target",
				Email:    "target@example.com",
				Role:     "user",
			},
		},
	})

	adminToken, err := utils.GenerateJWT(uuid.New(), "admin")
	if err != nil {
		t.Fatalf("generate admin token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPatch, "/v1/users/promote/"+targetID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+adminToken)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Message string              `json:"message"`
		User    entity.UserResponse `json:"user"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.User.Role != "admin" {
		t.Fatalf("role = %q", body.User.Role)
	}
}

func TestRateLimiterAnonymousUsesClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := utils.NewRateLimiter(2, time.Minute)
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "192.0.2.10:1234"
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, body = %s", i+1, rec.Code, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "192.0.2.10:1234"
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}

func TestRateLimiterAuthenticatedUsesJWTUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	limiter := utils.NewRateLimiter(2, time.Minute)
	router := gin.New()
	router.Use(limiter.Middleware())
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	userID := uuid.New()
	token, err := utils.GenerateJWT(userID, "user")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.RemoteAddr = "192.0.2.10:1234"
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d status = %d, body = %s", i+1, rec.Code, rec.Body.String())
		}
	}

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "198.51.100.25:9999"
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}
}
