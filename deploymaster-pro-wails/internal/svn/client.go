package svn

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Client SVN 命令行客户端封装
// 依赖系统安装的 svn CLI
// 通过执行 svn info 获取修订号
type Client struct {
	timeout time.Duration
}

// NewClient 创建 SVN 客户端
func NewClient(timeout time.Duration) *Client {
	return &Client{timeout: timeout}
}

// CheckAvailable 检查系统是否安装 svn CLI
func (c *Client) CheckAvailable() error {
	_, err := exec.LookPath("svn")
	if err != nil {
		return errors.New("svn client not found")
	}
	return nil
}

// Info 获取 SVN 资源信息（当前仅取修订号）
func (c *Client) Info(ctx context.Context, url, username, password string) (string, error) {
	if strings.TrimSpace(url) == "" {
		return "", errors.New("svn url is empty")
	}

	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	args := []string{
		"info",
		"--show-item",
		"revision",
		"--non-interactive",
		"--no-auth-cache",
		"--trust-server-cert",
	}

	if strings.TrimSpace(username) != "" {
		args = append(args, "--username", username)
	}
	if strings.TrimSpace(password) != "" {
		args = append(args, "--password", password)
	}

	args = append(args, url)

	cmd := exec.CommandContext(ctx, "svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("svn info failed: %s", msg)
	}

	rev := strings.TrimSpace(string(output))
	if rev == "" {
		return "", errors.New("svn info returned empty revision")
	}

	return rev, nil
}

// Export 将 SVN 资源导出到目标目录（不包含 .svn 元数据）
func (c *Client) Export(ctx context.Context, url, username, password, revision, dest string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("svn url is empty")
	}
	if strings.TrimSpace(dest) == "" {
		return errors.New("export destination is empty")
	}

	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	args := []string{
		"export",
		"--non-interactive",
		"--no-auth-cache",
		"--trust-server-cert",
		"--force",
	}

	if rev := normalizeRevision(revision); rev != "" {
		args = append(args, "--revision", rev)
	}
	if strings.TrimSpace(username) != "" {
		args = append(args, "--username", username)
	}
	if strings.TrimSpace(password) != "" {
		args = append(args, "--password", password)
	}

	args = append(args, url, dest)

	cmd := exec.CommandContext(ctx, "svn", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(output))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("svn export failed: %s", msg)
	}

	return nil
}

// CatToFile 将 SVN 文件资源写入到目标文件
func (c *Client) CatToFile(ctx context.Context, url, username, password, revision, destFile string) error {
	if strings.TrimSpace(url) == "" {
		return errors.New("svn url is empty")
	}
	if strings.TrimSpace(destFile) == "" {
		return errors.New("export destination is empty")
	}

	if c.timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, c.timeout)
		defer cancel()
	}

	if err := os.MkdirAll(filepath.Dir(destFile), 0755); err != nil {
		return err
	}

	dst, err := os.Create(destFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	args := []string{
		"cat",
		"--non-interactive",
		"--no-auth-cache",
		"--trust-server-cert",
	}

	if rev := normalizeRevision(revision); rev != "" {
		args = append(args, "--revision", rev)
	}
	if strings.TrimSpace(username) != "" {
		args = append(args, "--username", username)
	}
	if strings.TrimSpace(password) != "" {
		args = append(args, "--password", password)
	}

	args = append(args, url)

	cmd := exec.CommandContext(ctx, "svn", args...)
	var stderr bytes.Buffer
	cmd.Stdout = dst
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("svn cat failed: %s", msg)
	}

	return nil
}

func normalizeRevision(revision string) string {
	rev := strings.TrimSpace(revision)
	if rev == "" {
		return ""
	}
	if rev == "0" {
		return ""
	}
	return rev
}
