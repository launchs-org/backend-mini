# ISSUE-040 ビルドログエンドポイント

## 親 Issue
ISSUE-035

## 概要
k8s JobのPodログをストリーミングまたは一括取得するエンドポイントを実装する。

## 変更ファイル一覧

- `app/src/handler/build_handler.go`（編集）
    - **何を**: GetBuildLogsハンドラーの追加。DeploymentBuild IDからk8s JobのラベルでイコールなアクティブなPodを取得してログを返す。sinceパラメータで時間フィルタリングをサポートする。
    - **なぜ**: ビルドの進行状況確認のためにログ取得エンドポイントが必要なため
- `app/src/router/router.go`（編集）
    - **何を**: GET /api/v1/builds/:id/logsエンドポイントの登録。
    - **なぜ**: ビルドログエンドポイントをルーターに接続するため

## テスト確認項目

- [ ] GET /api/v1/builds/:id/logsでビルドログが取得できること
- [ ] sinceパラメータでログがフィルタされること
- [ ] Podが存在しない場合に空文字列が返ること
