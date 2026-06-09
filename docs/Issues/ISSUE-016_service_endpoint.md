# ISSUE-016 GET / PUT /service エンドポイント

## 親 Issue
ISSUE-015

## 概要
Service の取得・更新エンドポイントを実装する。PUT では `ports` を `pending_ports` に書き込む。

## 実装手順

### 1. `internal/handler/service.go` を作成

```go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/your-org/launchs/internal/model"
    "gorm.io/datatypes"
    "encoding/json"
)

func (h *Handler) GetService(c echo.Context) error {
    deploymentID := c.Param("id")
    var svc model.Service
    if err := h.DB.Where("deployment_id = ?", deploymentID).First(&svc).Error; err != nil {
        return echo.ErrNotFound
    }
    return c.JSON(http.StatusOK, svc)
}

func (h *Handler) UpdateService(c echo.Context) error {
    deploymentID := c.Param("id")
    var svc model.Service
    if err := h.DB.Where("deployment_id = ?", deploymentID).First(&svc).Error; err != nil {
        return echo.ErrNotFound
    }

    var req struct {
        Ports []struct {
            Protocol string `json:"protocol"`
            Port     int    `json:"port"`
        } `json:"ports"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    // バリデーション: port 範囲・重複チェック
    seen := map[string]bool{}
    for _, p := range req.Ports {
        if p.Port < 1 || p.Port > 65535 {
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": "port must be between 1 and 65535",
                "code":  "INVALID_PORT",
            })
        }
        key := fmt.Sprintf("%s:%d", p.Protocol, p.Port)
        if seen[key] {
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": fmt.Sprintf("duplicate port: %s %d", p.Protocol, p.Port),
                "code":  "DUPLICATE_PORT",
            })
        }
        seen[key] = true
    }

    portsJSON, _ := json.Marshal(req.Ports)
    svc.PendingPorts = datatypes.JSON(portsJSON)
    h.DB.Save(&svc)
    return c.JSON(http.StatusOK, svc)
}
```

### 2. ルーティング登録

```go
api.GET("/deployments/:id/service", h.GetService)
api.PUT("/deployments/:id/service", h.UpdateService)
```

## テスト確認項目

- [ ] `PUT /service` で `ports` を送ると `pending_ports` に入ること
- [ ] port が 0 や 65536 の場合にバリデーションエラーになること
- [ ] 同一 protocol + port の重複で 400 になること
- [ ] ports が空配列の場合にエラーになること
