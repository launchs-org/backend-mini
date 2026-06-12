# ISSUE-036 ビルド判定ロジック・HEAD SHA 解決

## 親 Issue
ISSUE-035

## 概要
apply 時にビルドが必要かどうかを判定し、commit_sha が HEAD の場合は GitHub API で実 SHA を取得する。

## 実装手順

### 1. `service/apply.go` に追加

```go
// HEAD SHA 解決
func resolveCommitSHA(ctx context.Context, repoURL, branch, sha string) (string, error) {
    if sha != "HEAD" { return sha, nil }

    // GitHub API: GET https://api.github.com/repos/{owner}/{repo}/commits/{branch}
    // パブリックリポジトリのみ対応（認証不要）
    parts := parseGithubURL(repoURL) // owner / repo を抽出
    url := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/%s", parts.Owner, parts.Repo, branch)

    resp, err := http.Get(url)
    if err != nil { return "", err }
    defer resp.Body.Close()

    var result struct {
        SHA string `json:"sha"`
    }
    json.NewDecoder(resp.Body).Decode(&result)
    return result.SHA, nil
}

// ビルド要否判定
func needsBuild(d models.Deployment) bool {
    if d.Type == models.DeploymentTypeImageURL { return false }

    return d.PendingGithubRepoURL != d.GithubRepoURL ||
        d.PendingGithubBranch != d.GithubBranch ||
        d.PendingGithubCommitSHA != d.GithubCommitSHA ||
        d.PendingGithubRepoDirectory != d.GithubRepoDirectory ||
        d.PendingDockerfilePath != d.DockerfilePath
}
```

## テスト確認項目

- [ ] `github_commit_sha = "HEAD"` が GitHub API から取得した実 SHA に変換されること
- [ ] image_url 型でビルド不要と判定されること
- [ ] GitHub 情報が変化していない場合にビルド不要と判定されること
- [ ] GitHub 情報が変化している場合にビルド必要と判定されること
