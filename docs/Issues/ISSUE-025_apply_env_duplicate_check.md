# ISSUE-025 apply サービスに env 重複チェック・ConfigMap/Secret 追加

## 親 Issue
ISSUE-021

## 概要
apply 時に env_var の実効キー重複チェックを行い、ConfigMap/Secret も k8s に apply する。
apply 後に env_var_mounts.status を applied に更新する。

## 実装手順

### `internal/service/apply.go` に追加

```go
// 1. env_var 実効キーの重複チェック
var mounts []model.EnvVarMount
tx.Preload("EnvVar").
    Where("deployment_id = ? AND status != ?", deploymentID, "deleting").
    Find(&mounts)

seen := map[string]bool{}
for _, m := range mounts {
    effectiveKey := m.OverrideKey
    if effectiveKey == "" { effectiveKey = m.EnvVar.Key }
    if seen[effectiveKey] {
        return fmt.Errorf("duplicate env key: %s", effectiveKey)
    }
    seen[effectiveKey] = true
}

// 2. ConfigMap / Secret を apply
configData, secretData := buildEnvData(mounts)

if len(configData) > 0 {
    k8s.ApplyConfigMap(ctx, s.K8s, project.Namespace, d.Name+"-env", configData)
}
if len(secretData) > 0 {
    k8s.ApplySecret(ctx, s.K8s, project.Namespace, d.Name+"-secret", secretData)
}

// 3. Deployment manifest に envFrom を追加
// manifest/generator.go の GenerateDeployment を拡張して
// ConfigMap/Secret の envFrom を Deployment spec に追加する

// 4. apply 後に env_var_mounts.status = applied に更新
tx.Model(&model.EnvVarMount{}).
    Where("deployment_id = ? AND status = ?", deploymentID, model.EnvVarMountStatusPending).
    Update("status", model.EnvVarMountStatusApplied)
```

## テスト確認項目

- [ ] 実効キーが重複する mount を持った状態で apply すると 400 になること
- [ ] apply 後に k8s ConfigMap が作成されること
- [ ] apply 後に k8s Secret が作成されること
- [ ] apply 後に env_var_mounts.status が applied になること
- [ ] k8s Deployment の envFrom に ConfigMap/Secret が含まれること
