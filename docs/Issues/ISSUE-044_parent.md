# ISSUE-044 [Phase10] 削除フロー完成

## Sub Issues
- [ ] ISSUE-045 Deployment 削除フロー（k8s リソース一括削除）
- [ ] ISSUE-046 Project 削除フロー（全リソース順次削除 → namespace 削除）

## 完了条件
- deployment 削除後に k8s の全関連リソースが削除されること
- project 削除後に namespace と全 DB レコードが削除されること
