package service

import (
	"app/models"
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"
)

// mockDeploymentRepository は DeploymentRepository のテスト用モック実装
type mockDeploymentRepository struct {
	createFunc              func(ctx context.Context, deployment *models.Deployment) error
	findByIDFunc            func(ctx context.Context, deploymentID string) (*models.Deployment, error)
	findByIDForUpdateFunc   func(ctx context.Context, tx *gorm.DB, deploymentID string) (*models.Deployment, error)
	findAllByProjectIDFunc  func(ctx context.Context, projectID string) ([]models.Deployment, error)
	saveFunc                func(ctx context.Context, deployment *models.Deployment) error
	updatesFunc             func(ctx context.Context, tx *gorm.DB, deployment *models.Deployment, values map[string]interface{}) error
}

func (mock *mockDeploymentRepository) Create(ctx context.Context, deployment *models.Deployment) error {
	return mock.createFunc(ctx, deployment) // モック関数を呼び出す
}

func (mock *mockDeploymentRepository) FindByID(ctx context.Context, deploymentID string) (*models.Deployment, error) {
	return mock.findByIDFunc(ctx, deploymentID) // モック関数を呼び出す
}

func (mock *mockDeploymentRepository) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, deploymentID string) (*models.Deployment, error) {
	if mock.findByIDForUpdateFunc != nil { // モック関数が設定されている場合は呼び出す
		return mock.findByIDForUpdateFunc(ctx, tx, deploymentID)
	}
	return nil, nil // デフォルトは nil を返す
}

func (mock *mockDeploymentRepository) FindAllByProjectID(ctx context.Context, projectID string) ([]models.Deployment, error) {
	return mock.findAllByProjectIDFunc(ctx, projectID) // モック関数を呼び出す
}

func (mock *mockDeploymentRepository) Save(ctx context.Context, deployment *models.Deployment) error {
	return mock.saveFunc(ctx, deployment) // モック関数を呼び出す
}

func (mock *mockDeploymentRepository) Updates(ctx context.Context, tx *gorm.DB, deployment *models.Deployment, values map[string]interface{}) error {
	if mock.updatesFunc != nil { // モック関数が設定されている場合は呼び出す
		return mock.updatesFunc(ctx, tx, deployment, values)
	}
	return nil // デフォルトは nil を返す
}

// mockServiceRepository は ServiceRepository のテスト用モック実装
type mockServiceRepository struct {
	createFunc func(ctx context.Context, service *models.Service) error
}

func (mock *mockServiceRepository) Create(ctx context.Context, service *models.Service) error {
	return mock.createFunc(ctx, service) // モック関数を呼び出す
}

// mockProjectRepository は ProjectRepository のテスト用モック実装（所有権チェック用）
type mockProjectRepository struct {
	findByIDNoTxFunc func(ctx context.Context, projectID string) (*models.Project, error)
}

func (mock *mockProjectRepository) Create(ctx context.Context, tx *gorm.DB, project *models.Project) error {
	return nil // 使用しない
}

func (mock *mockProjectRepository) FindByID(ctx context.Context, tx *gorm.DB, projectID string) (*models.Project, error) {
	return nil, nil // 使用しない
}

func (mock *mockProjectRepository) FindByIDNoTx(ctx context.Context, projectID string) (*models.Project, error) {
	if mock.findByIDNoTxFunc != nil { // モック関数が設定されている場合は呼び出す
		return mock.findByIDNoTxFunc(ctx, projectID)
	}
	return &models.Project{UserID: "test-user-id"}, nil // デフォルトは所有者として返す
}

func (mock *mockProjectRepository) FindAllByUserID(ctx context.Context, userID string) ([]*models.Project, error) {
	return nil, nil // 使用しない
}

func (mock *mockProjectRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, project *models.Project, status models.ProjectStatus) error {
	return nil // 使用しない
}

func (mock *mockProjectRepository) Save(ctx context.Context, project *models.Project) error {
	return nil // 使用しない
}

func (mock *mockProjectRepository) Delete(ctx context.Context, tx *gorm.DB, project *models.Project) error {
	return nil // 使用しない
}

