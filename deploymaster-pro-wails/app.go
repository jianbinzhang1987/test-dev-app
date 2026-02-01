package main

import (
	"bytes"
	"context"
	"deploymaster-pro-wails/internal"
	"deploymaster-pro-wails/internal/credential"
	"deploymaster-pro-wails/internal/node"
	"deploymaster-pro-wails/internal/ssh"
	"deploymaster-pro-wails/internal/svn"
	"deploymaster-pro-wails/internal/syncd"
	"deploymaster-pro-wails/internal/task"
	"deploymaster-pro-wails/internal/topology"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"al.essio.dev/pkg/shellescape"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	nodeService     *node.Service
	sshTester       *ssh.Tester
	topologyService *topology.Service
	credStore       *credential.Store
	svnService      *svn.Service
	svnClient       *svn.Client
	taskService     *task.Service
	dataDir         string
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
	a.dataDir = dataDir

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

	// 初始化 SVN 资源服务
	svnStorage, err := svn.NewJSONStorage(dataDir)
	if err != nil {
		log.Printf("Failed to create SVN storage: %v", err)
		return
	}
	a.svnService, err = svn.NewService(svnStorage)
	if err != nil {
		log.Printf("Failed to create SVN service: %v", err)
		return
	}
	a.svnClient = svn.NewClient(10 * time.Second)

	// 初始化任务编排服务
	taskStorage, err := task.NewJSONStorage(dataDir)
	if err != nil {
		log.Printf("Failed to create task storage: %v", err)
		return
	}
	a.taskService, err = task.NewService(taskStorage)
	if err != nil {
		log.Printf("Failed to create task service: %v", err)
		return
	}

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

// GetCredential 获取已保存的 SSH 密码明文
// 注意：仅用于用户显式查看时调用
func (a *App) GetCredential(nodeID, username string) (string, error) {
	if a.credStore == nil {
		return "", fmt.Errorf("credential store not initialized")
	}
	if nodeID == "" || username == "" {
		return "", fmt.Errorf("invalid credential key")
	}
	return a.credStore.GetPassword(nodeID, username)
}

// HasStoredKeyPassphrase 检查节点是否已存储密钥密码短语
func (a *App) HasStoredKeyPassphrase(nodeID string) bool {
	if a.credStore == nil || nodeID == "" {
		return false
	}

	return a.credStore.HasKeyPassphrase(nodeID)
}

// ===== SVN 资源管理 API =====

// GetSVNResources 获取所有 SVN 资源
func (a *App) GetSVNResources() []*internal.SVNResource {
	if a.svnService == nil {
		return []*internal.SVNResource{}
	}
	return a.svnService.ListResources()
}

// AddSVNResource 添加 SVN 资源
func (a *App) AddSVNResource(resource internal.SVNResource) (*internal.SVNResource, error) {
	if a.svnService == nil {
		return nil, fmt.Errorf("svn service not initialized")
	}
	if err := a.svnService.AddResource(&resource); err != nil {
		return nil, err
	}
	return &resource, nil
}

// UpdateSVNResource 更新 SVN 资源
func (a *App) UpdateSVNResource(resource internal.SVNResource) error {
	if a.svnService == nil {
		return fmt.Errorf("svn service not initialized")
	}
	return a.svnService.UpdateResource(&resource)
}

// DeleteSVNResource 删除 SVN 资源
func (a *App) DeleteSVNResource(resourceID string) error {
	if a.svnService == nil {
		return fmt.Errorf("svn service not initialized")
	}
	if a.credStore != nil {
		if res, err := a.svnService.GetResource(resourceID); err == nil {
			_ = a.credStore.DeleteSVNPassword(resourceID, res.Username)
		}
	}
	return a.svnService.DeleteResource(resourceID)
}

// SaveSVNCredential 保存 SVN 凭据
func (a *App) SaveSVNCredential(resourceID, username, password string, remember bool) error {
	if !remember || password == "" {
		return nil
	}
	if a.credStore == nil {
		return fmt.Errorf("credential store not initialized")
	}
	return a.credStore.SetSVNPassword(resourceID, username, password)
}

// HasStoredSVNCredential 检查 SVN 凭据是否存在
func (a *App) HasStoredSVNCredential(resourceID, username string) bool {
	if a.credStore == nil || resourceID == "" || username == "" {
		return false
	}
	return a.credStore.HasSVNPassword(resourceID, username)
}

// TestSVNConnection SVN 连接测试（仅检测，不更新资源）
func (a *App) TestSVNConnection(url, username, password, resourceID string) (*internal.SVNTestResult, error) {
	if a.svnClient == nil {
		return nil, fmt.Errorf("svn client not initialized")
	}
	if err := a.svnClient.CheckAvailable(); err != nil {
		_, _ = runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
			Title:   "SVN 客户端未安装",
			Message: "未检测到 svn 命令行客户端。请先安装 SVN（如：xcode-select --install 或 brew install svn）。",
			Type:    runtime.ErrorDialog,
		})
		return nil, err
	}
	if password == "" && a.credStore != nil && resourceID != "" && username != "" {
		if stored, err := a.credStore.GetSVNPassword(resourceID, username); err == nil {
			password = stored
		}
	}

	start := time.Now()
	rev, err := a.svnClient.Info(a.ctx, url, username, password)
	result := &internal.SVNTestResult{
		CheckedAt:  time.Now().Format(time.RFC3339),
		DurationMs: int(time.Since(start).Milliseconds()),
	}
	if err != nil {
		result.Ok = false
		result.Message = err.Error()
		return result, nil
	}
	result.Ok = true
	result.Revision = rev
	result.Message = "SVN 连接正常"
	return result, nil
}

