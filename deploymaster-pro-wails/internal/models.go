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
	LastChecked time.Time        `json:"lastChecked"` // 最后检测时间
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
