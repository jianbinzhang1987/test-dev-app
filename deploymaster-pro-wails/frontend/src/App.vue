<script setup lang="ts">
import { ref, computed } from 'vue';
import Sidebar from './components/Sidebar.vue';
import Header from './components/Header.vue';
import Dashboard from './pages/Dashboard.vue';
import SVNManager from './pages/SVNManager.vue';
import ServerManager from './pages/ServerManager.vue';
import TaskExecutor from './pages/TaskExecutor.vue';
import LogViewer from './pages/LogViewer.vue';
import { RemoteServer, SVNResource, DeploymentTask, TaskStatus } from './types';

// Global State
const activeTab = ref('dashboard');
const globalAutoOpenTaskModal = ref(false);

const servers = ref<RemoteServer[]>([
  { id: 'srv-1', name: '香港分发主控', ip: '43.154.21.90', port: 22, protocol: 'SFTP', isMaster: true, latency: 24, lastChecked: '10:45:00' },
  { id: 'srv-2', name: '内河-01 节点', ip: '10.0.4.12', port: 22, protocol: 'SFTP', isMaster: false, latency: 12, lastChecked: '10:45:00' },
  { id: 'srv-3', name: '内河-02 节点', ip: '10.0.4.15', port: 22, protocol: 'SFTP', isMaster: false, latency: 15, lastChecked: '10:45:00' }
]);

const resources = ref<SVNResource[]>([
  { id: 'res-1', name: '核心支付网关', url: 'svn://192.168.1.100/pay/trunk', revision: '8902', lastChecked: '2024-03-24 10:20:12', status: 'ready' as any, type: 'folder' },
  { id: 'res-2', name: '静态资源加速包', url: 'svn://192.168.1.100/cdn/assets', revision: '4521', lastChecked: '2024-03-23 15:44:02', status: 'ready' as any, type: 'folder' }
]);

const tasks = ref<DeploymentTask[]>([
  {
    id: 'task-1',
    name: '支付网关-例行更新-v2.1',
    svnResourceId: 'res-1',
    masterServerId: 'srv-1',
    slaveServerIds: ['srv-2', 'srv-3'],
    remotePath: '/var/www/pay',
    commands: ['npm install', 'npm run build', 'pm2 restart all'],
    status: TaskStatus.SUCCESS,
    progress: 100,
    logs: ['[系统] 任务已完成']
  }
]);

// Handlers
const handleGlobalNewDeployment = () => {
  activeTab.value = 'tasks';
  globalAutoOpenTaskModal.value = true;
};

const handleAddResource = (res: SVNResource) => resources.value.push(res);
const handleUpdateResource = (res: SVNResource) => {
  const idx = resources.value.findIndex(r => r.id === res.id);
  if (idx !== -1) resources.value[idx] = res;
};
const handleDeleteResource = (id: string) => resources.value = resources.value.filter(r => r.id !== id);

const handleAddServer = (srv: RemoteServer) => servers.value.push(srv);
const handleUpdateServer = (srv: RemoteServer) => {
  const idx = servers.value.findIndex(s => s.id === srv.id);
  if (idx !== -1) servers.value[idx] = srv;
};
const handleDeleteServer = (id: string) => servers.value = servers.value.filter(s => s.id !== id);

const handleAddTask = (task: DeploymentTask) => tasks.value.unshift(task);
const handleUpdateTask = (task: DeploymentTask) => {
  const idx = tasks.value.findIndex(t => t.id === task.id);
  if (idx !== -1) tasks.value[idx] = task;
};
</script>

