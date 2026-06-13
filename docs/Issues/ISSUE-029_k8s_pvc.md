# ISSUE-029 k8s PVC 生成・apply

## 親 Issue
ISSUE-026

## 実装手順

### `k8s/pvc.go` を作成

```go
package k8s

import (
    "context"
    "fmt"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func ApplyPVC(ctx context.Context, client *kubernetes.Clientset, namespace, name string, sizeMB int) error {
    storageSize := resource.MustParse(fmt.Sprintf("%dMi", sizeMB))
    pvc := &corev1.PersistentVolumeClaim{
        ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
        Spec: corev1.PersistentVolumeClaimSpec{
            AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
            Resources: corev1.VolumeResourceRequirements{
                Requests: corev1.ResourceList{
                    corev1.ResourceStorage: storageSize,
                },
            },
        },
    }
    _, err := client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        _, err = client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
    }
    // PVC は作成後変更不可なので Update しない
    return err
}

func DeletePVC(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
    return client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
```

## テスト確認項目

- [ ] `ApplyPVC` で k8s PVC が作成されること
- [ ] 既存 PVC を再 apply しても更新されないこと（エラーにもならないこと）
- [ ] size が正しく設定されること

### repository 層テスト

- [ ] `VolumeRepository.Save` で apply 後に `status = bound` に更新できること
- [ ] `VolumeMountRepository.FindAllByDeploymentID` で全マウントが取得できること
