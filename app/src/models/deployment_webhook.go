package models

import "time"

type DeploymentWebhook struct {
	ID           string `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	DeploymentID string `gorm:"type:uuid;not null;index"`
	Secret       string `gorm:"type:varchar(255);not null"` // HMAC 検証用シークレット（GitHub Webhook の Secret に設定する値）
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (DeploymentWebhook) TableName() string { return "deployment_webhooks" }
