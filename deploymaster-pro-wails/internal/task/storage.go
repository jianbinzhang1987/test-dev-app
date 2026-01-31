package task

import (
	"deploymaster-pro-wails/internal"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Storage 定义任务存储接口
// 包含任务、模板和执行历史
// 注意：仅存储配置与历史摘要，敏感凭据不存储
//
//go:generate echo "no codegen"
type Storage interface {
	Load() (*internal.TaskStore, error)
	Save(store *internal.TaskStore) error
}

// JSONStorage 基于JSON文件的存储实现
// 存储文件名：tasks.json
// 文件放置位置与节点/资源数据一致
type JSONStorage struct {
	filePath string
	mu       sync.RWMutex
}

// NewJSONStorage 创建任务存储实例
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "tasks.json")
	return &JSONStorage{filePath: filePath}, nil
}

// Load 从文件加载任务数据
func (s *JSONStorage) Load() (*internal.TaskStore, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return &internal.TaskStore{
			Tasks:     []*internal.TaskDefinition{},
			Templates: []*internal.TaskTemplate{},
			Runs:      []*internal.TaskRun{},
			UpdatedAt: time.Now(),
		}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var store internal.TaskStore
	if err := json.Unmarshal(data, &store); err != nil {
		return nil, err
	}

	if store.Tasks == nil {
		store.Tasks = []*internal.TaskDefinition{}
	}
	if store.Templates == nil {
		store.Templates = []*internal.TaskTemplate{}
	}
	if store.Runs == nil {
		store.Runs = []*internal.TaskRun{}
	}

	return &store, nil
}

// Save 保存任务数据到文件
func (s *JSONStorage) Save(store *internal.TaskStore) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	store.UpdatedAt = time.Now()

	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}

	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filePath)
}
