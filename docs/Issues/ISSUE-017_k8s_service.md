# ISSUE-017 k8s Service manifest 生成・apply

## 親 Issue
ISSUE-015

## 概要
k8s Service の manifest を生成し、apply する。

## 実装手順

### 1. `k8s/service.go` を作成

```go
package k8s

import (
    "context"
    "encoding/json"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/util/intstr"
    "k8s.io/client-go/kubernetes"
)

type ServicePort struct {
    Protocol string `json:"protocol"`
    Port     int    `json:"port"`
}

func ApplyService(ctx context.Context, client *kubernetes.Clientset, namespace, name string, portsJSON []byte) error {
    var ports []ServicePort
    json.Unmarshal(portsJSON, &ports)

    k8sPorts := make([]corev1.ServicePort, len(ports))
    for i, p := range ports {
        k8sPorts[i] = corev1.ServicePort{
            Name:       fmt.Sprintf("port-%d", p.Port),
            Protocol:   corev1.Protocol(p.Protocol),
            Port:       int32(p.Port),
            TargetPort: intstr.FromInt(p.Port),
        }
    }

    svc := &corev1.Service{
        ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
        Spec: corev1.ServiceSpec{
            Selector: map[string]string{"app": name},
            Ports:    k8sPorts,
        },
    }

    existing, err := client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        _, err = client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
        return err
    }
    svc.ResourceVersion = existing.ResourceVersion
    _, err = client.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
    return err
}

func DeleteService(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
    return client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
```

## テスト確認項目

- [ ] `ApplyService` で k8s Service が作成されること
- [ ] TCP / UDP 両方のポートが正しく設定されること
- [ ] 再 apply で更新されること

### repository 層テスト

- [ ] `ServiceRepository.Save` で apply 後の `current_ports` が更新されること
- [ ] `ServiceRepository.FindByDeploymentID` で存在しない deployment_id を渡すと `ErrRecordNotFound` が返ること
