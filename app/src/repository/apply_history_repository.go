package repository

import (
	"app/models"
	"context"

	"gorm.io/gorm"
)

// ApplyHistoryRepository は apply_history テーブルへのアクセスを定義するインターフェース
type ApplyHistoryRepository interface {
	Create(ctx context.Context, tx *gorm.DB, history *models.ApplyHistory) error                                    // apply_history を作成する
	UpdateStatus(ctx context.Context, tx *gorm.DB, history *models.ApplyHistory, status models.ApplyStatus) error  // apply_history のステータスを更新する
}

// applyHistoryRepositoryImpl は ApplyHistoryRepository の GORM 実装
type applyHistoryRepositoryImpl struct {
	db *gorm.DB // データベース接続
}

// NewApplyHistoryRepository は ApplyHistoryRepository の実装を返す
func NewApplyHistoryRepository(db *gorm.DB) ApplyHistoryRepository {
	return &applyHistoryRepositoryImpl{db: db} // 実装を生成して返す
}

// Create は apply_history レコードを作成する
func (repo *applyHistoryRepositoryImpl) Create(ctx context.Context, tx *gorm.DB, history *models.ApplyHistory) error {
	return tx.WithContext(ctx).Create(history).Error // tx を使って作成する
}

// UpdateStatus は apply_history のステータスとエラーメッセージを更新する
func (repo *applyHistoryRepositoryImpl) UpdateStatus(ctx context.Context, tx *gorm.DB, history *models.ApplyHistory, status models.ApplyStatus) error {
	return tx.WithContext(ctx).Model(history).Updates(map[string]interface{}{ // tx を使ってステータスを更新する
		"status":        status,               // ステータスを更新する
		"error_message": history.ErrorMessage, // エラーメッセージを更新する
	}).Error
}
