# ISSUE-028 Volume Mount CRUD エンドポイント

## 親 Issue
ISSUE-026

## 実装手順

### 1. `handler/volume_mount.go` を作成

```go
func (h *Handler) CreateVolumeMount(c echo.Context) error {
    deploymentID := c.Param("id")
    var req struct {
        VolumeID  string `json:"volume_id"`
        MountPath string `json:"mount_path"`
    }
    if err := c.Bind(&req); err != nil { return echo.ErrBadRequest }

    var count int64
    h.DB.Model(&models.VolumeMount{}).
        Where("volume_id = ? AND deployment_id = ?", req.VolumeID, deploymentID).
        Count(&count)
    if count > 0 {
        return c.JSON(http.StatusConflict, map[string]string{
            "error": "volume already mounted", "code": "ALREADY_MOUNTED",
        })
    }

    mount := models.VolumeMount{
        VolumeID:     req.VolumeID,
        DeploymentID: deploymentID,
        MountPath:    req.MountPath,
        Status:       models.VolumeMountStatusPending,
    }
    h.DB.Create(&mount)
    return c.JSON(http.StatusCreated, mount)
}

func (h *Handler) UpdateVolumeMount(c echo.Context) error {
    var mount models.VolumeMount
    if err := h.DB.First(&mount, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    var req struct {
        MountPath *string `json:"mount_path"`
    }
    if err := c.Bind(&req); err != nil { return echo.ErrBadRequest }
    if req.MountPath != nil {
        mount.PendingMountPath = *req.MountPath
        h.DB.Save(&mount)
    }
    return c.JSON(http.StatusOK, mount)
}
```

### 2. ルーティング登録

```go
api.GET("/deployments/:id/volume-mounts",  h.ListVolumeMounts)
api.POST("/deployments/:id/volume-mounts", h.CreateVolumeMount)
api.PUT("/volume-mounts/:id",              h.UpdateVolumeMount)
api.DELETE("/volume-mounts/:id",           h.DeleteVolumeMount)
```

## テスト確認項目

- [ ] 同じ volume を同じ deployment に2回 mount すると 409 になること
- [ ] `PUT /volume-mounts/:id` で `mount_path` を送ると `pending_mount_path` に入ること

### repository 層テスト

- [ ] `VolumeMountRepository.Create` でレコードが DB に保存されること
- [ ] 同一 deployment + volume_id の組み合わせで UNIQUE 制約エラーが返ること
- [ ] `VolumeMountRepository.Save` で `pending_mount_path` が更新されること
