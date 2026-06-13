package k8s

import (
	"context"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

// テスト用の Deployment manifest を生成するヘルパー関数
func newTestDeploymentManifest(name, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": name}, // セレクターを設定する
			},
		},
	}
}

// TestApplyDeployment_正常にDeploymentが作成される は ApplyDeployment で k8s に Deployment が作成されることを確認する
func TestApplyDeployment_正常にDeploymentが作成される(t *testing.T) {
	fakeClient := fake.NewSimpleClientset() // fake k8s クライアントを生成する
	ctx := context.Background()            // テスト用コンテキストを生成する

	deploymentManifest := newTestDeploymentManifest("test-deploy", "test-namespace") // テスト用 manifest を生成する

	err := ApplyDeployment(ctx, fakeClient, deploymentManifest) // Deployment を apply する
	if err != nil {
		t.Fatalf("ApplyDeployment() がエラーを返しました: %v", err) // エラーが返った場合はテスト失敗とする
	}

	createdDeployment, err := fakeClient.AppsV1().Deployments("test-namespace").Get(ctx, "test-deploy", metav1.GetOptions{}) // 作成された Deployment を取得する
	if err != nil {
		t.Fatalf("Deployment の取得に失敗しました: %v", err) // 取得失敗時はテスト失敗とする
	}
	if createdDeployment.Name != "test-deploy" { // Deployment 名を確認する
		t.Errorf("期待する Deployment 名: test-deploy, 実際: %s", createdDeployment.Name)
	}
}

// TestApplyDeployment_同名Deploymentを再applyすると更新される は同名の Deployment を再度 apply すると更新されることを確認する
func TestApplyDeployment_同名Deploymentを再applyすると更新される(t *testing.T) {
	fakeClient := fake.NewSimpleClientset() // fake k8s クライアントを生成する
	ctx := context.Background()            // テスト用コンテキストを生成する

	deploymentManifest := newTestDeploymentManifest("update-deploy", "test-namespace") // テスト用 manifest を生成する

	err := ApplyDeployment(ctx, fakeClient, deploymentManifest) // 1回目の apply（作成）
	if err != nil {
		t.Fatalf("1回目の ApplyDeployment() がエラーを返しました: %v", err) // 1回目は成功するべきなのでテスト失敗とする
	}

	updatedManifest := newTestDeploymentManifest("update-deploy", "test-namespace") // 更新用 manifest を生成する
	updatedManifest.Labels = map[string]string{"updated": "true"}                  // ラベルを追加して更新内容を確認できるようにする

	err = ApplyDeployment(ctx, fakeClient, updatedManifest) // 2回目の apply（更新）
	if err != nil {
		t.Fatalf("2回目の ApplyDeployment() がエラーを返しました: %v", err) // エラーが返った場合はテスト失敗とする
	}

	updatedDeployment, err := fakeClient.AppsV1().Deployments("test-namespace").Get(ctx, "update-deploy", metav1.GetOptions{}) // 更新後の Deployment を取得する
	if err != nil {
		t.Fatalf("更新後の Deployment 取得に失敗しました: %v", err) // 取得失敗時はテスト失敗とする
	}
	if updatedDeployment.Labels["updated"] != "true" { // 更新が反映されていることを確認する
		t.Errorf("Deployment の更新が反映されていません。Labels: %v", updatedDeployment.Labels)
	}
}

// TestDeleteDeployment_正常にDeploymentが削除される は DeleteDeployment で k8s から Deployment が削除されることを確認する
func TestDeleteDeployment_正常にDeploymentが削除される(t *testing.T) {
	fakeClient := fake.NewSimpleClientset() // fake k8s クライアントを生成する
	ctx := context.Background()            // テスト用コンテキストを生成する

	deploymentManifest := newTestDeploymentManifest("delete-deploy", "test-namespace") // テスト用 manifest を生成する

	err := ApplyDeployment(ctx, fakeClient, deploymentManifest) // 削除対象の Deployment を作成する
	if err != nil {
		t.Fatalf("事前の ApplyDeployment() がエラーを返しました: %v", err) // 前提条件の作成失敗時はテスト失敗とする
	}

	err = DeleteDeployment(ctx, fakeClient, "test-namespace", "delete-deploy") // Deployment を削除する
	if err != nil {
		t.Fatalf("DeleteDeployment() がエラーを返しました: %v", err) // エラーが返った場合はテスト失敗とする
	}

	_, err = fakeClient.AppsV1().Deployments("test-namespace").Get(ctx, "delete-deploy", metav1.GetOptions{}) // 削除後に取得を試みる
	if err == nil {
		t.Fatal("削除後も Deployment が存在しています") // 削除後に取得できた場合はテスト失敗とする
	}
}
