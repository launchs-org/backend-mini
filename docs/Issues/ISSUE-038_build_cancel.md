# ISSUE-038 ビルドキャンセル

## 親 Issue
ISSUE-035

## 概要
apply 時に既に building 中のビルドがある場合、k8s Job を削除してキャンセルする。

## 実装手順

### `service/build.go` を作成

```go
package service

func CancelRunningBuild(ctx context.Context, db *gorm.DB, k8sClient *kubernetes.Clientset, deploymentID string) error {
    // building 中のビルドを探す
    var build models.DeploymentBuild
    err := db.Where("deployment_id = ? AND status = ?", deploymentID, models.BuildStatusBuilding).
        First(&build).Error
    if err != nil { return nil } // building 中がなければ何もしない

    // k8s Job を削除
    if build.K8sJobName != "" {
        k8s.DeleteJob(ctx, k8sClient, build.K8sJobName)
    }

    // status を cancelled に更新
    db.Model(&build).Update("status", models.BuildStatusCancelled)

    // current_build_id は空のまま（次のビルドで上書き）
    return nil
}
```

## テスト確認項目

- [ ] building 中に apply を再度叩くと既存 Job が削除されること
- [ ] キャンセルされたビルドの `status = cancelled` になること
- [ ] building 中でない場合はキャンセル処理がスキップされること
