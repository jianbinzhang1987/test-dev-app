package internal

import "time"

// Protocol 定义支持的传输协议类型
type Protocol string

const (
	ProtocolSFTP Protocol = "SFTP"
	ProtocolSCP  Protocol = "SCP"
	ProtocolFTP  Protocol = "FTP"
)

// ConnectionStatus 定义节点连接状态
type ConnectionStatus string

const (
	StatusConnected    ConnectionStatus = "connected"
	StatusDisconnected ConnectionStatus = "disconnected"
	StatusTesting      ConnectionStatus = "testing"
	StatusError        ConnectionStatus = "error"
)

// 认证方式常量
type AuthMethod string

const (
	AuthMethodPassword AuthMethod = "password" // 密码认证
	AuthMethodKey      AuthMethod = "key"      // SSH密钥认证
	AuthMethodAgent    AuthMethod = "agent"    // SSH Agent认证
)

// Node 定义节点（服务器）的基本信息
type Node struct {
	ID       string   `json:"id"`       // 节点唯一标识
	Name     string   `json:"name"`     // 节点易记名称
	IP       string   `json:"ip"`       // IP地址
	Port     int      `json:"port"`     // 端口号
	Protocol Protocol `json:"protocol"` // 通信协议
	IsMaster bool     `json:"isMaster"` // 是否为主控节点

	// 认证相关字段
	Username   string     `json:"username,omitempty"`   // SSH用户名
	AuthMethod AuthMethod `json:"authMethod,omitempty"` // 认证方式 ("password", "key", "agent")
	KeyPath    string     `json:"keyPath,omitempty"`    // SSH私钥路径（仅当authMethod=key时）

	// 注意：密码和密钥密码短语不存储在此结构中
	// 它们通过系统密钥链管理（见 internal/credential/store.go）

	// 运行时状态（不持久化）
	Status *NodeStatus `json:"-"`
}

// NodeStatus 定义节点的运行时状态信息
type NodeStatus struct {
	Latency     int              `json:"latency"`     // 延迟(ms)
	LastChecked string           `json:"lastChecked"` // 最后检测时间
	Status      ConnectionStatus `json:"status"`      // 连接状态
	ErrorMsg    string           `json:"errorMsg"`    // 错误信息（如果有）
}

// TopologyData 定义拓扑结构数据，用于前端可视化
type TopologyData struct {
	Master *Node   `json:"master"` // 主控节点
	Slaves []*Node `json:"slaves"` // 从节点列表
	Total  int     `json:"total"`  // 总节点数
}

