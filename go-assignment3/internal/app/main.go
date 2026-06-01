package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/joho/godotenv"
	"github.com/yertaypert/go-assignment3/internal/handler"
	"github.com/yertaypert/go-assignment3/internal/middleware"
	"github.com/yertaypert/go-assignment3/internal/repository"
	"github.com/yertaypert/go-assignment3/internal/repository/_postgres"
	"github.com/yertaypert/go-assignment3/internal/usecase"
	"github.com/yertaypert/go-assignment3/pkg/modules"
)

func Run() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Config & DB connection
	dbConfig := initPostgreConfig()
	_postgre := _postgres.NewPGXDialect(ctx, dbConfig)

	// Repository Layer
	repositories := repository.NewRepositories(_postgre)

	// Usecase Layer
	userUsecase := usecase.NewUserUsecase(repositories.UserRepository)

	// Handler Layer
	userHandler := handler.NewUserHandler(userUsecase)

	// Setup the handlers
	getUsersHandler := http.HandlerFunc(userHandler.GetUsers)
	getUserByIDHandler := http.HandlerFunc(userHandler.GetUserByID)
	createUserHandler := http.HandlerFunc(userHandler.CreateUser)
	updateUserHandler := http.HandlerFunc(userHandler.UpdateUser)
	deleteUserHandler := http.HandlerFunc(userHandler.DeleteUser)

	// Wrap handlers in Middleware and Register them
	// For GET /users
	http.Handle("/users", middleware.Logging(middleware.Auth(getUsersHandler)))

	// For /users/get
	http.Handle("/users/get", middleware.Logging(middleware.Auth(getUserByIDHandler)))

	// For /users/create
	http.Handle("/users/create", middleware.Logging(middleware.Auth(createUserHandler)))

	// For /users/update
	http.Handle("/users/update", middleware.Logging(middleware.Auth(updateUserHandler)))

	// For /users/delete
	http.Handle("/users/delete", middleware.Logging(middleware.Auth(deleteUserHandler)))

	// HealthCheck
	http.Handle("/health", middleware.Logging(http.HandlerFunc(handler.Healthcheck)))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default fallback
	}
	println("Server starting on :", port, "...")
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}

func initPostgreConfig() *modules.PostgreConfig {
	return &modules.PostgreConfig{
		Host:        os.Getenv("DB_HOST"),
		Port:        os.Getenv("DB_PORT"),
		Username:    os.Getenv("DB_USERNAME"),
		Password:    os.Getenv("DB_PASSWORD"),
		DBName:      os.Getenv("DB_NAME"),
		SSLMode:     "disable",
		ExecTimeout: 5 * time.Second,
	}
}
