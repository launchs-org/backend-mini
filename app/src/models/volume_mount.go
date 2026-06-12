package models

import "time"

type VolumeMountStatus string

const (
	VolumeMountStatusPending  VolumeMountStatus = "pending"
	VolumeMountStatusMounted  VolumeMountStatus = "mounted"
	VolumeMountStatusDeleting VolumeMountStatus = "deleting"
)

// VolumeMount は volumes と deployments の中間テーブル
// UNIQUE 制約: (volume_id, deployment_id)
type VolumeMount struct {
	ID           string            `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	VolumeID     string            `gorm:"type:uuid;not null;index"`
	DeploymentID string            `gorm:"type:uuid;not null;index"`
	MountPath        string        `gorm:"type:varchar(255);not null"`
	PendingMountPath string        `gorm:"type:varchar(255)"`
	Status    VolumeMountStatus    `gorm:"type:varchar(32);not null;default:'pending'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (VolumeMount) TableName() string { return "volume_mounts" }
