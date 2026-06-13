# ISSUE-028 ボリュームマウントCRUD

## 親 Issue
ISSUE-026

## 概要
ボリュームをデプロイメントにマウントする設定のCRUDエンドポイントを実装する。マウントパスを指定してapply時にk8sのvolumeMountsとして反映される。

## 変更ファイル一覧

- `app/src/models/volume_mount.go`（編集）
    - **何を**: VolumeMountモデルの定義。DeploymentIDとVolumeIDへの外部キー、mount_pathフィールドを持つ。
    - **なぜ**: ボリュームとデプロイメントの多対多関係をDBで管理するため

- `app/src/repository/volume_mount_repository.go`（新規作成）
    - **何を**: VolumeMountRepositoryインターフェースと実装。Create・FindAllByDeploymentID・Deleteメソッドを持つ。
    - **なぜ**: ボリュームマウント設定のDB操作を抽象化するため

- `app/src/service/volume_service.go`（編集）
    - **何を**: ListVolumeMounts・CreateVolumeMount・DeleteVolumeMountメソッドの追加。同一DeploymentIDで同一mount_pathの重複を拒否する。
    - **なぜ**: ボリュームマウント管理のビジネスロジックをハンドラーから分離するため

- `app/src/handler/volume_handler.go`（編集）
    - **何を**: ListVolumeMounts・CreateVolumeMount・DeleteVolumeMountハンドラーの追加。
    - **なぜ**: ボリュームマウント設定管理のHTTPエントリーポイントが必要なため

- `app/src/router/router.go`（編集）
    - **何を**: GET/POST /api/v1/deployments/:id/volume-mounts、DELETE /api/v1/volume-mounts/:idエンドポイントの登録。
    - **なぜ**: ボリュームマウントエンドポイントをルーターに接続するため

## テスト確認項目

- [ ] POST /api/v1/deployments/:id/volume-mountsでマウント設定が作成できること
- [ ] GET /api/v1/deployments/:id/volume-mountsでマウント設定一覧が取得できること
- [ ] DELETE /api/v1/volume-mounts/:idでマウント設定が削除できること
- [ ] 同一DeploymentIDで同一mount_pathの重複が拒否されること

### repository 層テスト

- [ ] VolumeMountRepository.FindAllByDeploymentIDでマウント設定一覧が取得できること
