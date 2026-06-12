# ISSUE-027 Volume CRUD エンドポイント

## 親 Issue
ISSUE-026

## 概要
Volume の作成・取得・削除を実装する。作成時に quota チェックを行う。size_mb は作成後変更不可。

## 実装手順

### 1. `service/quota.go` を作成（volume quota チェック）

```go
package service

func CheckVolumeQuota(db *gorm.DB, accountID string, newSizeMB int) error {
    var quota models.AccountQuota
    db.Where("account_id = ?", accountID).First(&quota)

    var currentMB int64
    db.Model(&models.Volume{}).
        Joins("JOIN projects ON projects.id = volumes.project_id").
        Where("projects.account_id = ? AND volumes.status != ?", accountID, "deleted").
        Select("COALESCE(SUM(size_mb), 0)").Scan(&currentMB)

    if int(currentMB)+newSizeMB > quota.MaxVolumeMB {
        return fmt.Errorf("volume quota exceeded: current=%dMB, adding=%dMB, max=%dMB",
            currentMB, newSizeMB, quota.MaxVolumeMB)
    }
    return nil
}
```

### 2. `handler/volume.go` を作成

```go
func (h *Handler) CreateVolume(c echo.Context) error {
    projectID := c.Param("id")
    var req struct {
        Name   string `json:"name"`
        SizeMB int    `json:"size_mb"`
    }
    if err := c.Bind(&req); err != nil { return echo.ErrBadRequest }
    if req.SizeMB <= 0 {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "size_mb must be greater than 0", "code": "INVALID_SIZE",
        })
    }

    // project から account_id を取得して quota チェック
    var project models.Project
    h.DB.First(&project, "id = ?", projectID)
    if err := service.CheckVolumeQuota(h.DB, project.AccountID, req.SizeMB); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(), "code": "QUOTA_EXCEEDED",
        })
    }

    vol := models.Volume{
        ProjectID: projectID,
        Name:      req.Name,
        SizeMB:    req.SizeMB,
        Status:    models.VolumeStatusPending,
    }
    h.DB.Create(&vol)
    return c.JSON(http.StatusCreated, vol)
}

func (h *Handler) DeleteVolume(c echo.Context) error {
    var vol models.Volume
    if err := h.DB.First(&vol, "id = ?", c.Param("id")).Error; err != nil {
        return echo.ErrNotFound
    }
    // mount されている場合は削除不可
    var count int64
    h.DB.Model(&models.VolumeMount{}).Where("volume_id = ?", vol.ID).Count(&count)
    if count > 0 {
        return c.JSON(http.StatusConflict, map[string]string{
            "error": "volume is mounted", "code": "VOLUME_MOUNTED",
        })
    }
    h.DB.Model(&vol).Update("status", models.VolumeStatusDeleting)
    return c.NoContent(http.StatusNoContent)
}
```

### 3. ルーティング登録

```go
api.GET("/projects/:id/volumes",  h.ListVolumes)
api.POST("/projects/:id/volumes", h.CreateVolume)
api.DELETE("/volumes/:id",        h.DeleteVolume)
```

## テスト確認項目

- [ ] quota を超える volume 作成で 400 になること
- [ ] size_mb=0 で 400 になること
- [ ] mount されている volume を DELETE すると 409 になること
- [ ] volume 作成で `status = pending` になること
