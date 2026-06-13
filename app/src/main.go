package main

import (
	"app/config"
	"app/handler"
	"app/k8s"
	"app/middlewares"
	"app/repository"
	"app/router"
	"app/service"
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
	// 設定を初期化する
	cfg := config.NewEnvConfig()

	// ミドルウェア初期化（JWT公開鍵の読み込み）
	middlewares.Init()

	// データベース初期化・マイグレーション
	err := repository.Init()
	if err != nil {
		log.Fatalf("データベースの初期化に失敗しました: %v", err) // DB 初期化失敗時はアプリを終了する
	}

	// k8s クライアント初期化
	k8sClient, err := k8s.NewClient()
	if err != nil {
		log.Fatalf("k8s クライアントの作成に失敗しました: %v", err) // kubeconfig が存在しない場合などにエラーを出す
	}

	// dynamic クライアント初期化（Traefik CRD 用）
	dynamicClient, err := k8s.NewDynamicClient()
	if err != nil {
		log.Fatalf("dynamic クライアントの作成に失敗しました: %v", err) // dynamic クライアント作成失敗時にエラーを出す
	}

	// namespace 一覧を取得して k8s 接続確認を行う
	namespaceList, err := k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Fatalf("k8s クラスターへの接続に失敗しました: %v", err) // クラスター疎通に失敗した場合エラーを出す
	}
	log.Printf("k8s に接続しました: %d 個の namespace が見つかりました", len(namespaceList.Items)) // 接続確認ログを出す

	_ = dynamicClient // 将来のハンドラ初期化まで未使用警告を抑制する

	// quota ハンドラーを DI 組み立てする
	userQuotaRepo := repository.NewUserQuotaRepository(repository.Database) // quota リポジトリを生成する
	quotaServiceImpl := service.NewQuotaService(userQuotaRepo)              // quota サービスを生成する
	userQuotaHandler := handler.NewUserQuotaHandler(quotaServiceImpl)       // quota ハンドラーを生成する

	// Harbor クライアントを初期化する
	harborClient := k8s.NewHarborClient(
		cfg.GetHarborEndpoint(),    // Harbor エンドポイントを設定する
		cfg.GetHarborRobotName(),   // 管理用 robot アカウント名を設定する
		cfg.GetHarborRobotSecret(), // 管理用 robot アカウントのシークレットを設定する
	)

	// project ハンドラーを DI 組み立てする
	projectRepo := repository.NewProjectRepository(repository.Database)                                          // project リポジトリを生成する
	harborCredentialRepo := repository.NewHarborCredentialRepository(repository.Database)                        // harbor credential リポジトリを生成する
	projectServiceImpl := service.NewProjectService(repository.Database, projectRepo, harborCredentialRepo, k8sClient, harborClient) // project サービスを生成する
	projectHandler := handler.NewProjectHandler(projectServiceImpl)                                              // project ハンドラーを生成する

	// deployment ハンドラーを DI 組み立てする
	deploymentRepo := repository.NewDeploymentRepository(repository.Database)                     // deployment リポジトリを生成する
	serviceRepo := repository.NewServiceRepository(repository.Database)                           // service リポジトリを生成する
	deploymentServiceImpl := service.NewDeploymentService(deploymentRepo, serviceRepo)            // deployment サービスを生成する
	deploymentHandler := handler.NewDeploymentHandler(deploymentServiceImpl)                      // deployment ハンドラーを生成する

	// ルーターを生成してサーバーを起動する
	echoRouter := router.New(router.RouterOptions{
		UserQuotaHandler:  userQuotaHandler,  // quota ハンドラーを注入する
		ProjectHandler:    projectHandler,    // project ハンドラーを注入する
		DeploymentHandler: deploymentHandler, // deployment ハンドラーを注入する
	})
	if err := echoRouter.Start(":" + cfg.GetServerPort()); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("サーバーの起動に失敗しました", "error", err) // サーバー起動失敗時にエラーログを出す
	}
}
