# ISSUE-039 ビルド完了 Watcher → 自動 apply

## 親 Issue
ISSUE-035

## 概要
k8s Job の完了を Watch し、ビルドログを保存して自動 apply を実行する。

## 実装手順

### `watcher/` に job watcher を追加

```go
func WatchBuildJobs(ctx context.Context, db *gorm.DB, k8sClient *kubernetes.Clientset) {
    for {
        watcher, _ := k8sClient.BatchV1().Jobs("launchs-builds").Watch(ctx, metav1.ListOptions{})

        for event := range watcher.ResultChan() {
            job, ok := event.Object.(*batchv1.Job)
            if !ok { continue }

            buildID := job.Labels["launchs.org/build-id"]
            if buildID == "" { continue }

            var build models.DeploymentBuild
            db.First(&build, "id = ?", buildID)

            if event.Type == watch.Modified {
                // Job 完了チェック
                for _, cond := range job.Status.Conditions {
                    if cond.Type == batchv1.JobComplete && cond.Status == corev1.ConditionTrue {
                        // ビルドログを Pod から取得して保存
                        log := fetchJobLog(ctx, k8sClient, job)
                        builtImageURL := buildImageURL(build)

                        now := time.Now()
                        db.Model(&build).Updates(map[string]interface{}{
                            "status":          models.BuildStatusSucceeded,
                            "built_image_url": builtImageURL,
                            "build_log":       log,
                            "finished_at":     &now,
                        })

                        // current_build_id をセット
                        db.Model(&models.Deployment{}).
                            Where("id = ?", build.DeploymentID).
                            Update("current_build_id", build.ID)

                        // 自動 apply を実行
                        applySvc := &ApplyService{DB: db, K8s: k8sClient}
                        applySvc.Apply(ctx, build.DeploymentID)
                    }

                    if cond.Type == batchv1.JobFailed && cond.Status == corev1.ConditionTrue {
                        log := fetchJobLog(ctx, k8sClient, job)
                        db.Model(&build).Updates(map[string]interface{}{
                            "status":    models.BuildStatusFailed,
                            "build_log": log,
                        })
                    }
                }
            }
        }
        time.Sleep(1 * time.Second)
    }
}
```

## テスト確認項目

- [ ] ビルド完了後に `deployment_builds.status = succeeded` になること
- [ ] `built_image_url` が記録されること
- [ ] `current_build_id` に新しい build_id がセットされること
- [ ] ビルド完了後に自動で apply が走り `app_status = deploying` になること
- [ ] ビルドログが `build_log` に保存されること

### repository 層テスト

- [ ] `DeploymentBuildRepository.UpdateStatus` で `status = succeeded` に更新できること
- [ ] `DeploymentBuildRepository.Save` で `built_image_url` が保存されること
- [ ] `DeploymentRepository.Save` で `current_build_id` が更新されること
- [ ] `BuildLogRepository.Create` でログレコードが DB に保存されること
