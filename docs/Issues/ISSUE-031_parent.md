# ISSUE-031 [Phase7] Watcher プロセス

## Sub Issues
- [ ] ISSUE-032 Deployment Watcher（k8s_status 更新・app_status=running 検知）
- [ ] ISSUE-033 Service / IngressRoute / PVC Watcher
- [ ] ISSUE-034 Namespace Watcher（Project 削除完了検知）

## 完了条件
- apply 後に Pod が Ready になると `app_status = running` になること
- k8s_status が Watcher によってリアルタイムに更新されること
