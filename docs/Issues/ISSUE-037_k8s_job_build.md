# ISSUE-037 k8s Job でビルド実行

## 親 Issue
ISSUE-035

## 概要
ビルドを k8s Job として実行する。ビルド完了後 built_image_url をレジストリに push する。

## 実装手順

### 1. `k8s/job.go` を作成

```go
package k8s

import (
    "context"
    batchv1 "k8s.io/api/batch/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
)

func CreateBuildJob(ctx context.Context, client *kubernetes.Clientset, params BuildJobParams) (string, error) {
    jobName := fmt.Sprintf("build-%s", params.BuildID[:8])

    var buildCmd []string
    switch params.BuildType {
    case "dockerfile":
        buildCmd = []string{
            "sh", "-c",
            fmt.Sprintf(
                "git clone %s /workspace && cd /workspace/%s && docker build -f %s -t %s . && docker push %s",
                params.RepoURL, params.Directory, params.DockerfilePath,
                params.BuiltImageURL, params.BuiltImageURL,
            ),
        }
    case "railpack":
        buildCmd = []string{
            "sh", "-c",
            fmt.Sprintf(
                "git clone %s /workspace && cd /workspace/%s && railpack build && docker tag app %s && docker push %s",
                params.RepoURL, params.Directory,
                params.BuiltImageURL, params.BuiltImageURL,
            ),
        }
    }

    ttl := int32(3600) // 1時間後に Job を自動削除
    job := &batchv1.Job{
        ObjectMeta: metav1.ObjectMeta{
            Name:      jobName,
            Namespace: "launchs-builds",
            Labels: map[string]string{
                "launchs.org/build-id": params.BuildID,
            },
        },
        Spec: batchv1.JobSpec{
            TTLSecondsAfterFinished: &ttl,
            Template: corev1.PodTemplateSpec{
                Spec: corev1.PodSpec{
                    RestartPolicy: corev1.RestartPolicyNever,
                    Containers: []corev1.Container{{
                        Name:    "builder",
                        Image:   "docker:24-dind",
                        Command: buildCmd,
                    }},
                },
            },
        },
    }

    _, err := client.BatchV1().Jobs("launchs-builds").Create(ctx, job, metav1.CreateOptions{})
    return jobName, err
}

func DeleteJob(ctx context.Context, client *kubernetes.Clientset, jobName string) error {
    propagation := metav1.DeletePropagationBackground
    return client.BatchV1().Jobs("launchs-builds").Delete(
        ctx, jobName, metav1.DeleteOptions{PropagationPolicy: &propagation},
    )
}
```

## テスト確認項目

- [ ] k8s Job が正しいコマンドで作成されること
- [ ] `deployment_builds.k8s_job_name` に Job 名が保存されること
- [ ] Job 削除で関連 Pod も削除されること（PropagationPolicy=Background）

### repository 層テスト

- [ ] `DeploymentBuildRepository.Save` で `k8s_job_name` が保存されること
- [ ] `DeploymentBuildRepository.FindByID` で存在しない ID を渡すと `ErrRecordNotFound` が返ること
