# 节点拓扑功能使用指南

## 快速开始

### 1. 启动开发服务器

```bash
cd /Users/adolf/Desktop/code/test-dev-app/deploymaster-pro-wails
wails dev
```

### 2. 前端集成

当前前端页面 `ServerManager.vue` 使用的是模拟数据。要集成真实后端API，需要修改以下部分：

#### 导入Wails绑定

```typescript
// 在 ServerManager.vue 顶部添加
import { AddNode, UpdateNode, DeleteNode, GetNodes, TestNodeConnection } from '../../wailsjs/go/main/App';
import type { internal } from '../../wailsjs/go/models';
```

#### 替换事件发射为直接API调用

```typescript
// 原代码: emit('add', newSrv);
// 改为:
const newNode: internal.Node = {
  id: 'srv-' + Math.random().toString(36).substring(2, 9),
  name: currentSrv.value.name,
  ip: currentSrv.value.ip,
  port: currentSrv.value.port,
  protocol: currentSrv.value.protocol,
  isMaster: currentSrv.value.isMaster
};
await AddNode(newNode);
await loadServers(); // 刷新列表
```

#### 实现真实的连接测试

```typescript
const handleSinglePing = async (server: RemoteServer) => {
  testingId.value = server.id;
  
  try {
    // 弹窗或从配置获取SSH凭据
    const username = 'your-username';
    const password = 'your-password';
    
    const status = await TestNodeConnection(server.id, username, password);
    
    // 更新节点延迟信息
    emit('update', {
      ...server,
      latency: status.latency,
      lastChecked: new Date(status.lastChecked).toLocaleTimeString(),
    });
  } catch (error) {
    console.error('连接测试失败:', error);
  } finally {
    testingId.value = null;
  }
};
```

## API参考

### 节点管理

#### `AddNode(node: internal.Node): Promise<void>`
添加新节点到系统

**参数：**
```typescript
{
  id: string;           // 节点唯一ID
  name: string;        // 节点名称
  ip: string;          // IP地址
  port: number;        // 端口号
  protocol: Protocol;  // 'SFTP' | 'SCP' | 'FTP'
  isMaster: boolean;   // 是否为主节点
}
```

#### `UpdateNode(node: internal.Node): Promise<void>`
更新已存在的节点信息

#### `DeleteNode(nodeID: string): Promise<void>`
删除指定节点

#### `GetNodes(): Promise<internal.Node[]>`
获取所有节点列表

#### `GetNode(nodeID: string): Promise<internal.Node>`
获取单个节点详情

### 连接测试

#### `TestNodeConnection(nodeID: string, username: string, password: string): Promise<internal.NodeStatus>`
测试单个节点SSH连接

**返回：**
```typescript
{
  latency: number;           // 延迟（毫秒）
  lastChecked: string;       // ISO时间戳
  status: ConnectionStatus;  // 'connected' | 'disconnected' | 'testing' | 'error'
  errorMsg: string;         // 错误信息（如有）
}
```

#### `BatchTestConnections(username: string, password: string): Promise<Record<string, internal.NodeStatus>>`
批量测试所有节点连接

### 拓扑数据

#### `GetTopology(): Promise<internal.TopologyData>`
获取拓扑结构数据（用于可视化图表）

**返回：**
```typescript
{
  master: Node | null;  // 主节点
  slaves: Node[];       // 从节点列表
  total: number;        // 总节点数
}
```

## 数据存储

节点配置自动保存在：
- **macOS/Linux**: `~/.deploymaster/nodes.json`
- **Windows**: `%USERPROFILE%\.deploymaster\nodes.json`

## 运行测试

```bash
# 单元测试
go test -v ./internal/node/...

# 所有测试
go test -v ./...

# 带覆盖率
go test -v -cover ./internal/...
```

## 故障排查

### 问题：节点数据不保存
- 检查 `~/.deploymaster/` 目录权限
- 查看控制台日志中的错误信息

### 问题：SSH连接失败
- 确认目标服务器SSH服务已启动
- 验证防火墙规则允许22端口
- 检查用户名密码是否正确

### 问题：Wails绑定未生成
- 确保运行了 `wails dev` 或 `wails build`
- 检查 `frontend/wailsjs/go/main/` 目录

## 开发建议

1. **凭据管理**：不要在代码中硬编码密码，考虑：
   - 使用环境变量
   - 集成系统密钥链
   - 每次测试时提示用户输入

2. **错误处理**：所有API调用都应包裹在 try-catch 中

3. **加载状态**：在API调用期间显示加载指示器

4. **实时更新**：可使用Wails Events机制推送节点状态变化
