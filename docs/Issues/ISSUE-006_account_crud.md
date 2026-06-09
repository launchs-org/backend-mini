# ISSUE-006 Account CRUD

## 親 Issue
ISSUE-005

## 概要
Account の作成・取得と AccountQuota の取得・更新を実装する。

## 実装手順

### 1. `internal/handler/account.go` を作成

```go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "github.com/your-org/launchs/internal/model"
)

func (h *Handler) CreateAccount(c echo.Context) error {
    var req struct {
        Name string `json:"name" validate:"required"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    account := model.Account{Name: req.Name, Status: "active"}
    if err := h.DB.Create(&account).Error; err != nil {
        return echo.ErrInternalServerError
    }

    // quota をデフォルト値で作成
    quota := model.AccountQuota{
        AccountID:                account.ID,
        MaxProjects:              5,
        MaxDeployments:           20,
        MaxReplicasPerDeployment: 5,
        MaxVolumeMB:              10240,
    }
    h.DB.Create(&quota)

    return c.JSON(http.StatusCreated, account)
}

func (h *Handler) GetQuota(c echo.Context) error {
    accountID := c.Param("id")

    var quota model.AccountQuota
    if err := h.DB.Where("account_id = ?", accountID).First(&quota).Error; err != nil {
        return echo.ErrNotFound
    }

    // 現在の使用量を集計
    var currentDeployments int64
    h.DB.Model(&model.Deployment{}).
        Joins("JOIN projects ON projects.id = deployments.project_id").
        Where("projects.account_id = ? AND deployments.status != ?", accountID, "deleted").
        Count(&currentDeployments)

    var currentVolumeMB int64
    h.DB.Model(&model.Volume{}).
        Joins("JOIN projects ON projects.id = volumes.project_id").
        Where("projects.account_id = ? AND volumes.status != ?", accountID, "deleted").
        Select("COALESCE(SUM(size_mb), 0)").Scan(&currentVolumeMB)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "max_projects":                quota.MaxProjects,
        "max_deployments":             quota.MaxDeployments,
        "max_replicas_per_deployment": quota.MaxReplicasPerDeployment,
        "max_volume_mb":               quota.MaxVolumeMB,
        "current_deployments":         currentDeployments,
        "current_volume_mb":           currentVolumeMB,
    })
}

func (h *Handler) UpdateQuota(c echo.Context) error {
    accountID := c.Param("id")
    var req struct {
        MaxProjects              *int `json:"max_projects"`
        MaxDeployments           *int `json:"max_deployments"`
        MaxReplicasPerDeployment *int `json:"max_replicas_per_deployment"`
        MaxVolumeMB              *int `json:"max_volume_mb"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    var quota model.AccountQuota
    if err := h.DB.Where("account_id = ?", accountID).First(&quota).Error; err != nil {
        return echo.ErrNotFound
    }

    if req.MaxProjects != nil              { quota.MaxProjects = *req.MaxProjects }
    if req.MaxDeployments != nil           { quota.MaxDeployments = *req.MaxDeployments }
    if req.MaxReplicasPerDeployment != nil { quota.MaxReplicasPerDeployment = *req.MaxReplicasPerDeployment }
    if req.MaxVolumeMB != nil              { quota.MaxVolumeMB = *req.MaxVolumeMB }

    h.DB.Save(&quota)
    return c.JSON(http.StatusOK, quota)
}
```

### 2. ルーティング登録

`internal/router/router.go` に追加:

```go
api.POST("/accounts", h.CreateAccount)
api.GET("/accounts/:id/quota", h.GetQuota)
api.PUT("/accounts/:id/quota", h.UpdateQuota)
```

## テスト確認項目

- [ ] `POST /accounts` で Account と AccountQuota が作成されること
- [ ] `GET /accounts/:id/quota` で使用量も含めて返ること
- [ ] `PUT /accounts/:id/quota` で部分更新できること
- [ ] 存在しない account_id で 404 が返ること
