# ISSUE-015 [Phase4] Service / IngressRoute

## Sub Issues
- [ ] ISSUE-016 GET / PUT /service エンドポイント
- [ ] ISSUE-017 k8s Service manifest 生成・apply
- [ ] ISSUE-018 POST /ingress エンドポイント・ホスト名自動生成
- [ ] ISSUE-019 k8s IngressRoute (Traefik CRD) manifest 生成・apply
- [ ] ISSUE-020 apply サービスに Service / IngressRoute を追加

## 完了条件
- apply 後に k8s Service が作成されること
- POST /ingress でホスト名が自動生成されること
- apply 後に k8s IngressRoute が作成されること
