package service

import (
	"app/models"
	"app/repository"
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
)

var (
	applyTestDB     *gorm.DB   // テスト用 DB 接続（パッケージ内で共有する）
	applyTestDBOnce sync.Once  // 初期化を一度だけ実行するための Once
)

// setupApplyTestDB はテスト用の DB 接続とスキーマを準備する（パッケージ内で一度だけ初期化する）
func setupApplyTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	applyTestDBOnce.Do(func() { // 一度だけ実行する
		dsn := fmt.Sprintf(
			"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Tokyo",
			getApplyEnvOrDefault("DB_HOST", "localhost"),
			getApplyEnvOrDefault("DB_USER", "postgres"),
			getApplyEnvOrDefault("DB_PASSWORD", "postgres"),
			getApplyEnvOrDefault("DB_NAME", "postgres"),
			getApplyEnvOrDefault("DB_PORT", "5432"),
		)

		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{}) // DB に接続する
		if err != nil {
			return // 接続失敗時は nil のまま（テスト関数でスキップする）
		}

		// テストに必要なテーブルをマイグレーションする（一度だけ実行する）
		if migrateErr := db.AutoMigrate(
			&models.InstanceSize{},
			&models.UserQuota{},
			&models.Project{},
			&models.HarborCredential{},
			&models.Deployment{},
			&models.DeploymentBuild{},
			&models.ApplyHistory{},
			&models.DeploymentWebhook{},
			&models.Service{},
			&models.IngressRoute{},
			&models.EnvVar{},
			&models.EnvVarMount{},
			&models.Volume{},
			&models.VolumeMount{},
		); migrateErr != nil {
			return // マイグレーション失敗時は nil のまま
		}

		applyTestDB = db // 成功時のみセットする
	})

	if applyTestDB == nil { // DB が取得できない場合はスキップする
		t.Skip("DB に接続できないためテストをスキップします")
	}
	return applyTestDB
}

// getApplyEnvOrDefault は環境変数を取得し、未設定の場合はデフォルト値を返す
func getApplyEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value // 環境変数が設定されている場合はその値を返す
	}
	return defaultValue // 未設定の場合はデフォルト値を返す
}

// createApplyTestProject はテスト用の Project レコードを作成するヘルパー関数
func createApplyTestProject(t *testing.T, db *gorm.DB, namespace string) *models.Project {
	t.Helper()
	projectData := &models.Project{
		UserID:    "test-user-id",         // テスト用ユーザー ID を設定する
		Name:      "test-project-" + namespace, // テスト用プロジェクト名を設定する
		Namespace: namespace,              // テスト用 namespace を設定する
		Status:    models.ProjectStatusActive, // ステータスを active に設定する
	}
	if err := db.Create(projectData).Error; err != nil {
		t.Fatalf("テスト用 Project の作成に失敗しました: %v", err) // 作成失敗時はテスト失敗とする
	}
	t.Cleanup(func() { db.Unscoped().Delete(projectData) }) // テスト終了後にレコードを削除する
	return projectData
}

// createApplyTestDeployment はテスト用の Deployment レコードを作成するヘルパー関数
func createApplyTestDeployment(t *testing.T, db *gorm.DB, projectID string, name string) *models.Deployment {
	t.Helper()
	deploymentData := &models.Deployment{
		ProjectID:           projectID,                          // プロジェクト ID を設定する
		Name:                name,                               // デプロイメント名を設定する
		Type:                models.DeploymentTypeImageURL,      // タイプを設定する
		Status:              models.DeploymentStatusPending,     // ステータスを pending に設定する
		AppStatus:           models.AppStatusPending,            // アプリステータスを pending に設定する
		PendingImageURL:     "nginx:latest",                     // pending image_url を設定する
		PendingInstanceSize: "small",                            // pending instance_size を設定する
		PendingReplicas:     1,                                  // pending replicas を設定する
	}
	if err := db.Create(deploymentData).Error; err != nil {
		t.Fatalf("テスト用 Deployment の作成に失敗しました: %v", err) // 作成失敗時はテスト失敗とする
	}
	t.Cleanup(func() { db.Unscoped().Delete(deploymentData) }) // テスト終了後にレコードを削除する
	return deploymentData
}

