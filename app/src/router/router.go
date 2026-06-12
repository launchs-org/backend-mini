package router

import (
	"app/handler"
	"app/middlewares"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// RouterOptions はルーター生成に必要なハンドラーをまとめた構造体
type RouterOptions struct {
	UserQuotaHandler *handler.UserQuotaHandler // quota ハンドラー
}

// New はミドルウェアとルーティングを設定した Echo インスタンスを返す
func New(opts RouterOptions) *echo.Echo {
	router := echo.New() // Echo インスタンスを生成する

	router.Use(middleware.Logger())  // リクエストログミドルウェアを設定する
	router.Use(middleware.Recover()) // パニックリカバリミドルウェアを設定する

	// 認証必須の API グループを作成する
	apiGroup := router.Group("/api/v1", middlewares.RequireAuth)

	// quota エンドポイントを登録する
	apiGroup.GET("/users/quota", opts.UserQuotaHandler.GetQuota)    // quota 取得エンドポイント
	apiGroup.PUT("/users/quota", opts.UserQuotaHandler.UpdateQuota) // quota 更新エンドポイント

	return router
}
