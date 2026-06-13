package service

import (
	"app/k8s"
	"app/models"
	"app/repository"
	"context"
	"fmt"

	k8sclient "k8s.io/client-go/kubernetes"

	"gorm.io/gorm"
)

// ProjectService は Project の CRUD ビジネスロジックを定義するインターフェース
type ProjectService interface {
	CreateProject(ctx context.Context, userID string, req CreateProjectRequest) (*models.Project, error)        // project を作成する
	ListProjects(ctx context.Context, userID string) ([]*models.Project, error)                                 // project 一覧を取得する
	GetProject(ctx context.Context, projectID string) (*models.Project, error)                                  // project を取得する
	UpdateProject(ctx context.Context, projectID string, req UpdateProjectRequest) (*models.Project, error)     // project を更新する
	DeleteProject(ctx context.Context, projectID string) error                                                  // project を削除する
}

// CreateProjectRequest は POST /projects のリクエスト構造体
type CreateProjectRequest struct {
	Name string `json:"name"` // プロジェクト名（k8s namespace 名にもなる）
}

// UpdateProjectRequest は PUT /projects/:id のリクエスト構造体
type UpdateProjectRequest struct {
	Name *string `json:"name"` // nil の場合は更新しない
}

// projectServiceImpl は ProjectService の実装
type projectServiceImpl struct {
	db                          *gorm.DB                                   // データベース接続（トランザクション開始に使用する）
	projectRepo                 repository.ProjectRepository               // project リポジトリ
	harborCredentialRepo        repository.HarborCredentialRepository      // harbor credential リポジトリ
	k8sClient                   k8sclient.Interface                        // k8s クライアント
	harborClient                *k8s.HarborClient                          // Harbor API クライアント（管理用 robot）
}

// NewProjectService は ProjectService の実装を返す
func NewProjectService(
	db *gorm.DB,
	projectRepo repository.ProjectRepository,
	harborCredentialRepo repository.HarborCredentialRepository,
	k8sClient k8sclient.Interface,
	harborClient *k8s.HarborClient,
) ProjectService {
	return &projectServiceImpl{
		db:                   db,                   // DB 接続を注入する
		projectRepo:          projectRepo,          // project リポジトリを注入する
		harborCredentialRepo: harborCredentialRepo, // harbor credential リポジトリを注入する
		k8sClient:            k8sClient,            // k8s クライアントを注入する
		harborClient:         harborClient,         // Harbor クライアントを注入する
	}
}

// CreateProject は Project を作成し、Harbor project・robot account と k8s namespace を同時に作成する
// 外部リソース作成失敗時は補償処理で作成済みリソースを削除する
func (svc *projectServiceImpl) CreateProject(ctx context.Context, userID string, req CreateProjectRequest) (*models.Project, error) {
	var createdProject *models.Project

	// DB トランザクションを開始する
	err := svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Project レコードを作成する
		projectData := &models.Project{
			UserID:    userID,                           // ユーザーIDを設定する
			Name:      req.Name,                         // プロジェクト名を設定する
			Namespace: k8s.ToNamespaceName(req.Name),    // namespace 名として有効な文字列に変換する
			Status:    models.ProjectStatusProvisioning, // 初期ステータスを設定する
		}
		if err := svc.projectRepo.Create(ctx, tx, projectData); err != nil {
			return fmt.Errorf("project レコードの作成に失敗しました: %w", err)
		}

		// Harbor project を作成する（失敗時は DB ロールバック）
		if err := svc.harborClient.CreateHarborProject(ctx, req.Name); err != nil {
			return fmt.Errorf("harbor project の作成に失敗しました: %w", err)
		}

		// Harbor robot account を作成する（失敗時は管理用 robot で Harbor project を補償削除して DB ロールバック）
		robotCredential, err := svc.harborClient.CreateHarborRobotAccount(ctx, req.Name)
		if err != nil {
			// robot account 未作成なので管理用 robot で補償削除する
			_ = svc.harborClient.DeleteHarborProject(ctx, req.Name, svc.harborClient.AdminCredential()) // ベストエフォートで補償削除する
			return fmt.Errorf("harbor robot account の作成に失敗しました: %w", err)
		}

		// HarborCredential レコードを DB に保存する（失敗時は project 専用 robot で Harbor project を補償削除して DB ロールバック）
		credentialData := &models.HarborCredential{
			ProjectID:      projectData.ID,              // プロジェクト ID を設定する
			RobotName:      robotCredential.Name,        // robot アカウント名を設定する
			RobotSecret:    robotCredential.Secret,      // シークレットを設定する
			HarborEndpoint: svc.harborClient.Endpoint(), // エンドポイントを設定する
		}
		if err := svc.harborCredentialRepo.Create(ctx, tx, credentialData); err != nil {
			_ = svc.harborClient.DeleteHarborProject(ctx, req.Name, *robotCredential) // ベストエフォートで補償削除する
			return fmt.Errorf("harbor credential レコードの作成に失敗しました: %w", err)
		}

		// k8s namespace を作成する（失敗時は project 専用 robot で Harbor project を補償削除して DB ロールバック）
		if err := k8s.CreateNamespace(ctx, svc.k8sClient, projectData.Namespace); err != nil { // 変換済みの namespace 名を使う
			_ = svc.harborClient.DeleteHarborProject(ctx, req.Name, *robotCredential) // ベストエフォートで補償削除する
			return fmt.Errorf("k8s namespace の作成に失敗しました: %w", err)
		}

		// すべての作成が成功したら status を active に更新する
		if err := svc.projectRepo.UpdateStatus(ctx, tx, projectData, models.ProjectStatusActive); err != nil {
			_ = svc.harborClient.DeleteHarborProject(ctx, req.Name, *robotCredential)         // ベストエフォートで補償削除する
			_ = k8s.DeleteNamespace(ctx, svc.k8sClient, projectData.Namespace)                // ベストエフォートで補償削除する
			return fmt.Errorf("project ステータスの更新に失敗しました: %w", err)
		}
		projectData.Status = models.ProjectStatusActive // ローカルの値も更新する
		createdProject = projectData                    // 外側の変数に結果を格納する
		return nil
	})

	if err != nil {
		return nil, err // トランザクションエラーを返す
	}
	return createdProject, nil // 作成した project を返す
}

