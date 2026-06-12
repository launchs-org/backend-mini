package router

import (
	"app/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// New はミドルウェアとルーティングを設定した Echo インスタンスを返す
func New() *echo.Echo {
	router := echo.New() // Echo インスタンスを生成する

	router.Use(middleware.Logger())  // リクエストログミドルウェアを設定する
	router.Use(middleware.Recover()) // パニックリカバリミドルウェアを設定する

	// 認証必須の API グループを作成する
	apiGroup := router.Group("/api/v1", middlewares.RequireAuth)

	// Phase 2 以降で各ハンドラを main.go で初期化してここに登録する
	// apiGroup.GET("/projects", projectHandler.ListProjects)
	_ = apiGroup // 将来のルート追加まで未使用警告を抑制する

	return router
}
