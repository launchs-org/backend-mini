# ISSUE-040 ビルド履歴・ログ取得エンドポイント

## 親 Issue
ISSUE-035

## 実装手順

### `handler/build.go` を作成

```go
func (h *Handler) ListBuilds(c echo.Context) error {
    deploymentID := c.Param("id")
    var builds []models.DeploymentBuild
    h.DB.Where("deployment_id = ?", deploymentID).
        Order("created_at DESC").Find(&builds)
    return c.JSON(http.StatusOK, builds)
}

func (h *Handler) GetBuildLog(c echo.Context) error {
    buildID := c.Param("build_id")
    var build models.DeploymentBuild
    if err := h.DB.First(&build, "id = ?", buildID).Error; err != nil {
        return echo.ErrNotFound
    }

    since := c.QueryParam("since")
    until := c.QueryParam("until")
    log := filterLog(build.BuildLog, since, until)

    return c.JSON(http.StatusOK, map[string]string{"log": log})
}
```

### ルーティング登録

```go
api.GET("/deployments/:id/builds",                  h.ListBuilds)
api.GET("/deployments/:id/builds/:build_id/logs",   h.GetBuildLog)
```

## テスト確認項目

- [ ] ビルド履歴が新しい順で返ること
- [ ] since / until でログがフィルタされること
- [ ] 存在しない build_id で 404 が返ること

### repository 層テスト

- [ ] `DeploymentBuildRepository.FindAllByDeploymentID` で全履歴が新しい順で返ること
- [ ] `BuildLogRepository.FindByBuildID` で `since` / `until` によるフィルタが正しく動作すること
- [ ] `DeploymentBuildRepository.FindByID` で存在しない ID を渡すと `ErrRecordNotFound` が返ること
