package svn

import (
	"deploymaster-pro-wails/internal"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	// ErrResourceNotFound 资源不存在
	ErrResourceNotFound = errors.New("svn resource not found")
	// ErrResourceExists 资源已存在
	ErrResourceExists = errors.New("svn resource already exists")
)

// Storage 定义资源存储接口
// 复用节点的加密器用于凭据存储
// 注意：资源数据本身不加密
//
//go:generate echo "no codegen"
type Storage interface {
	Load() ([]*internal.SVNResource, error)
	Save(resources []*internal.SVNResource) error
}

// JSONStorage 基于JSON文件的存储实现
// 存储文件名：svn-resources.json
// 文件放置位置与节点数据一致
// 通过互斥锁保证并发安全
type JSONStorage struct {
	filePath string
	mu       sync.RWMutex
}

// NewJSONStorage 创建 SVN 资源存储实例
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "svn-resources.json")
	return &JSONStorage{
		filePath: filePath,
	}, nil
}

// Load 从文件加载资源数据
func (s *JSONStorage) Load() ([]*internal.SVNResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []*internal.SVNResource{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var collection internal.SVNResourceCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, err
	}

	return collection.Resources, nil
}

// Save 保存资源数据到文件
func (s *JSONStorage) Save(resources []*internal.SVNResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collection := internal.SVNResourceCollection{
		Resources: resources,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filePath)
}
