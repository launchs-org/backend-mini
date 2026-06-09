# ISSUE-021 [Phase5] Env Var / Env Var Mount

## Sub Issues
- [ ] ISSUE-022 Env Var CRUD エンドポイント
- [ ] ISSUE-023 Env Var Mount CRUD エンドポイント
- [ ] ISSUE-024 k8s ConfigMap / Secret 生成・apply
- [ ] ISSUE-025 apply サービスに env 重複チェック・ConfigMap/Secret 追加

## 完了条件
- env_var の CRUD が動くこと
- is_secret=true の value がマスクされること
- apply 後に k8s ConfigMap / Secret が作成されること
- env_var_mounts.status が applied になること
