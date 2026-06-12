# ISSUE-011 k8s Deployment manifest 生成・apply

## 親 Issue
ISSUE-009

## 概要
DB の値から k8s Deployment / ConfigMap / Secret manifest を生成し、k8s に server-side apply する。

## 実装手順

### 1. `k8s/manifest/generator.go` を作成

```go
package manifest

import (
    "app/models"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Generator struct {
    InstanceSizes map[string]models.InstanceSize
}

// GenerateDeployment: Deployment manifest を生成
func (g *Generator) GenerateDeployment(
    d models.Deployment,
    namespace string,
    imageURL string,
    envMounts []models.EnvVarMount,
    volumeMounts []models.VolumeMount,
) *appsv1.Deployment {

    size := g.InstanceSizes[d.InstanceSize]

    container := corev1.Container{
        Name:  "app",
        Image: imageURL,
        Resources: corev1.ResourceRequirements{
            Requests: corev1.ResourceList{
                corev1.ResourceCPU:    resource.MustParse(size.CPURequest),
                corev1.ResourceMemory: resource.MustParse(size.MemoryRequest),
            },
            Limits: corev1.ResourceList{
                corev1.ResourceCPU:    resource.MustParse(size.CPULimit),
                corev1.ResourceMemory: resource.MustParse(size.MemoryLimit),
            },
        },
    }

    // command / args
    if len(d.Command) > 0 { container.Command = d.Command }
    if len(d.Args) > 0    { container.Args = d.Args }

    // envFrom（Phase5 で拡張）
    // volumeMounts（Phase6 で拡張）

    replicas := d.Replicas

    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      d.Name,
            Namespace: namespace,
            Labels: map[string]string{
                "launchs.org/deployment-id": d.ID,
                "app":                       d.Name,
            },
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: &replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: map[string]string{"app": d.Name},
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: map[string]string{"app": d.Name},
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{container},
                },
            },
        },
    }
}
```

### 2. `k8s/deployment.go` を作成

```go
package k8s

import (
    "context"
    appsv1 "k8s.io/api/apps/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func ApplyDeployment(ctx context.Context, client *kubernetes.Clientset, dep *appsv1.Deployment) error {
    existing, err := client.AppsV1().Deployments(dep.Namespace).Get(ctx, dep.Name, metav1.GetOptions{})
    if err != nil {
        // 存在しない場合は Create
        _, err = client.AppsV1().Deployments(dep.Namespace).Create(ctx, dep, metav1.CreateOptions{})
        return err
    }
    // 存在する場合は Update
    dep.ResourceVersion = existing.ResourceVersion
    _, err = client.AppsV1().Deployments(dep.Namespace).Update(ctx, dep, metav1.UpdateOptions{})
    return err
}

func DeleteDeployment(ctx context.Context, client *kubernetes.Clientset, namespace, name string) error {
    return client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}
```

## テスト確認項目

- [ ] `GenerateDeployment` が正しい manifest を返すこと
- [ ] `instance_size = "small"` で cpu/memory が正しく設定されること
- [ ] `command` / `args` が空の場合は manifest に含まれないこと
- [ ] `ApplyDeployment` で k8s に Deployment が作成されること
- [ ] 同名の Deployment を再度 apply すると更新されること
