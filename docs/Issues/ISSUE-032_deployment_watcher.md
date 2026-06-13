# ISSUE-032 Deployment Watcher

## 親 Issue
ISSUE-031

## 概要
k8s Deploymentリソースの状態変化を監視して、DBのDeploymentのstatusとapp_statusを自動更新するWatcherを実装する。

## 変更ファイル一覧

- `app/src/k8s/deployment.go`（編集）
    - **何を**: WatchDeployments関数の追加。k8s watch APIを使ってDeploymentの追加・変更・削除イベントを監視する。launchs.org/deployment-idラベルでDeploymentを特定し、ReadyReplicasの状態からapp_statusを計算してDBを更新する。Deleted イベントでDBレコードを削除する。
    - **なぜ**: k8s側の実際のデプロイ状態をDBに反映するため
- `app/src/main.go`（編集）
    - **何を**: アプリ起動時にgoroutineでWatchDeployments()を起動する処理の追加。
    - **なぜ**: Watcherをバックグラウンドで常時起動させるため

## テスト確認項目

- [ ] k8s DeploymentがRunningになるとDBのapp_statusがrunningに更新されること
- [ ] k8s DeploymentがPendingのときDBのapp_statusがdeployingに更新されること
- [ ] k8s DeploymentがDeletedになるとDBレコードが削除されること
### repository 層テスト

- [ ] DeploymentRepository.UpdateAppStatusでapp_statusが更新できること