// RefreshSVNResource 刷新资源修订号与状态
func (a *App) RefreshSVNResource(resourceID string) (*internal.SVNResource, error) {
	if a.svnService == nil || a.svnClient == nil {
		return nil, fmt.Errorf("svn service not initialized")
	}
	res, err := a.svnService.GetResource(resourceID)
	if err != nil {
		return nil, err
	}

	password := ""
	if a.credStore != nil && res.Username != "" {
		if stored, err := a.credStore.GetSVNPassword(resourceID, res.Username); err == nil {
			password = stored
		}
	}

	rev, err := a.svnClient.Info(a.ctx, res.URL, res.Username, password)
	if err != nil {
		res.Status = internal.SVNStatusError
		res.LastChecked = time.Now().Format("2006-01-02 15:04")
		_ = a.svnService.UpdateResource(res)
		return res, nil
	}

	res.Revision = rev
	res.Status = internal.SVNStatusOnline
	res.LastChecked = time.Now().Format("2006-01-02 15:04")
	if err := a.svnService.UpdateResource(res); err != nil {
		return nil, err
	}
	return res, nil
}

// ===== 任务编排 API =====

// GetTasks 获取任务列表
func (a *App) GetTasks() []*internal.TaskDefinition {
	if a.taskService == nil {
		return []*internal.TaskDefinition{}
	}
	return a.taskService.ListTasks()
}

// AddTask 添加任务
func (a *App) AddTask(task internal.TaskDefinition) (*internal.TaskDefinition, error) {
	if a.taskService == nil {
		return nil, fmt.Errorf("task service not initialized")
	}
	return a.taskService.AddTask(&task)
}

// UpdateTask 更新任务
func (a *App) UpdateTask(task internal.TaskDefinition) error {
	if a.taskService == nil {
		return fmt.Errorf("task service not initialized")
	}
	return a.taskService.UpdateTask(&task)
}

// DeleteTask 删除任务
func (a *App) DeleteTask(taskID string) error {
	if a.taskService == nil {
		return fmt.Errorf("task service not initialized")
	}
	return a.taskService.DeleteTask(taskID)
}

// GetTaskTemplates 获取模板列表
func (a *App) GetTaskTemplates() []*internal.TaskTemplate {
	if a.taskService == nil {
		return []*internal.TaskTemplate{}
	}
	return a.taskService.ListTemplates()
}

