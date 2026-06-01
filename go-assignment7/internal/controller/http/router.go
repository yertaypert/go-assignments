package http

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	v1 "github.com/yertaypert/go-assignment7/internal/controller/http/v1"
	"github.com/yertaypert/go-assignment7/internal/usecase"
	"github.com/yertaypert/go-assignment7/utils"
)

func NewRouter(userUC usecase.UserInterface) *gin.Engine {
	router := gin.Default()
	router.Use(utils.NewRateLimiter(5, time.Minute).Middleware())

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiV1 := router.Group("/api/v1")
	v1.RegisterUserRoutes(apiV1, userUC)

	plainV1 := router.Group("/v1")
	v1.RegisterUserRoutes(plainV1, userUC)

	return router
}
