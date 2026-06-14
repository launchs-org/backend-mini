package manifest

import (
	"app/models"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Generator は manifest 生成に必要な設定を保持する構造体
type Generator struct {
	InstanceSizes map[string]models.InstanceSize // インスタンスサイズのマップ
}

// GenerateDeployment は DB の Deployment モデルから k8s Deployment manifest を生成する
func (generator *Generator) GenerateDeployment(
	deploymentData models.Deployment,
	namespace string,
	imageURL string,
	envMounts []models.EnvVarMount,
	volumeMounts []models.VolumeMount,
) *appsv1.Deployment {
	instanceSize := generator.InstanceSizes[deploymentData.InstanceSize] // インスタンスサイズを取得する

	container := corev1.Container{
		Name:  "app",   // コンテナ名を設定する
		Image: imageURL, // イメージ URL を設定する
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(instanceSize.CPURequest),    // CPU リクエストを設定する
				corev1.ResourceMemory: resource.MustParse(instanceSize.MemoryRequest), // メモリリクエストを設定する
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse(instanceSize.CPULimit),    // CPU リミットを設定する
				corev1.ResourceMemory: resource.MustParse(instanceSize.MemoryLimit), // メモリリミットを設定する
			},
		},
	}

	if len(deploymentData.Command) > 0 { // command が指定されている場合のみ設定する
		container.Command = deploymentData.Command
	}
	if len(deploymentData.Args) > 0 { // args が指定されている場合のみ設定する
		container.Args = deploymentData.Args
	}

	replicas := deploymentData.Replicas // レプリカ数を取得する

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentData.Name, // Deployment 名を設定する
			Namespace: namespace,           // namespace を設定する
			Labels: map[string]string{
				"launchs.org/deployment-id": deploymentData.ID,   // デプロイメント ID ラベルを設定する
				"app":                       deploymentData.Name,  // アプリ名ラベルを設定する
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas, // レプリカ数を設定する
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": deploymentData.Name}, // セレクターを設定する
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": deploymentData.Name}, // Pod ラベルを設定する
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container}, // コンテナを設定する
				},
			},
		},
	}
}

// GenerateService は DB の Service モデルから k8s Service manifest を生成する
func (generator *Generator) GenerateService(
	serviceData models.Service,
	deploymentName string,
	namespace string,
) *corev1.Service {
	port := serviceData.PendingPort // pending_port を使う
	if port == 0 {                  // pending が 0 の場合は current 値を使う
		port = serviceData.Port
	}
	targetPort := serviceData.PendingTargetPort // pending_target_port を使う
	if targetPort == 0 {                        // pending が 0 の場合は current 値を使う
		targetPort = serviceData.TargetPort
	}
	serviceType := corev1.ServiceType(serviceData.Type) // Service タイプを設定する
	if serviceType == "" {                              // 未設定の場合はデフォルトを ClusterIP にする
		serviceType = corev1.ServiceTypeClusterIP
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName + "-svc", // Service 名をデプロイメント名から生成する
			Namespace: namespace,               // namespace を設定する
			Labels: map[string]string{
				"launchs.org/service-id": serviceData.ID, // サービス ID ラベルを設定する
				"app":                    deploymentName,  // アプリ名ラベルを設定する
			},
		},
		Spec: corev1.ServiceSpec{
			Type: serviceType, // Service タイプを設定する
			Selector: map[string]string{
				"app": deploymentName, // Pod セレクターを設定する
			},
			Ports: []corev1.ServicePort{
				{
					Port:       int32(port),                    // 公開ポートを設定する
					TargetPort: intstr.FromInt32(int32(targetPort)), // ターゲットポートを設定する
					Protocol:   corev1.ProtocolTCP,             // プロトコルを TCP に設定する
				},
			},
		},
	}
}