// AddTaskTemplate 添加模板
func (a *App) AddTaskTemplate(tpl internal.TaskTemplate) (*internal.TaskTemplate, error) {
	if a.taskService == nil {
		return nil, fmt.Errorf("task service not initialized")
	}
	return a.taskService.AddTemplate(&tpl)
}

// UpdateTaskTemplate 更新模板
func (a *App) UpdateTaskTemplate(tpl internal.TaskTemplate) error {
	if a.taskService == nil {
		return fmt.Errorf("task service not initialized")
	}
	return a.taskService.UpdateTemplate(&tpl)
}

// DeleteTaskTemplate 删除模板
func (a *App) DeleteTaskTemplate(templateID string) error {
	if a.taskService == nil {
		return fmt.Errorf("task service not initialized")
	}
	return a.taskService.DeleteTemplate(templateID)
}

// GetTaskRuns 获取所有任务运行历史
func (a *App) GetTaskRuns() []*internal.TaskRun {
	if a.taskService == nil {
		return []*internal.TaskRun{}
	}
	return a.taskService.ListRuns()
}

// GetTaskRunsByTask 获取指定任务的运行历史
func (a *App) GetTaskRunsByTask(taskID string) []*internal.TaskRun {
	if a.taskService == nil {
		return []*internal.TaskRun{}
	}
	return a.taskService.ListRunsByTask(taskID)
}

// DeleteTaskRun 删除运行记录
func (a *App) DeleteTaskRun(runID string) error {
	if a.taskService == nil {
		return fmt.Errorf("task service not initialized")
	}
	return a.taskService.DeleteRun(runID)
}

// DeleteTaskRunsByTask 清空指定任务的运行历史
func (a *App) DeleteTaskRunsByTask(taskID string) error {
	if a.taskService == nil {
		return fmt.Errorf("task service not initialized")
	}
	return a.taskService.DeleteRunsByTask(taskID)
}

// CheckoutSVNResource 导出 SVN 资源到本地目录
// targetDir 为空时默认存储到 dataDir/svn-cache/<resourceID>
func (a *App) CheckoutSVNResource(resourceID, targetDir string) (string, error) {
	if a.svnService == nil || a.svnClient == nil {
		return "", fmt.Errorf("svn service not initialized")
	}
	if a.dataDir == "" {
		return "", fmt.Errorf("data directory not initialized")
	}

	res, err := a.svnService.GetResource(resourceID)
	if err != nil {
		return "", err
	}

	if targetDir == "" {
		targetDir = filepath.Join(a.dataDir, "svn-cache", resourceID)
	}

	password := ""
	if a.credStore != nil && res.Username != "" {
		if stored, err := a.credStore.GetSVNPassword(resourceID, res.Username); err == nil {
			password = stored
		}
	}

	exportDest := targetDir
	baseName := path.Base(strings.TrimRight(res.URL, "/"))
	if res.Type == internal.SVNResourceFile {
		if baseName == "" || baseName == "." || baseName == "/" {
			baseName = "package.bin"
		}
		if info, err := os.Stat(targetDir); err == nil && info.IsDir() {
			exportDest = filepath.Join(targetDir, baseName)
		}
		if targetDir == "" {
			exportDest = filepath.Join(a.dataDir, "svn-cache", resourceID, baseName)
		}
	} else {
		if baseName != "" && baseName != "." && baseName != "/" {
			exportDest = filepath.Join(targetDir, baseName)
		}
		if err := os.MkdirAll(exportDest, 0755); err != nil {
			return "", err
		}
	}

	res.Status = internal.SVNStatusSyncing
	res.LastChecked = time.Now().Format("2006-01-02 15:04")
	_ = a.svnService.UpdateResource(res)

	if res.Type == internal.SVNResourceFile {
		if err := a.svnClient.CatToFile(a.ctx, res.URL, res.Username, password, "", exportDest); err != nil {
			res.Status = internal.SVNStatusError
			res.LastChecked = time.Now().Format("2006-01-02 15:04")
			_ = a.svnService.UpdateResource(res)
			return "", err
		}
	} else if err := a.svnClient.Export(a.ctx, res.URL, res.Username, password, "", exportDest); err != nil {
		res.Status = internal.SVNStatusError
		res.LastChecked = time.Now().Format("2006-01-02 15:04")
		_ = a.svnService.UpdateResource(res)
		return "", err
	}

	res.Status = internal.SVNStatusOnline
	res.LastChecked = time.Now().Format("2006-01-02 15:04")
	if err := a.svnService.UpdateResource(res); err != nil {
		return "", err
	}

	return exportDest, nil
}

