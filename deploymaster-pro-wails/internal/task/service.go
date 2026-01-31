package task

import (
	"crypto/rand"
	"deploymaster-pro-wails/internal"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	// ErrTaskNotFound 任务不存在
	ErrTaskNotFound = errors.New("task not found")
	// ErrTaskExists 任务已存在
	ErrTaskExists = errors.New("task already exists")
	// ErrTemplateNotFound 模板不存在
	ErrTemplateNotFound = errors.New("template not found")
	// ErrTemplateExists 模板已存在
	ErrTemplateExists = errors.New("template already exists")
	// ErrRunNotFound 运行记录不存在
	ErrRunNotFound = errors.New("run not found")
)

// Service 任务服务
// 提供任务、模板、运行历史的持久化管理
type Service struct {
	storage   Storage
	tasks     []*internal.TaskDefinition
	templates []*internal.TaskTemplate
	runs      []*internal.TaskRun
	mu        sync.RWMutex
}

// NewService 创建任务服务实例
func NewService(storage Storage) (*Service, error) {
	store, err := storage.Load()
	if err != nil {
		return nil, err
	}
	return &Service{
		storage:   storage,
		tasks:     store.Tasks,
		templates: store.Templates,
		runs:      store.Runs,
	}, nil
}

func nowString() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

func newID(prefix string) string {
	buf := make([]byte, 8)
	_, _ = rand.Read(buf)
	return prefix + "-" + hex.EncodeToString(buf)
}

func (s *Service) saveLocked() error {
	store := &internal.TaskStore{
		Tasks:     s.tasks,
		Templates: s.templates,
		Runs:      s.runs,
		UpdatedAt: time.Now(),
	}
	return s.storage.Save(store)
}

// ===== Tasks =====

// ListTasks 返回所有任务
func (s *Service) ListTasks() []*internal.TaskDefinition {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]*internal.TaskDefinition{}, s.tasks...)
}

// GetTask 获取任务
func (s *Service) GetTask(taskID string) (*internal.TaskDefinition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, t := range s.tasks {
		if t.ID == taskID {
			return t, nil
		}
	}
	return nil, ErrTaskNotFound
}

// AddTask 添加任务
func (s *Service) AddTask(task *internal.TaskDefinition) (*internal.TaskDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.ID == "" {
		task.ID = newID("task")
	}
	for _, t := range s.tasks {
		if t.ID == task.ID {
			return nil, ErrTaskExists
		}
	}

	if task.Status == "" {
		task.Status = internal.TaskStatusIdle
	}
	if task.CreatedAt == "" {
		task.CreatedAt = nowString()
	}
	task.UpdatedAt = nowString()

	s.tasks = append([]*internal.TaskDefinition{task}, s.tasks...)
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return task, nil
}

// UpdateTask 更新任务
func (s *Service) UpdateTask(task *internal.TaskDefinition) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.tasks {
		if existing.ID != task.ID {
			continue
		}

		updated := *existing
		if task.Name != "" {
			updated.Name = task.Name
		}
		if task.SVNResourceID != "" {
			updated.SVNResourceID = task.SVNResourceID
		}
		if task.MasterServerID != "" {
			updated.MasterServerID = task.MasterServerID
		}
		if task.SlaveServerIDs != nil {
			updated.SlaveServerIDs = task.SlaveServerIDs
		}
		if task.RemotePath != "" {
			updated.RemotePath = task.RemotePath
		}
		if task.SlaveRemotePath != "" {
			updated.SlaveRemotePath = task.SlaveRemotePath
		}
		if task.SlaveRemotePaths != nil {
			updated.SlaveRemotePaths = task.SlaveRemotePaths
		}
		if task.Commands != nil {
			updated.Commands = task.Commands
		}
		if task.Status != "" {
			updated.Status = task.Status
		}
		if task.Progress >= 0 {
			updated.Progress = task.Progress
		}
		if task.LastRunAt != "" {
			updated.LastRunAt = task.LastRunAt
		}
		if task.TemplateID != "" {
			updated.TemplateID = task.TemplateID
		}
		if updated.CreatedAt == "" {
			updated.CreatedAt = nowString()
		}
		updated.UpdatedAt = nowString()

		s.tasks[i] = &updated
		return s.saveLocked()
	}
	return ErrTaskNotFound
}

// UpdateTaskState 更新任务运行状态
func (s *Service) UpdateTaskState(taskID string, status internal.TaskStatus, progress int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.tasks {
		if existing.ID != taskID {
			continue
		}
		updated := *existing
		if status != "" {
			updated.Status = status
		}
		if progress >= 0 {
			updated.Progress = progress
		}
		updated.LastRunAt = nowString()
		updated.UpdatedAt = nowString()
		s.tasks[i] = &updated
		return s.saveLocked()
	}
	return ErrTaskNotFound
}

// DeleteTask 删除任务
func (s *Service) DeleteTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, t := range s.tasks {
		if t.ID == taskID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrTaskNotFound
	}

	s.tasks = append(s.tasks[:idx], s.tasks[idx+1:]...)

	// 清理运行历史
	filtered := make([]*internal.TaskRun, 0, len(s.runs))
	for _, r := range s.runs {
		if r.TaskID != taskID {
			filtered = append(filtered, r)
		}
	}
	s.runs = filtered

	return s.saveLocked()
}

