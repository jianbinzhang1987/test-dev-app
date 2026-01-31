package main

import (
	"context"
	"deploymaster-pro-wails/internal"
	"deploymaster-pro-wails/internal/credential"
	"deploymaster-pro-wails/internal/node"
	"deploymaster-pro-wails/internal/ssh"
	"deploymaster-pro-wails/internal/topology"
	"fmt"
	"log"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	nodeService     *node.Service
	sshTester       *ssh.Tester
	topologyService *topology.Service
	credStore       *credential.Store
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// 初始化节点服务
	dataDir, err := node.GetDefaultDataDir()
	if err != nil {
		log.Printf("Failed to get data directory: %v", err)
		return
	}

	storage, err := node.NewJSONStorage(dataDir)
	if err != nil {
		log.Printf("Failed to create storage: %v", err)
		return
	}

	a.nodeService, err = node.NewService(storage)
	if err != nil {
		log.Printf("Failed to create node service: %v", err)
		return
	}

	// 初始化凭据存储
	a.credStore = credential.NewStore(dataDir, storage.GetCrypto())

	// 初始化SSH测试器（传入凭据存储）
	a.sshTester = ssh.NewTester(a.credStore)

	// 初始化拓扑服务
	a.topologyService = topology.NewService()

	log.Println("Node topology service initialized successfully")
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// ===== 节点管理 API =====

// AddNode 添加新节点
func (a *App) AddNode(node internal.Node) error {
	if a.nodeService == nil {
		return fmt.Errorf("node service not initialized")
	}
	return a.nodeService.AddNode(&node)
}

// UpdateNode 更新节点信息
func (a *App) UpdateNode(node internal.Node) error {
	if a.nodeService == nil {
		return fmt.Errorf("node service not initialized")
	}
	return a.nodeService.UpdateNode(&node)
}

// DeleteNode 删除节点
func (a *App) DeleteNode(nodeID string) error {
	if a.nodeService == nil {
		return fmt.Errorf("node service not initialized")
	}
	// 先尝试删除凭据（忽略失败）
	if a.credStore != nil {
		_ = a.credStore.DeleteAll(nodeID, "")
	}
	return a.nodeService.DeleteNode(nodeID)
}

// GetNodes 获取所有节点列表
func (a *App) GetNodes() []*internal.Node {
	if a.nodeService == nil {
		return []*internal.Node{}
	}
	return a.nodeService.ListNodes()
}

// GetNode 获取单个节点
func (a *App) GetNode(nodeID string) (*internal.Node, error) {
	if a.nodeService == nil {
		return nil, fmt.Errorf("node service not initialized")
	}
	return a.nodeService.GetNode(nodeID)
}

// ===== 连接测试 API =====

// TestNodeConnection 测试单个节点连接
func (a *App) TestNodeConnection(nodeID, username, password string) (*internal.NodeStatus, error) {
	if a.nodeService == nil || a.sshTester == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	node, err := a.nodeService.GetNode(nodeID)
	if err != nil {
		return nil, err
	}

	status := a.sshTester.TestConnection(node, username, password)
	return status, nil
}

// BatchTestConnections 批量测试节点连接
func (a *App) BatchTestConnections(username, password string) map[string]*internal.NodeStatus {
	if a.nodeService == nil || a.sshTester == nil {
		return map[string]*internal.NodeStatus{}
	}

	nodes := a.nodeService.ListNodes()
	return a.sshTester.BatchTestConnections(nodes, username, password)
}

// ===== 拓扑数据 API =====

// GetTopology 获取拓扑结构数据
func (a *App) GetTopology() *internal.TopologyData {
	if a.nodeService == nil || a.topologyService == nil {
		return &internal.TopologyData{
			Total:  0,
			Slaves: []*internal.Node{},
		}
	}

	nodes := a.nodeService.ListNodes()
	return a.topologyService.GetTopologyData(nodes)
}

// ===== 凭据管理 API =====

// SaveCredential 保存节点SSH密码凭据
// rememberPassword: 是否记住密码（存储到系统密钥链）
func (a *App) SaveCredential(nodeID, username, password string, rememberPassword bool) error {
	if !rememberPassword {
		// 用户选择不记住密码，不存储
		return nil
	}

	if a.credStore == nil {
		return fmt.Errorf("credential store not initialized")
	}

	return a.credStore.SetPassword(nodeID, username, password)
}

// SaveKeyPassphrase 保存SSH密钥的密码短语
func (a *App) SaveKeyPassphrase(nodeID, passphrase string, rememberPassphrase bool) error {
	if !rememberPassphrase || passphrase == "" {
		// 不记住或密码短语为空
		return nil
	}

	if a.credStore == nil {
		return fmt.Errorf("credential store not initialized")
	}

	return a.credStore.SetKeyPassphrase(nodeID, passphrase)
}

// DeleteCredential 删除节点的所有凭据
// 在删除节点或用户主动清除凭据时调用
func (a *App) DeleteCredential(nodeID, username string) error {
	if a.credStore == nil {
		return fmt.Errorf("credential store not initialized")
	}

	return a.credStore.DeleteAll(nodeID, username)
}

// HasStoredCredential 检查节点是否已存储密码凭据
// 用于UI显示"记住的密码"状态
func (a *App) HasStoredCredential(nodeID, username string) bool {
	if a.credStore == nil || nodeID == "" || username == "" {
		return false
	}

	return a.credStore.HasPassword(nodeID, username)
}

// HasStoredKeyPassphrase 检查节点是否已存储密钥密码短语
func (a *App) HasStoredKeyPassphrase(nodeID string) bool {
	if a.credStore == nil || nodeID == "" {
		return false
	}

	return a.credStore.HasKeyPassphrase(nodeID)
}

// TestConnectionWithCredentials 使用提供的凭据测试连接
// saveCredentials: 是否保存凭据到系统密钥链
// 返回连接状态，用于在添加/编辑节点时验证凭据
func (a *App) TestConnectionWithCredentials(nodeID, username, password, keyPassphrase string, saveCredentials bool) (*internal.NodeStatus, error) {
	if a.nodeService == nil || a.sshTester == nil {
		return nil, fmt.Errorf("services not initialized")
	}

	node, err := a.nodeService.GetNode(nodeID)
	if err != nil {
		return nil, err
	}

	// 测试连接
	status := a.sshTester.TestConnectionWithCredentials(node, username, password, keyPassphrase)

	// 如果连接成功且用户选择保存凭据
	if status.Status == internal.StatusConnected && saveCredentials {
		switch node.AuthMethod {
		case internal.AuthMethodPassword:
			_ = a.credStore.SetPassword(nodeID, username, password)
		case internal.AuthMethodKey:
			if keyPassphrase != "" {
				_ = a.credStore.SetKeyPassphrase(nodeID, keyPassphrase)
			}
		}
	}

	return status, nil
}

// SelectKeyFile 弹出文件选择对话框让用户选择 SSH 私钥文件
func (a *App) SelectKeyFile() (string, error) {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择 SSH 私钥文件",
		Filters: []runtime.FileFilter{
			{
				DisplayName: "所有文件 (*.*)",
				Pattern:     "*.*",
			},
		},
	})
	if err != nil {
		return "", err
	}
	return selection, nil
}
