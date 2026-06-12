package models

import "time"

type EnvVar struct {
	ID        string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID string `gorm:"type:uuid;not null;index"`
	Key       string `gorm:"type:varchar(255);not null"`
	Value     string `gorm:"type:text"`    // 即時更新。k8s 反映は apply 時
	IsSecret  bool   `gorm:"not null;default:false"` // true → k8s Secret / UI マスク
	Status    string `gorm:"type:varchar(32);not null;default:'active'"` // active / deleting
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (EnvVar) TableName() string { return "env_vars" }
