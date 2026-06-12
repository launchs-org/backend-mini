package models

import (
	"time"

	"gorm.io/datatypes"
)

type ApplyStatus string

const (
	ApplyStatusApplied ApplyStatus = "applied"
	ApplyStatusFailed  ApplyStatus = "failed"
)

type ApplyHistory struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DeploymentID string         `gorm:"type:uuid;not null;index"`
	Manifests    datatypes.JSON `gorm:"type:jsonb;not null"` // 生成した k8s manifest 全スナップショット
	Status       ApplyStatus    `gorm:"type:varchar(32);not null"`
	ErrorMessage string         `gorm:"type:text"`
	AppliedAt    time.Time      // POST /apply が叩かれた時刻
}

func (ApplyHistory) TableName() string { return "apply_history" }
