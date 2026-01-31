package svn

import (
	"deploymaster-pro-wails/internal"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Service SVN 资源管理服务
// 负责资源增删改查与本地持久化
// 不直接处理凭据存储
type Service struct {
	storage   Storage
	resources []*internal.SVNResource
	mu        sync.RWMutex
}

// NewService 创建 SVN 服务
func NewService(storage Storage) (*Service, error) {
	s := &Service{
		storage:   storage,
		resources: make([]*internal.SVNResource, 0),
	}

	if err := s.loadResources(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) loadResources() error {
	resources, err := s.storage.Load()
	if err != nil {
		return err
	}

	changed := false
	for _, r := range resources {
		if r.ID == "" {
			r.ID = uuid.NewString()
			changed = true
		}
		if r.Revision == "" {
			r.Revision = "HEAD"
			changed = true
		}
		if r.Status == "" {
			r.Status = internal.SVNStatusOnline
			changed = true
		}
		if r.LastChecked == "" {
			r.LastChecked = time.Now().Format("2006-01-02 15:04")
			changed = true
		}
	}

	s.resources = resources

	if changed {
		return s.saveResources()
	}

	return nil
}

func (s *Service) saveResources() error {
	return s.storage.Save(s.resources)
}

// AddResource 添加新资源
func (s *Service) AddResource(resource *internal.SVNResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if resource.ID == "" {
		resource.ID = uuid.NewString()
	}
	if resource.Revision == "" {
		resource.Revision = "HEAD"
	}
	if resource.Status == "" {
		resource.Status = internal.SVNStatusOnline
	}
	if resource.LastChecked == "" {
		resource.LastChecked = time.Now().Format("2006-01-02 15:04")
	}

	for _, r := range s.resources {
		if r.ID == resource.ID {
			return ErrResourceExists
		}
	}

	s.resources = append(s.resources, resource)
	return s.saveResources()
}

// UpdateResource 更新资源信息
func (s *Service) UpdateResource(resource *internal.SVNResource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.resources {
		if r.ID == resource.ID {
			if resource.Revision == "" {
				resource.Revision = r.Revision
			}
			if resource.Status == "" {
				resource.Status = r.Status
			}
			if resource.LastChecked == "" {
				resource.LastChecked = time.Now().Format("2006-01-02 15:04")
			}
			s.resources[i] = resource
			return s.saveResources()
		}
	}

	return ErrResourceNotFound
}

// DeleteResource 删除资源
func (s *Service) DeleteResource(resourceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.resources {
		if r.ID == resourceID {
			s.resources = append(s.resources[:i], s.resources[i+1:]...)
			return s.saveResources()
		}
	}

	return ErrResourceNotFound
}

// GetResource 获取单个资源
func (s *Service) GetResource(resourceID string) (*internal.SVNResource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, r := range s.resources {
		if r.ID == resourceID {
			return r, nil
		}
	}

	return nil, ErrResourceNotFound
}

// ListResources 获取所有资源
func (s *Service) ListResources() []*internal.SVNResource {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*internal.SVNResource, len(s.resources))
	copy(result, s.resources)
	return result
}
