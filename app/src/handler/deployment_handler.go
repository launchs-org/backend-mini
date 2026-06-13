package handler

import (
	"app/service"
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// DeploymentHandler は Deployment CRUD の HTTP ハンドラーを提供する
type DeploymentHandler struct {
	deploymentService service.DeploymentService // deployment サービスのインターフェース
}

// NewDeploymentHandler は DeploymentHandler を生成して返す
func NewDeploymentHandler(deploymentService service.DeploymentService) *DeploymentHandler {
	return &DeploymentHandler{
		deploymentService: deploymentService, // 依存を注入する
	}
}

// ListDeployments は GET /projects/:id/deployments のハンドラー
func (deploymentHandler *DeploymentHandler) ListDeployments(echoCtx echo.Context) error {
	projectID := echoCtx.Param("id") // パスパラメータから project ID を取得する

	deploymentList, err := deploymentHandler.deploymentService.ListDeployments(echoCtx.Request().Context(), projectID) // サービスを呼び出して一覧を取得する
	if err != nil {                                                                                                      // エラーが発生した場合
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusOK, deploymentList) // deployment 一覧を返す
}

// CreateDeployment は POST /projects/:id/deployments のハンドラー
func (deploymentHandler *DeploymentHandler) CreateDeployment(echoCtx echo.Context) error {
	projectID := echoCtx.Param("id") // パスパラメータから project ID を取得する

	var requestBody service.CreateDeploymentRequest             // リクエストボディの構造体を定義する
	if err := echoCtx.Bind(&requestBody); err != nil {         // リクエストをバインドする
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストが不正です",
		})
	}
	requestBody.ProjectID = projectID // パスパラメータの project ID をセットする

	deploymentData, err := deploymentHandler.deploymentService.CreateDeployment(echoCtx.Request().Context(), requestBody) // サービスを呼び出して deployment を作成する
	if err != nil {                                                                                                         // エラーが発生した場合
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusCreated, deploymentData) // 作成した deployment を返す
}

// GetDeployment は GET /deployments/:id のハンドラー
func (deploymentHandler *DeploymentHandler) GetDeployment(echoCtx echo.Context) error {
	deploymentID := echoCtx.Param("id") // パスパラメータから deployment ID を取得する

	deploymentData, err := deploymentHandler.deploymentService.GetDeployment(echoCtx.Request().Context(), deploymentID) // サービスを呼び出して deployment を取得する
	if err != nil {                                                                                                       // エラーが発生した場合
		if errors.Is(err, gorm.ErrRecordNotFound) { // レコードが存在しない場合は 404 を返す
			return echoCtx.JSON(http.StatusNotFound, map[string]string{
				"error": "リソースが見つかりません",
			})
		}
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusOK, deploymentData) // deployment を返す
}

// UpdateDeployment は PUT /deployments/:id のハンドラー
func (deploymentHandler *DeploymentHandler) UpdateDeployment(echoCtx echo.Context) error {
	deploymentID := echoCtx.Param("id") // パスパラメータから deployment ID を取得する

	var requestBody service.UpdateDeploymentRequest             // リクエストボディの構造体を定義する
	if err := echoCtx.Bind(&requestBody); err != nil {         // リクエストをバインドする
		return echoCtx.JSON(http.StatusBadRequest, map[string]string{
			"error": "リクエストが不正です",
		})
	}

	deploymentData, err := deploymentHandler.deploymentService.UpdateDeployment(echoCtx.Request().Context(), deploymentID, requestBody) // サービスを呼び出して deployment を更新する
	if err != nil {                                                                                                                       // エラーが発生した場合
		if errors.Is(err, gorm.ErrRecordNotFound) { // レコードが存在しない場合は 404 を返す
			return echoCtx.JSON(http.StatusNotFound, map[string]string{
				"error": "リソースが見つかりません",
			})
		}
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusOK, deploymentData) // 更新後の deployment を返す
}

// DeleteDeployment は DELETE /deployments/:id のハンドラー
func (deploymentHandler *DeploymentHandler) DeleteDeployment(echoCtx echo.Context) error {
	deploymentID := echoCtx.Param("id") // パスパラメータから deployment ID を取得する

	deploymentData, err := deploymentHandler.deploymentService.DeleteDeployment(echoCtx.Request().Context(), deploymentID) // サービスを呼び出して deployment を削除する
	if err != nil {                                                                                                          // エラーが発生した場合
		if errors.Is(err, gorm.ErrRecordNotFound) { // レコードが存在しない場合は 404 を返す
			return echoCtx.JSON(http.StatusNotFound, map[string]string{
				"error": "リソースが見つかりません",
			})
		}
		return echoCtx.JSON(http.StatusInternalServerError, map[string]string{
			"error": "内部サーバーエラー",
		})
	}
	return echoCtx.JSON(http.StatusOK, deploymentData) // 更新後の deployment を返す
}
