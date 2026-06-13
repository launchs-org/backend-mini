package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type DeploymentType string

const (
	DeploymentTypeImageURL   DeploymentType = "image_url"
	DeploymentTypeDockerfile DeploymentType = "dockerfile"
	DeploymentTypeRailpack   DeploymentType = "railpack"
)

type DeploymentStatus string

const (
	DeploymentStatusPending  DeploymentStatus = "pending" // 初回作成・未 apply
	DeploymentStatusRunning  DeploymentStatus = "running"
	DeploymentStatusFailed   DeploymentStatus = "failed"
	DeploymentStatusDeleting DeploymentStatus = "deleting"
)

type AppStatus string

const (
	AppStatusPending   AppStatus = "pending"
	AppStatusBuilding  AppStatus = "building"
	AppStatusDeploying AppStatus = "deploying"
	AppStatusRunning   AppStatus = "running"
	AppStatusError     AppStatus = "error"
)

type Deployment struct {
	ID        string         `gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ProjectID string         `gorm:"type:uuid;not null;index"`
	Name      string         `gorm:"type:varchar(63);not null"`
	Type      DeploymentType `gorm:"type:varchar(32);not null"` // 作成後変更不可

	// --- image_url 専用 ---
	ImageURL        string `gorm:"type:text"`
	PendingImageURL string `gorm:"type:text"`

	// --- dockerfile / railpack 共通（GitHub）---
	GithubRepoURL        string `gorm:"type:text"`
	PendingGithubRepoURL string `gorm:"type:text"`

	GithubBranch        string `gorm:"type:varchar(255)"`
	PendingGithubBranch string `gorm:"type:varchar(255)"`

	// "HEAD" 指定可。apply 時に最新 SHA を取得して上書き
	GithubCommitSHA        string `gorm:"type:varchar(40)"`
	PendingGithubCommitSHA string `gorm:"type:varchar(40)"`

	// ビルド作業ディレクトリ。このディレクトリに CD した状態でビルドを開始する
	GithubRepoDirectory        string `gorm:"type:varchar(255);default:'./'"`
	PendingGithubRepoDirectory string `gorm:"type:varchar(255)"`

	// --- dockerfile 専用 ---
	DockerfilePath        string `gorm:"type:varchar(255);default:'./Dockerfile'"`
	PendingDockerfilePath string `gorm:"type:varchar(255)"`

	// --- ビルド管理 ---
	// nil = ビルドなし。完了時に build_id をセット
	CurrentBuildID *string          `gorm:"type:uuid"`
	CurrentBuild   *DeploymentBuild `gorm:"foreignKey:CurrentBuildID"`

	// --- デプロイ設定 ---
	InstanceSize        string `gorm:"type:varchar(16);not null;default:'small'"`
	PendingInstanceSize string `gorm:"type:varchar(16)"`

	Replicas        int32 `gorm:"not null;default:1"`
	PendingReplicas int32

	// --- 起動設定 ---
	Command        pq.StringArray `gorm:"type:text[]"` // k8s command（ENTRYPOINT 上書き）
	PendingCommand pq.StringArray `gorm:"type:text[]"`

	Args        pq.StringArray `gorm:"type:text[]"` // k8s args（CMD 上書き）
	PendingArgs pq.StringArray `gorm:"type:text[]"`

	// --- ステータス ---
	Status    DeploymentStatus `gorm:"type:varchar(32);not null;default:'pending'"`
	AppStatus AppStatus        `gorm:"type:varchar(32);not null;default:'pending'"`
	K8sStatus datatypes.JSON   `gorm:"type:jsonb"` // null = 未同期
	AppliedAt *time.Time

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Deployment) TableName() string { return "deployments" }
