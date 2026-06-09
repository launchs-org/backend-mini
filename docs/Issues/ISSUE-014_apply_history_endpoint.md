# ISSUE-014 apply-history 取得エンドポイント

## 親 Issue
ISSUE-009

## 概要
apply 履歴の一覧・詳細取得エンドポイントを実装する。

## 実装手順

### 1. `internal/handler/apply_history.go` を作成

```go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/your-org/launchs/internal/model"
)

func (h *Handler) ListApplyHistory(c echo.Context) error {
    deploymentID := c.Param("id")
    var histories []model.ApplyHistory
    h.DB.Where("deployment_id = ?", deploymentID).
        Order("applied_at DESC").
        Find(&histories)
    return c.JSON(http.StatusOK, histories)
}

func (h *Handler) GetApplyHistory(c echo.Context) error {
    var history model.ApplyHistory
    if err := h.DB.First(&history,
        "id = ? AND deployment_id = ?",
        c.Param("history_id"), c.Param("id"),
    ).Error; err != nil {
        return echo.ErrNotFound
    }
    return c.JSON(http.StatusOK, history)
}
```

### 2. ルーティング登録

```go
api.GET("/deployments/:id/apply-history",             h.ListApplyHistory)
api.GET("/deployments/:id/apply-history/:history_id", h.GetApplyHistory)
```

## テスト確認項目

- [ ] apply 後に履歴一覧に1件追加されること
- [ ] 失敗した apply の `error_message` が返ること
- [ ] 詳細取得で `manifests` の中身が返ること
- [ ] 別の deployment の history_id で 404 になること
