# ISSUE-048 Pod ログ取得エンドポイント

## 親 Issue
ISSUE-047

## 実装手順

### `handler/log.go` を作成

```go
package handler

import (
    "bytes"
    "io"
    "time"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *Handler) GetPodLogs(c echo.Context) error {
    deploymentID := c.Param("id")

    var d models.Deployment
    if err := h.DB.First(&d, "id = ?", deploymentID).Error; err != nil {
        return echo.ErrNotFound
    }
    var project models.Project
    h.DB.First(&project, "id = ?", d.ProjectID)

    // Pod を deployment name のラベルで特定
    pods, err := h.K8s.CoreV1().Pods(project.Namespace).List(
        c.Request().Context(),
        metav1.ListOptions{LabelSelector: "app=" + d.Name},
    )
    if err != nil || len(pods.Items) == 0 {
        return c.JSON(http.StatusOK, map[string]string{"logs": ""})
    }

    opts := &corev1.PodLogOptions{}

    if since := c.QueryParam("since"); since != "" {
        t, _ := time.Parse(time.RFC3339, since)
        sinceTime := metav1.NewTime(t)
        opts.SinceTime = &sinceTime
    }

    container := c.QueryParam("container")
    if container != "" { opts.Container = container }

    req := h.K8s.CoreV1().Pods(project.Namespace).GetLogs(pods.Items[0].Name, opts)
    stream, err := req.Stream(c.Request().Context())
    if err != nil {
        return echo.ErrInternalServerError
    }
    defer stream.Close()

    var buf bytes.Buffer
    io.Copy(&buf, stream)

    return c.JSON(http.StatusOK, map[string]string{"logs": buf.String()})
}
```

### ルーティング登録

```go
api.GET("/deployments/:id/logs", h.GetPodLogs)
```

## テスト確認項目

- [ ] running な deployment のログが取得できること
- [ ] `since` パラメータでログがフィルタされること
- [ ] Pod が存在しない場合に空文字列が返ること
- [ ] `container` パラメータで特定コンテナのログが取得できること
