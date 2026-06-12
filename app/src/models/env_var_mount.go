package models

import "time"

type EnvVarMountStatus string

const (
	EnvVarMountStatusPending  EnvVarMountStatus = "pending"  // mount 済み・未 apply
	EnvVarMountStatusApplied  EnvVarMountStatus = "applied"  // apply 済み
	EnvVarMountStatusDeleting EnvVarMountStatus = "deleting"
)

// EnvVarMount は env_vars と deployments の中間テーブル
// UNIQUE 制約: (env_var_id, deployment_id)
// apply 前チェック: 実効キー (COALESCE(override_key, key)) の重複をエラーで弾く
type EnvVarMount struct {
	ID           string            `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EnvVarID     string            `gorm:"type:uuid;not null;index"`
	DeploymentID string            `gorm:"type:uuid;not null;index"`

	// NULL = 元の key をそのまま使用。指定時はコンテナ側でこの名前でマウント
	OverrideKey        string `gorm:"type:varchar(255)"`
	PendingOverrideKey string `gorm:"type:varchar(255)"`

	Status    EnvVarMountStatus `gorm:"type:varchar(32);not null;default:'pending'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (EnvVarMount) TableName() string { return "env_var_mounts" }
