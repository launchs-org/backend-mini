package k8s

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ApplyDeployment は k8s に Deployment を作成または更新する
func ApplyDeployment(ctx context.Context, client kubernetes.Interface, deploymentManifest *appsv1.Deployment) error {
	existing, err := client.AppsV1().Deployments(deploymentManifest.Namespace).Get(ctx, deploymentManifest.Name, metav1.GetOptions{}) // 既存の Deployment を取得する
	if err != nil {
		// 存在しない場合は新規作成する
		_, err = client.AppsV1().Deployments(deploymentManifest.Namespace).Create(ctx, deploymentManifest, metav1.CreateOptions{})
		return err
	}
	// 既存の ResourceVersion を設定して更新する（k8s の楽観的並行性制御のため）
	deploymentManifest.ResourceVersion = existing.ResourceVersion
	_, err = client.AppsV1().Deployments(deploymentManifest.Namespace).Update(ctx, deploymentManifest, metav1.UpdateOptions{})
	return err
}

// DeleteDeployment は k8s から Deployment を削除する
func DeleteDeployment(ctx context.Context, client kubernetes.Interface, namespace, name string) error {
	return client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{}) // Deployment を削除する
}
