# ISSUE-034 Namespace Watcher

## 親 Issue
ISSUE-031

## 概要
Namespace の削除完了を Watch し、project DB レコードを削除する。

## 実装手順

### `watcher/namespace.go`

```go
func WatchNamespaces(ctx context.Context, db *gorm.DB, client *kubernetes.Clientset) {
    // Namespace の Watch
    // Deleted イベント → launchs.org/managed ラベルが付いている場合のみ処理
    // namespace 名から project を特定して DB レコード削除
}
```

## テスト確認項目

- [ ] namespace が削除されると `projects` レコードが DB から削除されること
- [ ] launchs.org/managed ラベルがない namespace は無視されること

### repository 層テスト

- [ ] `ProjectRepository.FindByNamespace` で namespace 名からプロジェクトが取得できること
- [ ] `ProjectRepository.Delete` で namespace 削除時にプロジェクトレコードが DB から削除されること
