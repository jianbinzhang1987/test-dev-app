package node

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
	// ErrNodeNotFound 节点不存在
	ErrNodeNotFound = errors.New("node not found")
	// ErrNodeExists 节点已存在
	ErrNodeExists = errors.New("node already exists")
)

// Storage 定义节点存储接口
type Storage interface {
	Load() ([]*internal.Node, error)
	Save(nodes []*internal.Node) error
	GetCrypto() *Crypto
}

// JSONStorage 基于JSON文件的存储实现
type JSONStorage struct {
	filePath string
	crypto   *Crypto
	mu       sync.RWMutex
}

// NewJSONStorage 创建JSON存储实例
func NewJSONStorage(dataDir string) (*JSONStorage, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	crypto, err := NewCrypto(dataDir)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(dataDir, "nodes.json")
	return &JSONStorage{
		filePath: filePath,
		crypto:   crypto,
	}, nil
}

// GetDefaultDataDir 获取默认数据目录
func GetDefaultDataDir() (string, error) {
	// 1. 优先尝试便携模式：检查程序同级目录下是否有 data 目录
	exePath, err := os.Executable()
	if err == nil {
		portableDir := filepath.Join(filepath.Dir(exePath), "data")
		// 如果 data 目录已存在，或者我们可以创建它，则使用便携模式
		if _, err := os.Stat(portableDir); err == nil {
			return portableDir, nil
		}
	}

	// 2. 回退到用户家目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, ".deploymaster"), nil
}

// Load 从文件加载节点数据
func (s *JSONStorage) Load() ([]*internal.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 如果文件不存在，返回空列表
	if _, err := os.Stat(s.filePath); os.IsNotExist(err) {
		return []*internal.Node{}, nil
	}

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, err
	}

	var collection internal.NodeCollection
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, err
	}

	return collection.Nodes, nil
}

// Save 保存节点数据到文件
func (s *JSONStorage) Save(nodes []*internal.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	collection := internal.NodeCollection{
		Nodes:     nodes,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	// 使用临时文件 + 原子重命名确保数据安全
	tmpFile := s.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0644); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.filePath)
}

// GetCrypto 获取加密器实例
func (s *JSONStorage) GetCrypto() *Crypto {
	return s.crypto
}
