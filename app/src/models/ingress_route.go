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
	ID           string             `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`            // UUID 主キー
	DeploymentID string             `gorm:"type:uuid;not null;uniqueIndex"                  json:"deployment_id"` // Deployment と 1:1

	Host       string `gorm:"type:varchar(253);not null" json:"host"`        // ホスト名（自動生成: {deployment_name}-{uuid8}.launchs.org）
	PathPrefix string `gorm:"type:varchar(255);not null" json:"path_prefix"` // パスプレフィックス

	TLSEnabled          bool   `gorm:"not null;default:false"   json:"tls_enabled"`          // TLS を有効にするかどうか
	CertificateResolver string `gorm:"type:varchar(255)"        json:"certificate_resolver"` // 証明書リゾルバー名

	Port int `gorm:"not null" json:"port"` // 転送先の TCP ポート

	// pending_* フィールド: apply 実行前の未適用設定値
	PendingHost                string `gorm:"type:varchar(253)"  json:"pending_host"`                 // 未 apply のホスト名
	PendingPathPrefix          string `gorm:"type:varchar(255)"  json:"pending_path_prefix"`          // 未 apply のパスプレフィックス
	PendingPort                int    `gorm:"type:int"           json:"pending_port"`                 // 未 apply のポート番号
	PendingTLSEnabled          *bool  `gorm:"type:boolean"       json:"pending_tls_enabled"`          // 未 apply の TLS 設定
	PendingCertificateResolver string `gorm:"type:varchar(255)"  json:"pending_certificate_resolver"` // 未 apply の証明書リゾルバー

	Status    IngressRouteStatus `gorm:"type:varchar(32);not null;default:'pending'" json:"status"`     // ステータス
	K8sStatus datatypes.JSON     `gorm:"type:jsonb"                                   json:"k8s_status"` // null = 未同期
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

func (IngressRoute) TableName() string { return "ingress_routes" }
