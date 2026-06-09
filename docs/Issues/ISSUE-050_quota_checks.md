# ISSUE-050 deployment 作成・更新時の quota チェック

## 親 Issue
ISSUE-049

## 実装手順

### `internal/service/quota.go` を完成

```go
func CheckDeploymentQuota(db *gorm.DB, accountID string) error {
    var quota model.AccountQuota
    db.Where("account_id = ?", accountID).First(&quota)

    var count int64
    db.Model(&model.Deployment{}).
        Joins("JOIN projects ON projects.id = deployments.project_id").
        Where("projects.account_id = ? AND deployments.status NOT IN ?",
            accountID, []string{"deleted", "deleting"}).
        Count(&count)

    if int(count) >= quota.MaxDeployments {
        return fmt.Errorf("deployment quota exceeded: max=%d", quota.MaxDeployments)
    }
    return nil
}

func CheckReplicasQuota(db *gorm.DB, accountID string, replicas int32) error {
    var quota model.AccountQuota
    db.Where("account_id = ?", accountID).First(&quota)

    if int(replicas) > quota.MaxReplicasPerDeployment {
        return fmt.Errorf("replicas exceed limit: max=%d", quota.MaxReplicasPerDeployment)
    }
    return nil
}
```

### 各ハンドラーに組み込む

- `CreateDeployment`: `CheckDeploymentQuota` を呼ぶ
- `UpdateDeployment` で replicas 変更時: `CheckReplicasQuota` を呼ぶ
- `apply` 時にも replicas の quota チェックを行う

## テスト確認項目

- [ ] `max_deployments` を超える deployment 作成で 400 になること
- [ ] `max_replicas_per_deployment` を超える replicas 設定で 400 になること
- [ ] quota 更新後に新しい制限が反映されること
