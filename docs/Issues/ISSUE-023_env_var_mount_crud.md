# ISSUE-023 Env Var Mount CRUD エンドポイント

## 親 Issue
ISSUE-021

## 概要
env_var を deployment にマウント・アンマウントするエンドポイントを実装する。
override_key の設定と pending_override_key への書き込みを行う。

## 実装手順

### 1. `internal/handler/env_var_mount.go` を作成

```go
package handler

func (h *Handler) ListEnvMounts(c echo.Context) error {
    deploymentID := c.Param("id")
    var mounts []model.EnvVarMount
    h.DB.Preload("EnvVar").
        Where("deployment_id = ? AND status != ?", deploymentID, "deleting").
        Find(&mounts)
    return c.JSON(http.StatusOK, mounts)
}

func (h *Handler) CreateEnvMount(c echo.Context) error {
    deploymentID := c.Param("id")
    var req struct {
        EnvVarID    string  `json:"env_var_id"`
        OverrideKey *string `json:"override_key"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    // 重複チェック
    var count int64
    h.DB.Model(&model.EnvVarMount{}).
        Where("env_var_id = ? AND deployment_id = ?", req.EnvVarID, deploymentID).
        Count(&count)
    if count > 0 {
        return c.JSON(http.StatusConflict, map[string]string{
            "error": "env_var already mounted",
            "code":  "ALREADY_MOUNTED",
        })
    }

    overrideKey := ""
    if req.OverrideKey != nil { overrideKey = *req.OverrideKey }

    mount := model.EnvVarMount{
        EnvVarID:     req.EnvVarID,
        DeploymentID: deploymentID,
        OverrideKey:  overrideKey,
        Status:       model.EnvVarMountStatusPending,
    }
    h.DB.Create(&mount)
    return c.JSON(http.StatusCreated, mount)
}

func (h *Handler) UpdateEnvMount(c echo.Context) error {
    var mount model.EnvVarMount
    if err := h.DB.First(&mount, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    var req struct {
        OverrideKey *string `json:"override_key"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }
    if req.OverrideKey != nil {
        mount.PendingOverrideKey = *req.OverrideKey
        h.DB.Save(&mount)
    }
    return c.JSON(http.StatusOK, mount)
}

func (h *Handler) DeleteEnvMount(c echo.Context) error {
    var mount model.EnvVarMount
    if err := h.DB.First(&mount, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    h.DB.Model(&mount).Update("status", model.EnvVarMountStatusDeleting)
    return c.NoContent(http.StatusNoContent)
}
```

### 2. ルーティング登録

```go
api.GET("/deployments/:id/env-mounts",  h.ListEnvMounts)
api.POST("/deployments/:id/env-mounts", h.CreateEnvMount)
api.PUT("/env-mounts/:id",              h.UpdateEnvMount)
api.DELETE("/env-mounts/:id",           h.DeleteEnvMount)
```

## テスト確認項目

- [ ] 同一 deployment に同じ env_var_id を2回 mount すると 409 になること
- [ ] `PUT /env-mounts/:id` で `override_key` を送ると `pending_override_key` に入ること
- [ ] `DELETE /env-mounts/:id` で `status = deleting` になること
