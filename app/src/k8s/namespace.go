package k8s

import (
	"context"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var invalidNamespaceChars = regexp.MustCompile(`[^a-z0-9-]`) // namespace に使用できない文字にマッチする正規表現

// ToNamespaceName はプロジェクト名を k8s namespace 名として有効な文字列に変換する
// 大文字を小文字に変換し、英小文字・数字・ハイフン以外の文字をハイフンに置換する
func ToNamespaceName(name string) string {
	lower := strings.ToLower(name)                             // 大文字を小文字に変換する
	replaced := invalidNamespaceChars.ReplaceAllString(lower, "-") // 使用不可文字をハイフンに置換する
	trimmed := strings.Trim(replaced, "-")                     // 先頭・末尾のハイフンを除去する
	return trimmed
}

// CreateNamespace は指定した名前の k8s namespace を作成する
func CreateNamespace(ctx context.Context, client kubernetes.Interface, name string) error {
	namespaceObj := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name, // namespace 名を設定する
			Labels: map[string]string{
				"launchs.org/managed": "true", // このサービスが管理する namespace であることを示すラベル
			},
		},
	}
	_, err := client.CoreV1().Namespaces().Create(ctx, namespaceObj, metav1.CreateOptions{}) // namespace を作成する
	return err
}

// DeleteNamespace は指定した名前の k8s namespace を削除する
func DeleteNamespace(ctx context.Context, client kubernetes.Interface, name string) error {
	return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{}) // namespace を削除する
}