// NodeCollection 节点集合，用于持久化存储
type NodeCollection struct {
	Nodes     []*Node   `json:"nodes"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// ===== SVN 资源管理模型 =====

// SVNResourceType 定义资源类型
type SVNResourceType string

const (
	SVNResourceFile   SVNResourceType = "file"
	SVNResourceFolder SVNResourceType = "folder"
)

// SVNResourceStatus 定义资源状态
type SVNResourceStatus string

const (
	SVNStatusOnline  SVNResourceStatus = "online"
	SVNStatusError   SVNResourceStatus = "error"
	SVNStatusSyncing SVNResourceStatus = "syncing"
)

// SVNResource 定义 SVN 资源信息
type SVNResource struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Name        string            `json:"name"`
	Type        SVNResourceType   `json:"type"`
	Revision    string            `json:"revision"`
	Status      SVNResourceStatus `json:"status"`
	LastChecked string            `json:"lastChecked"`
	Size        string            `json:"size,omitempty"`
	Username    string            `json:"username,omitempty"`
}

// SVNResourceCollection SVN 资源集合
type SVNResourceCollection struct {
	Resources []*SVNResource `json:"resources"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// SVNTestResult SVN 连接测试结果
type SVNTestResult struct {
	Ok         bool   `json:"ok"`
	Revision   string `json:"revision,omitempty"`
	Message    string `json:"message,omitempty"`
	DurationMs int    `json:"durationMs,omitempty"`
	CheckedAt  string `json:"checkedAt"`
}

// ===== 任务编排模型 =====

// TaskStatus 任务状态枚举
type TaskStatus string

const (
	TaskStatusIdle        TaskStatus = "IDLE"
	TaskStatusDownloading TaskStatus = "DOWNLOADING"
	TaskStatusUploading   TaskStatus = "UPLOADING"
	TaskStatusSyncing     TaskStatus = "SYNCING"
	TaskStatusExecuting   TaskStatus = "EXECUTING"
	TaskStatusSuccess     TaskStatus = "SUCCESS"
	TaskStatusFailed      TaskStatus = "FAILED"
)

// TaskRunRequest 任务执行请求
type TaskRunRequest struct {
	TaskID           string            `json:"taskId"`
	TaskName         string            `json:"taskName,omitempty"`
	SVNResourceID    string            `json:"svnResourceId"`
	MasterServerID   string            `json:"masterServerId"`
	SlaveServerIDs   []string          `json:"slaveServerIds"`
	RemotePath       string            `json:"remotePath"`
	SlaveRemotePath  string            `json:"slaveRemotePath,omitempty"`
	SlaveRemotePaths map[string]string `json:"slaveRemotePaths,omitempty"`
	Commands         []string          `json:"commands"`
}

// TaskEvent 任务状态事件
type TaskEvent struct {
	TaskID   string     `json:"taskId"`
	RunID    string     `json:"runId,omitempty"`
	Status   TaskStatus `json:"status"`
	Progress int        `json:"progress"`
	Log      string     `json:"log"`
}

// ===== 任务编排数据模型 =====

// TaskDefinition 任务编排定义
// 只存储配置与状态，不包含敏感凭据
type TaskDefinition struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	SVNResourceID    string            `json:"svnResourceId"`
	MasterServerID   string            `json:"masterServerId"`
	SlaveServerIDs   []string          `json:"slaveServerIds"`
	RemotePath       string            `json:"remotePath"`
	SlaveRemotePath  string            `json:"slaveRemotePath,omitempty"`
	SlaveRemotePaths map[string]string `json:"slaveRemotePaths,omitempty"`
	Commands         []string          `json:"commands"`
	Status           TaskStatus        `json:"status"`
	Progress         int               `json:"progress"`
	CreatedAt        string            `json:"createdAt"`
	UpdatedAt        string            `json:"updatedAt"`
	LastRunAt        string            `json:"lastRunAt,omitempty"`
	TemplateID       string            `json:"templateId,omitempty"`
}

// TaskTemplate 任务模板
type TaskTemplate struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	SVNResourceID    string            `json:"svnResourceId"`
	MasterServerID   string            `json:"masterServerId"`
	SlaveServerIDs   []string          `json:"slaveServerIds"`
	RemotePath       string            `json:"remotePath"`
	SlaveRemotePath  string            `json:"slaveRemotePath,omitempty"`
	SlaveRemotePaths map[string]string `json:"slaveRemotePaths,omitempty"`
	Commands         []string          `json:"commands"`
	SourceTaskID     string            `json:"sourceTaskId,omitempty"`
	CreatedAt        string            `json:"createdAt"`
	UpdatedAt        string            `json:"updatedAt"`
}

// TaskRun 任务执行历史
type TaskRun struct {
	ID         string     `json:"id"`
	TaskID     string     `json:"taskId"`
	TaskName   string     `json:"taskName"`
	Status     TaskStatus `json:"status"`
	Progress   int        `json:"progress"`
	StartedAt  string     `json:"startedAt"`
	FinishedAt string     `json:"finishedAt,omitempty"`
	Logs       []string   `json:"logs"`
}

// TaskStore 任务持久化存储集合
type TaskStore struct {
	Tasks     []*TaskDefinition `json:"tasks"`
	Templates []*TaskTemplate   `json:"templates"`
	Runs      []*TaskRun        `json:"runs"`
	UpdatedAt time.Time         `json:"updatedAt"`
}
