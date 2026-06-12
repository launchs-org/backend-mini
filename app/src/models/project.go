package models

import (
	"time"

	"gorm.io/datatypes"
)

type ProjectStatus string

const (
	ProjectStatusProvisioning ProjectStatus = "provisioning"
	ProjectStatusActive       ProjectStatus = "active"
	ProjectStatusDeleting     ProjectStatus = "deleting"
)

// ステータス遷移: provisioning → active / active → deleting
type Project struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    string         `gorm:"type:varchar(255);not null;index"`
	Name      string         `gorm:"type:varchar(63);not null;uniqueIndex"`
	Namespace string         `gorm:"type:varchar(63);not null;uniqueIndex"`
	Status    ProjectStatus  `gorm:"type:varchar(32);not null;default:'provisioning'"`
	K8sStatus datatypes.JSON `gorm:"type:jsonb"` // null = 未同期
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Project) TableName() string { return "projects" }
