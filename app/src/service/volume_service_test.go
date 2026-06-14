package service

import (
	"app/models"
	"context"
	"testing"

	"gorm.io/gorm"
)

// mockVolumeRepository は VolumeRepository のテスト用モック実装
type mockVolumeRepository struct {
	createFunc             func(ctx context.Context, tx *gorm.DB, volume *models.Volume) error
	findByIDFunc           func(ctx context.Context, volumeID string) (*models.Volume, error)
	findAllByProjectIDFunc func(ctx context.Context, projectID string) ([]*models.Volume, error)
	deleteFunc             func(ctx context.Context, tx *gorm.DB, volume *models.Volume) error
}

func (mock *mockVolumeRepository) Create(ctx context.Context, tx *gorm.DB, volume *models.Volume) error {
	return mock.createFunc(ctx, tx, volume) // モック関数を呼び出す
}

func (mock *mockVolumeRepository) FindByID(ctx context.Context, volumeID string) (*models.Volume, error) {
	if mock.findByIDFunc != nil { // モック関数が設定されている場合は呼び出す
		return mock.findByIDFunc(ctx, volumeID)
	}
	return &models.Volume{ID: volumeID, ProjectID: "project-id-1"}, nil // デフォルトは volume を返す
}

func (mock *mockVolumeRepository) FindAllByProjectID(ctx context.Context, projectID string) ([]*models.Volume, error) {
	if mock.findAllByProjectIDFunc != nil { // モック関数が設定されている場合は呼び出す
		return mock.findAllByProjectIDFunc(ctx, projectID)
	}
	return []*models.Volume{}, nil // デフォルトは空一覧を返す
}

func (mock *mockVolumeRepository) Delete(ctx context.Context, tx *gorm.DB, volume *models.Volume) error {
	if mock.deleteFunc != nil { // モック関数が設定されている場合は呼び出す
		return mock.deleteFunc(ctx, tx, volume)
	}
	return nil // デフォルトは nil を返す
}

// TestListVolumes_正常に一覧が取得される は ListVolumes が volume 一覧を返すことを確認する
func TestListVolumes_正常に一覧が取得される_service(t *testing.T) {
	expectedList := []*models.Volume{
		{ID: "volume-id-1", ProjectID: "project-id-1", Name: "vol-a", SizeMB: 512},  // volume 1件目
		{ID: "volume-id-2", ProjectID: "project-id-1", Name: "vol-b", SizeMB: 1024}, // volume 2件目
	}

	volumeRepo := &mockVolumeRepository{
		findAllByProjectIDFunc: func(ctx context.Context, projectID string) ([]*models.Volume, error) {
			return expectedList, nil // 一覧を返す
		},
	}
	projectRepo := &mockProjectRepository{
		findByIDNoTxFunc: func(ctx context.Context, projectID string) (*models.Project, error) {
			return &models.Project{ID: projectID, UserID: "test-user-id"}, nil // 所有者として返す
		},
	}

	svc := NewVolumeService(nil, volumeRepo, projectRepo) // サービスを生成する（db は未使用）

	result, err := svc.ListVolumes(context.Background(), "test-user-id", "project-id-1") // 一覧を取得する
	if err != nil {
		t.Fatalf("ListVolumes がエラーを返しました: %v", err)
	}
	if len(result) != 2 { // 2 件返ることを確認する
		t.Errorf("期待する件数: 2, 実際の件数: %d", len(result))
	}
}

// TestListVolumes_他ユーザーはErrForbiddenを返す は所有者でないユーザーが ErrForbidden を受け取ることを確認する
func TestListVolumes_他ユーザーはErrForbiddenを返す_service(t *testing.T) {
	volumeRepo := &mockVolumeRepository{}
	projectRepo := &mockProjectRepository{
		findByIDNoTxFunc: func(ctx context.Context, projectID string) (*models.Project, error) {
			return &models.Project{ID: projectID, UserID: "owner-user-id"}, nil // 別ユーザーが所有するプロジェクトを返す
		},
	}

	svc := NewVolumeService(nil, volumeRepo, projectRepo) // サービスを生成する

	_, err := svc.ListVolumes(context.Background(), "other-user-id", "project-id-1") // 他ユーザーとして一覧を取得する
	if err != ErrForbidden {                                                           // ErrForbidden であることを確認する
		t.Errorf("期待するエラー: ErrForbidden, 実際のエラー: %v", err)
	}
}

