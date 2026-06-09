# ISSUE-006 User Quota 取得・更新

## 親 Issue
ISSUE-005

## 概要
`user_id` ごとの quota 取得・更新を実装する。

認証は別サービスが担当し、JWT から取り出した `user_id`（UUID文字列）だけがこのサービスに渡ってくる。
Account テーブルは持たず、`user_quotas` テーブルを `user_id` で直接管理する。
レコードが存在しない場合は初回アクセス時にデフォルト値で upsert する。

## レイヤー構成

```
controller/user_quota_controller.go   ← HTTP ハンドラ
service/quota_service.go              ← quota チェックロジック
repository/user_quota_repository.go   ← DB アクセス（GORM）
model/user_quota.go                   ← UserQuota モデル
model/project.go                      ← Project モデル（UserID フィールド）
```

## ルーティング

```
GET /api/v1/users/:user_id/quota
PUT /api/v1/users/:user_id/quota
```

## レスポンス例

```json
// GET /users/:user_id/quota
{
  "user_id": "uuid",
  "max_projects": 5,
  "max_deployments": 20,
  "max_replicas_per_deployment": 5,
  "max_volume_mb": 10240,
  "current_projects": 2,
  "current_deployments": 3,
  "current_volume_mb": 2048
}
```

```json
// PUT /users/:user_id/quota  (部分更新)
{ "max_deployments": 30, "max_volume_mb": 20480 }
```

## テスト確認項目

- [ ] `GET /users/:user_id/quota` で quota と現在使用量が返ること
- [ ] quota レコードが存在しない user_id でも 200 でデフォルト値が返ること（auto-create）
- [ ] `PUT /users/:user_id/quota` で部分更新できること
