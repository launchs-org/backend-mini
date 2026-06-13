package repository

import (
	"app/models"
	"context"

	"gorm.io/gorm"
)

// DeploymentRepository は deployments テーブルへのアクセスを定義するインターフェース
type DeploymentRepository interface {
	Create(ctx context.Context, deployment *models.Deployment) error                               // deployment を作成する
	FindByID(ctx context.Context, deploymentID string) (*models.Deployment, error)                 // deployment を ID で取得する
	FindAllByProjectID(ctx context.Context, projectID string) ([]models.Deployment, error)         // projectID に紐づく deployment 一覧を取得する
	Save(ctx context.Context, deployment *models.Deployment) error                                 // deployment を保存する
}

// deploymentRepositoryImpl は DeploymentRepository の GORM 実装
type deploymentRepositoryImpl struct {
	db *gorm.DB // データベース接続
}

// NewDeploymentRepository は DeploymentRepository の実装を返す
func NewDeploymentRepository(db *gorm.DB) DeploymentRepository {
	return &deploymentRepositoryImpl{db: db} // 実装を生成して返す
}

// Create は deployment レコードを作成する
func (repo *deploymentRepositoryImpl) Create(ctx context.Context, deployment *models.Deployment) error {
	return repo.db.WithContext(ctx).Create(deployment).Error // db を使って作成する
}

// FindByID は deploymentID に対応する deployment を返す
func (repo *deploymentRepositoryImpl) FindByID(ctx context.Context, deploymentID string) (*models.Deployment, error) {
	var deploymentData models.Deployment                                                          // deployment を格納する変数を定義する
	if err := repo.db.WithContext(ctx).First(&deploymentData, "id = ?", deploymentID).Error; err != nil { // db から deployment を取得する
		return nil, err // 取得エラーを返す
	}
	return &deploymentData, nil // deployment を返す
}

// FindAllByProjectID は projectID に紐づく deployment 一覧を返す
func (repo *deploymentRepositoryImpl) FindAllByProjectID(ctx context.Context, projectID string) ([]models.Deployment, error) {
	var deploymentList []models.Deployment                                                                              // deployment 一覧を格納するスライスを定義する
	if err := repo.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&deploymentList).Error; err != nil { // db から deployment 一覧を取得する
		return nil, err // 取得エラーを返す
	}
	return deploymentList, nil // deployment 一覧を返す
}

// Save は deployment レコードを保存する
func (repo *deploymentRepositoryImpl) Save(ctx context.Context, deployment *models.Deployment) error {
	return repo.db.WithContext(ctx).Save(deployment).Error // db を使って保存する
}

// ServiceRepository は services テーブルへのアクセスを定義するインターフェース
type ServiceRepository interface {
	Create(ctx context.Context, service *models.Service) error // service を作成する
}

// serviceRepositoryImpl は ServiceRepository の GORM 実装
type serviceRepositoryImpl struct {
	db *gorm.DB // データベース接続
}

// NewServiceRepository は ServiceRepository の実装を返す
func NewServiceRepository(db *gorm.DB) ServiceRepository {
	return &serviceRepositoryImpl{db: db} // 実装を生成して返す
}

// Create は service レコードを作成する
func (repo *serviceRepositoryImpl) Create(ctx context.Context, service *models.Service) error {
	return repo.db.WithContext(ctx).Create(service).Error // db を使って作成する
}
