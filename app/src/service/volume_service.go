package service

import (
	"app/models"
	"app/repository"
	"context"

	"gorm.io/gorm"
)

// VolumeService はボリューム CRUD のビジネスロジックを定義するインターフェース
type VolumeService interface {
	ListVolumes(ctx context.Context, userID string, projectID string) ([]*models.Volume, error)                          // ボリューム一覧を取得する
	CreateVolume(ctx context.Context, userID string, projectID string, req CreateVolumeRequest) (*models.Volume, error) // ボリュームを作成する
	DeleteVolume(ctx context.Context, userID string, volumeID string) error                                              // ボリュームを削除する
}

// CreateVolumeRequest は POST /projects/:id/volumes のリクエスト構造体
type CreateVolumeRequest struct {
	Name   string `json:"name"`    // ボリューム名
	SizeMB int    `json:"size_mb"` // ボリュームサイズ（MB）
}

// volumeServiceImpl は VolumeService の実装
type volumeServiceImpl struct {
	db          *gorm.DB                     // データベース接続（トランザクション開始に使用する）
	volumeRepo  repository.VolumeRepository  // volume リポジトリ
	projectRepo repository.ProjectRepository // project リポジトリ（認可チェックに使用する）
}

// NewVolumeService は VolumeService の実装を返す
func NewVolumeService(
	db *gorm.DB,
	volumeRepo repository.VolumeRepository,
	projectRepo repository.ProjectRepository,
) VolumeService {
	return &volumeServiceImpl{
		db:          db,          // DB 接続を注入する
		volumeRepo:  volumeRepo,  // volume リポジトリを注入する
		projectRepo: projectRepo, // project リポジトリを注入する
	}
}

// checkProjectOwner は projectID に対応する project の UserID と userID を比較し、不一致の場合は ErrForbidden を返す
func (svc *volumeServiceImpl) checkProjectOwner(ctx context.Context, userID string, projectID string) error {
	projectData, err := svc.projectRepo.FindByIDNoTx(ctx, projectID) // project を取得する
	if err != nil {
		return err // 取得エラーを返す
	}
	if projectData.UserID != userID { // ユーザー ID が一致しない場合は forbidden を返す
		return ErrForbidden // アクセス拒否エラーを返す
	}
	return nil // 認可チェック成功を返す
}

// ListVolumes は projectID に紐づく volume 一覧を返す
func (svc *volumeServiceImpl) ListVolumes(ctx context.Context, userID string, projectID string) ([]*models.Volume, error) {
	if err := svc.checkProjectOwner(ctx, userID, projectID); err != nil { // 認可チェックを行う
		return nil, err // エラーを返す
	}
	return svc.volumeRepo.FindAllByProjectID(ctx, projectID) // リポジトリ経由で一覧を取得する
}

// CreateVolume は volume を作成する
func (svc *volumeServiceImpl) CreateVolume(ctx context.Context, userID string, projectID string, req CreateVolumeRequest) (*models.Volume, error) {
	if err := svc.checkProjectOwner(ctx, userID, projectID); err != nil { // 認可チェックを行う
		return nil, err // エラーを返す
	}

	var createdVolume *models.Volume                                             // 結果格納用変数を宣言する
	err := svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {        // トランザクションを開始する
		volumeData := &models.Volume{
			ProjectID: projectID, // プロジェクト ID を設定する
			Name:      req.Name,  // ボリューム名を設定する
			SizeMB:    req.SizeMB, // サイズを設定する
			Status:    models.VolumeStatusPending, // ステータスを pending に設定する
		}
		if err := svc.volumeRepo.Create(ctx, tx, volumeData); err != nil { // tx を渡してリポジトリに委譲する
			return err // エラーでロールバックする
		}
		createdVolume = volumeData // 作成した volume を格納する
		return nil                 // コミットする
	})
	return createdVolume, err // 結果を返す
}

// DeleteVolume は volumeID に対応する volume を削除する
func (svc *volumeServiceImpl) DeleteVolume(ctx context.Context, userID string, volumeID string) error {
	volumeData, err := svc.volumeRepo.FindByID(ctx, volumeID) // volume を取得する
	if err != nil {
		return err // 取得エラーを返す
	}

	if err := svc.checkProjectOwner(ctx, userID, volumeData.ProjectID); err != nil { // 認可チェックを行う
		return err // エラーを返す
	}

	return svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { // トランザクションを開始する
		return svc.volumeRepo.Delete(ctx, tx, volumeData) // tx を渡してリポジトリに委譲する
	})
}
