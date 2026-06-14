package handler

import (
	"app/service"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// VolumeHandler はボリューム CRUD の HTTP ハンドラーを提供する
type VolumeHandler struct {
	volumeService service.VolumeService // volume サービスのインターフェース
}

// NewVolumeHandler は VolumeHandler を生成して返す
func NewVolumeHandler(volumeService service.VolumeService) *VolumeHandler {
	return &VolumeHandler{
		volumeService: volumeService, // volume サービスを注入する
	}
}

// ListVolumes は GET /api/v1/projects/:id/volumes のハンドラー
func (volumeHandler *VolumeHandler) ListVolumes(echoCtx echo.Context) error {
	userID := echoCtx.Get("UserID").(string) // ミドルウェアがセットした UserID を取得する
	projectID := echoCtx.Param("id")         // パスパラメータから project ID を取得する

	volumeList, err := volumeHandler.volumeService.ListVolumes(echoCtx.Request().Context(), userID, projectID) // サービスを呼び出して一覧を取得する
	if err != nil {
		if errors.Is(err, service.ErrForbidden) { // 権限エラーの場合は 403 を返す
			return echoCtx.JSON(http.StatusForbidden, map[string]string{
				"error": "アクセスが拒否されました",
			})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) { // レコードが存在しない場合は 404 を返す
			return echoCtx.JSON(http.StatusNotFound, map[string]string{
				"error": "リソースが見つかりません",
			})
		}
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusOK, volumeList) // volume 一覧を返す
}

// CreateVolume は POST /api/v1/projects/:id/volumes のハンドラー
func (volumeHandler *VolumeHandler) CreateVolume(echoCtx echo.Context) error {
	userID := echoCtx.Get("UserID").(string) // ミドルウェアがセットした UserID を取得する
	projectID := echoCtx.Param("id")         // パスパラメータから project ID を取得する

	var requestBody service.CreateVolumeRequest               // リクエストボディの構造体を定義する
	if err := echoCtx.Bind(&requestBody); err != nil {        // リクエストをバインドする
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストが不正です",
		})
	}
	if requestBody.Name == "" { // 必須フィールドのバリデーションを行う
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "name は必須です",
		})
	}
	if requestBody.SizeMB <= 0 { // size_mb のバリデーションを行う
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "size_mb は 1 以上の値が必要です",
		})
	}

	volumeData, err := volumeHandler.volumeService.CreateVolume(echoCtx.Request().Context(), userID, projectID, requestBody) // サービスを呼び出して volume を作成する
	if err != nil {
		if errors.Is(err, service.ErrForbidden) { // 権限エラーの場合は 403 を返す
			return echoCtx.JSON(http.StatusForbidden, map[string]string{
				"error": "アクセスが拒否されました",
			})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) { // レコードが存在しない場合は 404 を返す
			return echoCtx.JSON(http.StatusNotFound, map[string]string{
				"error": "リソースが見つかりません",
			})
		}
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusCreated, volumeData) // 作成結果を返す
}

// DeleteVolume は DELETE /api/v1/volumes/:id のハンドラー
func (volumeHandler *VolumeHandler) DeleteVolume(echoCtx echo.Context) error {
	userID := echoCtx.Get("UserID").(string) // ミドルウェアがセットした UserID を取得する
	volumeID := echoCtx.Param("id")          // パスパラメータから volume ID を取得する

	err := volumeHandler.volumeService.DeleteVolume(echoCtx.Request().Context(), userID, volumeID) // サービスを呼び出して volume を削除する
	if err != nil {
		if errors.Is(err, service.ErrForbidden) { // 権限エラーの場合は 403 を返す
			return echoCtx.JSON(http.StatusForbidden, map[string]string{
				"error": "アクセスが拒否されました",
			})
		}
		if errors.Is(err, gorm.ErrRecordNotFound) { // レコードが存在しない場合は 404 を返す
			return echoCtx.JSON(http.StatusNotFound, map[string]string{
				"error": "リソースが見つかりません",
			})
		}
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.NoContent(http.StatusNoContent) // 削除成功時は 204 を返す
}