// TestCreateDeployment_正常に作成されpendingフィールドに値が入る は POST で全フィールドが pending_*** に入ることを確認する
func TestCreateDeployment_正常に作成されpendingフィールドに値が入る(t *testing.T) {
	var capturedDeployment *models.Deployment // キャプチャした deployment を格納する変数を定義する

	deploymentRepo := &mockDeploymentRepository{
		createFunc: func(ctx context.Context, deployment *models.Deployment) error {
			capturedDeployment = deployment // deployment をキャプチャする
			deployment.ID = "deployment-id-1" // ID を付与する
			return nil
		},
	}
	serviceRepo := &mockServiceRepository{
		createFunc: func(ctx context.Context, svc *models.Service) error {
			return nil // 正常終了を返す
		},
	}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	req := CreateDeploymentRequest{
		ProjectID:    "project-id-1",  // プロジェクト ID を設定する
		Name:         "my-app",        // デプロイメント名を設定する
		Type:         "image_url",     // タイプを設定する
		ImageURL:     "nginx:latest",  // image_url を設定する
		InstanceSize: "small",         // インスタンスサイズを設定する
		Replicas:     2,               // レプリカ数を設定する
	}

	result, err := deploymentSvc.CreateDeployment(context.Background(), req) // サービスを実行する
	if err != nil {
		t.Fatalf("CreateDeployment がエラーを返しました: %v", err)
	}
	if result.Status != models.DeploymentStatusPending { // status が pending であることを確認する
		t.Errorf("期待する status: pending, 実際の status: %s", result.Status)
	}
	if result.AppStatus != models.AppStatusPending { // app_status が pending であることを確認する
		t.Errorf("期待する app_status: pending, 実際の app_status: %s", result.AppStatus)
	}
	if capturedDeployment.PendingImageURL != "nginx:latest" { // pending_image_url が設定されていることを確認する
		t.Errorf("期待する pending_image_url: nginx:latest, 実際の pending_image_url: %s", capturedDeployment.PendingImageURL)
	}
	if capturedDeployment.PendingInstanceSize != "small" { // pending_instance_size が設定されていることを確認する
		t.Errorf("期待する pending_instance_size: small, 実際の pending_instance_size: %s", capturedDeployment.PendingInstanceSize)
	}
	if capturedDeployment.PendingReplicas != 2 { // pending_replicas が設定されていることを確認する
		t.Errorf("期待する pending_replicas: 2, 実際の pending_replicas: %d", capturedDeployment.PendingReplicas)
	}
}

// TestCreateDeployment_デフォルト値が適用される はデフォルト値が正しく設定されることを確認する
func TestCreateDeployment_デフォルト値が適用される(t *testing.T) {
	var capturedDeployment *models.Deployment // キャプチャした deployment を格納する変数を定義する

	deploymentRepo := &mockDeploymentRepository{
		createFunc: func(ctx context.Context, deployment *models.Deployment) error {
			capturedDeployment = deployment // deployment をキャプチャする
			return nil
		},
	}
	serviceRepo := &mockServiceRepository{
		createFunc: func(ctx context.Context, svc *models.Service) error {
			return nil // 正常終了を返す
		},
	}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	req := CreateDeploymentRequest{
		ProjectID: "project-id-1", // 最低限の情報のみ設定する
		Name:      "my-app",
		Type:      "image_url",
	}

	_, err := deploymentSvc.CreateDeployment(context.Background(), req) // サービスを実行する
	if err != nil {
		t.Fatalf("CreateDeployment がエラーを返しました: %v", err)
	}
	if capturedDeployment.PendingInstanceSize != "small" { // instance_size のデフォルト値を確認する
		t.Errorf("期待する pending_instance_size: small, 実際の pending_instance_size: %s", capturedDeployment.PendingInstanceSize)
	}
	if capturedDeployment.PendingReplicas != 1 { // replicas のデフォルト値を確認する
		t.Errorf("期待する pending_replicas: 1, 実際の pending_replicas: %d", capturedDeployment.PendingReplicas)
	}
	if capturedDeployment.PendingDockerfilePath != "./Dockerfile" { // dockerfile_path のデフォルト値を確認する
		t.Errorf("期待する pending_dockerfile_path: ./Dockerfile, 実際の pending_dockerfile_path: %s", capturedDeployment.PendingDockerfilePath)
	}
	if capturedDeployment.PendingGithubRepoDirectory != "./" { // build_directory のデフォルト値を確認する
		t.Errorf("期待する pending_github_repo_directory: ./, 実際の pending_github_repo_directory: %s", capturedDeployment.PendingGithubRepoDirectory)
	}
}

