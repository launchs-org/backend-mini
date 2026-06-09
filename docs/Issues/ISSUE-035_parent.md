# ISSUE-035 [Phase8] ビルド機能 (dockerfile / railpack)

## Sub Issues
- [ ] ISSUE-036 ビルド判定ロジック・HEAD SHA 解決
- [ ] ISSUE-037 k8s Job でビルド実行
- [ ] ISSUE-038 ビルドキャンセル（既存 Job の削除）
- [ ] ISSUE-039 ビルド完了 Watcher → 自動 apply
- [ ] ISSUE-040 ビルド履歴・ログ取得エンドポイント

## 完了条件
- apply 時に GitHub 情報が変化していると k8s Job が作成されること
- HEAD が実 SHA に解決されること
- ビルド完了後に自動で apply が走ること
- ビルドログが取得できること
