# ISSUE-049 [Phase12] Quota 管理の完成

## Sub Issues
- [ ] ISSUE-050 deployment 作成・更新・apply 時の quota チェック

## 完了条件
- `max_projects` を超える project 作成でエラーになること
- `max_deployments` を超える deployment 作成でエラーになること
- `max_replicas_per_deployment` を超える replicas 設定でエラーになること
- `max_volume_mb` を超える volume 作成でエラーになること
