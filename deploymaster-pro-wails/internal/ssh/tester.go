package ssh

import (
	"deploymaster-pro-wails/internal"
	"deploymaster-pro-wails/internal/credential"
	"fmt"
	"sync"
	"time"
)

// Tester SSH连接测试器
type Tester struct {
	timeout   time.Duration
	credStore *credential.Store
}

// NewTester 创建连接测试器
func NewTester(credStore *credential.Store) *Tester {
	return &Tester{
		timeout:   10 * time.Second,
		credStore: credStore,
	}
}

// TestConnection 测试单个节点连接
// 根据节点的认证方式自动选择合适的认证方法
// 如果node.AuthMethod未设置，则使用传入的username/password进行密码认证
func (t *Tester) TestConnection(node *internal.Node, username, password string) *internal.NodeStatus {
	status := &internal.NodeStatus{
		LastChecked: time.Now(),
		Status:      internal.StatusTesting,
	}

	// 记录开始时间
	startTime := time.Now()

	var client *Client
	var err error

	// 根据认证方式创建客户端
	switch node.AuthMethod {
	case internal.AuthMethodKey:
		// SSH密钥认证
		if node.KeyPath == "" {
			status.Status = internal.StatusError
			status.ErrorMsg = "密钥认证模式但未提供密钥路径"
			return status
		}

		// 尝试从凭据存储读取密钥密码短语
		passphrase := ""
		if t.credStore != nil {
			passphrase, _ = t.credStore.GetKeyPassphrase(node.ID)
		}

		client, err = NewClientWithKeyFile(node.Username, node.KeyPath, passphrase)
		if err != nil {
			status.Status = internal.StatusError
			status.ErrorMsg = fmt.Sprintf("创建SSH密钥客户端失败: %v", err)
			return status
		}

	case internal.AuthMethodAgent:
		// SSH Agent认证
		client, err = NewClientWithAgent(node.Username)
		if err != nil {
			status.Status = internal.StatusError
			status.ErrorMsg = fmt.Sprintf("创建SSH Agent客户端失败: %v", err)
			return status
		}

	default:
		// 密码认证（默认）
		actualUsername := username
		actualPassword := password

		// 如果有凭据存储，尝试从中读取
		if t.credStore != nil && node.ID != "" && node.Username != "" {
			storedPassword, err := t.credStore.GetPassword(node.ID, node.Username)
			if err == nil && storedPassword != "" {
				actualUsername = node.Username
				actualPassword = storedPassword
			}
		}

		client = NewClient(actualUsername, actualPassword)
	}

	// 尝试连接
	err = client.Connect(node.IP, node.Port)
	if err != nil {
		status.Status = internal.StatusError
		status.ErrorMsg = fmt.Sprintf("连接失败: %v", err)
		return status
	}
	defer client.Close()

	// 计算延迟（SSH握手时间）
	latency := time.Since(startTime).Milliseconds()

	// 执行简单命令验证连接
	_, err = client.ExecuteCommand("echo 'ping'")
	if err != nil {
		status.Status = internal.StatusError
		status.ErrorMsg = fmt.Sprintf("命令执行失败: %v", err)
		return status
	}

	// 连接成功
	status.Status = internal.StatusConnected
	status.Latency = int(latency)
	status.ErrorMsg = ""

	return status
}

// TestConnectionWithCredentials 使用提供的凭据测试连接
// 用于首次测试或更新凭据时使用
func (t *Tester) TestConnectionWithCredentials(node *internal.Node, username, password, keyPassphrase string) *internal.NodeStatus {
	status := &internal.NodeStatus{
		LastChecked: time.Now(),
		Status:      internal.StatusTesting,
	}

	startTime := time.Now()

	var client *Client
	var err error

	switch node.AuthMethod {
	case internal.AuthMethodKey:
		if node.KeyPath == "" {
			status.Status = internal.StatusError
			status.ErrorMsg = "密钥认证模式但未提供密钥路径"
			return status
		}

		client, err = NewClientWithKeyFile(username, node.KeyPath, keyPassphrase)
		if err != nil {
			status.Status = internal.StatusError
			status.ErrorMsg = fmt.Sprintf("密钥认证失败: %v", err)
			return status
		}

	case internal.AuthMethodAgent:
		client, err = NewClientWithAgent(username)
		if err != nil {
			status.Status = internal.StatusError
			status.ErrorMsg = fmt.Sprintf("Agent认证失败: %v", err)
			return status
		}

	default:
		client = NewClient(username, password)
	}

	err = client.Connect(node.IP, node.Port)
	if err != nil {
		status.Status = internal.StatusError
		status.ErrorMsg = fmt.Sprintf("连接失败: %v", err)
		return status
	}
	defer client.Close()

	latency := time.Since(startTime).Milliseconds()

	_, err = client.ExecuteCommand("echo 'ping'")
	if err != nil {
		status.Status = internal.StatusError
		status.ErrorMsg = fmt.Sprintf("命令执行失败: %v", err)
		return status
	}

	status.Status = internal.StatusConnected
	status.Latency = int(latency)
	status.ErrorMsg = ""

	return status
}

// BatchTestConnections 批量测试多个节点连接
func (t *Tester) BatchTestConnections(nodes []*internal.Node, username, password string) map[string]*internal.NodeStatus {
	results := make(map[string]*internal.NodeStatus)
	resultsMu := sync.Mutex{}

	// 使用goroutine并发测试
	var wg sync.WaitGroup

	for _, node := range nodes {
		wg.Add(1)

		go func(n *internal.Node) {
			defer wg.Done()

			status := t.TestConnection(n, username, password)

			resultsMu.Lock()
			results[n.ID] = status
			resultsMu.Unlock()
		}(node)
	}

	wg.Wait()
	return results
}

// QuickPing 快速Ping测试（仅测试TCP连接，不进行SSH认证）
func (t *Tester) QuickPing(node *internal.Node) *internal.NodeStatus {
	status := &internal.NodeStatus{
		LastChecked: time.Now(),
		Status:      internal.StatusTesting,
	}

	startTime := time.Now()

	// 这里可以使用net.DialTimeout进行TCP连接测试
	// 暂时使用简单的SSH连接测试
	// 在实际生产中，可以优化为仅测试TCP端口可达性

	client := NewClient("test", "test") // 临时凭据
	err := client.Connect(node.IP, node.Port)

	if err != nil {
		status.Status = internal.StatusDisconnected
		status.ErrorMsg = fmt.Sprintf("连接不可达: %v", err)
		return status
	}

	client.Close()
	latency := time.Since(startTime).Milliseconds()

	status.Status = internal.StatusConnected
	status.Latency = int(latency)

	return status
}
