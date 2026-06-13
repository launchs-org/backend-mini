# ISSUE-045 Deployment 削除フロー

## 親 Issue
ISSUE-044

## 実装手順

### `service/delete.go` を作成

```go
package service

func DeleteDeploymentResources(ctx context.Context, db *gorm.DB, k8sClient *kubernetes.Clientset, dc dynamic.Interface, deploymentID string) error {
    var d models.Deployment
    if err := db.Preload("Service").First(&d, "id = ?", deploymentID).Error; err != nil {
        return err
    }
    var project models.Project
    db.First(&project, "id = ?", d.ProjectID)

    // k8s リソースを削除
    k8s.DeleteDeployment(ctx, k8sClient, project.Namespace, d.Name)
    k8s.DeleteService(ctx, k8sClient, project.Namespace, d.Name)
    k8s.DeleteIngressRoute(ctx, dc, project.Namespace, d.Name)
    k8s.DeleteConfigMap(ctx, k8sClient, project.Namespace, d.Name+"-env")
    k8s.DeleteSecret(ctx, k8sClient, project.Namespace, d.Name+"-secret")

    // volume_mounts / env_var_mounts も deleting に
    db.Model(&models.VolumeMount{}).
        Where("deployment_id = ?", deploymentID).
        Update("status", "deleting")
    db.Model(&models.EnvVarMount{}).
        Where("deployment_id = ?", deploymentID).
        Update("status", "deleting")

    // Watcher が k8s 削除完了を検知して DB レコードを削除する
    return nil
}
```

### Watcher 側に削除完了検知を追加

```go
// watcher/deployment.go の Deleted イベント処理を拡張
case watch.Deleted:
    // 関連レコードを全て削除
    db.Where("deployment_id = ?", deploymentID).Delete(&models.EnvVarMount{})
    db.Where("deployment_id = ?", deploymentID).Delete(&models.VolumeMount{})
    db.Where("deployment_id = ?", deploymentID).Delete(&models.ApplyHistory{})
    db.Where("deployment_id = ?", deploymentID).Delete(&models.DeploymentBuild{})
    // service / ingress は cascade で削除されるか別途処理
    db.Delete(&models.Deployment{}, "id = ?", deploymentID)
```

## テスト確認項目

- [ ] deployment 削除後に k8s Deployment が削除されること
- [ ] k8s Service / IngressRoute も削除されること
- [ ] Watcher が削除完了を検知して DB レコードが削除されること
- [ ] 削除中（deleting）の deployment に apply すると 409 になること

### repository 層テスト

- [ ] `DeploymentRepository.UpdateStatus` で `status = deleting` に更新できること
- [ ] `DeploymentRepository.Delete` でレコードが DB から削除されること
- [ ] `ServiceRepository.FindByDeploymentID` で削除対象のサービスが取得できること
