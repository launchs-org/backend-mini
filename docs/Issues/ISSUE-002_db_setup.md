# ISSUE-002 DB 接続・AutoMigrate セットアップ

## 親 Issue
ISSUE-001

## 概要
PostgreSQL への接続と全モデルの AutoMigrate を実装する。

## 実装手順

### 1. 依存パッケージ追加

```bash
go get gorm.io/gorm
go get gorm.io/driver/postgres
go get gorm.io/datatypes
go get github.com/lib/pq
go get github.com/google/uuid
```

### 2. `config/config.go` を作成

```go
package config

import (
    "os"
)

type Config struct {
    DatabaseDSN  string
    RegistryHost string
    ServerPort   string
}

func Load() *Config {
    return &Config{
        DatabaseDSN:  getEnv("DATABASE_DSN", "host=localhost user=postgres password=postgres dbname=launchs port=5432 sslmode=disable"),
        RegistryHost: getEnv("REGISTRY_HOST", "registry.launchs.org"),
        ServerPort:   getEnv("SERVER_PORT", "8080"),
    }
}

func getEnv(key, fallback string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return fallback
}
```

### 3. 全モデルを作成

`models/` 以下に各ファイルを作成する。
モデルの詳細は設計書 `02_tables.md` を参照。

作成するファイル:
- `instance_size.go`
- `account.go`
- `account_quota.go`
- `project.go`
- `deployment.go`
- `deployment_build.go`
- `apply_history.go`
- `service.go`
- `ingress_route.go`
- `env_var.go`
- `env_var_mount.go`
- `volume.go`
- `volume_mount.go`
- `webhook.go`

各モデルの共通事項:
- `ID` は `uuid` 型、`default:gen_random_uuid()`
- `CreatedAt` / `UpdatedAt` は `time.Time`（GORM が自動管理）

### 4. `repository/db.go` を作成

```go
package repository

import (
    "app/models"
    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func New(dsn string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        return nil, err
    }
    return db, nil
}

func AutoMigrate(db *gorm.DB) error {
    return db.AutoMigrate(
        &models.InstanceSize{},
        &models.Account{},
        &models.AccountQuota{},
        &models.Project{},
        &models.Deployment{},
        &models.DeploymentBuild{},
        &models.ApplyHistory{},
        &models.Service{},
        &models.IngressRoute{},
        &models.EnvVar{},
        &models.EnvVarMount{},
        &models.Volume{},
        &models.VolumeMount{},
        &models.DeploymentWebhook{},
    )
}

func SeedInstanceSizes(db *gorm.DB) error {
    sizes := []models.InstanceSize{
        {Size: "micro",  CPURequest: "50m",   CPULimit: "200m",  MemoryRequest: "64Mi",   MemoryLimit: "128Mi"},
        {Size: "small",  CPURequest: "100m",  CPULimit: "500m",  MemoryRequest: "128Mi",  MemoryLimit: "256Mi"},
        {Size: "medium", CPURequest: "250m",  CPULimit: "1000m", MemoryRequest: "256Mi",  MemoryLimit: "512Mi"},
        {Size: "large",  CPURequest: "500m",  CPULimit: "2000m", MemoryRequest: "512Mi",  MemoryLimit: "1024Mi"},
        {Size: "xlarge", CPURequest: "1000m", CPULimit: "4000m", MemoryRequest: "1024Mi", MemoryLimit: "2048Mi"},
    }
    return db.FirstOrCreate(&sizes).Error
}
```

### 5. `main.go` に DB 初期化を追加

```go
cfg := config.Load()

database, err := db.New(cfg.DatabaseDSN)
if err != nil {
    log.Fatalf("failed to connect db: %v", err)
}

if err := db.AutoMigrate(database); err != nil {
    log.Fatalf("failed to migrate: %v", err)
}

if err := db.SeedInstanceSizes(database); err != nil {
    log.Fatalf("failed to seed: %v", err)
}
```

## テスト確認項目

- [ ] `go run .` でエラーなく起動すること
- [ ] PostgreSQL に接続できること
- [ ] 全テーブルが作成されること（`\dt` で確認）
- [ ] `instance_sizes` に5件のマスターデータが入っていること
- [ ] 2回起動しても AutoMigrate がエラーにならないこと（冪等性）
