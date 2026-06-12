package main

import (
	"app/middlewares"
	"app/repository"
	"errors"
	"log"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// ミドルウェア初期化
	middlewares.Init()

	// リポジトリ初期化
	err := repository.Init()

	// エラー処理
	if err != nil {
		log.Fatal(err)
	}

	// Echo instance
	router := echo.New()

	// Middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recover())

	// Routes
	router.GET("/", hello)

	// Start server
	if err := router.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("failed to start server", "error", err)
	}
}

// Handler
func hello(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "Hello, World!")
}
