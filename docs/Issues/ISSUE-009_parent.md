# ISSUE-009 [Phase3] Deployment CRUD + apply (image_url)

## 概要
Deployment の CRUD と apply のコアロジックを実装する。
まず image_url 型のみを対象とし、ビルドは Phase8 で追加する。
このフェーズが全フェーズの核心。

## Sub Issues

- [ ] ISSUE-010 Deployment モデル・CRUD
- [ ] ISSUE-011 k8s Deployment manifest 生成
- [ ] ISSUE-012 apply サービス
- [ ] ISSUE-013 POST /apply エンドポイント
- [ ] ISSUE-014 apply-history 取得エンドポイント

## 完了条件

- Deployment を作成して pending_*** に値が入ること
- apply を叩くと k8s Deployment が作成されること
- apply 後に pending_*** が空になること
- apply_history が記録されること
