# ISSUE-001 [Phase1] プロジェクト基盤セットアップ

## 概要

Echo の Hello World が動いている状態から、DB 接続・k8s クライアント初期化・AutoMigrate までを整備する。
このフェーズが完了すると全フェーズの土台が整う。

## Sub Issues

- [ ] ISSUE-002 DB 接続・AutoMigrate セットアップ
- [ ] ISSUE-003 k8s クライアント初期化
- [ ] ISSUE-004 ディレクトリ構成・Router 整備

## 完了条件

- サーバーが起動して DB に接続できること
- k8s クラスターに接続できること
- 全モデルの AutoMigrate が通ること