<template>
  <div class="flex flex-col h-full bg-white text-slate-900 font-sans overflow-hidden">
    <!-- Window Header (Custom Titlebar area) -->
    <div class="h-10 bg-slate-900 flex items-center justify-between px-4 shrink-0 select-none">
      <div class="flex items-center space-x-3">
        <div class="flex space-x-1.5">
          <div class="w-2.5 h-2.5 rounded-full bg-red-400"></div>
          <div class="w-2.5 h-2.5 rounded-full bg-amber-400"></div>
          <div class="w-2.5 h-2.5 rounded-full bg-emerald-400"></div>
        </div>
        <div class="flex items-center space-x-2 text-white/40 text-[10px] uppercase font-black tracking-widest ml-4">
          <i class="fa-solid fa-rocket text-blue-500"></i>
          <span>DeployMaster Pro</span>
          <span class="text-white/20">|</span>
          <span class="text-blue-400">v2.5.0-Release</span>
        </div>
      </div>
      <div class="flex items-center space-x-6 text-white/30 text-[9px] font-bold">
        <span class="flex items-center space-x-1.5"><i class="fa-solid fa-shield-halved"></i> <span>SECURE
            MODE</span></span>
        <span class="flex items-center space-x-1.5"><i class="fa-solid fa-server"></i> <span>CLUSTER: ON</span></span>
      </div>
    </div>

    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar Navigation -->
      <Sidebar :activeTab="activeTab" @update:activeTab="(v) => activeTab = v" />

      <div class="flex-1 flex flex-col min-w-0 overflow-hidden bg-slate-50">
        <!-- Page Header -->
        <Header :activeTab="activeTab" @newDeployment="handleGlobalNewDeployment" />

        <!-- Main Content Area -->
        <main class="flex-1 overflow-y-auto p-6 scroll-smooth">
          <Dashboard v-if="activeTab === 'dashboard'" :tasks="tasks" :servers="servers" />

          <SVNManager v-else-if="activeTab === 'svn'" :resources="resources" @add="handleAddResource"
            @update="handleUpdateResource" @delete="handleDeleteResource" />

          <ServerManager v-else-if="activeTab === 'servers'" :servers="servers" @add="handleAddServer"
            @update="handleUpdateServer" @delete="handleDeleteServer" />

          <TaskExecutor v-else-if="activeTab === 'tasks'" :tasks="tasks" :servers="servers" :resources="resources"
            :autoOpenModal="globalAutoOpenTaskModal" @addTask="handleAddTask" @updateTask="handleUpdateTask"
            @modalClose="globalAutoOpenTaskModal = false" />

          <LogViewer v-else-if="activeTab === 'logs'" :tasks="tasks" />
        </main>

        <!-- Status Footer -->
        <footer
          class="h-6 bg-white border-t border-slate-200 px-4 flex items-center justify-between shrink-0 text-[10px] text-slate-500">
          <div class="flex items-center space-x-4">
            <span class="flex items-center space-x-1.5"><span
                class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span> <span
                class="font-bold">系统就绪</span></span>
            <span class="text-slate-200">|</span>
            <span class="opacity-70 font-medium">后端引擎: <span class="text-blue-500">Go/Wails v2.11</span></span>
          </div>
          <div class="flex items-center space-x-4 uppercase font-bold tracking-tighter opacity-70">
            <span>MEM: 124MB</span>
            <span>CPU: 0.8%</span>
            <span class="text-slate-300">|</span>
            <span>2024-03-24 11:32:04</span>
          </div>
        </footer>
      </div>
    </div>
  </div>
</template>

<style>
/* Global scrollbar styling for a consistent premium look */
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #cbd5e1;
  border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #94a3b8;
}

/* Base animations */
.animate-in {
  animation-duration: 0.3s;
  animation-fill-mode: both;
}

.fade-in {
  animation-name: fadeIn;
}

.zoom-in-95 {
  animation-name: zoomIn95;
}

.slide-in-from-right {
  animation-name: slideInRight;
}

@keyframes fadeIn {
  from {
    opacity: 0;
  }

  to {
    opacity: 1;
  }
}

@keyframes zoomIn95 {
  from {
    opacity: 0;
    transform: scale(0.95);
  }

  to {
    opacity: 1;
    transform: scale(1);
  }
}

@keyframes slideInRight {
  from {
    transform: translateX(100%);
  }

  to {
    transform: translateX(0);
  }
}
</style>
