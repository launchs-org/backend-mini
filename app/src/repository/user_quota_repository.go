package repository

import (
	"app/models"
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserQuotaRepository は user_quotas テーブルへのアクセスを定義するインターフェース
type UserQuotaRepository interface {
	GetOrCreate(ctx context.Context, userID string) (*models.UserQuota, error)       // quota を取得し存在しなければ作成する
	Update(ctx context.Context, userID string, updates map[string]interface{}) (*models.UserQuota, error) // quota を部分更新する
	CountProjects(ctx context.Context, userID string) (int, error)                   // ユーザーのプロジェクト数を集計する
	CountDeployments(ctx context.Context, userID string) (int, error)                // ユーザーのデプロイメント数を集計する
	SumVolumeMB(ctx context.Context, userID string) (int, error)                     // ユーザーの合計ボリューム容量を集計する
}

// userQuotaRepositoryImpl は UserQuotaRepository の GORM 実装
type userQuotaRepositoryImpl struct {
	db *gorm.DB // データベース接続
}

// NewUserQuotaRepository は UserQuotaRepository の実装を返す
func NewUserQuotaRepository(db *gorm.DB) UserQuotaRepository {
	return &userQuotaRepositoryImpl{db: db} // 実装を生成して返す
}

// GetOrCreate は userID に対応する quota を返す。存在しない場合はデフォルト値で作成する
func (repo *userQuotaRepositoryImpl) GetOrCreate(ctx context.Context, userID string) (*models.UserQuota, error) {
	quotaData := &models.UserQuota{
		UserID: userID, // upsert のキーとして userID を設定する
	}
	result := repo.db.WithContext(ctx).
		Where(models.UserQuota{UserID: userID}).
		Attrs(models.UserQuota{ // レコードが存在しない場合のみデフォルト値を適用する
			MaxProjects:              5,
			MaxDeployments:           20,
			MaxReplicasPerDeployment: 5,
			MaxVolumeMB:              10240,
		}).
		FirstOrCreate(quotaData) // 存在すれば取得、なければ作成する
	if result.Error != nil {
		return nil, result.Error // DB エラーを返す
	}
	return quotaData, nil // quota データを返す
}

// Update は userID の quota を部分更新して更新後のレコードを返す
func (repo *userQuotaRepositoryImpl) Update(ctx context.Context, userID string, updates map[string]interface{}) (*models.UserQuota, error) {
	result := repo.db.WithContext(ctx).
		Model(&models.UserQuota{}).
		Clauses(clause.Returning{}).
		Where("user_id = ?", userID).
		Updates(updates) // 指定フィールドのみ更新する
	if result.Error != nil {
		return nil, result.Error // DB エラーを返す
	}

	// 更新後のレコードを取得する
	quotaData, err := repo.GetOrCreate(ctx, userID)
	if err != nil {
		return nil, err // 取得エラーを返す
	}
	return quotaData, nil // 更新後の quota データを返す
}

// CountProjects はユーザーが所有するプロジェクト数を返す
func (repo *userQuotaRepositoryImpl) CountProjects(ctx context.Context, userID string) (int, error) {
	var projectCount int64
	result := repo.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("user_id = ? AND status != ?", userID, models.ProjectStatusDeleting).
		Count(&projectCount) // 削除中以外のプロジェクト数を集計する
	if result.Error != nil {
		return 0, result.Error // DB エラーを返す
	}
	return int(projectCount), nil // プロジェクト数を返す
}

// CountDeployments はユーザーが所有するデプロイメント数を返す
func (repo *userQuotaRepositoryImpl) CountDeployments(ctx context.Context, userID string) (int, error) {
	var deploymentCount int64
	result := repo.db.WithContext(ctx).
		Model(&models.Deployment{}).
		Joins("JOIN projects ON projects.id = deployments.project_id").
		Where("projects.user_id = ? AND deployments.status != ?", userID, models.DeploymentStatusDeleting).
		Count(&deploymentCount) // 削除中以外のデプロイメント数を集計する
	if result.Error != nil {
		return 0, result.Error // DB エラーを返す
	}
	return int(deploymentCount), nil // デプロイメント数を返す
}

// SumVolumeMB はユーザーが使用中の合計ボリューム容量（MB）を返す
func (repo *userQuotaRepositoryImpl) SumVolumeMB(ctx context.Context, userID string) (int, error) {
	var totalMB *int64
	result := repo.db.WithContext(ctx).
		Model(&models.Volume{}).
		Joins("JOIN projects ON projects.id = volumes.project_id").
		Where("projects.user_id = ?", userID).
		Select("COALESCE(SUM(volumes.size_mb), 0)").
		Scan(&totalMB) // ボリューム容量の合計を集計する
	if result.Error != nil {
		return 0, result.Error // DB エラーを返す
	}
	if totalMB == nil {
		return 0, nil // ボリュームが存在しない場合は 0 を返す
	}
	return int(*totalMB), nil // 合計容量を返す
}
