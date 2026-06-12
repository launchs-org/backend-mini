# ISSUE-007 k8s Namespace 作成・削除

## 親 Issue
ISSUE-005

## 概要
k8s namespace の作成・削除ロジックを実装する。Project の作成・削除から呼び出す。

## 実装手順

### 1. `k8s/namespace.go` を作成

```go
package k8s

import (
    "context"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func CreateNamespace(ctx context.Context, client *kubernetes.Clientset, name string) error {
    ns := &corev1.Namespace{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
            Labels: map[string]string{
                "launchs.org/managed": "true",
            },
        },
    }
    _, err := client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
    return err
}

func DeleteNamespace(ctx context.Context, client *kubernetes.Clientset, name string) error {
    return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}
```

## テスト確認項目

- [ ] `CreateNamespace` で k8s に namespace が作られること
- [ ] `DeleteNamespace` で k8s から namespace が削除されること
- [ ] 同名 namespace を2回作成するとエラーになること