// ensureInstanceSize はテスト用の InstanceSize レコードが存在することを確認するヘルパー関数
func ensureInstanceSize(t *testing.T, db *gorm.DB) {
	t.Helper()
	instanceSizeData := &models.InstanceSize{
		Size:          "small",    // サイズ名を設定する
		CPURequest:    "100m",     // CPU リクエストを設定する
		CPULimit:      "500m",     // CPU リミットを設定する
		MemoryRequest: "128Mi",    // メモリリクエストを設定する
		MemoryLimit:   "512Mi",    // メモリリミットを設定する
	}
	db.Where("size = ?", "small").FirstOrCreate(instanceSizeData) // 存在しない場合のみ作成する
}

// TestApplyService_Apply_正常にapplyされk8sDeploymentが作成される は apply 後に k8s Deployment が作成されることを確認する
func TestApplyService_Apply_正常にapplyされk8sDeploymentが作成される(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-1") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-1") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	fakeK8sClient := fake.NewSimpleClientset()                   // fake k8s クライアントを生成する
	deploymentRepo := repository.NewDeploymentRepository(db)     // リポジトリを生成する
	applyHistoryRepo := repository.NewApplyHistoryRepository(db) // apply_history リポジトリを生成する
	projectRepo := repository.NewProjectRepository(db)           // project リポジトリを生成する
	applyService := NewApplyService(db, fakeK8sClient, deploymentRepo, applyHistoryRepo, projectRepo) // サービスを生成する

	result, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する
	if err != nil {
		t.Fatalf("Apply がエラーを返しました: %v", err)
	}
	if result == nil { // 結果が返ることを確認する
		t.Fatal("Apply の結果が nil です")
	}
	if result.Status != "applied" { // ステータスが applied であることを確認する
		t.Errorf("期待する status: applied, 実際の status: %s", result.Status)
	}

	// k8s に Deployment が作成されていることを確認する
	k8sDeployment, err := fakeK8sClient.AppsV1().Deployments(projectData.Namespace).Get(
		context.Background(), deploymentData.Name, metav1.GetOptions{},
	)
	if err != nil { // k8s Deployment が作成されていることを確認する
		t.Fatalf("k8s Deployment が作成されていません: %v", err)
	}
	if k8sDeployment.Name != deploymentData.Name { // Deployment 名が一致することを確認する
		t.Errorf("期待する Deployment 名: %s, 実際の Deployment 名: %s", deploymentData.Name, k8sDeployment.Name)
	}
}

// TestApplyService_Apply_apply後にpendingフィールドが空になる は apply 後に pending_*** がクリアされることを確認する
func TestApplyService_Apply_apply後にpendingフィールドが空になる(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-2") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-2") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	fakeK8sClient := fake.NewSimpleClientset()                   // fake k8s クライアントを生成する
	deploymentRepo := repository.NewDeploymentRepository(db)     // リポジトリを生成する
	applyHistoryRepo := repository.NewApplyHistoryRepository(db) // apply_history リポジトリを生成する
	projectRepo := repository.NewProjectRepository(db)           // project リポジトリを生成する
	applyService := NewApplyService(db, fakeK8sClient, deploymentRepo, applyHistoryRepo, projectRepo) // サービスを生成する

	_, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する
	if err != nil {
		t.Fatalf("Apply がエラーを返しました: %v", err)
	}

	// DB から取得して pending フィールドが空になっていることを確認する
	var fetchedDeployment models.Deployment
	db.First(&fetchedDeployment, "id = ?", deploymentData.ID) // apply 後のレコードを取得する
	if fetchedDeployment.PendingImageURL != "" { // pending_image_url がクリアされていることを確認する
		t.Errorf("pending_image_url がクリアされていません: %s", fetchedDeployment.PendingImageURL)
	}
	if fetchedDeployment.PendingInstanceSize != "" { // pending_instance_size がクリアされていることを確認する
		t.Errorf("pending_instance_size がクリアされていません: %s", fetchedDeployment.PendingInstanceSize)
	}
	if fetchedDeployment.PendingReplicas != 0 { // pending_replicas がクリアされていることを確認する
		t.Errorf("pending_replicas がクリアされていません: %d", fetchedDeployment.PendingReplicas)
	}
}

