package handler

import (
	"app/models"
	"app/service"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// mockDeploymentService は DeploymentService のテスト用モック実装
type mockDeploymentService struct {
	listDeploymentsFunc  func(ctx context.Context, projectID string) ([]models.Deployment, error)
	createDeploymentFunc func(ctx context.Context, req service.CreateDeploymentRequest) (*models.Deployment, error)
	getDeploymentFunc    func(ctx context.Context, deploymentID string) (*models.Deployment, error)
	updateDeploymentFunc func(ctx context.Context, deploymentID string, req service.UpdateDeploymentRequest) (*models.Deployment, error)
	deleteDeploymentFunc func(ctx context.Context, deploymentID string) (*models.Deployment, error)
}

func (mock *mockDeploymentService) ListDeployments(ctx context.Context, projectID string) ([]models.Deployment, error) {
	return mock.listDeploymentsFunc(ctx, projectID) // モック関数を呼び出す
}

func (mock *mockDeploymentService) CreateDeployment(ctx context.Context, req service.CreateDeploymentRequest) (*models.Deployment, error) {
	return mock.createDeploymentFunc(ctx, req) // モック関数を呼び出す
}

func (mock *mockDeploymentService) GetDeployment(ctx context.Context, deploymentID string) (*models.Deployment, error) {
	return mock.getDeploymentFunc(ctx, deploymentID) // モック関数を呼び出す
}

func (mock *mockDeploymentService) UpdateDeployment(ctx context.Context, deploymentID string, req service.UpdateDeploymentRequest) (*models.Deployment, error) {
	return mock.updateDeploymentFunc(ctx, deploymentID, req) // モック関数を呼び出す
}

func (mock *mockDeploymentService) DeleteDeployment(ctx context.Context, deploymentID string) (*models.Deployment, error) {
	return mock.deleteDeploymentFunc(ctx, deploymentID) // モック関数を呼び出す
}

// setupDeploymentEchoContext はテスト用の Echo コンテキストを生成するヘルパー関数
func setupDeploymentEchoContext(method, path, body string, params map[string]string) (echo.Context, *httptest.ResponseRecorder) {
	echoInstance := echo.New()                                            // Echo インスタンスを生成する
	bodyReader := strings.NewReader(body)                                 // リクエストボディを設定する
	request := httptest.NewRequest(method, path, bodyReader)             // テスト用リクエストを生成する
	request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON) // Content-Type を JSON に設定する
	responseRecorder := httptest.NewRecorder()                            // テスト用レスポンスレコーダーを生成する
	echoCtx := echoInstance.NewContext(request, responseRecorder)         // Echo コンテキストを生成する

	if len(params) > 0 { // パスパラメータが存在する場合は設定する
		paramNames := make([]string, 0, len(params))
		paramValues := make([]string, 0, len(params))
		for paramName, paramValue := range params {
			paramNames = append(paramNames, paramName)
			paramValues = append(paramValues, paramValue)
		}
		echoCtx.SetParamNames(paramNames...)   // パラメータ名を設定する
		echoCtx.SetParamValues(paramValues...) // パラメータ値を設定する
	}

	return echoCtx, responseRecorder
}

