package k8s

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ApplyService は k8s に Service を作成または更新する
func ApplyService(ctx context.Context, client kubernetes.Interface, serviceManifest *corev1.Service) error {
	existing, err := client.CoreV1().Services(serviceManifest.Namespace).Get(ctx, serviceManifest.Name, metav1.GetOptions{}) // 既存の Service を取得する
	if err != nil {
		// 存在しない場合は新規作成する
		_, err = client.CoreV1().Services(serviceManifest.Namespace).Create(ctx, serviceManifest, metav1.CreateOptions{})
		return err
	}
	// 既存の ResourceVersion と ClusterIP を引き継いで更新する（k8s の楽観的並行性制御および ClusterIP 不変制約のため）
	serviceManifest.ResourceVersion = existing.ResourceVersion
	serviceManifest.Spec.ClusterIP = existing.Spec.ClusterIP
	_, err = client.CoreV1().Services(serviceManifest.Namespace).Update(ctx, serviceManifest, metav1.UpdateOptions{})
	return err
}

// DeleteService は k8s から Service を削除する
func DeleteService(ctx context.Context, client kubernetes.Interface, namespace, name string) error {
	return client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{}) // Service を削除する
}