// TestCreateDeployment_Serviceレコードも作成される は Service レコードが同時に作成されることを確認する
func TestCreateDeployment_Serviceレコードも作成される(t *testing.T) {
	var capturedService *models.Service // キャプチャした service を格納する変数を定義する

	deploymentRepo := &mockDeploymentRepository{
		createFunc: func(ctx context.Context, deployment *models.Deployment) error {
			deployment.ID = "deployment-id-1" // ID を付与する
			return nil
		},
	}
	serviceRepo := &mockServiceRepository{
		createFunc: func(ctx context.Context, svc *models.Service) error {
			capturedService = svc // service をキャプチャする
			return nil
		},
	}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	req := CreateDeploymentRequest{ProjectID: "project-id-1", Name: "my-app", Type: "image_url"}

	_, err := deploymentSvc.CreateDeployment(context.Background(), req) // サービスを実行する
	if err != nil {
		t.Fatalf("CreateDeployment がエラーを返しました: %v", err)
	}
	if capturedService == nil { // Service レコードが作成されていることを確認する
		t.Fatal("Service レコードが作成されていません")
	}
	if capturedService.DeploymentID != "deployment-id-1" { // deployment_id が設定されていることを確認する
		t.Errorf("期待する deployment_id: deployment-id-1, 実際の deployment_id: %s", capturedService.DeploymentID)
	}
	if capturedService.Status != models.ServiceStatusPending { // status が pending であることを確認する
		t.Errorf("期待する status: pending, 実際の status: %s", capturedService.Status)
	}
}

// TestCreateDeployment_Deployment作成失敗時にエラーを返す は Deployment 作成失敗時にエラーが返ることを確認する
func TestCreateDeployment_Deployment作成失敗時にエラーを返す(t *testing.T) {
	deploymentRepo := &mockDeploymentRepository{
		createFunc: func(ctx context.Context, deployment *models.Deployment) error {
			return errors.New("DB エラー") // 作成失敗を返す
		},
	}
	serviceRepo := &mockServiceRepository{
		createFunc: func(ctx context.Context, svc *models.Service) error {
			return nil
		},
	}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	req := CreateDeploymentRequest{ProjectID: "project-id-1", Name: "my-app", Type: "image_url"}

	_, err := deploymentSvc.CreateDeployment(context.Background(), req) // サービスを実行する
	if err == nil { // エラーが返ることを確認する
		t.Fatal("エラーが返るべきですが nil が返りました")
	}
}

// TestCreateDeployment_Service作成失敗時にエラーを返す は Service 作成失敗時にエラーが返ることを確認する
func TestCreateDeployment_Service作成失敗時にエラーを返す(t *testing.T) {
	deploymentRepo := &mockDeploymentRepository{
		createFunc: func(ctx context.Context, deployment *models.Deployment) error {
			deployment.ID = "deployment-id-1" // ID を付与する
			return nil
		},
	}
	serviceRepo := &mockServiceRepository{
		createFunc: func(ctx context.Context, svc *models.Service) error {
			return errors.New("DB エラー") // Service 作成失敗を返す
		},
	}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	req := CreateDeploymentRequest{ProjectID: "project-id-1", Name: "my-app", Type: "image_url"}

	_, err := deploymentSvc.CreateDeployment(context.Background(), req) // サービスを実行する
	if err == nil { // エラーが返ることを確認する
		t.Fatal("エラーが返るべきですが nil が返りました")
	}
}

// TestUpdateDeployment_送ったフィールドのみpendingが更新される は送ったフィールドのみ更新されることを確認する
func TestUpdateDeployment_送ったフィールドのみpendingが更新される(t *testing.T) {
	originalDeployment := &models.Deployment{
		ID:                  "deployment-id-1",
		PendingImageURL:     "nginx:1.24",  // 更新前の値を設定する
		PendingInstanceSize: "small",
		PendingReplicas:     1,
	}

	var savedDeployment *models.Deployment // 保存された deployment を格納する変数を定義する
	deploymentRepo := &mockDeploymentRepository{
		findByIDFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return originalDeployment, nil // 元の deployment を返す
		},
		saveFunc: func(ctx context.Context, deployment *models.Deployment) error {
			savedDeployment = deployment // 保存された deployment をキャプチャする
			return nil
		},
	}
	serviceRepo := &mockServiceRepository{}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	newImageURL := "nginx:1.25"                                        // 更新する image_url を設定する
	req := UpdateDeploymentRequest{
		ImageURL: &newImageURL, // image_url のみ送る
	}

	_, err := deploymentSvc.UpdateDeployment(context.Background(), "test-user-id", "deployment-id-1", req) // サービスを実行する
	if err != nil {
		t.Fatalf("UpdateDeployment がエラーを返しました: %v", err)
	}
	if savedDeployment.PendingImageURL != "nginx:1.25" { // image_url が更新されていることを確認する
		t.Errorf("期待する pending_image_url: nginx:1.25, 実際の pending_image_url: %s", savedDeployment.PendingImageURL)
	}
	if savedDeployment.PendingInstanceSize != "small" { // 送っていない instance_size が変化していないことを確認する
		t.Errorf("instance_size は変化しないはずですが変化しています: %s", savedDeployment.PendingInstanceSize)
	}
	if savedDeployment.PendingReplicas != 1 { // 送っていない replicas が変化していないことを確認する
		t.Errorf("replicas は変化しないはずですが変化しています: %d", savedDeployment.PendingReplicas)
	}
}

