# ISSUE-010 Deployment CRUD

## 親 Issue
ISSUE-009

## 概要
Deployment の POST / GET / PUT / DELETE を実装する。
POST・PUT ではリクエストのフィールドを `pending_***` カラムに書き込む。

## 実装手順

### 1. `handler/deployment.go` を作成

#### POST /projects/:id/deployments

```go
func (h *Handler) CreateDeployment(c echo.Context) error {
    projectID := c.Param("id")

    var req struct {
        Name           string `json:"name"`
        Type           string `json:"type"` // image_url / dockerfile / railpack
        // image_url 専用
        ImageURL       string `json:"image_url"`
        // GitHub 共通
        GithubRepoURL        string `json:"github_repo_url"`
        GithubBranch         string `json:"github_branch"`
        GithubCommitSHA      string `json:"github_commit_sha"`
        GithubRepoDirectory  string `json:"build_directory"`
        // dockerfile 専用
        DockerfilePath string `json:"dockerfile_path"`
        // デプロイ設定
        InstanceSize   string `json:"instance_size"`
        Replicas       int32  `json:"replicas"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    // デフォルト値
    if req.InstanceSize == "" { req.InstanceSize = "small" }
    if req.Replicas == 0 { req.Replicas = 1 }
    if req.DockerfilePath == "" { req.DockerfilePath = "./Dockerfile" }
    if req.GithubRepoDirectory == "" { req.GithubRepoDirectory = "./" }

    d := models.Deployment{
        ProjectID: projectID,
        Name:      req.Name,
        Type:      models.DeploymentType(req.Type),
        Status:    models.DeploymentStatusPending,
        AppStatus: models.AppStatusPending,
        // 全て pending_*** に入れる
        PendingImageURL:             req.ImageURL,
        PendingGithubRepoURL:        req.GithubRepoURL,
        PendingGithubBranch:         req.GithubBranch,
        PendingGithubCommitSHA:      req.GithubCommitSHA,
        PendingGithubRepoDirectory:  req.GithubRepoDirectory,
        PendingDockerfilePath:       req.DockerfilePath,
        PendingInstanceSize:         req.InstanceSize,
        PendingReplicas:             req.Replicas,
    }

    if err := h.DB.Create(&d).Error; err != nil {
        return echo.ErrInternalServerError
    }

    // Service レコードも同時に作成（ports は空）
    svc := models.Service{
        DeploymentID: d.ID,
        Status:       models.ServiceStatusPending,
    }
    h.DB.Create(&svc)

    return c.JSON(http.StatusCreated, d)
}
```

#### PUT /deployments/:id

```go
func (h *Handler) UpdateDeployment(c echo.Context) error {
    var d models.Deployment
    if err := h.DB.First(&d, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }

    var req struct {
        ImageURL              *string `json:"image_url"`
        GithubRepoURL         *string `json:"github_repo_url"`
        GithubBranch          *string `json:"github_branch"`
        GithubCommitSHA       *string `json:"github_commit_sha"`
        GithubRepoDirectory   *string `json:"build_directory"`
        DockerfilePath        *string `json:"dockerfile_path"`
        InstanceSize          *string `json:"instance_size"`
        Replicas              *int32  `json:"replicas"`
        Command               []string `json:"command"`
        Args                  []string `json:"args"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    // 送られてきたフィールドのみ pending_*** に書き込む
    if req.ImageURL != nil             { d.PendingImageURL = *req.ImageURL }
    if req.GithubRepoURL != nil        { d.PendingGithubRepoURL = *req.GithubRepoURL }
    if req.GithubBranch != nil         { d.PendingGithubBranch = *req.GithubBranch }
    if req.GithubCommitSHA != nil      { d.PendingGithubCommitSHA = *req.GithubCommitSHA }
    if req.GithubRepoDirectory != nil  { d.PendingGithubRepoDirectory = *req.GithubRepoDirectory }
    if req.DockerfilePath != nil       { d.PendingDockerfilePath = *req.DockerfilePath }
    if req.InstanceSize != nil         { d.PendingInstanceSize = *req.InstanceSize }
    if req.Replicas != nil             { d.PendingReplicas = *req.Replicas }
    if req.Command != nil              { d.PendingCommand = req.Command }
    if req.Args != nil                 { d.PendingArgs = req.Args }

    h.DB.Save(&d)
    return c.JSON(http.StatusOK, d)
}
```

### 2. ルーティング登録

```go
api.GET("/projects/:id/deployments",  h.ListDeployments)
api.POST("/projects/:id/deployments", h.CreateDeployment)
api.GET("/deployments/:id",           h.GetDeployment)
api.PUT("/deployments/:id",           h.UpdateDeployment)
api.DELETE("/deployments/:id",        h.DeleteDeployment)
```

## テスト確認項目

- [ ] `POST /deployments` で `status = pending`、`app_status = pending` で作成されること
- [ ] `POST /deployments` でリクエストの値が全て `pending_***` に入ること
- [ ] `POST /deployments` で Service レコードも同時に作成されること
- [ ] `PUT /deployments` で送ったフィールドのみ `pending_***` が更新されること
- [ ] `PUT /deployments` で送らなかったフィールドは変化しないこと
- [ ] `type` カラムは PUT で変更できないこと
- [ ] `DELETE /deployments/:id` で `status = deleting` になること

### repository 層テスト

- [ ] `DeploymentRepository.Create` でレコードが DB に保存されること
- [ ] `DeploymentRepository.FindByID` で存在しない ID を渡すと `ErrRecordNotFound` が返ること
- [ ] `DeploymentRepository.Save` で `pending_***` フィールドが正しく更新されること
- [ ] `DeploymentRepository.Delete` でレコードが DB から削除されること
