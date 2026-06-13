package service

import (
	"app/k8s"
	"app/k8s/manifest"
	"app/models"
	"app/repository"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	k8sclient "k8s.io/client-go/kubernetes"
)

// ErrAlreadyApplying は apply 中の deployment に再 apply しようとした場合のエラー
var ErrAlreadyApplying = errors.New("already applying")

// ApplyServiceInterface は apply サービスのインターフェース
type ApplyServiceInterface interface {
	Apply(ctx context.Context, userID string, deploymentID string) (*ApplyResult, error) // apply を実行する
}

// ApplyService は apply のコアロジックを実装するサービス
type ApplyService struct {
	DB                *gorm.DB                          // データベース接続（トランザクション管理用）
	K8s               k8sclient.Interface               // k8s クライアント
	DeploymentRepo    repository.DeploymentRepository   // deployment リポジトリ
	ApplyHistoryRepo  repository.ApplyHistoryRepository // apply_history リポジトリ
	ProjectRepository repository.ProjectRepository      // project リポジトリ
}

// ApplyResult は Apply 処理の結果を表す構造体
type ApplyResult struct {
	ApplyHistoryID string // apply_history の ID
	Status         string // apply の結果ステータス
	BuildID        string // ビルドが必要な場合に設定（Phase8 で使用）
}

// NewApplyService は ApplyService を生成して返す
func NewApplyService(
	db *gorm.DB,
	k8sClient k8sclient.Interface,
	deploymentRepo repository.DeploymentRepository,
	applyHistoryRepo repository.ApplyHistoryRepository,
	projectRepository repository.ProjectRepository,
) *ApplyService {
	return &ApplyService{ // 依存を注入して返す
		DB:               db,
		K8s:              k8sClient,
		DeploymentRepo:   deploymentRepo,
		ApplyHistoryRepo: applyHistoryRepo,
		ProjectRepository: projectRepository,
	}
}

