# ISSUE-043 GitHub push イベント受信・apply トリガー

## 親 Issue
ISSUE-041

## 実装手順

### `service/webhook.go` を作成

```go
func HandleGithubPush(ctx context.Context, db *gorm.DB, deploymentID, signature string, body []byte) error {
    // 1. Webhook シークレット取得
    var webhook models.DeploymentWebhook
    if err := db.Where("deployment_id = ?", deploymentID).First(&webhook).Error; err != nil {
        return fmt.Errorf("webhook not found")
    }

    // 2. HMAC-SHA256 署名検証
    mac := hmac.New(sha256.New, []byte(webhook.Secret))
    mac.Write(body)
    expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    if !hmac.Equal([]byte(signature), []byte(expected)) {
        return fmt.Errorf("invalid signature")
    }

    // 3. push イベントのブランチを取得
    var payload struct {
        Ref string `json:"ref"` // "refs/heads/main"
        HeadCommit struct {
            ID      string `json:"id"`
            Message string `json:"message"`
        } `json:"head_commit"`
    }
    json.Unmarshal(body, &payload)

    pushedBranch := strings.TrimPrefix(payload.Ref, "refs/heads/")

    // 4. deployment の github_branch と一致するか確認
    var d models.Deployment
    db.First(&d, "id = ?", deploymentID)
    if d.GithubBranch != pushedBranch {
        return nil // スキップ
    }

    // 5. pending_github_commit_sha を push の SHA に更新して apply
    db.Model(&d).Update("pending_github_commit_sha", payload.HeadCommit.ID)

    applySvc := &ApplyService{DB: db}
    _, err := applySvc.Apply(ctx, deploymentID)
    return err
}
```

### ルーティング登録（/api/v1 の外）

```go
e.POST("/webhooks/:deployment_id/github", h.ReceiveGithubWebhook)
```

## テスト確認項目

- [ ] 正しい署名で apply がトリガーされること
- [ ] 不正な署名で 401 が返ること
- [ ] branch が一致しない push では apply がトリガーされないこと
- [ ] push の commit SHA が `pending_github_commit_sha` に設定されること