// ExecuteTask 执行任务流水线（下载->上传->同步->执行）
// 通过事件推送任务进度与日志：task:event
func (a *App) ExecuteTask(req internal.TaskRunRequest) error {
	if a.svnService == nil || a.svnClient == nil || a.nodeService == nil {
		return fmt.Errorf("services not initialized")
	}
	if req.TaskID == "" {
		return fmt.Errorf("taskId is required")
	}

	go a.runTask(req)
	return nil
}

func (a *App) runTask(req internal.TaskRunRequest) {
	taskName := req.TaskName
	if taskName == "" && a.taskService != nil {
		if task, err := a.taskService.GetTask(req.TaskID); err == nil {
			taskName = task.Name
		}
	}
	if taskName == "" {
		taskName = req.TaskID
	}

	runID := ""
	if a.taskService != nil {
		if run, err := a.taskService.CreateRun(req.TaskID, taskName); err == nil && run != nil {
			runID = run.ID
		}
	}

	emit := func(status internal.TaskStatus, progress int, logLine string) {
		logWithTime := fmt.Sprintf("[%s] %s", time.Now().Format("2006-01-02 15:04:05"), logLine)
		runtime.EventsEmit(a.ctx, "task:event", internal.TaskEvent{
			TaskID:   req.TaskID,
			RunID:    runID,
			Status:   status,
			Progress: progress,
			Log:      logWithTime,
		})
		if a.taskService != nil {
			_ = a.taskService.UpdateTaskState(req.TaskID, status, progress)
			if runID != "" {
				_ = a.taskService.AppendRunLog(runID, status, progress, logWithTime)
			}
		}
	}

	emit(internal.TaskStatusDownloading, 5, "[信息] 启动自动化分发流水线...")

	resource, err := a.svnService.GetResource(req.SVNResourceID)
	if err != nil {
		emit(internal.TaskStatusFailed, 5, "[错误] 未找到 SVN 资源，任务终止。")
		return
	}

	cacheDir := filepath.Join(a.dataDir, "svn-cache", req.SVNResourceID)
	emit(internal.TaskStatusDownloading, 15, "正在建立 SVN 连接，准备拉取最新内容 (HEAD) ...")

	password := ""
	if a.credStore != nil && resource.Username != "" {
		if stored, err := a.credStore.GetSVNPassword(req.SVNResourceID, resource.Username); err == nil {
			password = stored
		}
	}

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		emit(internal.TaskStatusFailed, 15, fmt.Sprintf("[错误] 创建缓存目录失败：%v", err))
		return
	}

	exportDest := cacheDir
	baseName := path.Base(strings.TrimRight(resource.URL, "/"))
	if resource.Type == internal.SVNResourceFile {
		if baseName == "" || baseName == "." || baseName == "/" {
			baseName = "package.bin"
		}
		exportDest = filepath.Join(cacheDir, baseName)
	} else if baseName != "" && baseName != "." && baseName != "/" {
		exportDest = filepath.Join(cacheDir, baseName)
	}

	if resource.Type == internal.SVNResourceFile {
		if err := a.svnClient.CatToFile(a.ctx, resource.URL, resource.Username, password, "", exportDest); err != nil {
			emit(internal.TaskStatusFailed, 15, fmt.Sprintf("[错误] SVN 检出失败：%v", err))
			return
		}
	} else if err := a.svnClient.Export(a.ctx, resource.URL, resource.Username, password, "", exportDest); err != nil {
		emit(internal.TaskStatusFailed, 15, fmt.Sprintf("[错误] SVN 检出失败：%v", err))
		return
	}
	emit(internal.TaskStatusDownloading, 30, fmt.Sprintf("SVN 资源检出完成。缓存路径: %s", exportDest))

	master, err := a.nodeService.GetNode(req.MasterServerID)
	if err != nil {
		emit(internal.TaskStatusFailed, 30, "[错误] 未找到主控节点，任务终止。")
		return
	}

	remoteTarget := req.RemotePath
	if strings.TrimSpace(remoteTarget) == "" {
		remoteTarget = "/tmp/deploymaster"
	}
	if resource.Type == internal.SVNResourceFile {
		baseName = path.Base(strings.TrimRight(resource.URL, "/"))
		if baseName == "" || baseName == "." || baseName == "/" {
			baseName = "package.bin"
		}
		remoteTarget = filepath.ToSlash(filepath.Join(remoteTarget, baseName))
	}

	emit(internal.TaskStatusUploading, 45, fmt.Sprintf("正在通过 %s 上传资源至主控机: %s", master.Protocol, remoteTarget))
	if err := a.uploadToNode(master, exportDest, remoteTarget); err != nil {
		emit(internal.TaskStatusFailed, 45, fmt.Sprintf("[错误] 上传至主控机失败：%v", err))
		return
	}
	emit(internal.TaskStatusUploading, 55, fmt.Sprintf("主控机资源上传完成：%s", remoteTarget))

	slaveTargetBase := req.SlaveRemotePath
	if strings.TrimSpace(slaveTargetBase) == "" {
		slaveTargetBase = req.RemotePath
	}
	if strings.TrimSpace(slaveTargetBase) == "" {
		slaveTargetBase = "/tmp/deploymaster"
	}

	emit(internal.TaskStatusSyncing, 65, fmt.Sprintf("主控机开始同步 %d 台从机...", len(req.SlaveServerIDs)))
	emit(internal.TaskStatusSyncing, 68, "准备主控机临时同步服务 /tmp/deploymaster-syncd（自动校验版本，必要时覆盖上传）")
	syncdLogs, err := a.syncFromMaster(master, req.SlaveServerIDs, remoteTarget, slaveTargetBase, req.SlaveRemotePaths, resource.Type == internal.SVNResourceFile, baseName)
	if err != nil {
		msg := err.Error()
		if strings.Contains(strings.ToLower(msg), "permission denied") {
			msg = msg + "（请检查从机目标目录权限，或改用可写目录如 /tmp）"
		}
		if strings.HasPrefix(msg, "从机同步失败：") {
			emit(internal.TaskStatusFailed, 65, fmt.Sprintf("[错误] %s", msg))
		} else {
			emit(internal.TaskStatusFailed, 65, fmt.Sprintf("[错误] 从机同步失败：%s", msg))
		}
		return
	}
	if len(syncdLogs) > 0 {
		progressSteps := []int{69, 70, 71, 72, 73, 74}
		for i, line := range syncdLogs {
			p := 70
			if i < len(progressSteps) {
				p = progressSteps[i]
			}
			emit(internal.TaskStatusSyncing, p, line)
		}
	}
	emit(internal.TaskStatusSyncing, 75, "临时同步服务执行完成，已清理 /tmp/deploymaster-syncd")
	emit(internal.TaskStatusSyncing, 77, "主控机同步从机完成。")

	emit(internal.TaskStatusExecuting, 85, "正在启动远程自定义脚本执行序列...")
	if err := a.executeCommandsOnNodes(req.Commands, req.MasterServerID, req.SlaveServerIDs); err != nil {
		emit(internal.TaskStatusFailed, 85, fmt.Sprintf("[错误] 远程脚本执行失败：%v", err))
		return
	}

	emit(internal.TaskStatusSuccess, 100, "✓ 任务执行成功。所有节点已同步至最新状态。")
}

