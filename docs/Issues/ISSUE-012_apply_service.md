# ISSUE-012 apply サービス（ロック・pending 昇格・apply_history）

## 親 Issue
ISSUE-009

## 概要
apply のコアロジックを実装する。SELECT FOR UPDATE によるロック、pending_*** の昇格、apply_history の記録を行う。

## 実装手順

### 1. `service/apply.go` を作成

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "app/k8s"
    "app/k8s/manifest"
    "app/models"
    "gorm.io/gorm"
    k8sclient "k8s.io/client-go/kubernetes"
)

type ApplyService struct {
    DB        *gorm.DB
    K8s       *k8sclient.Clientset
    Generator *manifest.Generator
}

type ApplyResult struct {
    ApplyHistoryID string
    Status         string
    BuildID        string // ビルドが必要な場合に設定（Phase8 で使用）
}

func (s *ApplyService) Apply(ctx context.Context, deploymentID string) (*ApplyResult, error) {
    var result *ApplyResult

    err := s.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        // 1. SELECT FOR UPDATE でロック取得
        var d models.Deployment
        if err := tx.Set("gorm:query_option", "FOR UPDATE").
            First(&d, "id = ?", deploymentID).Error; err != nil {
            return fmt.Errorf("deployment not found: %w", err)
        }

        // 2. Project 取得（namespace のため）
        var project models.Project
        if err := tx.First(&project, "id = ?", d.ProjectID).Error; err != nil {
            return fmt.Errorf("project not found: %w", err)
        }

        // 3. env_var 実効キーの重複チェック（Phase5 で実装）

        // 4. pending_*** から使用する値を決定
        imageURL := d.PendingImageURL
        if imageURL == "" { imageURL = d.ImageURL }

        instanceSize := d.PendingInstanceSize
        if instanceSize == "" { instanceSize = d.InstanceSize }

        replicas := d.PendingReplicas
        if replicas == 0 { replicas = d.Replicas }
        if replicas == 0 { replicas = 1 }

        // ビルド要否判定（Phase8 で実装）
        // image_url 型はビルド不要

        // 5. manifest 生成
        // instance_size マスターを取得
        var size models.InstanceSize
        tx.First(&size, "size = ?", instanceSize)

        dForManifest := d
        dForManifest.InstanceSize = instanceSize
        dForManifest.Replicas = replicas
        dForManifest.Command = d.PendingCommand
        if len(dForManifest.Command) == 0 { dForManifest.Command = d.Command }
        dForManifest.Args = d.PendingArgs
        if len(dForManifest.Args) == 0 { dForManifest.Args = d.Args }

        gen := &manifest.Generator{
            InstanceSizes: map[string]models.InstanceSize{instanceSize: size},
        }
        depManifest := gen.GenerateDeployment(dForManifest, project.Namespace, imageURL, nil, nil)

        // 6. apply_history INSERT
        manifestJSON, _ := json.Marshal(depManifest)
        history := models.ApplyHistory{
            DeploymentID: deploymentID,
            Manifests:    manifestJSON,
            Status:       models.ApplyStatusApplied,
            AppliedAt:    time.Now(),
        }
        tx.Create(&history)

        // 7. k8s apply
        if err := k8s.ApplyDeployment(ctx, s.K8s, depManifest); err != nil {
            history.Status = models.ApplyStatusFailed
            history.ErrorMessage = err.Error()
            tx.Save(&history)
            return fmt.Errorf("k8s apply: %w", err)
        }

        // 8. pending_*** を空に → current 値に昇格
        now := time.Now()
        updates := map[string]interface{}{
            "image_url":               imageURL,
            "pending_image_url":       "",
            "instance_size":           instanceSize,
            "pending_instance_size":   "",
            "replicas":                replicas,
            "pending_replicas":        0,
            "github_repo_url":         d.PendingGithubRepoURL,
            "pending_github_repo_url": "",
            "github_branch":           d.PendingGithubBranch,
            "pending_github_branch":   "",
            "github_commit_sha":       d.PendingGithubCommitSHA,
            "pending_github_commit_sha": "",
            "github_repo_directory":   d.PendingGithubRepoDirectory,
            "pending_github_repo_directory": "",
            "dockerfile_path":         d.PendingDockerfilePath,
            "pending_dockerfile_path": "",
            "command":                 dForManifest.Command,
            "pending_command":         nil,
            "args":                    dForManifest.Args,
            "pending_args":            nil,
            "status":                  models.DeploymentStatusRunning,
            "app_status":              models.AppStatusDeploying,
            "applied_at":              &now,
        }
        tx.Model(&d).Updates(updates)

        result = &ApplyResult{
            ApplyHistoryID: history.ID,
            Status:         "applied",
        }
        return nil
    })

    return result, err
}
```

## テスト確認項目

- [ ] apply 後に k8s Deployment が作成されること
- [ ] apply 後に `pending_***` が空になること
- [ ] apply 後に `current` 値が更新されること
- [ ] apply 後に `status = running`、`app_status = deploying` になること
- [ ] apply_history が1件作成されること
- [ ] k8s apply 失敗時に `apply_history.status = failed` になること
- [ ] k8s apply 失敗時に `pending_***` がそのまま残ること
- [ ] 同一 deployment に並行 apply を投げると2つ目がロック待ちになること
