# ISSUE-032 Deployment Watcher

## 親 Issue
ISSUE-031

## 概要
k8s Deployment の変化を Watch し、k8s_status と app_status を DB に反映する。
RetryWatcher を使って切断時に自動再接続する。

## 実装手順

### 1. `cmd/watcher/main.go` を作成

```go
package main

import (
    "context"
    "log"
    "app/config"
    "app/repository"
    "app/k8s"
    "app/watcher"
)

func main() {
    cfg := config.Load()
    database, _ := db.New(cfg.DatabaseDSN)
    k8sClient, _ := k8s.NewClient()

    ctx := context.Background()
    log.Println("Starting watcher...")

    go watcher.WatchDeployments(ctx, database, k8sClient)
    go watcher.WatchServices(ctx, database, k8sClient)
    go watcher.WatchPVCs(ctx, database, k8sClient)
    go watcher.WatchNamespaces(ctx, database, k8sClient)

    select {} // ブロック
}
```

### 2. `watcher/deployment.go` を作成

```go
package watcher

import (
    "context"
    "encoding/json"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/watch"
    "k8s.io/client-go/kubernetes"
    "gorm.io/gorm"
    "app/models"
)

func WatchDeployments(ctx context.Context, db *gorm.DB, client *kubernetes.Clientset) {
    for {
        watcher, err := client.AppsV1().Deployments("").Watch(ctx, metav1.ListOptions{
            LabelSelector: "launchs.org/managed=true",
        })
        if err != nil {
            time.Sleep(5 * time.Second)
            continue
        }

        for event := range watcher.ResultChan() {
            dep, ok := event.Object.(*appsv1.Deployment)
            if !ok { continue }

            deploymentID := dep.Labels["launchs.org/deployment-id"]
            if deploymentID == "" { continue }

            statusJSON, _ := json.Marshal(dep.Status)

            switch event.Type {
            case watch.Modified, watch.Added:
                updates := map[string]interface{}{
                    "k8s_status": statusJSON,
                }
                // Pod が全て Ready → app_status = running
                if dep.Status.ReadyReplicas > 0 &&
                    dep.Status.ReadyReplicas == dep.Status.Replicas {
                    updates["app_status"] = models.AppStatusRunning
                    updates["status"] = models.DeploymentStatusRunning
                }
                db.Model(&models.Deployment{}).
                    Where("id = ?", deploymentID).
                    Updates(updates)

            case watch.Deleted:
                db.Model(&models.Deployment{}).
                    Where("id = ?", deploymentID).
                    Update("status", models.DeploymentStatusFailed)
            }
        }

        // watcher が切れたら再接続
        time.Sleep(1 * time.Second)
    }
}
```

## テスト確認項目

- [ ] apply 後に Watcher が `k8s_status` を更新すること
- [ ] Pod が Ready になると `app_status = running` になること
- [ ] Watcher プロセスが k8s 切断後に再接続すること