// Apply は deployment に対して apply を実行する
func (applyService *ApplyService) Apply(ctx context.Context, userID string, deploymentID string) (*ApplyResult, error) {
	var applyResult *ApplyResult // 結果を格納する変数を定義する

	err := applyService.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error { // トランザクションを開始する
		// 1. SELECT FOR UPDATE でロックを取得する
		deploymentData, err := applyService.DeploymentRepo.FindByIDForUpdate(ctx, tx, deploymentID) // FOR UPDATE ロック付きで deployment を取得する
		if err != nil {
			return fmt.Errorf("deployment not found: %w", err) // 取得エラーを返す
		}

		// 所有権を確認する（トランザクション外で取得）
		ownerProjectData, ownerErr := applyService.ProjectRepository.FindByIDNoTx(ctx, deploymentData.ProjectID) // project を取得する
		if ownerErr != nil {
			return fmt.Errorf("project not found: %w", ownerErr) // 取得エラーを返す
		}
		if ownerProjectData.UserID != userID { // UserID が一致しない場合は禁止エラーを返す
			return ErrForbidden
		}

		// apply 中の deployment への二重 apply を防ぐ
		if deploymentData.AppStatus == models.AppStatusDeploying { // 既に apply 中の場合は競合エラーを返す
			return ErrAlreadyApplying
		}

		// 2. Project を取得する（namespace 解決のため）
		projectData, err := applyService.ProjectRepository.FindByID(ctx, tx, deploymentData.ProjectID) // ProjectRepository 経由で project を取得する
		if err != nil {
			return fmt.Errorf("project not found: %w", err) // 取得エラーを返す
		}

		// 3. pending_*** から使用する実効値を決定する
		imageURL := deploymentData.PendingImageURL // pending の image_url を使う
		if imageURL == "" {                        // pending が空の場合は current 値を使う
			imageURL = deploymentData.ImageURL
		}

		instanceSize := deploymentData.PendingInstanceSize // pending の instance_size を使う
		if instanceSize == "" {                            // pending が空の場合は current 値を使う
			instanceSize = deploymentData.InstanceSize
		}

		replicas := deploymentData.PendingReplicas // pending の replicas を使う
		if replicas == 0 {                         // pending が 0 の場合は current 値を使う
			replicas = deploymentData.Replicas
		}
		if replicas == 0 { // current も 0 の場合はデフォルト値を設定する
			replicas = 1
		}

		command := deploymentData.PendingCommand // pending の command を使う
		if len(command) == 0 {                   // pending が空の場合は current 値を使う
			command = deploymentData.Command
		}

		args := deploymentData.PendingArgs // pending の args を使う
		if len(args) == 0 {                // pending が空の場合は current 値を使う
			args = deploymentData.Args
		}

		// 4. instance_size マスターを取得してマニフェスト生成用データを組み立てる
		var instanceSizeData models.InstanceSize                               // instance_size を格納する変数を定義する
		tx.WithContext(ctx).First(&instanceSizeData, "size = ?", instanceSize) // instance_size マスターを取得する

		deploymentForManifest := *deploymentData          // manifest 生成用にコピーする
		deploymentForManifest.InstanceSize = instanceSize // 実効 instance_size を設定する
		deploymentForManifest.Replicas = replicas         // 実効 replicas を設定する
		deploymentForManifest.Command = command           // 実効 command を設定する
		deploymentForManifest.Args = args                 // 実効 args を設定する

		// 5. k8s Deployment マニフェストを生成する
		manifestGenerator := &manifest.Generator{ // マニフェストジェネレーターを生成する
			InstanceSizes: map[string]models.InstanceSize{instanceSize: instanceSizeData},
		}
		deploymentManifest := manifestGenerator.GenerateDeployment(deploymentForManifest, projectData.Namespace, imageURL, nil, nil) // マニフェストを生成する

		// 6. apply_history を INSERT する
		manifestJSON, _ := json.Marshal(deploymentManifest) // マニフェストを JSON にシリアライズする
		applyHistoryRecord := &models.ApplyHistory{         // apply_history レコードを生成する
			DeploymentID: deploymentID,
			Manifests:    manifestJSON,
			Status:       models.ApplyStatusApplied, // 初期ステータスは applied とする
			AppliedAt:    time.Now(),
		}
		if err := applyService.ApplyHistoryRepo.Create(ctx, tx, applyHistoryRecord); err != nil { // apply_history を作成する
			return fmt.Errorf("apply_history create: %w", err) // 作成エラーを返す
		}

		// 7. k8s に apply する
		if err := k8s.ApplyDeployment(ctx, applyService.K8s, deploymentManifest); err != nil { // k8s に Deployment を apply する
			applyHistoryRecord.Status = models.ApplyStatusFailed                                                                                  // k8s apply 失敗時はステータスを failed に変更する
			applyHistoryRecord.ErrorMessage = err.Error()                                                                                         // エラーメッセージを記録する
			if updateErr := applyService.ApplyHistoryRepo.UpdateStatus(ctx, tx, applyHistoryRecord, models.ApplyStatusFailed); updateErr != nil { // ステータスを更新する
				return fmt.Errorf("apply_history update: %w", updateErr) // 更新エラーを返す
			}
			return fmt.Errorf("k8s apply: %w", err) // k8s apply エラーを返す
		}

		// 8. pending_*** を空にして current 値に昇格させる
		appliedAt := time.Now() // apply 完了時刻を記録する
		updates := map[string]interface{}{
			"image_url":                     imageURL,                                  // image_url を昇格する
			"pending_image_url":             "",                                        // pending_image_url をクリアする
			"instance_size":                 instanceSize,                              // instance_size を昇格する
			"pending_instance_size":         "",                                        // pending_instance_size をクリアする
			"replicas":                      replicas,                                  // replicas を昇格する
			"pending_replicas":              0,                                         // pending_replicas をクリアする
			"github_repo_url":               deploymentData.PendingGithubRepoURL,       // github_repo_url を昇格する
			"pending_github_repo_url":       "",                                        // pending_github_repo_url をクリアする
			"github_branch":                 deploymentData.PendingGithubBranch,        // github_branch を昇格する
			"pending_github_branch":         "",                                        // pending_github_branch をクリアする
			"github_commit_sha":             deploymentData.PendingGithubCommitSHA,     // github_commit_sha を昇格する
			"pending_github_commit_sha":     "",                                        // pending_github_commit_sha をクリアする
			"github_repo_directory":         deploymentData.PendingGithubRepoDirectory, // github_repo_directory を昇格する
			"pending_github_repo_directory": "",                                        // pending_github_repo_directory をクリアする
			"dockerfile_path":               deploymentData.PendingDockerfilePath,      // dockerfile_path を昇格する
			"pending_dockerfile_path":       "",                                        // pending_dockerfile_path をクリアする
			"command":                       command,                                   // command を昇格する
			"pending_command":               nil,                                       // pending_command をクリアする
			"args":                          args,                                      // args を昇格する
			"pending_args":                  nil,                                       // pending_args をクリアする
			"status":                        models.DeploymentStatusRunning,            // status を running に更新する
			"app_status":                    models.AppStatusDeploying,                 // app_status を deploying に更新する
			"applied_at":                    &appliedAt,                                // applied_at を更新する
		}
		if err := applyService.DeploymentRepo.Updates(ctx, tx, deploymentData, updates); err != nil { // deployment を更新する
			return fmt.Errorf("deployment updates: %w", err) // 更新エラーを返す
		}

		applyResult = &ApplyResult{ // 結果を設定する
			ApplyHistoryID: applyHistoryRecord.ID,
			Status:         "applied",
		}
		return nil // トランザクションをコミットする
	})

	return applyResult, err // 結果とエラーを返す
}
