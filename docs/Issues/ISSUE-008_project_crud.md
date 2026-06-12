# ISSUE-008 Project CRUD

## 親 Issue
ISSUE-005

## 概要
Project の作成・取得・更新・削除を実装する。作成時に k8s namespace も作成する。

## 実装手順

### 1. `service/project.go` を作成

```go
package service

import (
    "context"
    "fmt"
    "app/k8s"
    "app/models"
    "gorm.io/gorm"
    k8sclient "k8s.io/client-go/kubernetes"
)

type ProjectService struct {
    DB  *gorm.DB
    K8s *k8sclient.Clientset
}

func (s *ProjectService) Create(ctx context.Context, accountID, name string) (*models.Project, error) {
    // namespace 名 = project 名（lowercase）
    namespace := name

    project := models.Project{
        AccountID: accountID,
        Name:      name,
        Namespace: namespace,
        Status:    models.ProjectStatusProvisioning,
    }

    if err := s.DB.Create(&project).Error; err != nil {
        return nil, fmt.Errorf("db insert: %w", err)
    }

    // k8s namespace 作成
    if err := k8s.CreateNamespace(ctx, s.K8s, namespace); err != nil {
        // namespace 作成失敗でも DB レコードは残す（Watcher or リトライで対応）
        return &project, fmt.Errorf("k8s namespace: %w", err)
    }

    // namespace 作成成功 → active に更新
    s.DB.Model(&project).Update("status", models.ProjectStatusActive)
    project.Status = models.ProjectStatusActive

    return &project, nil
}

func (s *ProjectService) Delete(ctx context.Context, id string) error {
    var project models.Project
    if err := s.DB.First(&project, "id = ?", id).Error; err != nil {
        return err
    }

    // deleting に更新（実際の削除は Phase10 で完成）
    return s.DB.Model(&project).Update("status", models.ProjectStatusDeleting).Error
}
```

### 2. `handler/project.go` を作成

```go
package handler

import (
    "net/http"
    "github.com/labstack/echo/v4"
    "app/models"
    "app/service"
)

func (h *Handler) ListProjects(c echo.Context) error {
    var projects []models.Project
    h.DB.Find(&projects)
    return c.JSON(http.StatusOK, projects)
}

func (h *Handler) CreateProject(c echo.Context) error {
    var req struct {
        AccountID string `json:"account_id" validate:"required"`
        Name      string `json:"name"       validate:"required"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    svc := &service.ProjectService{DB: h.DB, K8s: h.K8s}
    project, err := svc.Create(c.Request().Context(), req.AccountID, req.Name)
    if err != nil {
        return echo.ErrInternalServerError
    }
    return c.JSON(http.StatusCreated, project)
}

func (h *Handler) GetProject(c echo.Context) error {
    var project models.Project
    if err := h.DB.First(&project, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    return c.JSON(http.StatusOK, project)
}

func (h *Handler) UpdateProject(c echo.Context) error {
    var project models.Project
    if err := h.DB.First(&project, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }

    var req struct {
        Name *string `json:"name"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }
    if req.Name != nil {
        project.Name = *req.Name
    }

    h.DB.Save(&project)
    return c.JSON(http.StatusOK, project)
}

func (h *Handler) DeleteProject(c echo.Context) error {
    svc := &service.ProjectService{DB: h.DB, K8s: h.K8s}
    if err := svc.Delete(c.Request().Context(), c.Param("id")); err != nil {
        return echo.ErrNotFound
    }
    return c.NoContent(http.StatusAccepted)
}
```

### 3. ルーティング登録

```go
api.GET("/projects",     h.ListProjects)
api.POST("/projects",    h.CreateProject)
api.GET("/projects/:id", h.GetProject)
api.PUT("/projects/:id", h.UpdateProject)
api.DELETE("/projects/:id", h.DeleteProject)
```

## テスト確認項目

- [ ] `POST /projects` で project が作成されること
- [ ] `POST /projects` 後に k8s namespace が作成されること
- [ ] `POST /projects` 後に `status = active` になること
- [ ] 同名の project を作成すると 500（UNIQUE 制約違反）になること
- [ ] `GET /projects/:id` で詳細が取得できること
- [ ] `PUT /projects/:id` で name が更新できること
- [ ] `DELETE /projects/:id` で `status = deleting` になること