// TestUpdateDeployment_存在しないdeploymentはエラーを返す は存在しない deployment ID でエラーが返ることを確認する
func TestUpdateDeployment_存在しないdeploymentはエラーを返す(t *testing.T) {
	deploymentRepo := &mockDeploymentRepository{
		findByIDFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return nil, gorm.ErrRecordNotFound // レコードが存在しないエラーを返す
		},
	}
	serviceRepo := &mockServiceRepository{}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する
	req := UpdateDeploymentRequest{}

	_, err := deploymentSvc.UpdateDeployment(context.Background(), "test-user-id", "nonexistent", req) // サービスを実行する
	if err == nil { // エラーが返ることを確認する
		t.Fatal("エラーが返るべきですが nil が返りました")
	}
}

// TestDeleteDeployment_statusがdeletingになる は status が deleting に変更されることを確認する
func TestDeleteDeployment_statusがdeletingになる(t *testing.T) {
	originalDeployment := &models.Deployment{
		ID:     "deployment-id-1",
		Status: models.DeploymentStatusRunning, // 更新前の status を設定する
	}

	var savedDeployment *models.Deployment // 保存された deployment を格納する変数を定義する
	deploymentRepo := &mockDeploymentRepository{
		findByIDFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return originalDeployment, nil // 元の deployment を返す
		},
		saveFunc: func(ctx context.Context, deployment *models.Deployment) error {
			savedDeployment = deployment // 保存された deployment をキャプチャする
			return nil
		},
	}
	serviceRepo := &mockServiceRepository{}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する

	result, err := deploymentSvc.DeleteDeployment(context.Background(), "test-user-id", "deployment-id-1") // サービスを実行する
	if err != nil {
		t.Fatalf("DeleteDeployment がエラーを返しました: %v", err)
	}
	if result.Status != models.DeploymentStatusDeleting { // status が deleting であることを確認する
		t.Errorf("期待する status: deleting, 実際の status: %s", result.Status)
	}
	if savedDeployment.Status != models.DeploymentStatusDeleting { // 保存された status が deleting であることを確認する
		t.Errorf("保存された status が deleting ではありません: %s", savedDeployment.Status)
	}
}

// TestDeleteDeployment_存在しないdeploymentはエラーを返す は存在しない deployment ID でエラーが返ることを確認する
func TestDeleteDeployment_存在しないdeploymentはエラーを返す(t *testing.T) {
	deploymentRepo := &mockDeploymentRepository{
		findByIDFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return nil, gorm.ErrRecordNotFound // レコードが存在しないエラーを返す
		},
	}
	serviceRepo := &mockServiceRepository{}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する

	_, err := deploymentSvc.DeleteDeployment(context.Background(), "test-user-id", "nonexistent") // サービスを実行する
	if err == nil { // エラーが返ることを確認する
		t.Fatal("エラーが返るべきですが nil が返りました")
	}
}

// TestListDeployments_正常に一覧が返る は deployment 一覧が返ることを確認する
func TestListDeployments_正常に一覧が返る(t *testing.T) {
	expectedList := []models.Deployment{
		{ID: "deployment-id-1", Name: "app-1", ProjectID: "project-id-1"},
		{ID: "deployment-id-2", Name: "app-2", ProjectID: "project-id-1"},
	}

	deploymentRepo := &mockDeploymentRepository{
		findAllByProjectIDFunc: func(ctx context.Context, projectID string) ([]models.Deployment, error) {
			return expectedList, nil // 期待する一覧を返す
		},
	}
	serviceRepo := &mockServiceRepository{}

	deploymentSvc := NewDeploymentService(deploymentRepo, serviceRepo, &mockProjectRepository{}) // サービスを生成する

	result, err := deploymentSvc.ListDeployments(context.Background(), "project-id-1") // サービスを実行する
	if err != nil {
		t.Fatalf("ListDeployments がエラーを返しました: %v", err)
	}
	if len(result) != 2 { // deployment 数を確認する
		t.Errorf("期待する deployment 数: 2, 実際の deployment 数: %d", len(result))
	}
}
