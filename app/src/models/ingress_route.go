package models

import (
	"time"

	"gorm.io/datatypes"
)

type IngressRouteStatus string

const (
	IngressRouteStatusPending  IngressRouteStatus = "pending"
	IngressRouteStatusActive   IngressRouteStatus = "active"
	IngressRouteStatusDeleting IngressRouteStatus = "deleting"
)

type IngressRoute struct {
	ID        string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ServiceID string             `gorm:"type:uuid;not null;uniqueIndex"` // Service と 1:1

	// 作成後変更不可。自動生成: {deployment_name}-{uuid8}.launchs.org
	Host       string `gorm:"type:varchar(253);not null;uniqueIndex"`
	PathPrefix string `gorm:"type:varchar(255);not null;default:'/'"`

	Port int `gorm:"not null"` // 転送先の TCP ポート（Service の ports のうちいずれか）

	Status    IngressRouteStatus `gorm:"type:varchar(32);not null;default:'pending'"`
	K8sStatus datatypes.JSON     `gorm:"type:jsonb"` // null = 未同期
	CreatedAt time.Time
}

func (IngressRoute) TableName() string { return "ingress_routes" }
