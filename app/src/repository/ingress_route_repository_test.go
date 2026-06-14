package repository

import (
	"app/models"
	"context"
	"testing"
)

// TestIngressRouteRepository_Create_正常に作成される は IngressRoute レコードが作成されることを確認する
func TestIngressRouteRepository_Create_正常に作成される(t *testing.T) {
	db := setupTestDB(t)                     // テスト用 DB を準備する
	projectData := createTestProject(t, db) // テスト用 Project を作成する

	// テスト用 Deployment を作成する
	deploymentData := &models.Deployment{
		ProjectID: projectData.ID,                 // プロジェクト ID を設定する
		Name:      "test-app-ingress-create",      // デプロイメント名を設定する
		Type:      models.DeploymentTypeImageURL,  // タイプを設定する
		Status:    models.DeploymentStatusPending, // ステータスを設定する
		AppStatus: models.AppStatusPending,        // アプリステータスを設定する
	}
	if err := db.Create(deploymentData).Error; err != nil { // テスト用レコードを作成する
		t.Fatalf("テスト用 Deployment の作成に失敗しました: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(deploymentData) }) // テスト終了後にレコードを削除する

	repo := NewIngressRouteRepository(db) // リポジトリを生成する
	ingressRouteData := &models.IngressRoute{
		DeploymentID: deploymentData.ID,             // deployment_id を設定する
		Host:         "example.launchs.org",         // ホスト名を設定する
		PathPrefix:   "/",                           // パスプレフィックスを設定する
		Port:         8080,                          // ポート番号を設定する
		TLSEnabled:   false,                         // TLS 無効を設定する
		Status:       models.IngressRouteStatusPending, // ステータスを設定する
	}

	err := repo.Create(context.Background(), ingressRouteData) // リポジトリを実行する
	if err != nil {
		t.Fatalf("Create がエラーを返しました: %v", err)
	}
	if ingressRouteData.ID == "" { // ID が付与されていることを確認する
		t.Error("作成後に ID が設定されていません")
	}
	t.Cleanup(func() { db.Unscoped().Delete(ingressRouteData) }) // テスト終了後にレコードを削除する

	// DB から取得して値を確認する
	var fetchedIngressRoute models.IngressRoute
	db.First(&fetchedIngressRoute, "id = ?", ingressRouteData.ID) // 作成したレコードを取得する
	if fetchedIngressRoute.Host != "example.launchs.org" {        // host が一致することを確認する
		t.Errorf("期待する host: example.launchs.org, 実際の host: %s", fetchedIngressRoute.Host)
	}
	if fetchedIngressRoute.Port != 8080 { // port が一致することを確認する
		t.Errorf("期待する port: 8080, 実際の port: %d", fetchedIngressRoute.Port)
	}
	if fetchedIngressRoute.Status != models.IngressRouteStatusPending { // status が pending であることを確認する
		t.Errorf("期待する status: pending, 実際の status: %s", fetchedIngressRoute.Status)
	}
}

// TestIngressRouteRepository_FindByDeploymentID_正常に取得される は FindByDeploymentID で ingress_route が取得されることを確認する
func TestIngressRouteRepository_FindByDeploymentID_正常に取得される(t *testing.T) {
	db := setupTestDB(t)                     // テスト用 DB を準備する
	projectData := createTestProject(t, db) // テスト用 Project を作成する

	// テスト用 Deployment を作成する
	deploymentData := &models.Deployment{
		ProjectID: projectData.ID,                    // プロジェクト ID を設定する
		Name:      "test-app-ingress-find",           // デプロイメント名を設定する
		Type:      models.DeploymentTypeImageURL,     // タイプを設定する
		Status:    models.DeploymentStatusPending,    // ステータスを設定する
		AppStatus: models.AppStatusPending,           // アプリステータスを設定する
	}
	if err := db.Create(deploymentData).Error; err != nil { // テスト用レコードを作成する
		t.Fatalf("テスト用 Deployment の作成に失敗しました: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(deploymentData) }) // テスト終了後にレコードを削除する

	// テスト用 IngressRoute を直接 DB に作成する
	ingressRouteData := &models.IngressRoute{
		DeploymentID: deploymentData.ID,             // deployment_id を設定する
		Host:         "find-test.launchs.org",       // ホスト名を設定する
		PathPrefix:   "/api",                        // パスプレフィックスを設定する
		Port:         9090,                          // ポート番号を設定する
		Status:       models.IngressRouteStatusActive, // ステータスを設定する
	}
	if err := db.Create(ingressRouteData).Error; err != nil { // テスト用レコードを作成する
		t.Fatalf("テスト用 IngressRoute の作成に失敗しました: %v", err)
	}
	t.Cleanup(func() { db.Unscoped().Delete(ingressRouteData) }) // テスト終了後にレコードを削除する

	repo := NewIngressRouteRepository(db) // リポジトリを生成する

	result, err := repo.FindByDeploymentID(context.Background(), deploymentData.ID) // リポジトリを実行する
	if err != nil {
		t.Fatalf("FindByDeploymentID がエラーを返しました: %v", err)
	}
	if result.Host != "find-test.launchs.org" { // host が一致することを確認する
		t.Errorf("期待する host: find-test.launchs.org, 実際の host: %s", result.Host)
	}
	if result.Port != 9090 { // port が一致することを確認する
		t.Errorf("期待する port: 9090, 実際の port: %d", result.Port)
	}
	if result.PathPrefix != "/api" { // path_prefix が一致することを確認する
		t.Errorf("期待する path_prefix: /api, 実際の path_prefix: %s", result.PathPrefix)
	}
}
