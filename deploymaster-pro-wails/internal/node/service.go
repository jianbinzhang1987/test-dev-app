package node

import (
	"deploymaster-pro-wails/internal"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// Service 节点管理服务
type Service struct {
	storage Storage
	nodes   []*internal.Node
	mu      sync.RWMutex
}

// NewService 创建节点服务实例
func NewService(storage Storage) (*Service, error) {
	s := &Service{
		storage: storage,
		nodes:   make([]*internal.Node, 0),
	}

	// 从存储加载节点数据
	if err := s.loadNodes(); err != nil {
		return nil, err
	}

	return s, nil
}

// loadNodes 从存储加载节点数据
func (s *Service) loadNodes() error {
	nodes, err := s.storage.Load()
	if err != nil {
		return err
	}

	// 为缺失 ID 的旧数据补全 ID 并持久化
	changed := false
	for _, n := range nodes {
		if n.ID == "" {
			n.ID = uuid.NewString()
			changed = true
		}
	}

	s.nodes = nodes

	if changed {
		return s.saveNodes()
	}

	return nil
}

// saveNodes 保存节点数据到存储
func (s *Service) saveNodes() error {
	return s.storage.Save(s.nodes)
}

// AddNode 添加新节点
func (s *Service) AddNode(node *internal.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 后端兜底生成 ID，避免空 ID 导致删除/更新失败
	if node.ID == "" {
		node.ID = uuid.NewString()
	}

	// 检查ID是否已存在
	for _, n := range s.nodes {
		if n.ID == node.ID {
			return ErrNodeExists
		}
	}

	s.nodes = append(s.nodes, node)
	return s.saveNodes()
}

// UpdateNode 更新节点信息
func (s *Service) UpdateNode(node *internal.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, n := range s.nodes {
		if n.ID == node.ID {
			s.nodes[i] = node
			return s.saveNodes()
		}
	}

	return ErrNodeNotFound
}

// DeleteNode 删除节点
func (s *Service) DeleteNode(nodeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, n := range s.nodes {
		if n.ID == nodeID {
			// 删除节点（保持顺序）
			s.nodes = append(s.nodes[:i], s.nodes[i+1:]...)
			return s.saveNodes()
		}
	}

	return ErrNodeNotFound
}

// GetNode 获取单个节点
func (s *Service) GetNode(nodeID string) (*internal.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, n := range s.nodes {
		if n.ID == nodeID {
			return n, nil
		}
	}

	return nil, ErrNodeNotFound
}

// ListNodes 获取所有节点列表
func (s *Service) ListNodes() []*internal.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 返回副本，避免外部修改
	result := make([]*internal.Node, len(s.nodes))
	copy(result, s.nodes)
	return result
}

// GetMasterNode 获取主控节点
func (s *Service) GetMasterNode() (*internal.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, n := range s.nodes {
		if n.IsMaster {
			return n, nil
		}
	}

	return nil, errors.New("no master node configured")
}

// GetSlaveNodes 获取所有从节点
func (s *Service) GetSlaveNodes() []*internal.Node {
	s.mu.RLock()
	defer s.mu.RUnlock()

	slaves := make([]*internal.Node, 0)
	for _, n := range s.nodes {
		if !n.IsMaster {
			slaves = append(slaves, n)
		}
	}

	return slaves
}