// ===== Templates =====

// ListTemplates 返回所有模板
func (s *Service) ListTemplates() []*internal.TaskTemplate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]*internal.TaskTemplate{}, s.templates...)
}

// AddTemplate 添加模板
func (s *Service) AddTemplate(tpl *internal.TaskTemplate) (*internal.TaskTemplate, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tpl.ID == "" {
		tpl.ID = newID("tpl")
	}
	for _, t := range s.templates {
		if t.ID == tpl.ID {
			return nil, ErrTemplateExists
		}
	}
	if tpl.CreatedAt == "" {
		tpl.CreatedAt = nowString()
	}
	tpl.UpdatedAt = nowString()

	s.templates = append([]*internal.TaskTemplate{tpl}, s.templates...)
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return tpl, nil
}

// UpdateTemplate 更新模板
func (s *Service) UpdateTemplate(tpl *internal.TaskTemplate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.templates {
		if existing.ID != tpl.ID {
			continue
		}

		updated := *existing
		if tpl.Name != "" {
			updated.Name = tpl.Name
		}
		if tpl.SVNResourceID != "" {
			updated.SVNResourceID = tpl.SVNResourceID
		}
		if tpl.MasterServerID != "" {
			updated.MasterServerID = tpl.MasterServerID
		}
		if tpl.SlaveServerIDs != nil {
			updated.SlaveServerIDs = tpl.SlaveServerIDs
		}
		if tpl.RemotePath != "" {
			updated.RemotePath = tpl.RemotePath
		}
		if tpl.SlaveRemotePath != "" {
			updated.SlaveRemotePath = tpl.SlaveRemotePath
		}
		if tpl.SlaveRemotePaths != nil {
			updated.SlaveRemotePaths = tpl.SlaveRemotePaths
		}
		if tpl.Commands != nil {
			updated.Commands = tpl.Commands
		}
		if tpl.SourceTaskID != "" {
			updated.SourceTaskID = tpl.SourceTaskID
		}
		updated.UpdatedAt = nowString()
		if updated.CreatedAt == "" {
			updated.CreatedAt = nowString()
		}

		s.templates[i] = &updated
		return s.saveLocked()
	}
	return ErrTemplateNotFound
}

// DeleteTemplate 删除模板
func (s *Service) DeleteTemplate(templateID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, t := range s.templates {
		if t.ID == templateID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrTemplateNotFound
	}

	s.templates = append(s.templates[:idx], s.templates[idx+1:]...)
	return s.saveLocked()
}

// ===== Runs =====

// CreateRun 创建运行记录
func (s *Service) CreateRun(taskID, taskName string) (*internal.TaskRun, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run := &internal.TaskRun{
		ID:        newID("run"),
		TaskID:    taskID,
		TaskName:  taskName,
		Status:    internal.TaskStatusIdle,
		Progress:  0,
		StartedAt: nowString(),
		Logs:      []string{},
	}

	s.runs = append([]*internal.TaskRun{run}, s.runs...)
	if err := s.saveLocked(); err != nil {
		return nil, err
	}
	return run, nil
}

// AppendRunLog 追加运行日志并更新状态
func (s *Service) AppendRunLog(runID string, status internal.TaskStatus, progress int, logLine string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, r := range s.runs {
		if r.ID != runID {
			continue
		}
		updated := *r
		if status != "" {
			updated.Status = status
		}
		if progress >= 0 {
			updated.Progress = progress
		}
		if logLine != "" {
			updated.Logs = append(updated.Logs, logLine)
		}
		if status == internal.TaskStatusSuccess || status == internal.TaskStatusFailed {
			if updated.FinishedAt == "" {
				updated.FinishedAt = nowString()
			}
		}

		s.runs[i] = &updated
		return s.saveLocked()
	}
	return ErrRunNotFound
}

// ListRuns 返回所有运行记录
func (s *Service) ListRuns() []*internal.TaskRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]*internal.TaskRun{}, s.runs...)
}

// ListRunsByTask 返回指定任务的运行记录
func (s *Service) ListRunsByTask(taskID string) []*internal.TaskRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*internal.TaskRun, 0)
	for _, r := range s.runs {
		if r.TaskID == taskID {
			result = append(result, r)
		}
	}
	return result
}

// DeleteRun 删除运行记录
func (s *Service) DeleteRun(runID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, r := range s.runs {
		if r.ID == runID {
			idx = i
			break
		}
	}
	if idx == -1 {
		return ErrRunNotFound
	}

	s.runs = append(s.runs[:idx], s.runs[idx+1:]...)
	return s.saveLocked()
}

// DeleteRunsByTask 删除任务下全部历史
func (s *Service) DeleteRunsByTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filtered := make([]*internal.TaskRun, 0, len(s.runs))
	for _, r := range s.runs {
		if r.TaskID != taskID {
			filtered = append(filtered, r)
		}
	}
	s.runs = filtered
	return s.saveLocked()
}
