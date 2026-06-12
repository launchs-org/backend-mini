# ISSUE-022 Env Var CRUD エンドポイント

## 親 Issue
ISSUE-021

## 概要
Project に属する env_var の作成・取得・更新・削除を実装する。
is_secret=true の場合 GET で value を "***" にマスクする。

## 実装手順

### 1. `handler/env_var.go` を作成

```go
package handler

func (h *Handler) ListEnvVars(c echo.Context) error {
    projectID := c.Param("id")
    var vars []models.EnvVar
    h.DB.Where("project_id = ? AND status != ?", projectID, "deleted").Find(&vars)

    // is_secret=true の value をマスク
    for i := range vars {
        if vars[i].IsSecret {
            vars[i].Value = "***"
        }
    }
    return c.JSON(http.StatusOK, vars)
}

func (h *Handler) CreateEnvVar(c echo.Context) error {
    projectID := c.Param("id")
    var req struct {
        Key      string `json:"key"`
        Value    string `json:"value"`
        IsSecret bool   `json:"is_secret"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    // key のバリデーション（英数字・アンダースコアのみ）
    for _, ch := range req.Key {
        if !((ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') ||
            (ch >= '0' && ch <= '9') || ch == '_') {
            return c.JSON(http.StatusBadRequest, map[string]string{
                "error": "key must be alphanumeric or underscore",
                "code":  "INVALID_KEY",
            })
        }
    }

    ev := models.EnvVar{
        ProjectID: projectID,
        Key:       req.Key,
        Value:     req.Value,
        IsSecret:  req.IsSecret,
        Status:    "active",
    }
    h.DB.Create(&ev)
    return c.JSON(http.StatusCreated, ev)
}

func (h *Handler) UpdateEnvVar(c echo.Context) error {
    var ev models.EnvVar
    if err := h.DB.First(&ev, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    var req struct {
        Value *string `json:"value"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }
    if req.Value != nil {
        ev.Value = *req.Value
        h.DB.Save(&ev)
    }
    return c.JSON(http.StatusOK, ev)
}

func (h *Handler) DeleteEnvVar(c echo.Context) error {
    var ev models.EnvVar
    if err := h.DB.First(&ev, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    // mount されている場合は削除不可
    var count int64
    h.DB.Model(&models.EnvVarMount{}).Where("env_var_id = ?", ev.ID).Count(&count)
    if count > 0 {
        return c.JSON(http.StatusConflict, map[string]string{
            "error": "env_var is mounted to deployments",
            "code":  "ENV_VAR_MOUNTED",
        })
    }
    h.DB.Model(&ev).Update("status", "deleting")
    return c.NoContent(http.StatusNoContent)
}
```

### 2. ルーティング登録

```go
api.GET("/projects/:id/env-vars",  h.ListEnvVars)
api.POST("/projects/:id/env-vars", h.CreateEnvVar)
api.PUT("/env-vars/:id",           h.UpdateEnvVar)
api.DELETE("/env-vars/:id",        h.DeleteEnvVar)
```

## テスト確認項目

- [ ] `is_secret=true` の env_var を GET すると value が `"***"` になること
- [ ] key に記号が含まれると 400 になること
- [ ] mount されている env_var を DELETE すると 409 になること
- [ ] value の PUT が即時反映されること
