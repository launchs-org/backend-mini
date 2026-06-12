# ISSUE-019 k8s Traefik IngressRoute CRD apply

## 親 Issue
ISSUE-015

## 概要
Traefik の IngressRoute CRD を dynamic client で apply する。

## 実装手順

### 1. `k8s/ingress.go` を作成

Traefik IngressRoute は CRD なので `dynamic.Interface` を使う。

```go
package k8s

import (
    "context"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/client-go/dynamic"
)

var ingressRouteGVR = schema.GroupVersionResource{
    Group:    "traefik.io",
    Version:  "v1alpha1",
    Resource: "ingressroutes",
}

func ApplyIngressRoute(ctx context.Context, dc dynamic.Interface, namespace, deploymentName, host, pathPrefix, serviceName string, port int) error {
    obj := &unstructured.Unstructured{
        Object: map[string]interface{}{
            "apiVersion": "traefik.io/v1alpha1",
            "kind":       "IngressRoute",
            "metadata": map[string]interface{}{
                "name":      deploymentName,
                "namespace": namespace,
            },
            "spec": map[string]interface{}{
                "entryPoints": []interface{}{"web", "websecure"},
                "routes": []interface{}{
                    map[string]interface{}{
                        "match": fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", host, pathPrefix),
                        "kind":  "Rule",
                        "services": []interface{}{
                            map[string]interface{}{
                                "name": serviceName,
                                "port": port,
                            },
                        },
                    },
                },
            },
        },
    }

    _, err := dc.Resource(ingressRouteGVR).Namespace(namespace).Apply(
        ctx,
        deploymentName,
        obj,
        metav1.ApplyOptions{FieldManager: "launchs", Force: true},
    )
    return err
}

func DeleteIngressRoute(ctx context.Context, dc dynamic.Interface, namespace, name string) error {
    return dc.Resource(ingressRouteGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
```

## テスト確認項目

- [ ] `ApplyIngressRoute` で k8s に IngressRoute が作成されること
- [ ] `Host` と `PathPrefix` が正しく設定されること
- [ ] 再 apply で更新されること（Force: true）
