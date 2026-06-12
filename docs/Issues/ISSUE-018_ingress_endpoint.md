# ISSUE-018 POST /ingress エンドポイント・ホスト名自動生成

## 親 Issue
ISSUE-015

## 概要
IngressRoute を手動で作成するエンドポイントを実装する。ホスト名を自動生成する。

## 実装手順

### 1. ホスト名生成ロジック

```go
// {deployment_name}-{uuid8}.launchs.org
func generateHost(deploymentName string) string {
    uid := uuid.New().String()[:8]
    return fmt.Sprintf("%s-%s.launchs.org", deploymentName, uid)
}
```

### 2. `handler/ingress.go` を作成

```go
func (h *Handler) CreateIngress(c echo.Context) error {
    deploymentID := c.Param("id")

    // deployment 取得
    var d models.Deployment
    if err := h.DB.First(&d, "id = ?", deploymentID).Error; err != nil {
        return echo.ErrNotFound
    }

    // service 取得
    var svc models.Service
    if err := h.DB.Where("deployment_id = ?", deploymentID).First(&svc).Error; err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "service not found", "code": "SERVICE_NOT_FOUND",
        })
    }

    // 既に IngressRoute が存在する場合は 409
    var existing models.IngressRoute
    if err := h.DB.Where("service_id = ?", svc.ID).First(&existing).Error; err == nil {
        return c.JSON(http.StatusConflict, map[string]string{
            "error": "ingress already exists", "code": "INGRESS_EXISTS",
        })
    }

    var req struct {
        Port int `json:"port"`
    }
    if err := c.Bind(&req); err != nil {
        return echo.ErrBadRequest
    }

    // port が service.ports の TCP ポートに存在するか確認
    // （バリデーションロジックは省略）

    ingress := models.IngressRoute{
        ServiceID:  svc.ID,
        Host:       generateHost(d.Name),
        PathPrefix: "/",
        Port:       req.Port,
        Status:     models.IngressRouteStatusPending,
    }
    h.DB.Create(&ingress)
    return c.JSON(http.StatusCreated, ingress)
}

func (h *Handler) GetIngress(c echo.Context) error {
    deploymentID := c.Param("id")
    var svc models.Service
    h.DB.Where("deployment_id = ?", deploymentID).First(&svc)
    var ingress models.IngressRoute
    if err := h.DB.Where("service_id = ?", svc.ID).First(&ingress).Error; err != nil {
        return echo.ErrNotFound
    }
    return c.JSON(http.StatusOK, ingress)
}
```

### 3. ルーティング登録

```go
api.GET("/deployments/:id/ingress",    h.GetIngress)
api.POST("/deployments/:id/ingress",   h.CreateIngress)
api.DELETE("/deployments/:id/ingress", h.DeleteIngress)
```

## テスト確認項目

- [ ] `POST /ingress` でホスト名が `{name}-{uuid8}.launchs.org` 形式で生成されること
- [ ] 同じ deployment に2回 POST すると 409 になること
- [ ] service が存在しない場合に 400 になること
- [ ] UNIQUE 制約でホスト名の重複が防がれること