func (a *App) uploadToNode(node *internal.Node, localPath, remotePath string) error {
	client, err := a.createSSHClient(node)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(node.IP, node.Port); err != nil {
		return err
	}

	sftpClient, err := client.NewSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	remote := remotePath
	if strings.TrimSpace(remote) == "" {
		remote = "/tmp/deploymaster"
	}
	return ssh.UploadPath(sftpClient, localPath, remote)
}

func (a *App) executeCommandsOnNodes(commands []string, masterID string, slaveIDs []string) error {
	if len(commands) == 0 {
		return nil
	}

	ids := append([]string{masterID}, slaveIDs...)
	for _, id := range ids {
		node, err := a.nodeService.GetNode(id)
		if err != nil {
			return err
		}
		if err := a.executeCommandsOnNode(node, commands); err != nil {
			return err
		}
	}
	return nil
}

type syncdSlave struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	User       string `json:"user"`
	Password   string `json:"password"`
	RemotePath string `json:"remotePath"`
}

type syncdPayload struct {
	Version    string       `json:"version"`
	Checksum   string       `json:"checksum,omitempty"`
	BinarySize int          `json:"binarySize,omitempty"`
	SourcePath string       `json:"sourcePath"`
	RemotePath string       `json:"remotePath"`
	Slaves     []syncdSlave `json:"slaves"`
}