// TestCreateVolume_正常にvolumeが作成される は CreateVolume が volume を作成することを確認する
func TestCreateVolume_正常にvolumeが作成される_service(t *testing.T) {
	var capturedVolume *models.Volume // キャプチャした volume を格納する変数を定義する

	volumeRepo := &mockVolumeRepository{
		createFunc: func(ctx context.Context, tx *gorm.DB, volume *models.Volume) error {
			capturedVolume = volume       // volume をキャプチャする
			volume.ID = "new-volume-id"  // ID を付与する
			return nil
		},
	}
	projectRepo := &mockProjectRepository{
		findByIDNoTxFunc: func(ctx context.Context, projectID string) (*models.Project, error) {
			return &models.Project{ID: projectID, UserID: "test-user-id"}, nil // 所有者として返す
		},
	}

	db := setupApplyTestDB(t)                             // テスト用 DB を準備する
	svc := NewVolumeService(db, volumeRepo, projectRepo)  // サービスを生成する

	req := CreateVolumeRequest{Name: "my-volume", SizeMB: 512} // リクエストを定義する
	result, err := svc.CreateVolume(context.Background(), "test-user-id", "project-id-1", req) // volume を作成する
	if err != nil {
		t.Fatalf("CreateVolume がエラーを返しました: %v", err)
	}
	if result.ID != "new-volume-id" { // ID が付与されていることを確認する
		t.Errorf("期待する ID: new-volume-id, 実際の ID: %s", result.ID)
	}
	if capturedVolume.Name != "my-volume" { // 名前が正しく設定されていることを確認する
		t.Errorf("期待する Name: my-volume, 実際の Name: %s", capturedVolume.Name)
	}
	if capturedVolume.SizeMB != 512 { // サイズが正しく設定されていることを確認する
		t.Errorf("期待する SizeMB: 512, 実際の SizeMB: %d", capturedVolume.SizeMB)
	}
	if capturedVolume.ProjectID != "project-id-1" { // ProjectID が正しく設定されていることを確認する
		t.Errorf("期待する ProjectID: project-id-1, 実際の ProjectID: %s", capturedVolume.ProjectID)
	}
	if capturedVolume.Status != models.VolumeStatusPending { // ステータスが pending であることを確認する
		t.Errorf("期待する Status: pending, 実際の Status: %s", capturedVolume.Status)
	}
}

// TestCreateVolume_他ユーザーはErrForbiddenを返す は所有者でないユーザーが ErrForbidden を受け取ることを確認する
func TestCreateVolume_他ユーザーはErrForbiddenを返す_service(t *testing.T) {
	volumeRepo := &mockVolumeRepository{}
	projectRepo := &mockProjectRepository{
		findByIDNoTxFunc: func(ctx context.Context, projectID string) (*models.Project, error) {
			return &models.Project{ID: projectID, UserID: "owner-user-id"}, nil // 別ユーザーが所有するプロジェクトを返す
		},
	}

	svc := NewVolumeService(nil, volumeRepo, projectRepo) // サービスを生成する

	req := CreateVolumeRequest{Name: "vol", SizeMB: 512} // リクエストを定義する
	_, err := svc.CreateVolume(context.Background(), "other-user-id", "project-id-1", req) // 他ユーザーとして作成する
	if err != ErrForbidden {                                                                 // ErrForbidden であることを確認する
		t.Errorf("期待するエラー: ErrForbidden, 実際のエラー: %v", err)
	}
}

// TestDeleteVolume_他ユーザーはErrForbiddenを返す は所有者でないユーザーが ErrForbidden を受け取ることを確認する
func TestDeleteVolume_他ユーザーはErrForbiddenを返す_service(t *testing.T) {
	volumeRepo := &mockVolumeRepository{
		findByIDFunc: func(ctx context.Context, volumeID string) (*models.Volume, error) {
			return &models.Volume{ID: volumeID, ProjectID: "project-id-1"}, nil // volume を返す
		},
	}
	projectRepo := &mockProjectRepository{
		findByIDNoTxFunc: func(ctx context.Context, projectID string) (*models.Project, error) {
			return &models.Project{ID: projectID, UserID: "owner-user-id"}, nil // 別ユーザーが所有するプロジェクトを返す
		},
	}

	svc := NewVolumeService(nil, volumeRepo, projectRepo) // サービスを生成する

	err := svc.DeleteVolume(context.Background(), "other-user-id", "volume-id-1") // 他ユーザーとして削除する
	if err != ErrForbidden {                                                        // ErrForbidden であることを確認する
		t.Errorf("期待するエラー: ErrForbidden, 実際のエラー: %v", err)
	}
}
