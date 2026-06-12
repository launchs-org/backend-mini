# ISSUE-020 apply サービスに Service / IngressRoute を追加

## 親 Issue
ISSUE-015

## 概要
ISSUE-012 の apply サービスを拡張し、Service と IngressRoute も k8s に apply する。

## 実装手順

### `service/apply.go` に追加

```go
// Service の apply（apply.go の k8s apply セクションに追加）

// Service の pending_ports を取得
var svcModel models.Service
tx.Where("deployment_id = ?", deploymentID).First(&svcModel)

ports := svcModel.PendingPorts
if ports == nil { ports = svcModel.Ports }

if ports != nil {
    if err := k8s.ApplyService(ctx, s.K8s, project.Namespace, d.Name, ports); err != nil {
        // Service apply 失敗時の処理
    }
    // pending_ports → ports に昇格
    tx.Model(&svcModel).Updates(map[string]interface{}{
        "ports":         ports,
        "pending_ports": nil,
        "status":        models.ServiceStatusActive,
    })
}

// IngressRoute の apply
var ingressModel models.IngressRoute
if err := tx.Where("service_id = ?", svcModel.ID).First(&ingressModel).Error; err == nil {
    k8s.ApplyIngressRoute(
        ctx, s.DynamicClient,
        project.Namespace, d.Name,
        ingressModel.Host, ingressModel.PathPrefix,
        d.Name, ingressModel.Port,
    )
    tx.Model(&ingressModel).Update("status", models.IngressRouteStatusActive)
}
```

## テスト確認項目

- [ ] apply 後に k8s Service が作成されること
- [ ] apply 後に `services.pending_ports` が空になること
- [ ] IngressRoute が存在する場合、apply 後に k8s IngressRoute が作成されること
- [ ] Service も IngressRoute もない状態で apply がエラーにならないこと
