# ISSUE-050 deployment 作成・更新・apply 時の quota チェック

## 親 Issue
ISSUE-049

## 概要
`service/quota_service.go` に quota チェック関数を実装し、各 controller に組み込む。
`user_id` は Echo コンテキストから `ctx.Get("UserID")` で取得する（`RequireAuth` ミドルウェアがセット済み）。

## チェック一覧

| チェック関数 | 呼び出し箇所 | エラー時レスポンス |
|---|---|---|
| `CheckProjectQuota(userID)` | `POST /projects` | 400 |
| `CheckDeploymentQuota(userID)` | `POST /projects/:id/deployments` | 400 |
| `CheckReplicasQuota(userID, replicas)` | `POST /projects/:id/deployments`、`PUT /deployments/:id`（replicas 変更時）、`POST /deployments/:id/apply` | 400 |
| `CheckVolumeQuota(userID, sizeMB)` | `POST /projects/:id/volumes` | 400 |

## service/quota_service.go の責務

- `repository.UserQuotaRepository` を通じて `user_quotas` と使用量カウントを取得する
- quota を超えていた場合は専用の sentinel error を返す（controller 側で 400 に変換）

## エラーレスポンス例

```json
{ "error": "project quota exceeded", "code": "PROJECT_QUOTA_EXCEEDED" }
{ "error": "deployment quota exceeded", "code": "DEPLOYMENT_QUOTA_EXCEEDED" }
{ "error": "replicas exceed limit", "code": "REPLICAS_QUOTA_EXCEEDED" }
{ "error": "volume storage quota exceeded", "code": "VOLUME_QUOTA_EXCEEDED" }
```

## テスト確認項目

- [ ] `max_projects` を超える project 作成で 400 になること
- [ ] `max_deployments` を超える deployment 作成で 400 になること
- [ ] `max_replicas_per_deployment` を超える replicas 設定で 400 になること
- [ ] `max_volume_mb` を超える volume 作成で 400 になること
- [ ] quota 更新後に新しい制限が即時反映されること
