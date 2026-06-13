# ISSUE-013 POST /deployments/:id/apply エンドポイント

## 親 Issue
ISSUE-009

## 概要
apply サービスを呼び出す HTTP エンドポイントを実装する。

## 実装手順

### 1. `handler/deployment.go` に追加

```go
func (h *Handler) ApplyDeployment(c echo.Context) error {
    deploymentID := c.Param("id")

    // deployment の存在確認
    var d models.Deployment
    if err := h.DB.First(&d, "id = ?", deploymentID).Error; err != nil {
        return echo.ErrNotFound
    }

    // deleting 中は apply 不可
    if d.Status == models.DeploymentStatusDeleting {
        return c.JSON(http.StatusConflict, map[string]string{
            "error": "deployment is being deleted",
            "code":  "DEPLOYMENT_DELETING",
        })
    }

    // instance_sizes をロード
    var sizes []models.InstanceSize
    h.DB.Find(&sizes)
    sizeMap := make(map[string]models.InstanceSize)
    for _, s := range sizes {
        sizeMap[s.Size] = s
    }

    svc := &service.ApplyService{
        DB:        h.DB,
        K8s:       h.K8s,
        Generator: &manifest.Generator{InstanceSizes: sizeMap},
    }

    result, err := svc.Apply(c.Request().Context(), deploymentID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": err.Error(),
            "code":  "APPLY_FAILED",
        })
    }

    return c.JSON(http.StatusOK, result)
}
```

### 2. ルーティング登録

```go
api.POST("/deployments/:id/apply", h.ApplyDeployment)
```

## テスト確認項目

- [ ] `POST /apply` で 200 と `apply_history_id` が返ること
- [ ] 存在しない deployment_id で 404 が返ること
- [ ] `status = deleting` の deployment に apply すると 409 が返ること
- [ ] pending_*** が全て空の場合でも apply が通ること（現在値を維持して再 apply）

### repository 層テスト

- [ ] `DeploymentRepository.FindByID` で存在しない ID を渡すと `ErrRecordNotFound` が返ること
- [ ] `DeploymentRepository.FindByID` で `status = deleting` のレコードが正しく取得されること
