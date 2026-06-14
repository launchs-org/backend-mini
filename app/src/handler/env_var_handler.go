package handler

import (
	"app/middlewares"
	"app/service"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// EnvVarHandler は環境変数 CRUD の HTTP ハンドラーを提供する
type EnvVarHandler struct {
	envVarService service.EnvVarService // env_var サービスのインターフェース
}

// NewEnvVarHandler は EnvVarHandler を生成して返す
func NewEnvVarHandler(envVarService service.EnvVarService) *EnvVarHandler {
	return &EnvVarHandler{
		envVarService: envVarService, // 依存を注入する
	}
}

// maskedValue は is_secret=true のときに返すマスク文字列
const maskedValue = "***"

// ListEnvVars は GET /api/v1/projects/:id/env-vars のハンドラー
func (envVarHandler *EnvVarHandler) ListEnvVars(echoCtx echo.Context) error {
	userClaim := echoCtx.Get("claim").(*middlewares.AccessTokenClaim) // JWT クレームを取得する
	projectID := echoCtx.Param("id")                                  // パスパラメータから project ID を取得する

	envVarList, err := envVarHandler.envVarService.ListEnvVars(echoCtx.Request().Context(), userClaim.UserID, projectID) // サービスを呼び出して一覧を取得する
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

	responseList := make([]map[string]interface{}, 0, len(envVarList)) // レスポンス一覧を生成する
	for _, envVar := range envVarList {
		value := envVar.Value // 値を取得する
		if envVar.IsSecret {
			value = maskedValue // シークレットの場合はマスクする
		}
		responseList = append(responseList, map[string]interface{}{
			"id":         envVar.ID,        // ID を設定する
			"project_id": envVar.ProjectID, // プロジェクト ID を設定する
			"key":        envVar.Key,       // キーを設定する
			"value":      value,            // マスク済みの値を設定する
			"is_secret":  envVar.IsSecret,  // シークレットフラグを設定する
		})
	}
	return echoCtx.JSON(http.StatusOK, responseList) // env_var 一覧を返す
}

// CreateEnvVar は POST /api/v1/projects/:id/env-vars のハンドラー
func (envVarHandler *EnvVarHandler) CreateEnvVar(echoCtx echo.Context) error {
	userClaim := echoCtx.Get("claim").(*middlewares.AccessTokenClaim) // JWT クレームを取得する
	projectID := echoCtx.Param("id")                                  // パスパラメータから project ID を取得する

	var requestBody service.CreateEnvVarRequest               // リクエストボディの構造体を定義する
	if err := echoCtx.Bind(&requestBody); err != nil {        // リクエストをバインドする
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストが不正です",
		})
	}
	if requestBody.Key == "" { // 必須フィールドのバリデーションを行う
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "key は必須です",
		})
	}

	envVarData, err := envVarHandler.envVarService.CreateEnvVar(echoCtx.Request().Context(), userClaim.UserID, projectID, requestBody) // サービスを呼び出して env_var を作成する
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

	value := envVarData.Value // 値を取得する
	if envVarData.IsSecret {
		value = maskedValue // シークレットの場合はマスクする
	}
	return echoCtx.JSON(http.StatusCreated, map[string]interface{}{
		"id":         envVarData.ID,        // ID を設定する
		"project_id": envVarData.ProjectID, // プロジェクト ID を設定する
		"key":        envVarData.Key,       // キーを設定する
		"value":      value,                // マスク済みの値を設定する
		"is_secret":  envVarData.IsSecret,  // シークレットフラグを設定する
	})
}

// UpdateEnvVar は PUT /api/v1/env-vars/:id のハンドラー
func (envVarHandler *EnvVarHandler) UpdateEnvVar(echoCtx echo.Context) error {
	userClaim := echoCtx.Get("claim").(*middlewares.AccessTokenClaim) // JWT クレームを取得する
	envVarID := echoCtx.Param("id")                                   // パスパラメータから env_var ID を取得する

	var requestBody service.UpdateEnvVarRequest               // リクエストボディの構造体を定義する
	if err := echoCtx.Bind(&requestBody); err != nil {        // リクエストをバインドする
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストが不正です",
		})
	}

	envVarData, err := envVarHandler.envVarService.UpdateEnvVar(echoCtx.Request().Context(), userClaim.UserID, envVarID, requestBody) // サービスを呼び出して env_var を更新する
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

	value := envVarData.Value // 値を取得する
	if envVarData.IsSecret {
		value = maskedValue // シークレットの場合はマスクする
	}
	return echoCtx.JSON(http.StatusOK, map[string]interface{}{
		"id":         envVarData.ID,        // ID を設定する
		"project_id": envVarData.ProjectID, // プロジェクト ID を設定する
		"key":        envVarData.Key,       // キーを設定する
		"value":      value,                // マスク済みの値を設定する
		"is_secret":  envVarData.IsSecret,  // シークレットフラグを設定する
	})
}

// DeleteEnvVar は DELETE /api/v1/env-vars/:id のハンドラー
func (envVarHandler *EnvVarHandler) DeleteEnvVar(echoCtx echo.Context) error {
	userClaim := echoCtx.Get("claim").(*middlewares.AccessTokenClaim) // JWT クレームを取得する
	envVarID := echoCtx.Param("id")                                   // パスパラメータから env_var ID を取得する

	err := envVarHandler.envVarService.DeleteEnvVar(echoCtx.Request().Context(), userClaim.UserID, envVarID) // サービスを呼び出して env_var を削除する
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
