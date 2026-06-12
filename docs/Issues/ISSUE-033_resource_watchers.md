# ISSUE-033 Service / IngressRoute / PVC Watcher

## 親 Issue
ISSUE-031

## 概要
Service・PVC の k8s_status を Watch して DB に反映する。
PVC が削除されたら DB レコードも削除する。

## 実装手順

### `watcher/service.go`

```go
func WatchServices(ctx context.Context, db *gorm.DB, client *kubernetes.Clientset) {
    // Service の Watch
    // k8s_status を services テーブルに反映
    // deployment_name ラベルから service レコードを特定
}
```

### `watcher/pvc.go`

```go
func WatchPVCs(ctx context.Context, db *gorm.DB, client *kubernetes.Clientset) {
    // PVC の Watch
    // Bound → volumes.status = bound
    // Deleted → volumes レコードを DB から削除
}
```

## テスト確認項目

- [ ] PVC が Bound になると `volumes.status = bound` になること
- [ ] PVC が削除されると `volumes` レコードが DB から削除されること
