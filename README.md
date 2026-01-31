# DeployMaster Pro

跨平台自动化部署客户端的代码与文档集合，涵盖 React 高保真原型和 Go/Wails 桌面端实现。目标：从 SVN 拉取发布包，经 Master/Slave 拓扑并行分发至多节点，并执行远程脚本，提升企业级发布效率与可观测性。

## 仓库结构
- `deploymaster-pro/`：React + TypeScript + Vite 的 Web 高保真原型（静态模拟数据）。
- `deploymaster-pro-wails/`：Go 1.21 + Wails v2 + Vue 3 的桌面端实现，含节点管理与 SSH 连通性测试。
- `assets/`：设计/素材占位目录。
- `需求文档.md`：产品需求规格说明书（PRD）。
- `架构文档.md`：技术架构设计与分层说明。

## 核心能力概览
- SVN 资源管理：版本锁定、连通性测试、文件/目录检出（计划接入真实驱动）。
- 节点拓扑：Master/Slave 模型，多协议（SFTP/SCP/FTP）配置与状态监控。
- 任务编排：下载 → 上传 Master → 同步 Slaves → 远程脚本执行的流水线，状态机驱动。
- 实时观测：进度条、流式日志、错误中断与审计侧边栏（前端已实现 UI）。

## 快速开始
### React Web 原型 (`deploymaster-pro`)
1) 安装依赖：`cd deploymaster-pro && npm install`
2) 本地启动：`npm run dev`（默认 http://localhost:5173）
3) 生产构建：`npm run build`；预览产物：`npm run preview`
> 此版本使用前端模拟数据，适合演示交互与样式。

### Wails 桌面端 (`deploymaster-pro-wails`)
前置：Go 1.21+、Node 18+、`wails` CLI 已安装 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`).
1) 安装前端依赖：`cd deploymaster-pro-wails/frontend && npm install`
2) 开发模式：回到 `deploymaster-pro-wails`，执行 `wails dev`（热重载，Web 端口默认 34115）。
3) 构建桌面包：`wails build`（产物位于 `deploymaster-pro-wails/build/bin/`）。
4) 节点数据持久化：默认存储在 `~/.deploymaster/nodes.json`（macOS/Linux）或 `%USERPROFILE%\.deploymaster\nodes.json`（Windows）。

## 代码实现要点
- Go 后端
  - `internal/node`: JSON 持久化的节点 CRUD（`nodes.json` 原子写入）。
  - `internal/ssh`: 轻量 SSH 连接与批量连通性测试（握手延迟计时）。
  - `internal/topology`: 主/从拓扑数据汇总，供前端可视化。
  - `app.go`: Wails 绑定导出节点管理、连通性测试、拓扑查询等 API。
- Vue 前端（Wails）
  - `frontend/src/composables/useNodeService.ts`: 与 Wails 绑定交互的响应式封装，统一节点数据模型。
  - 页面分区：Dashboard、SVNManager、ServerManager、TaskExecutor、LogViewer；`ServerManager` 已接入真实节点 API，其余使用模拟数据待对接。
- React 原型
  - `deploymaster-pro/App.tsx`：同名页面与交互布局的纯前端版本，便于快速演示 UI/UX。

## 常用命令速查
- Go 单元测试：`cd deploymaster-pro-wails && go test ./...`
- 仅节点模块测试：`go test ./internal/node/...`
- 前端类型检查/构建（Wails）：`cd deploymaster-pro-wails/frontend && npm run build`

## 参考文档
- 产品需求：`需求文档.md`
- 技术架构：`架构文档.md`
- 节点拓扑接入指引：`deploymaster-pro-wails/NODE_TOPOLOGY_README.md`

## 下一步可选事项
- 将 SVN/SSH 真正驱动接入 Wails 后端，并替换前端模拟数据。
- 为任务状态机与日志流添加集成测试与 e2e 回归。
- 增加环境/密钥管理（系统钥匙串或安全存储）。
