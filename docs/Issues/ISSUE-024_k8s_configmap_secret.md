# ISSUE-024 k8s ConfigMap / Secret 生成・apply

## 親 Issue
ISSUE-021

## 概要
env_var_mounts から ConfigMap / Secret を生成して k8s に apply する。
is_secret=false → ConfigMap、is_secret=true → Secret に振り分ける。

## 実装手順

### 1. `k8s/configmap.go` を作成

```go
package k8s

import (
    "context"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func ApplyConfigMap(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data map[string]string) error {
    cm := &corev1.ConfigMap{
        ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
        Data:       data,
    }
    existing, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
    if err != nil {
        _, err = client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
        return err
    }
    cm.ResourceVersion = existing.ResourceVersion
    _, err = client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
    return err
}
```

### 2. `k8s/secret.go` を作成

```go
package k8s

func ApplySecret(ctx context.Context, client *kubernetes.Clientset, namespace, name string, data map[string]string) error {
    stringData := make(map[string][]byte)
    for k, v := range data { stringData[k] = []byte(v) }

    secret := &corev1.Secret{
        ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
        Data:       stringData,
    }
    // 同様に Create or Update
    ...
}
```

### 3. env_var_mounts から ConfigMap/Secret データを構築するヘルパー

```go
// service/apply.go に追加
func buildEnvData(mounts []models.EnvVarMount) (configData map[string]string, secretData map[string]string) {
    configData = map[string]string{}
    secretData = map[string]string{}
    for _, m := range mounts {
        // 実効キー: override_key が空なら env_var.key を使う
        effectiveKey := m.OverrideKey
        if effectiveKey == "" {
            effectiveKey = m.EnvVar.Key
        }
        if m.EnvVar.IsSecret {
            secretData[effectiveKey] = m.EnvVar.Value
        } else {
            configData[effectiveKey] = m.EnvVar.Value
        }
    }
    return
}
```

## テスト確認項目

- [ ] is_secret=false の env_var が ConfigMap に入ること
- [ ] is_secret=true の env_var が Secret に入ること
- [ ] override_key が設定されている場合、そのキー名で ConfigMap/Secret に入ること
- [ ] 再 apply で ConfigMap/Secret が更新されること
