package models

import (
	"time"

	"gorm.io/datatypes"
)

type ServiceStatus string

const (
	ServiceStatusPending  ServiceStatus = "pending"
	ServiceStatusActive   ServiceStatus = "active"
	ServiceStatusDeleting ServiceStatus = "deleting"
)

// ServicePort は ports カラム（jsonb）の要素型
// 例: [{"protocol": "TCP", "port": 8080}, {"protocol": "UDP", "port": 9090}]
type ServicePort struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

type Service struct {
	ID           string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DeploymentID string         `gorm:"type:uuid;not null;uniqueIndex"`
	Ports        datatypes.JSON `gorm:"type:jsonb"` // 現在稼働中
	PendingPorts datatypes.JSON `gorm:"type:jsonb"` // 未 apply
	Status       ServiceStatus  `gorm:"type:varchar(32);not null;default:'pending'"`
	K8sStatus    datatypes.JSON `gorm:"type:jsonb"` // null = 未同期
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Service) TableName() string { return "services" }
