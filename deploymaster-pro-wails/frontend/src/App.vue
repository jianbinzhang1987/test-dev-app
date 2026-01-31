<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';
import Sidebar from './components/Sidebar.vue';
import Header from './components/Header.vue';
import Dashboard from './pages/Dashboard.vue';
import SVNManager from './pages/SVNManager.vue';
import ServerManager from './pages/ServerManager.vue';
import TaskExecutor from './pages/TaskExecutor.vue';
import LogViewer from './pages/LogViewer.vue';
import { RemoteServer, SVNResource, DeploymentTask, TaskStatus, TaskRun, TaskTemplate } from './types';
import { useNodeService } from './composables/useNodeService';
import { useSvnService } from './composables/useSvnService';
import { useTaskService } from './composables/useTaskService';
import { EventsOn } from '../wailsjs/runtime/runtime';

// Global State
const activeTab = ref('dashboard');
const selectedLogTaskId = ref<string | null>(null);
const globalAutoOpenTaskModal = ref(false);

// 使用真实的节点服务
const nodeService = useNodeService();
const svnService = useSvnService();
const taskService = useTaskService();

// 在组件挂载时加载节点数据
onMounted(async () => {
  await nodeService.loadNodes();
  await svnService.loadResources();
  await taskService.loadTasks();
  await taskService.loadTemplates();
  await taskService.loadRuns();
});

const tasks = taskService.tasks;
const templates = taskService.templates;
const runs = taskService.runs;

// Handlers
const handleGlobalNewDeployment = () => {
  activeTab.value = 'tasks';
  globalAutoOpenTaskModal.value = true;
};

const handleAddResource = async (
  res: SVNResource,
  creds?: { username?: string; password?: string; remember?: boolean }
) => {
  await svnService.addResource(res, creds);
};
const handleUpdateResource = async (
  res: SVNResource,
  creds?: { username?: string; password?: string; remember?: boolean }
) => {
  await svnService.updateResource(res, creds);
};
const handleDeleteResource = async (id: string) => {
  await svnService.deleteResource(id);
};

const handleSVNTestConnection = async (payload: { url: string; username: string; password: string; resourceId?: string }) => {
  return svnService.testConnection(payload.url, payload.username, payload.password, payload.resourceId || '');
};

const handleSVNRefreshAll = async () => {
  await svnService.refreshAll();
};

// 节点管理处理器 - 现在使用真实API
const handleAddServer = async (srv: RemoteServer) => {
  try {
    await nodeService.addNode(srv);
  } catch (err) {
    console.error('添加服务器失败:', err);
  }
};

const handleUpdateServer = async (srv: RemoteServer) => {
  try {
    await nodeService.updateNode(srv);
  } catch (err) {
    console.error('更新服务器失败:', err);
  }
};

const handleDeleteServer = async (id: string) => {
  try {
    await nodeService.deleteNode(id);
  } catch (err) {
    console.error('删除服务器失败:', err);
  }
};

const handleAddTask = async (task: Partial<DeploymentTask>) => {
  await taskService.addTask(task);
};
const handleSaveTask = async (task: Partial<DeploymentTask>) => {
  await taskService.updateTask({ ...task, progress: -1 });
};
const handleUpdateTask = (task: DeploymentTask) => {
  const idx = tasks.value.findIndex(t => t.id === task.id);
  if (idx !== -1) tasks.value[idx] = task;
};
const handleDeleteTask = async (taskId: string) => {
  await taskService.deleteTask(taskId);
};

const handleCreateTemplate = async (tpl: Partial<TaskTemplate>) => {
  await taskService.addTemplate(tpl);
};

const handleDeleteTemplate = async (templateId: string) => {
  await taskService.deleteTemplate(templateId);
};

const handleDeleteRun = async (runId: string) => {
  await taskService.deleteRun(runId);
};

const handleDeleteRunsByTask = async (taskId: string) => {
  await taskService.deleteRunsByTask(taskId);
};

const handleViewLogs = (taskId: string) => {
  selectedLogTaskId.value = taskId;
  activeTab.value = 'logs';
};

let unsubscribeTaskEvents: (() => void) | null = null;
onMounted(() => {
  unsubscribeTaskEvents = EventsOn('task:event', (event: any) => {
    const task = tasks.value.find(t => t.id === event.taskId);
    if (task) {
      handleUpdateTask({
        ...task,
        status: event.status,
        progress: event.progress,
      });
    }

    if (event.runId) {
      const existing = runs.value.find(r => r.id === event.runId);
      if (existing) {
        existing.status = event.status;
        existing.progress = event.progress;
        existing.logs = [...(existing.logs || []), event.log];
        if ((event.status === TaskStatus.SUCCESS || event.status === TaskStatus.FAILED) && !existing.finishedAt) {
          existing.finishedAt = new Date().toLocaleString();
        }
      } else {
        const run: TaskRun = {
          id: event.runId,
          taskId: event.taskId,
          taskName: task?.name || '未知任务',
          status: event.status,
          progress: event.progress,
          startedAt: new Date().toLocaleString(),
          logs: [event.log],
        };
        runs.value.unshift(run);
      }
    }
  });
});

onBeforeUnmount(() => {
  if (unsubscribeTaskEvents) {
    unsubscribeTaskEvents();
    unsubscribeTaskEvents = null;
  }
});
</script>

<template>
  <div class="flex flex-col h-full bg-white text-slate-900 font-sans overflow-hidden">
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar Navigation -->
      <Sidebar :activeTab="activeTab" @update:activeTab="(v) => activeTab = v" />

      <div class="flex-1 flex flex-col min-w-0 overflow-hidden bg-slate-50">
        <!-- Page Header -->
        <Header :activeTab="activeTab" @newDeployment="handleGlobalNewDeployment" />

        <!-- Main Content Area -->
        <main class="flex-1 overflow-y-auto p-6 scroll-smooth">
          <Dashboard v-if="activeTab === 'dashboard'" :tasks="tasks" :servers="nodeService.servers.value" :runs="runs"
            @viewAllRuns="activeTab = 'logs'" />

          <SVNManager v-else-if="activeTab === 'svn'" :resources="svnService.resources.value"
            :loading="svnService.loading.value" :testConnection="handleSVNTestConnection"
            :refreshAll="handleSVNRefreshAll" @add="handleAddResource" @update="handleUpdateResource"
            @delete="handleDeleteResource" />

          <ServerManager v-else-if="activeTab === 'servers'" :servers="nodeService.servers.value"
            :loading="nodeService.loading.value" @update-list="nodeService.loadNodes" @delete="handleDeleteServer"
            @test="nodeService.testConnection" />

          <TaskExecutor v-else-if="activeTab === 'tasks'" :tasks="tasks" :templates="templates"
            :servers="nodeService.servers.value" :resources="svnService.resources.value"
            :autoOpenModal="globalAutoOpenTaskModal" @addTask="handleAddTask" @saveTask="handleSaveTask"
            @updateTask="handleUpdateTask" @deleteTask="handleDeleteTask" @createTemplate="handleCreateTemplate"
            @deleteTemplate="handleDeleteTemplate" @modalClose="globalAutoOpenTaskModal = false" @viewLogs="handleViewLogs" />

          <LogViewer v-else-if="activeTab === 'logs'" :runs="runs" :selectedTaskId="selectedLogTaskId"
            @deleteRun="handleDeleteRun" @deleteRunsByTask="handleDeleteRunsByTask" />
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
