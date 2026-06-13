# ISSUE-030 apply サービスに Volume Mount 追加

## 親 Issue
ISSUE-026

## 実装手順

### `service/apply.go` に追加

```go
// volume_mounts を取得して PVC apply と Deployment volumeMounts を設定
var volMounts []models.VolumeMount
tx.Preload("Volume").
    Where("deployment_id = ? AND status != ?", deploymentID, "deleting").
    Find(&volMounts)

for _, vm := range volMounts {
    // PVC を apply
    k8s.ApplyPVC(ctx, s.K8s, project.Namespace, vm.Volume.Name, vm.Volume.SizeMB)

    // pending_mount_path → mount_path に昇格
    mountPath := vm.PendingMountPath
    if mountPath == "" { mountPath = vm.MountPath }
    tx.Model(&vm).Updates(map[string]interface{}{
        "mount_path":         mountPath,
        "pending_mount_path": "",
        "status":             models.VolumeMountStatusMounted,
    })
}

// manifest generator に volume_mounts を渡して
// Deployment の spec.volumes と spec.containers[].volumeMounts を設定
```

## テスト確認項目

- [ ] apply 後に k8s PVC が作成されること
- [ ] apply 後に `pending_mount_path` が空になること
- [ ] apply 後に Deployment の volumeMounts に設定されること

### repository 層テスト

- [ ] `VolumeMountRepository.Save` で apply 後に `pending_mount_path` が空になること
- [ ] `VolumeMountRepository.FindAllByDeploymentID` で全ボリュームマウントが取得できること
