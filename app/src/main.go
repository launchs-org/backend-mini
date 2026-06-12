package main

import (
	"app/k8s"
	"app/middlewares"
	"app/repository"
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	// k8s クライアント初期化
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("k8s クライアントの作成に失敗しました: %v", err) // kubeconfig が存在しない場合などにエラーを出す
	}

	// dynamic クライアント初期化（Traefik CRD 用）
	_, err = k8s.NewDynamicClient()
	if err != nil {
		log.Fatalf("dynamic クライアントの作成に失敗しました: %v", err) // dynamic クライアント作成失敗時にエラーを出す
	}

	// namespace 一覧を取得して k8s 接続確認を行う
	namespaceList, err := k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("k8s クラスターへの接続に失敗しました: %v", err) // クラスター疎通に失敗した場合エラーを出す
	}
	log.Printf("k8s に接続しました: %d 個の namespace が見つかりました", len(namespaceList.Items)) // 接続確認ログを出す

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