// TestCreateDeployment_正常にdeploymentが作成される は POST で status=pending、全フィールドが pending_*** に入ることを確認する
func TestCreateDeployment_正常にdeploymentが作成される(t *testing.T) {
	expectedDeployment := &models.Deployment{
		ID:                  "deployment-id-1",           // deployment ID を設定する
		ProjectID:           "project-id-1",              // project ID を設定する
		Name:                "my-app",                    // deployment 名を設定する
		Type:                models.DeploymentTypeImageURL, // deployment タイプを設定する
		Status:              models.DeploymentStatusPending, // ステータスを pending に設定する
		AppStatus:           models.AppStatusPending,     // アプリステータスを pending に設定する
		PendingImageURL:     "nginx:latest",              // pending image_url を設定する
		PendingInstanceSize: "small",                     // pending instance_size を設定する
		PendingReplicas:     1,                           // pending replicas を設定する
	}

	mockSvc := &mockDeploymentService{
		createDeploymentFunc: func(ctx context.Context, req service.CreateDeploymentRequest) (*models.Deployment, error) {
			return expectedDeployment, nil // 作成した deployment を返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                                       // ハンドラーを生成する
	requestJSON := `{"name":"my-app","type":"image_url","image_url":"nginx:latest","instance_size":"small","replicas":1}`                    // リクエスト JSON を定義する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodPost, "/api/v1/projects/project-id-1/deployments", requestJSON, map[string]string{"id": "project-id-1"}) // テスト用コンテキストを生成する

	err := deploymentHandler.CreateDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusCreated { // 201 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusCreated, responseRecorder.Code)
	}

	var actualDeployment models.Deployment
	if err := json.NewDecoder(responseRecorder.Body).Decode(&actualDeployment); err != nil { // レスポンスをデコードする
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}
	if actualDeployment.Status != models.DeploymentStatusPending { // status が pending であることを確認する
		t.Errorf("期待する status: pending, 実際の status: %s", actualDeployment.Status)
	}
	if actualDeployment.AppStatus != models.AppStatusPending { // app_status が pending であることを確認する
		t.Errorf("期待する app_status: pending, 実際の app_status: %s", actualDeployment.AppStatus)
	}
	if actualDeployment.PendingImageURL != "nginx:latest" { // pending_image_url が設定されていることを確認する
		t.Errorf("期待する pending_image_url: nginx:latest, 実際の pending_image_url: %s", actualDeployment.PendingImageURL)
	}
}

// TestCreateDeployment_サービスエラーで500になる はサービスエラー時に 500 が返ることを確認する
func TestCreateDeployment_サービスエラーで500になる(t *testing.T) {
	mockSvc := &mockDeploymentService{
		createDeploymentFunc: func(ctx context.Context, req service.CreateDeploymentRequest) (*models.Deployment, error) {
			return nil, errors.New("DB エラー") // エラーを返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                           // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodPost, "/api/v1/projects/project-id-1/deployments", `{"name":"my-app","type":"image_url"}`, map[string]string{"id": "project-id-1"}) // テスト用コンテキストを生成する

	err := deploymentHandler.CreateDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusInternalServerError { // 500 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestUpdateDeployment_送ったフィールドのみpendingが更新される は PUT で送ったフィールドのみ更新されることを確認する
func TestUpdateDeployment_送ったフィールドのみpendingが更新される(t *testing.T) {
	updatedImageURL := "nginx:1.25"
	expectedDeployment := &models.Deployment{
		ID:              "deployment-id-1",            // deployment ID を設定する
		PendingImageURL: updatedImageURL,              // 更新後の pending_image_url を設定する
		PendingReplicas: 0,                            // replicas は送っていないので変化しない
	}

	var capturedRequest service.UpdateDeploymentRequest // キャプチャしたリクエストを格納する変数を定義する
	mockSvc := &mockDeploymentService{
		updateDeploymentFunc: func(ctx context.Context, deploymentID string, req service.UpdateDeploymentRequest) (*models.Deployment, error) {
			capturedRequest = req          // リクエストをキャプチャする
			return expectedDeployment, nil // 更新後の deployment を返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                                                   // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodPut, "/api/v1/deployments/deployment-id-1", `{"image_url":"nginx:1.25"}`, map[string]string{"id": "deployment-id-1"}) // テスト用コンテキストを生成する

	err := deploymentHandler.UpdateDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusOK { // 200 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusOK, responseRecorder.Code)
	}
	if capturedRequest.ImageURL == nil || *capturedRequest.ImageURL != updatedImageURL { // image_url が送られていることを確認する
		t.Errorf("image_url が正しく送られていません: %v", capturedRequest.ImageURL)
	}
	if capturedRequest.Replicas != nil { // replicas は送っていないので nil であることを確認する
		t.Errorf("replicas は送っていないので nil であるべきです: %v", capturedRequest.Replicas)
	}
	if capturedRequest.InstanceSize != nil { // instance_size は送っていないので nil であることを確認する
		t.Errorf("instance_size は送っていないので nil であるべきです: %v", capturedRequest.InstanceSize)
	}
}

// TestUpdateDeployment_存在しないdeploymentは404になる は存在しない deployment ID で 404 が返ることを確認する
func TestUpdateDeployment_存在しないdeploymentは404になる(t *testing.T) {
	mockSvc := &mockDeploymentService{
		updateDeploymentFunc: func(ctx context.Context, deploymentID string, req service.UpdateDeploymentRequest) (*models.Deployment, error) {
			return nil, gorm.ErrRecordNotFound // レコードが存在しないエラーを返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                                        // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodPut, "/api/v1/deployments/nonexistent", `{"image_url":"nginx:latest"}`, map[string]string{"id": "nonexistent"}) // テスト用コンテキストを生成する

	err := deploymentHandler.UpdateDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusNotFound { // 404 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusNotFound, responseRecorder.Code)
	}
}

// TestDeleteDeployment_statusがdeletingになる は DELETE で status が deleting に変更されることを確認する
func TestDeleteDeployment_statusがdeletingになる(t *testing.T) {
	expectedDeployment := &models.Deployment{
		ID:     "deployment-id-1",               // deployment ID を設定する
		Status: models.DeploymentStatusDeleting, // status が deleting であることを確認する
	}

	mockSvc := &mockDeploymentService{
		deleteDeploymentFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return expectedDeployment, nil // deleting 状態の deployment を返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                             // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodDelete, "/api/v1/deployments/deployment-id-1", "", map[string]string{"id": "deployment-id-1"}) // テスト用コンテキストを生成する

	err := deploymentHandler.DeleteDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusOK { // 200 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusOK, responseRecorder.Code)
	}

	var actualDeployment models.Deployment
	if err := json.NewDecoder(responseRecorder.Body).Decode(&actualDeployment); err != nil { // レスポンスをデコードする
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}
	if actualDeployment.Status != models.DeploymentStatusDeleting { // status が deleting であることを確認する
		t.Errorf("期待する status: deleting, 実際の status: %s", actualDeployment.Status)
	}
}

// TestDeleteDeployment_存在しないdeploymentは404になる は存在しない deployment ID で 404 が返ることを確認する
func TestDeleteDeployment_存在しないdeploymentは404になる(t *testing.T) {
	mockSvc := &mockDeploymentService{
		deleteDeploymentFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return nil, gorm.ErrRecordNotFound // レコードが存在しないエラーを返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                            // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodDelete, "/api/v1/deployments/nonexistent", "", map[string]string{"id": "nonexistent"}) // テスト用コンテキストを生成する

	err := deploymentHandler.DeleteDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusNotFound { // 404 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusNotFound, responseRecorder.Code)
	}
}

// TestCreateDeployment_Service作成失敗時に500になる は Service 作成失敗時に 500 が返ることを確認する
func TestCreateDeployment_Service作成失敗時に500になる(t *testing.T) {
	mockSvc := &mockDeploymentService{
		createDeploymentFunc: func(ctx context.Context, req service.CreateDeploymentRequest) (*models.Deployment, error) {
			return nil, errors.New("Service レコードの作成に失敗しました") // Service 作成失敗を返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                                        // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodPost, "/api/v1/projects/project-id-1/deployments", `{"name":"my-app","type":"image_url"}`, map[string]string{"id": "project-id-1"}) // テスト用コンテキストを生成する

	err := deploymentHandler.CreateDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusInternalServerError { // 500 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusInternalServerError, responseRecorder.Code)
	}
}

// TestGetDeployment_正常にdeployment詳細が返る は GET /deployments/:id で詳細が取得できることを確認する
func TestGetDeployment_正常にdeployment詳細が返る(t *testing.T) {
	expectedDeployment := &models.Deployment{
		ID:        "deployment-id-1",              // deployment ID を設定する
		Name:      "my-app",                       // deployment 名を設定する
		ProjectID: "project-id-1",                 // project ID を設定する
		Status:    models.DeploymentStatusRunning, // status を設定する
	}

	mockSvc := &mockDeploymentService{
		getDeploymentFunc: func(ctx context.Context, deploymentID string) (*models.Deployment, error) {
			return expectedDeployment, nil // 期待する deployment を返す
		},
	}

	deploymentHandler := NewDeploymentHandler(mockSvc)                                                                              // ハンドラーを生成する
	echoCtx, responseRecorder := setupDeploymentEchoContext(http.MethodGet, "/api/v1/deployments/deployment-id-1", "", map[string]string{"id": "deployment-id-1"}) // テスト用コンテキストを生成する

	err := deploymentHandler.GetDeployment(echoCtx) // ハンドラーを実行する
	if err != nil {
		t.Fatalf("ハンドラーがエラーを返しました: %v", err)
	}
	if responseRecorder.Code != http.StatusOK { // 200 が返ることを確認する
		t.Errorf("期待するステータスコード: %d, 実際のステータスコード: %d", http.StatusOK, responseRecorder.Code)
	}

	var actualDeployment models.Deployment
	if err := json.NewDecoder(responseRecorder.Body).Decode(&actualDeployment); err != nil { // レスポンスをデコードする
		t.Fatalf("レスポンスのデコードに失敗しました: %v", err)
	}
	if actualDeployment.ID != "deployment-id-1" { // deployment ID を確認する
		t.Errorf("期待する ID: deployment-id-1, 実際の ID: %s", actualDeployment.ID)
	}
}