// TestApplyService_Apply_apply後にcurrent値が更新される は apply 後に current フィールドが昇格されることを確認する
func TestApplyService_Apply_apply後にcurrent値が更新される(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-3") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-3") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	fakeK8sClient := fake.NewSimpleClientset()                   // fake k8s クライアントを生成する
	deploymentRepo := repository.NewDeploymentRepository(db)     // リポジトリを生成する
	applyHistoryRepo := repository.NewApplyHistoryRepository(db) // apply_history リポジトリを生成する
	projectRepo := repository.NewProjectRepository(db)           // project リポジトリを生成する
	applyService := NewApplyService(db, fakeK8sClient, deploymentRepo, applyHistoryRepo, projectRepo) // サービスを生成する

	_, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する
	if err != nil {
		t.Fatalf("Apply がエラーを返しました: %v", err)
	}

	// DB から取得して current 値が昇格されていることを確認する
	var fetchedDeployment models.Deployment
	db.First(&fetchedDeployment, "id = ?", deploymentData.ID) // apply 後のレコードを取得する
	if fetchedDeployment.ImageURL != "nginx:latest" { // image_url が昇格されていることを確認する
		t.Errorf("期待する image_url: nginx:latest, 実際の image_url: %s", fetchedDeployment.ImageURL)
	}
	if fetchedDeployment.InstanceSize != "small" { // instance_size が昇格されていることを確認する
		t.Errorf("期待する instance_size: small, 実際の instance_size: %s", fetchedDeployment.InstanceSize)
	}
	if fetchedDeployment.Replicas != 1 { // replicas が昇格されていることを確認する
		t.Errorf("期待する replicas: 1, 実際の replicas: %d", fetchedDeployment.Replicas)
	}
}

// TestApplyService_Apply_apply後にstatusがrunningappstatusがdeployingになる は apply 後のステータス遷移を確認する
func TestApplyService_Apply_apply後にstatusがrunningappstatusがdeployingになる(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-4") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-4") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	fakeK8sClient := fake.NewSimpleClientset()                   // fake k8s クライアントを生成する
	deploymentRepo := repository.NewDeploymentRepository(db)     // リポジトリを生成する
	applyHistoryRepo := repository.NewApplyHistoryRepository(db) // apply_history リポジトリを生成する
	projectRepo := repository.NewProjectRepository(db)           // project リポジトリを生成する
	applyService := NewApplyService(db, fakeK8sClient, deploymentRepo, applyHistoryRepo, projectRepo) // サービスを生成する

	_, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する
	if err != nil {
		t.Fatalf("Apply がエラーを返しました: %v", err)
	}

	// DB から取得してステータスを確認する
	var fetchedDeployment models.Deployment
	db.First(&fetchedDeployment, "id = ?", deploymentData.ID) // apply 後のレコードを取得する
	if fetchedDeployment.Status != models.DeploymentStatusRunning { // status が running であることを確認する
		t.Errorf("期待する status: running, 実際の status: %s", fetchedDeployment.Status)
	}
	if fetchedDeployment.AppStatus != models.AppStatusDeploying { // app_status が deploying であることを確認する
		t.Errorf("期待する app_status: deploying, 実際の app_status: %s", fetchedDeployment.AppStatus)
	}
	if fetchedDeployment.AppliedAt == nil { // applied_at が設定されていることを確認する
		t.Error("applied_at が設定されていません")
	}
}

// TestApplyService_Apply_applyHistoryが1件作成される は apply 後に apply_history が1件作成されることを確認する
func TestApplyService_Apply_applyHistoryが1件作成される(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-5") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-5") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	fakeK8sClient := fake.NewSimpleClientset()                   // fake k8s クライアントを生成する
	deploymentRepo := repository.NewDeploymentRepository(db)     // リポジトリを生成する
	applyHistoryRepo := repository.NewApplyHistoryRepository(db) // apply_history リポジトリを生成する
	projectRepo := repository.NewProjectRepository(db)           // project リポジトリを生成する
	applyService := NewApplyService(db, fakeK8sClient, deploymentRepo, applyHistoryRepo, projectRepo) // サービスを生成する

	result, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する
	if err != nil {
		t.Fatalf("Apply がエラーを返しました: %v", err)
	}
	if result.ApplyHistoryID == "" { // apply_history の ID が返ることを確認する
		t.Error("ApplyHistoryID が設定されていません")
	}

	// apply_history が1件作成されていることを確認する
	var applyHistoryCount int64
	db.Model(&models.ApplyHistory{}).Where("deployment_id = ?", deploymentData.ID).Count(&applyHistoryCount) // 件数を取得する
	if applyHistoryCount != 1 { // 1件作成されていることを確認する
		t.Errorf("期待する apply_history 件数: 1, 実際の件数: %d", applyHistoryCount)
	}

	// apply_history の status が applied であることを確認する
	var applyHistory models.ApplyHistory
	db.First(&applyHistory, "deployment_id = ?", deploymentData.ID) // apply_history を取得する
	if applyHistory.Status != models.ApplyStatusApplied { // status が applied であることを確認する
		t.Errorf("期待する apply_history status: applied, 実際の status: %s", applyHistory.Status)
	}
}

