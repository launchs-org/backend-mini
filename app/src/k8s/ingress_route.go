package k8s

import (
	"app/models"
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// traefikIngressRouteGVR は Traefik IngressRoute CRD の GroupVersionResource を定義する
var traefikIngressRouteGVR = schema.GroupVersionResource{
	Group:    "traefik.io",
	Version:  "v1alpha1",
	Resource: "ingressroutes",
}

// buildRouterRule は IngressRoute のルールマッチ文字列を生成する
func buildRouterRule(host, pathPrefix string) string {
	return fmt.Sprintf("Host(`%s`) && PathPrefix(`%s`)", host, pathPrefix) // ホストとパスプレフィックスのルールを生成する
}

// buildIngressRouteManifest は Traefik IngressRoute の unstructured マニフェストを生成する
func buildIngressRouteManifest(ingressRouteData models.IngressRoute, namespace, serviceName string, servicePort int) *unstructured.Unstructured {
	host := ingressRouteData.PendingHost // pending_host を使う
	if host == "" {                      // pending が空の場合は current 値を使う
		host = ingressRouteData.Host
	}
	pathPrefix := ingressRouteData.PendingPathPrefix // pending_path_prefix を使う
	if pathPrefix == "" {                            // pending が空の場合は current 値を使う
		pathPrefix = ingressRouteData.PathPrefix
	}
	if pathPrefix == "" { // pathPrefix が未設定の場合はデフォルト値を使う
		pathPrefix = "/"
	}
	port := ingressRouteData.PendingPort // pending_port を使う
	if port == 0 {                       // pending が 0 の場合は current 値を使う
		port = ingressRouteData.Port
	}

	routeRule := buildRouterRule(host, pathPrefix) // ルールを生成する

	routeSpec := map[string]interface{}{
		"kind":  "Rule",
		"match": routeRule, // ルールを設定する
		"services": []interface{}{
			map[string]interface{}{
				"name": serviceName,  // サービス名を設定する
				"port": int64(port),  // ポートを設定する
			},
		},
	}

	spec := map[string]interface{}{
		"entryPoints": []interface{}{"web", "websecure"}, // エントリーポイントを設定する
		"routes":      []interface{}{routeSpec},          // ルートを設定する
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "traefik.io/v1alpha1",
			"kind":       "IngressRoute",
			"metadata": map[string]interface{}{
				"name":      ingressRouteData.ID, // IngressRoute 名を設定する（deployment ID を使う）
				"namespace": namespace,           // namespace を設定する
				"labels": map[string]interface{}{
					"launchs-managed": "true", // launchs が管理するリソースであることを示すラベル
					"generated":       "1",    // 自動生成リソースであることを示すラベル
				},
			},
			"spec": spec,
		},
	}
}

// ApplyIngressRoute は Traefik IngressRoute を作成または更新する
func ApplyIngressRoute(ctx context.Context, client dynamic.Interface, ingressRouteData models.IngressRoute, namespace, serviceName string, servicePort int) error {
	manifest := buildIngressRouteManifest(ingressRouteData, namespace, serviceName, servicePort) // マニフェストを生成する

	existing, err := client.Resource(traefikIngressRouteGVR).Namespace(namespace).Get(ctx, manifest.GetName(), metav1.GetOptions{}) // 既存の IngressRoute を取得する
	if err != nil {
		// 存在しない場合は新規作成する
		_, err = client.Resource(traefikIngressRouteGVR).Namespace(namespace).Create(ctx, manifest, metav1.CreateOptions{})
		return err
	}

	// 既存の IngressRoute を更新する
	manifest.SetResourceVersion(existing.GetResourceVersion()) // 楽観的並行性制御のため ResourceVersion を引き継ぐ
	_, err = client.Resource(traefikIngressRouteGVR).Namespace(namespace).Update(ctx, manifest, metav1.UpdateOptions{})
	return err
}

// DeleteIngressRoute は Traefik IngressRoute を削除する
func DeleteIngressRoute(ctx context.Context, client dynamic.Interface, namespace, name string) error {
	return client.Resource(traefikIngressRouteGVR).Namespace(namespace).Delete(ctx, name, metav1.DeleteOptions{}) // IngressRoute を削除する
}