// ListProjects は userID に紐づく project 一覧を返す
func (svc *projectServiceImpl) ListProjects(ctx context.Context, userID string) ([]*models.Project, error) {
	return svc.projectRepo.FindAllByUserID(ctx, userID) // リポジトリ経由で取得する
}

// GetProject は projectID に対応する project を返す
func (svc *projectServiceImpl) GetProject(ctx context.Context, projectID string) (*models.Project, error) {
	return svc.projectRepo.FindByID(ctx, svc.db, projectID) // リポジトリ経由で取得する
}

// UpdateProject は projectID の project 名を部分更新する
func (svc *projectServiceImpl) UpdateProject(ctx context.Context, projectID string, req UpdateProjectRequest) (*models.Project, error) {
	projectData, err := svc.projectRepo.FindByID(ctx, svc.db, projectID) // リポジトリ経由で取得する
	if err != nil {
		return nil, err // 取得エラーを返す
	}

	if req.Name != nil {
		projectData.Name = *req.Name // 名前を更新する
	}

	if err := svc.projectRepo.Save(ctx, projectData); err != nil { // リポジトリ経由で保存する
		return nil, err // 保存エラーを返す
	}
	return projectData, nil // 更新後の project を返す
}

// DeleteProject は project を deleting 状態にし、Harbor project 削除・k8s namespace 削除・DB 削除を実行する
func (svc *projectServiceImpl) DeleteProject(ctx context.Context, projectID string) error {
	// DB トランザクションを開始する
	return svc.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// project を取得する
		projectData, err := svc.projectRepo.FindByID(ctx, tx, projectID)
		if err != nil {
			return fmt.Errorf("project の取得に失敗しました: %w", err)
		}

		// status を deleting に更新する
		if err := svc.projectRepo.UpdateStatus(ctx, tx, projectData, models.ProjectStatusDeleting); err != nil {
			return fmt.Errorf("project ステータスの更新に失敗しました: %w", err)
		}

		// DB から project 専用の Harbor 認証情報を取得する
		credentialData, err := svc.harborCredentialRepo.FindByProjectID(ctx, tx, projectID)
		if err != nil {
			return fmt.Errorf("harbor credential の取得に失敗しました: %w", err)
		}

		// project 作成時に生成した専用 robot の認証情報で Harbor project を削除する
		if err := svc.harborClient.DeleteHarborProject(ctx, projectData.Name, k8s.HarborRobotCredential{
			Name:   credentialData.RobotName,   // DB に保存した robot 名を使う
			Secret: credentialData.RobotSecret, // DB に保存したシークレットを使う
		}); err != nil {
			return fmt.Errorf("harbor project の削除に失敗しました: %w", err)
		}

		// k8s namespace を削除する
		if err := k8s.DeleteNamespace(ctx, svc.k8sClient, projectData.Namespace); err != nil {
			return fmt.Errorf("k8s namespace の削除に失敗しました: %w", err)
		}

		// HarborCredential レコードを削除する
		if err := svc.harborCredentialRepo.DeleteByProjectID(ctx, tx, projectID); err != nil {
			return fmt.Errorf("harbor credential レコードの削除に失敗しました: %w", err)
		}

		// Project レコードを削除する
		if err := svc.projectRepo.Delete(ctx, tx, projectData); err != nil {
			return fmt.Errorf("project レコードの削除に失敗しました: %w", err)
		}

		return nil
	})
}
