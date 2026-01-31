export enum TaskStatus {
  IDLE = 'IDLE',
  DOWNLOADING = 'DOWNLOADING',
  UPLOADING = 'UPLOADING',
  SYNCING = 'SYNCING',
  EXECUTING = 'EXECUTING',
  SUCCESS = 'SUCCESS',
  FAILED = 'FAILED'
}

export interface RemoteServer {
  id: string;
  name: string;
  ip: string;
  port: number;
  protocol: 'SFTP' | 'FTP' | 'SCP';
  isMaster: boolean;

  // 认证相关字段
  username?: string;           // SSH用户名
  authMethod?: 'password' | 'key' | 'agent';  // 认证方式
  keyPath?: string;            // SSH私钥路径（仅key模式）

  // 运行时状态
  latency?: number; // 延迟(ms) - 兼容字段
  delay?: number;   // 延迟(ms) - UI显示字段
  status?: 'connected' | 'disconnected' | 'testing';
  lastChecked?: string; // 最后检测时间
}

export interface SVNResource {
  id: string;
  url: string;
  name: string;
  type: 'file' | 'folder';
  revision: string;
  status: 'online' | 'error' | 'syncing';
  lastChecked: string;
  size?: string;
  username?: string;
}

export interface SVNTestResult {
  ok: boolean;
  revision?: string;
  message?: string;
  durationMs?: number;
  checkedAt?: string;
}

export interface DeploymentTask {
  id: string;
  name: string;
  svnResourceId: string;
  masterServerId: string;
  slaveServerIds: string[];
  remotePath: string;
  slaveRemotePath?: string;
  slaveRemotePaths?: Record<string, string>;
  commands: string[];
  status: TaskStatus;
  progress: number;
  createdAt?: string;
  updatedAt?: string;
  lastRunAt?: string;
  templateId?: string;
}

export interface TaskTemplate {
  id: string;
  name: string;
  svnResourceId: string;
  masterServerId: string;
  slaveServerIds: string[];
  remotePath: string;
  slaveRemotePath?: string;
  slaveRemotePaths?: Record<string, string>;
  commands: string[];
  sourceTaskId?: string;
  createdAt?: string;
  updatedAt?: string;
}

export interface TaskRun {
  id: string;
  taskId: string;
  taskName: string;
  status: TaskStatus;
  progress: number;
  startedAt: string;
  finishedAt?: string;
  logs: string[];
}