// mockFailingK8sClient は k8s apply を失敗させるためのモック用 fake クライアントを準備するヘルパー
// fake.NewSimpleClientset では apply 失敗をシミュレートできないため、
// k8s.Interface を満たすカスタムモックを定義する

// applyHistoryMockDeploymentRepository は Apply テスト専用の DeploymentRepository モック
type applyHistoryMockDeploymentRepository struct {
	deploymentData *models.Deployment // 返す deployment データ
	updatedValues  map[string]interface{} // Updates で渡された values を記録する
}

func (mock *applyHistoryMockDeploymentRepository) Create(ctx context.Context, deployment *models.Deployment) error {
	return nil // 使用しない
}

func (mock *applyHistoryMockDeploymentRepository) FindByID(ctx context.Context, deploymentID string) (*models.Deployment, error) {
	return mock.deploymentData, nil // deployment を返す
}

func (mock *applyHistoryMockDeploymentRepository) FindByIDForUpdate(ctx context.Context, tx *gorm.DB, deploymentID string) (*models.Deployment, error) {
	return mock.deploymentData, nil // FOR UPDATE で deployment を返す
}

func (mock *applyHistoryMockDeploymentRepository) FindAllByProjectID(ctx context.Context, projectID string) ([]models.Deployment, error) {
	return nil, nil // 使用しない
}

func (mock *applyHistoryMockDeploymentRepository) Save(ctx context.Context, deployment *models.Deployment) error {
	return nil // 使用しない
}

func (mock *applyHistoryMockDeploymentRepository) Updates(ctx context.Context, tx *gorm.DB, deployment *models.Deployment, values map[string]interface{}) error {
	mock.updatedValues = values // 渡された values を記録する
	return nil
}

// applyHistoryMockProjectRepository は Apply テスト専用の ProjectRepository モック
type applyHistoryMockProjectRepository struct {
	projectData *models.Project // 返す project データ
}

func (mock *applyHistoryMockProjectRepository) Create(ctx context.Context, tx *gorm.DB, project *models.Project) error {
	return nil // 使用しない
}

func (mock *applyHistoryMockProjectRepository) FindByID(ctx context.Context, tx *gorm.DB, projectID string) (*models.Project, error) {
	return mock.projectData, nil // project を返す
}

func (mock *applyHistoryMockProjectRepository) FindAllByUserID(ctx context.Context, userID string) ([]*models.Project, error) {
	return nil, nil // 使用しない
}

func (mock *applyHistoryMockProjectRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, project *models.Project, status models.ProjectStatus) error {
	return nil // 使用しない
}

func (mock *applyHistoryMockProjectRepository) Save(ctx context.Context, project *models.Project) error {
	return nil // 使用しない
}

func (mock *applyHistoryMockProjectRepository) Delete(ctx context.Context, tx *gorm.DB, project *models.Project) error {
	return nil // 使用しない
}

// applyHistoryMockRepository は Apply テスト専用の ApplyHistoryRepository モック
type applyHistoryMockRepository struct {
	createdHistory  *models.ApplyHistory // Create で渡された history を記録する
	updatedStatus   models.ApplyStatus   // UpdateStatus で渡された status を記録する
	updatedHistory  *models.ApplyHistory // UpdateStatus で渡された history を記録する
}

func (mock *applyHistoryMockRepository) Create(ctx context.Context, tx *gorm.DB, history *models.ApplyHistory) error {
	history.ID = "apply-history-id-1"  // テスト用 ID を付与する
	mock.createdHistory = history       // 記録する
	return nil
}

func (mock *applyHistoryMockRepository) UpdateStatus(ctx context.Context, tx *gorm.DB, history *models.ApplyHistory, status models.ApplyStatus) error {
	mock.updatedStatus = status   // 更新されたステータスを記録する
	mock.updatedHistory = history // 更新された history を記録する
	return nil
}

