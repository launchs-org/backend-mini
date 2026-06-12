package models

import (
	"time"

	"gorm.io/datatypes"
)

type VolumeStatus string

const (
	VolumeStatusPending  VolumeStatus = "pending"  // 作成済み・未 apply
	VolumeStatusBound    VolumeStatus = "bound"
	VolumeStatusDeleting VolumeStatus = "deleting"
)

type Volume struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID string         `gorm:"type:uuid;not null;index"`
	Name      string         `gorm:"type:varchar(63);not null"`
	SizeMB    int            `gorm:"not null"` // 作成後変更不可。PVC ReclaimPolicy = Delete
	Status    VolumeStatus   `gorm:"type:varchar(32);not null;default:'pending'"`
	K8sStatus datatypes.JSON `gorm:"type:jsonb"` // null = 未同期
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Volume) TableName() string { return "volumes" }
