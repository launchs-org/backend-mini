# ISSUE-046 Project 削除フロー

## 親 Issue
ISSUE-044

## 実装手順

### `internal/service/project.go` の Delete メソッドを完成

```go
func (s *ProjectService) Delete(ctx context.Context, id string) error {
    var project model.Project
    if err := s.DB.First(&project, "id = ?", id).Error; err != nil {
        return err
    }

    // project を deleting に
    s.DB.Model(&project).Update("status", model.ProjectStatusDeleting)

    // 配下の全 deployment を deleting に
    var deployments []model.Deployment
    s.DB.Where("project_id = ?", id).Find(&deployments)

    for _, d := range deployments {
        s.DB.Model(&d).Update("status", model.DeploymentStatusDeleting)
        // k8s リソースを削除
        DeleteDeploymentResources(ctx, s.DB, s.K8s, s.DynamicClient, d.ID)
    }

    // env_vars / volumes を deleting に
    s.DB.Model(&model.EnvVar{}).Where("project_id = ?", id).Update("status", "deleting")
    s.DB.Model(&model.Volume{}).Where("project_id = ?", id).Update("status", "deleting")

    // 全リソース削除完了を監視する goroutine を起動
    go s.waitAndDeleteNamespace(ctx, id, project.Namespace)

    return nil
}

func (s *ProjectService) waitAndDeleteNamespace(ctx context.Context, projectID, namespace string) {
    for {
        time.Sleep(5 * time.Second)

        var count int64
        s.DB.Model(&model.Deployment{}).
            Where("project_id = ? AND status != ?", projectID, "deleted").
            Count(&count)

        if count == 0 {
            // 全リソース削除完了 → namespace を削除
            k8s.DeleteNamespace(ctx, s.K8s, namespace)
            return
        }
    }
}
```

## テスト確認項目

- [ ] project 削除後に全 deployment が `status = deleting` になること
- [ ] 全リソース削除後に k8s namespace が削除されること
- [ ] Watcher が namespace 削除を検知して project DB レコードが削除されること
- [ ] 削除中の project に新規 deployment を作成すると 409 になること