// TestApplyService_Apply_k8sapply失敗時にapplyHistorystatusがfailedになる は k8s apply 失敗時の挙動を確認する
func TestApplyService_Apply_k8sapply失敗時にapplyHistorystatusがfailedになる(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-6") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-6") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	// k8s apply を失敗させるために Get/Create が失敗する fake クライアントを使う
	// fake クライアントに Create 前に既存の Deployment を仕込み、Update をエラーにする方法は難しいため
	// ここでは実際の DB + モックリポジトリを組み合わせる

	applyHistoryMock := &applyHistoryMockRepository{}             // apply_history モックを生成する
	deploymentMock := &applyHistoryMockDeploymentRepository{      // deployment モックを生成する
		deploymentData: deploymentData,
	}
	projectMock := &applyHistoryMockProjectRepository{            // project モックを生成する
		projectData: projectData,
	}

	// テスト用 k8s Deployment を fake に追加して Update を失敗させるオブジェクトを用意する
	existingDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentData.Name,
			Namespace: projectData.Namespace,
		},
	}
	fakeK8sClient := fake.NewSimpleClientset(existingDeployment) // 既存の Deployment を持つ fake クライアントを生成する

	// k8s apply が失敗する Reactor を追加する
	fakeK8sClient.Fake.PrependReactor("update", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("k8s update failed: simulated error") // エラーを返す
	})

	applyService := NewApplyService(db, fakeK8sClient, deploymentMock, applyHistoryMock, projectMock) // サービスを生成する

	_, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する（失敗が期待される）
	if err == nil { // エラーが返ることを確認する
		t.Fatal("k8s apply 失敗時にエラーが返るべきですが nil が返りました")
	}

	// apply_history の status が failed に更新されていることを確認する
	if applyHistoryMock.updatedStatus != models.ApplyStatusFailed { // status が failed であることを確認する
		t.Errorf("期待する apply_history status: failed, 実際の status: %s", applyHistoryMock.updatedStatus)
	}
}

// TestApplyService_Apply_k8sapply失敗時にpendingフィールドがそのまま残る は k8s apply 失敗時に pending フィールドが変更されないことを確認する
func TestApplyService_Apply_k8sapply失敗時にpendingフィールドがそのまま残る(t *testing.T) {
	db := setupApplyTestDB(t)                                    // テスト用 DB を準備する
	ensureInstanceSize(t, db)                                    // InstanceSize を準備する
	projectData := createApplyTestProject(t, db, "test-ns-apply-7") // テスト用 Project を作成する
	deploymentData := createApplyTestDeployment(t, db, projectData.ID, "test-deploy-apply-7") // テスト用 Deployment を作成する
	t.Cleanup(func() {
		db.Unscoped().Where("deployment_id = ?", deploymentData.ID).Delete(&models.ApplyHistory{}) // apply_history を削除する
	})

	// k8s apply を失敗させる Reactor 付き fake クライアントを生成する
	fakeK8sClient := fake.NewSimpleClientset()                   // fake k8s クライアントを生成する
	fakeK8sClient.Fake.PrependReactor("create", "deployments", func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("k8s create failed: simulated error") // Create をエラーにする
	})

	projectMock := &applyHistoryMockProjectRepository{            // project モックを生成する
		projectData: projectData,
	}

	deploymentRepo := repository.NewDeploymentRepository(db)     // リポジトリを生成する
	applyHistoryRepo := repository.NewApplyHistoryRepository(db) // apply_history リポジトリを生成する
	applyService := NewApplyService(db, fakeK8sClient, deploymentRepo, applyHistoryRepo, projectMock) // サービスを生成する

	_, err := applyService.Apply(context.Background(), deploymentData.ID) // apply を実行する（失敗が期待される）
	if err == nil { // エラーが返ることを確認する
		t.Fatal("k8s apply 失敗時にエラーが返るべきですが nil が返りました")
	}

	// DB から取得して pending フィールドが変更されていないことを確認する
	var fetchedDeployment models.Deployment
	db.First(&fetchedDeployment, "id = ?", deploymentData.ID) // apply 後のレコードを取得する
	if fetchedDeployment.PendingImageURL != "nginx:latest" { // pending_image_url がそのままであることを確認する
		t.Errorf("k8s apply 失敗時に pending_image_url が変更されています: %s", fetchedDeployment.PendingImageURL)
	}
	if fetchedDeployment.Status != models.DeploymentStatusPending { // status が変更されていないことを確認する
		t.Errorf("k8s apply 失敗時に status が変更されています: %s", fetchedDeployment.Status)
	}
}

// 未使用変数のコンパイルエラーを防ぐためのダミー変数
var _ = datatypes.JSON(`{}`)
var _ = time.Now