func (a *App) ensureSyncdOnMaster(client *ssh.Client, remotePath string) (string, string, bool, int, string, error) {
	arch := "amd64"
	osName := "unknown"
	if output, err := client.ExecuteCommand("uname -s"); err == nil {
		osName = strings.TrimSpace(strings.ToLower(output))
		switch osName {
		case "linux", "darwin":
		default:
			return "", osName, false, 0, "", fmt.Errorf("主控机系统暂不支持同步服务：仅支持 Linux/macOS")
		}
	}
	if output, err := client.ExecuteCommand("uname -m"); err == nil {
		rawArch := strings.TrimSpace(strings.ToLower(output))
		switch rawArch {
		case "x86_64", "amd64":
			arch = "amd64"
		case "aarch64", "arm64":
			arch = "arm64"
		default:
			return "", osName, false, 0, "", fmt.Errorf("主控机架构暂不支持同步服务：仅支持 amd64/arm64")
		}
	}

	output, err := client.ExecuteCommand(remotePath + " --version")
	if err == nil && strings.TrimSpace(output) == syncd.Version {
		return arch, osName, false, 0, "", nil
	}

	var bin []byte
	if osName == "darwin" {
		if arch == "arm64" {
			bin = syncd.GetDarwinARM64()
		} else {
			bin = syncd.GetDarwinAMD64()
		}
	} else {
		if arch == "arm64" {
			bin = syncd.GetLinuxARM64()
		} else {
			bin = syncd.GetLinuxAMD64()
		}
	}
	if len(bin) == 0 {
		return "", osName, false, 0, "", fmt.Errorf("syncd binary not embedded")
	}

	sftpClient, err := client.NewSFTPClient()
	if err != nil {
		return "", osName, false, 0, "", err
	}
	defer sftpClient.Close()

	dst, err := sftpClient.Create(remotePath)
	if err != nil {
		return "", osName, false, 0, "", err
	}
	if _, err := io.Copy(dst, bytes.NewReader(bin)); err != nil {
		_ = dst.Close()
		return "", osName, false, 0, "", err
	}
	if err := dst.Close(); err != nil {
		return "", osName, false, 0, "", err
	}

	if _, err := client.ExecuteCommand("chmod +x " + shellescape.Quote(remotePath)); err != nil {
		return "", osName, false, 0, "", fmt.Errorf("chmod syncd failed: %w", err)
	}

	checksum := fmt.Sprintf("%08x", crc32.ChecksumIEEE(bin))
	return arch, osName, true, len(bin), checksum, nil
}

