# ISSUE-003 k8s クライアント初期化

## 親 Issue
ISSUE-001

## 概要
`~/.kube/config` を使って k8s クライアントを初期化する。

## 実装手順

### 1. 依存パッケージ追加

```bash
go get k8s.io/client-go@latest
go get k8s.io/api@latest
go get k8s.io/apimachinery@latest
```

### 2. `internal/k8s/client.go` を作成

```go
package k8s

import (
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
    "path/filepath"
)

func NewClient() (*kubernetes.Clientset, error) {
    kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, err
    }
    return kubernetes.NewForConfig(config)
}
```

### 3. Traefik CRD 用クライアントを追加

IngressRoute は Traefik の CRD なので `dynamic client` を使う。

```bash
go get k8s.io/client-go/dynamic
```

```go
// internal/k8s/client.go に追加
import "k8s.io/client-go/dynamic"

func NewDynamicClient() (dynamic.Interface, error) {
    kubeconfig := filepath.Join(homedir.HomeDir(), ".kube", "config")
    config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
    if err != nil {
        return nil, err
    }
    return dynamic.NewForConfig(config)
}
```

### 4. `cmd/api/main.go` に k8s クライアント初期化を追加

```go
k8sClient, err := k8s.NewClient()
if err != nil {
    log.Fatalf("failed to create k8s client: %v", err)
}

dynamicClient, err := k8s.NewDynamicClient()
if err != nil {
    log.Fatalf("failed to create dynamic client: %v", err)
}
```

### 5. 接続確認用の起動ログを追加

```go
// namespace 一覧を取得して接続確認
nsList, err := k8sClient.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
if err != nil {
    log.Fatalf("failed to connect k8s: %v", err)
}
log.Printf("k8s connected: %d namespaces found", len(nsList.Items))
```

## テスト確認項目

- [ ] 起動時に k8s クラスターに接続できること
- [ ] 起動ログに namespace 数が表示されること
- [ ] `~/.kube/config` が存在しない場合に分かりやすいエラーが出ること
