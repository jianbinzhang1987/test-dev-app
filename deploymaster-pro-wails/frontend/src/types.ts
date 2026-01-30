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
  latency?: number; // 延迟(ms)
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
}

export interface DeploymentTask {
  id: string;
  name: string;
  svnResourceId: string;
  masterServerId: string;
  slaveServerIds: string[];
  remotePath: string;
  commands: string[];
  status: TaskStatus;
  progress: number;
  logs: string[];
}
