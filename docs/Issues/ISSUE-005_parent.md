# ISSUE-005 [Phase2] Account / Project CRUD

## 概要
Account・Project の CRUD と k8s namespace 作成・削除を実装する。

## Sub Issues

- [ ] ISSUE-006 Account CRUD
- [ ] ISSUE-007 Project 作成・取得・更新
- [ ] ISSUE-008 Project 削除（deleting ステータス遷移）
- [ ] ISSUE-009 k8s namespace 作成・削除

## 完了条件

- Project を作成すると k8s namespace が作られること
- Project の CRUD が全て動くこと
- Account の quota 取得・更新が動くこと