func (a *App) syncFromMaster(master *internal.Node, slaveIDs []string, remotePath string, slaveRemotePath string, slaveRemotePaths map[string]string, isFile bool, baseName string) ([]string, error) {
	if len(slaveIDs) == 0 {
		return []string{}, nil
	}

	client, err := a.createSSHClient(master)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	if err := client.Connect(master.IP, master.Port); err != nil {
		return nil, err
	}

	syncdPath := "/tmp/deploymaster-syncd"
	arch, osName, updated, binSize, checksum, err := a.ensureSyncdOnMaster(client, syncdPath)
	if err != nil {
		return nil, fmt.Errorf("部署同步服务失败：%v", err)
	}
	defer func() {
		_, _ = client.ExecuteCommand("rm -f " + shellescape.Quote(syncdPath))
	}()
	logs := make([]string, 0, 16)
	logs = append(logs, fmt.Sprintf("同步服务路径：%s", syncdPath))
	osArchNote := fmt.Sprintf("主控机系统检测：%s/%s", osName, arch)
	var syncdNote string
	if updated {
		syncdNote = fmt.Sprintf("同步服务已更新：%s (version=%s, arch=%s)", syncdPath, syncd.Version, arch)
	} else {
		syncdNote = fmt.Sprintf("同步服务已就绪：%s (version=%s, arch=%s)", syncdPath, syncd.Version, arch)
	}
	logs = append(logs, osArchNote)
	logs = append(logs, syncdNote)
	if binSize > 0 && checksum != "" {
		logs = append(logs, fmt.Sprintf("同步服务校验：size=%dB crc32=%s", binSize, checksum))
	}

	if output, err := client.ExecuteCommand("df -k /tmp | tail -n +2 | awk '{print $4\"K\"\"/\"$2\"K\"\"(\"$5\" used)\"}'"); err == nil {
		info := strings.TrimSpace(output)
		if info != "" {
			logs = append(logs, fmt.Sprintf("/tmp 磁盘占用：%s", info))
		}
	}

	src := remotePath
	if strings.TrimSpace(src) == "" {
		src = "/tmp/deploymaster"
	}
	src = filepath.ToSlash(src)

	slaves := make([]syncdSlave, 0, len(slaveIDs))
	slaveNames := make([]string, 0, len(slaveIDs))
	dest := slaveRemotePath
	if strings.TrimSpace(dest) == "" {
		dest = remotePath
	}
	if strings.TrimSpace(dest) == "" {
		dest = "/tmp/deploymaster"
	}
	if isFile && baseName != "" {
		dest = filepath.ToSlash(filepath.Join(dest, baseName))
	} else {
		dest = filepath.ToSlash(dest)
	}

	for _, slaveID := range slaveIDs {
		slave, err := a.nodeService.GetNode(slaveID)
		if err != nil {
			return nil, err
		}
		user := slave.Username
		if strings.TrimSpace(user) == "" {
			user = "root"
		}

		if slave.AuthMethod != internal.AuthMethodPassword {
			return nil, fmt.Errorf("从机同步失败：主控机同步服务仅支持密码认证，从机 %s 请改为密码认证或改用客户端直传模式", slave.Name)
		}

		password := ""
		if a.credStore != nil {
			if stored, err := a.credStore.GetPassword(slave.ID, user); err == nil {
				password = stored
			}
		}
		if strings.TrimSpace(password) == "" {
			return nil, fmt.Errorf("从机同步失败：未找到从机 %s 的密码，请先保存密码", slave.Name)
		}

		slaveDest := dest
		if len(slaveRemotePaths) > 0 {
			if custom, ok := slaveRemotePaths[slaveID]; ok && strings.TrimSpace(custom) != "" {
				if isFile && baseName != "" {
					slaveDest = filepath.ToSlash(filepath.Join(custom, baseName))
				} else {
					slaveDest = filepath.ToSlash(custom)
				}
			}
		}

		slaves = append(slaves, syncdSlave{
			ID:         slave.ID,
			Name:       slave.Name,
			Host:       slave.IP,
			Port:       slave.Port,
			User:       user,
			Password:   password,
			RemotePath: slaveDest,
		})
		name := slave.Name
		if strings.TrimSpace(name) == "" {
			name = slave.IP
		}
		slaveNames = append(slaveNames, name)
	}

	sort.Strings(slaveNames)
	logs = append(logs, fmt.Sprintf("同步目标从机：%s", strings.Join(slaveNames, ", ")))

	payload := syncdPayload{
		Version:    syncd.Version,
		Checksum:   checksum,
		BinarySize: binSize,
		SourcePath: src,
		RemotePath: dest,
		Slaves:     slaves,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	payloadB64 := base64.StdEncoding.EncodeToString(raw)
	timeoutSeconds := int((time.Duration(len(slaves)) * 120 * time.Second).Seconds())
	cmd := fmt.Sprintf("%s --payload %s", shellescape.Quote(syncdPath), shellescape.Quote(payloadB64))
	if _, err := client.ExecuteCommand("command -v timeout"); err == nil {
		cmd = fmt.Sprintf("timeout %ds %s --payload %s", timeoutSeconds, shellescape.Quote(syncdPath), shellescape.Quote(payloadB64))
	} else {
		logs = append(logs, "注意：主控机未安装 timeout，无法设置同步超时保护")
	}
	logs = append(logs, fmt.Sprintf("同步执行开始：%s", time.Now().Format("2006-01-02 15:04:05")))
	if output, err := client.ExecuteCommand(cmd); err != nil {
		msg := strings.TrimSpace(output)
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("从机同步失败：%s", msg)
	}

	logs = append(logs, fmt.Sprintf("同步执行结束：%s", time.Now().Format("2006-01-02 15:04:05")))
	logs = append(logs, fmt.Sprintf("同步耗时预估：%ds（按 %d 台从机计算）", timeoutSeconds, len(slaves)))
	return logs, nil
}

func (a *App) executeCommandsOnNode(node *internal.Node, commands []string) error {
	client, err := a.createSSHClient(node)
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Connect(node.IP, node.Port); err != nil {
		return err
	}

	for _, cmd := range commands {
		if strings.TrimSpace(cmd) == "" {
			continue
		}
		if _, err := client.ExecuteCommand(cmd); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) createSSHClient(node *internal.Node) (*ssh.Client, error) {
	username := node.Username
	if strings.TrimSpace(username) == "" {
		username = "root"
	}

	switch node.AuthMethod {
	case internal.AuthMethodKey:
		passphrase := ""
		if a.credStore != nil {
			if stored, err := a.credStore.GetKeyPassphrase(node.ID); err == nil {
				passphrase = stored
			}
		}
		return ssh.NewClientWithKeyFile(username, node.KeyPath, passphrase)
	case internal.AuthMethodAgent:
		return ssh.NewClientWithAgent(username)
	default:
		password := ""
		if a.credStore != nil {
			if stored, err := a.credStore.GetPassword(node.ID, username); err == nil {
				password = stored
			}
		}
		if strings.TrimSpace(password) == "" {
			return nil, fmt.Errorf("missing password for node %s", node.Name)
		}
		return ssh.NewClient(username, password), nil
	}
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

// ===== 通用弹窗 API =====

// ShowMessageDialog 显示系统消息弹窗
// dialogType: info | warning | error | question
func (a *App) ShowMessageDialog(title, message, dialogType string) error {
	if a.ctx == nil {
		return fmt.Errorf("app context not initialized")
	}
	_, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Title:   title,
		Message: message,
		Type:    parseDialogType(dialogType),
	})
	return err
}

// ConfirmDialog 显示确认弹窗，返回用户是否确认
func (a *App) ConfirmDialog(title, message string) (bool, error) {
	if a.ctx == nil {
		return false, fmt.Errorf("app context not initialized")
	}
	result, err := runtime.MessageDialog(a.ctx, runtime.MessageDialogOptions{
		Title:         title,
		Message:       message,
		Type:          runtime.QuestionDialog,
		Buttons:       []string{"取消", "确定"},
		DefaultButton: "确定",
		CancelButton:  "取消",
	})
	if err != nil {
		return false, err
	}
	return result == "确定", nil
}

func parseDialogType(dialogType string) runtime.DialogType {
	switch strings.ToLower(strings.TrimSpace(dialogType)) {
	case "warning":
		return runtime.WarningDialog
	case "error":
		return runtime.ErrorDialog
	case "question":
		return runtime.QuestionDialog
	default:
		return runtime.InfoDialog
	}
}
