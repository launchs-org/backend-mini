package models

import "time"

// UserQuota はユーザーごとのリソース上限を管理する
// 認証は別サービスが担当し、user_id（UUID 文字列）のみがこのサービスに渡る
// レコードが存在しない場合は初回アクセス時に upsert で作成する
type UserQuota struct {
	ID                       string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID                   string    `gorm:"type:varchar(255);not null;uniqueIndex"`
	MaxProjects              int       `gorm:"not null;default:5"`
	MaxDeployments           int       `gorm:"not null;default:20"`
	MaxReplicasPerDeployment int       `gorm:"not null;default:5"`
	MaxVolumeMB              int       `gorm:"not null;default:10240"`
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

func (UserQuota) TableName() string { return "user_quotas" }
