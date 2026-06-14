package k8s

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ApplyPVC は k8s に PersistentVolumeClaim を作成または更新する
func ApplyPVC(ctx context.Context, client kubernetes.Interface, pvcManifest *corev1.PersistentVolumeClaim) error {
	existing, err := client.CoreV1().PersistentVolumeClaims(pvcManifest.Namespace).Get(ctx, pvcManifest.Name, metav1.GetOptions{}) // 既存の PVC を取得する
	if err != nil {
		// 存在しない場合は新規作成する
		_, err = client.CoreV1().PersistentVolumeClaims(pvcManifest.Namespace).Create(ctx, pvcManifest, metav1.CreateOptions{})
		return err
	}
	// 既存の ResourceVersion を設定して更新する（k8s の楽観的並行性制御のため）
	pvcManifest.ResourceVersion = existing.ResourceVersion
	_, err = client.CoreV1().PersistentVolumeClaims(pvcManifest.Namespace).Update(ctx, pvcManifest, metav1.UpdateOptions{})
	return err
}

// DeletePVC は k8s から PersistentVolumeClaim を削除する
func DeletePVC(ctx context.Context, client kubernetes.Interface, namespace, name string) error {
	return client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{}) // PVC を削除する
}

// BuildPVCManifest は Volume モデルの情報から PVC マニフェストを生成する
func BuildPVCManifest(namespace, name string, sizeMB int, storageClassName string) *corev1.PersistentVolumeClaim {
	storageRequest := resource.MustParse(fmt.Sprintf("%dMi", sizeMB)) // SizeMB を MiB 単位の Quantity に変換する
	pvcManifest := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce, // 単一ノードからの読み書きアクセスを許可する
			},
			Resources: corev1.VolumeResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: storageRequest, // ストレージ容量を設定する
				},
			},
		},
	}
	if storageClassName != "" {
		pvcManifest.Spec.StorageClassName = &storageClassName // StorageClass が指定されている場合に設定する
	}
	return pvcManifest
}
