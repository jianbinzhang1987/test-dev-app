package ssh

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pkg/sftp"
)

// NewSFTPClient 创建 SFTP 客户端
func (c *Client) NewSFTPClient() (*sftp.Client, error) {
	if c.client == nil {
		return nil, fmt.Errorf("not connected")
	}
	return sftp.NewClient(c.client)
}

// UploadPath 上传本地路径到远端
// localPath 可以是文件或目录，remotePath 为目标目录或文件路径
func UploadPath(client *sftp.Client, localPath, remotePath string) error {
	info, err := os.Stat(localPath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return uploadDir(client, localPath, remotePath)
	}

	return uploadFile(client, localPath, remotePath)
}

func uploadDir(client *sftp.Client, localDir, remoteDir string) error {
	if err := client.MkdirAll(remoteDir); err != nil {
		return err
	}

	return filepath.WalkDir(localDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		remotePath := filepath.ToSlash(filepath.Join(remoteDir, rel))
		if d.IsDir() {
			return client.MkdirAll(remotePath)
		}
		return uploadFile(client, path, remotePath)
	})
}

func uploadFile(client *sftp.Client, localFile, remoteFile string) error {
	if err := client.MkdirAll(filepath.ToSlash(filepath.Dir(remoteFile))); err != nil {
		return err
	}

	src, err := os.Open(localFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := client.Create(remoteFile)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
