# ISSUE-004 ディレクトリ構成・Router 整備

## 親 Issue
ISSUE-001

## 概要
全フェーズで使うルーティング基盤と依存性注入の仕組みを整える。

## 実装手順

### 1. Handler に渡す依存性をまとめる構造体を作成

`internal/handler/handler.go`

```go
package handler

import (
    "gorm.io/gorm"
    "k8s.io/client-go/dynamic"
    "k8s.io/client-go/kubernetes"
    "github.com/your-org/launchs/internal/config"
)

type Handler struct {
    DB            *gorm.DB
    K8s           *kubernetes.Clientset
    DynamicClient dynamic.Interface
    Config        *config.Config
}

func New(db *gorm.DB, k8s *kubernetes.Clientset, dc dynamic.Interface, cfg *config.Config) *Handler {
    return &Handler{
        DB:            db,
        K8s:           k8s,
        DynamicClient: dc,
        Config:        cfg,
    }
}
```

### 2. `internal/router/router.go` を作成

```go
package router

import (
    "github.com/labstack/echo/v4"
    "github.com/labstack/echo/v4/middleware"
    "github.com/your-org/launchs/internal/handler"
)

func New(h *handler.Handler) *echo.Echo {
    e := echo.New()
    e.Use(middleware.Logger())
    e.Use(middleware.Recover())

    api := e.Group("/api/v1")

    // Phase2 以降で各グループを追加していく
    // api.GET("/projects", h.ListProjects)

    return e
}
```

### 3. `cmd/api/main.go` を整理

```go
package main

import (
    "log"
    "github.com/your-org/launchs/internal/config"
    "github.com/your-org/launchs/internal/db"
    "github.com/your-org/launchs/internal/handler"
    "github.com/your-org/launchs/internal/k8s"
    "github.com/your-org/launchs/internal/router"
)

func main() {
    cfg := config.Load()

    database, err := db.New(cfg.DatabaseDSN)
    if err != nil {
        log.Fatalf("db: %v", err)
    }
    if err := db.AutoMigrate(database); err != nil {
        log.Fatalf("migrate: %v", err)
    }
    if err := db.SeedInstanceSizes(database); err != nil {
        log.Fatalf("seed: %v", err)
    }

    k8sClient, err := k8s.NewClient()
    if err != nil {
        log.Fatalf("k8s: %v", err)
    }
    dynamicClient, err := k8s.NewDynamicClient()
    if err != nil {
        log.Fatalf("dynamic: %v", err)
    }

    h := handler.New(database, k8sClient, dynamicClient, cfg)
    e := router.New(h)
    e.Logger.Fatal(e.Start(":" + cfg.ServerPort))
}
```

## テスト確認項目

- [ ] `go build ./...` がエラーなく通ること
- [ ] サーバーが起動すること
- [ ] 存在しないパスへのリクエストで 404 が返ること
