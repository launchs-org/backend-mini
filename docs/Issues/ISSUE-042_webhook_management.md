# ISSUE-042 Webhook 登録・削除 API

## 親 Issue
ISSUE-041

## 実装手順

### `handler/webhook.go` を作成

```go
func (h *Handler) CreateWebhook(c echo.Context) error {
    deploymentID := c.Param("id")

    // 既に存在する場合は 409
    var existing models.DeploymentWebhook
    if err := h.DB.Where("deployment_id = ?", deploymentID).First(&existing).Error; err == nil {
        return c.JSON(http.StatusConflict, map[string]string{"error": "webhook already exists"})
    }

    // シークレットを自動生成（32バイトのランダム文字列）
    secretBytes := make([]byte, 32)
    rand.Read(secretBytes)
    secret := "whsec_" + hex.EncodeToString(secretBytes)

    webhook := models.DeploymentWebhook{
        DeploymentID: deploymentID,
        Secret:       secret,
    }
    h.DB.Create(&webhook)

    webhookURL := fmt.Sprintf("https://api.launchs.org/webhooks/%s/github", deploymentID)
    return c.JSON(http.StatusCreated, map[string]string{
        "id":          webhook.ID,
        "webhook_url": webhookURL,
        "secret":      secret, // 初回のみ返す
    })
}
```

### ルーティング登録

```go
api.GET("/deployments/:id/webhook",    h.GetWebhook)
api.POST("/deployments/:id/webhook",   h.CreateWebhook)
api.DELETE("/deployments/:id/webhook", h.DeleteWebhook)
```

## テスト確認項目

- [ ] Webhook 作成でシークレットが返ること
- [ ] 同じ deployment に2回 POST すると 409 になること
