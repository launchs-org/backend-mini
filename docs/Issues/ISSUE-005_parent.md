# ISSUE-005 [Phase2] Project CRUD

## 概要
Project の CRUD と k8s namespace 作成・削除を実装する。

認証は別サービスが担当し、JWT から取り出した `user_id`（UUID文字列）をそのまま使う。
Account テーブルは持たない。

## Sub Issues

- [ ] ISSUE-006 User Quota 取得・更新
- [ ] ISSUE-007 Project 作成・取得・更新
- [ ] ISSUE-008 Project 削除（deleting ステータス遷移）
- [ ] ISSUE-009 k8s namespace 作成・削除

## 完了条件

- Project を作成すると k8s namespace が作られること
- Project の CRUD が全て動くこと
- User の quota 取得・更新が動くこと
