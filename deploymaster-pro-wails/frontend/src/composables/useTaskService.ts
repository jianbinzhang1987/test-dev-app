import { ref } from 'vue';
import {
  GetTasks,
  AddTask,
  UpdateTask,
  DeleteTask,
  GetTaskTemplates,
  AddTaskTemplate,
  UpdateTaskTemplate,
  DeleteTaskTemplate,
  GetTaskRuns,
  GetTaskRunsByTask,
  DeleteTaskRun,
  DeleteTaskRunsByTask,
} from '../../wailsjs/go/main/App';
import { internal } from '../../wailsjs/go/models';
import { DeploymentTask, TaskTemplate, TaskRun } from '../types';

const tasks = ref<DeploymentTask[]>([]);
const templates = ref<TaskTemplate[]>([]);
const runs = ref<TaskRun[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const modelToTask = (task: internal.TaskDefinition): DeploymentTask => ({
  id: task.id,
  name: task.name,
  svnResourceId: task.svnResourceId,
  masterServerId: task.masterServerId,
  slaveServerIds: task.slaveServerIds || [],
  remotePath: task.remotePath,
  slaveRemotePath: task.slaveRemotePath,
  slaveRemotePaths: task.slaveRemotePaths || {},
  commands: task.commands || [],
  status: task.status as any,
  progress: task.progress ?? 0,
  createdAt: task.createdAt,
  updatedAt: task.updatedAt,
  lastRunAt: task.lastRunAt,
  templateId: task.templateId,
});

const modelToTemplate = (tpl: internal.TaskTemplate): TaskTemplate => ({
  id: tpl.id,
  name: tpl.name,
  svnResourceId: tpl.svnResourceId,
  masterServerId: tpl.masterServerId,
  slaveServerIds: tpl.slaveServerIds || [],
  remotePath: tpl.remotePath,
  slaveRemotePath: tpl.slaveRemotePath,
  slaveRemotePaths: tpl.slaveRemotePaths || {},
  commands: tpl.commands || [],
  sourceTaskId: tpl.sourceTaskId,
  createdAt: tpl.createdAt,
  updatedAt: tpl.updatedAt,
});

const modelToRun = (run: internal.TaskRun): TaskRun => ({
  id: run.id,
  taskId: run.taskId,
  taskName: run.taskName,
  status: run.status as any,
  progress: run.progress ?? 0,
  startedAt: run.startedAt,
  finishedAt: run.finishedAt,
  logs: run.logs || [],
});

export function useTaskService() {
  const loadTasks = async () => {
    loading.value = true;
    error.value = null;
    try {
      const list = await GetTasks();
      tasks.value = list.map(modelToTask);
    } catch (err: any) {
      error.value = `加载任务失败: ${err.message || err}`;
      console.error('加载任务失败:', err);
    } finally {
      loading.value = false;
    }
  };

  const addTask = async (task: Partial<DeploymentTask>) => {
    loading.value = true;
    try {
      const created = await AddTask(internal.TaskDefinition.createFrom(task));
      await loadTasks();
      return created;
    } catch (err: any) {
      error.value = `新增任务失败: ${err.message || err}`;
      console.error('新增任务失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const updateTask = async (task: Partial<DeploymentTask>) => {
    loading.value = true;
    try {
      if (!task.id) throw new Error('更新任务失败：缺少任务 ID');
      await UpdateTask(internal.TaskDefinition.createFrom(task));
      await loadTasks();
    } catch (err: any) {
      error.value = `更新任务失败: ${err.message || err}`;
      console.error('更新任务失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const deleteTask = async (id: string) => {
    loading.value = true;
    try {
      await DeleteTask(id);
      tasks.value = tasks.value.filter(t => t.id !== id);
      await loadTasks();
    } catch (err: any) {
      error.value = `删除任务失败: ${err.message || err}`;
      console.error('删除任务失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const loadTemplates = async () => {
    loading.value = true;
    error.value = null;
    try {
      const list = await GetTaskTemplates();
      templates.value = list.map(modelToTemplate);
    } catch (err: any) {
      error.value = `加载模板失败: ${err.message || err}`;
      console.error('加载模板失败:', err);
    } finally {
      loading.value = false;
    }
  };

  const addTemplate = async (tpl: Partial<TaskTemplate>) => {
    loading.value = true;
    try {
      const created = await AddTaskTemplate(internal.TaskTemplate.createFrom(tpl));
      await loadTemplates();
      return created;
    } catch (err: any) {
      error.value = `新增模板失败: ${err.message || err}`;
      console.error('新增模板失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const updateTemplate = async (tpl: Partial<TaskTemplate>) => {
    loading.value = true;
    try {
      if (!tpl.id) throw new Error('更新模板失败：缺少模板 ID');
      await UpdateTaskTemplate(internal.TaskTemplate.createFrom(tpl));
      await loadTemplates();
    } catch (err: any) {
      error.value = `更新模板失败: ${err.message || err}`;
      console.error('更新模板失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const deleteTemplate = async (id: string) => {
    loading.value = true;
    try {
      await DeleteTaskTemplate(id);
      templates.value = templates.value.filter(t => t.id !== id);
      await loadTemplates();
    } catch (err: any) {
      error.value = `删除模板失败: ${err.message || err}`;
      console.error('删除模板失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const loadRuns = async () => {
    loading.value = true;
    error.value = null;
    try {
      const list = await GetTaskRuns();
      runs.value = list.map(modelToRun);
    } catch (err: any) {
      error.value = `加载运行历史失败: ${err.message || err}`;
      console.error('加载运行历史失败:', err);
    } finally {
      loading.value = false;
    }
  };

  const loadRunsByTask = async (taskId: string) => {
    loading.value = true;
    error.value = null;
    try {
      const list = await GetTaskRunsByTask(taskId);
      runs.value = list.map(modelToRun);
    } catch (err: any) {
      error.value = `加载运行历史失败: ${err.message || err}`;
      console.error('加载运行历史失败:', err);
    } finally {
      loading.value = false;
    }
  };

  const deleteRun = async (runId: string) => {
    loading.value = true;
    try {
      await DeleteTaskRun(runId);
      runs.value = runs.value.filter(r => r.id !== runId);
    } catch (err: any) {
      error.value = `删除运行历史失败: ${err.message || err}`;
      console.error('删除运行历史失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const deleteRunsByTask = async (taskId: string) => {
    loading.value = true;
    try {
      await DeleteTaskRunsByTask(taskId);
      runs.value = runs.value.filter(r => r.taskId !== taskId);
    } catch (err: any) {
      error.value = `清空运行历史失败: ${err.message || err}`;
      console.error('清空运行历史失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  return {
    tasks,
    templates,
    runs,
    loading,
    error,
    loadTasks,
    addTask,
    updateTask,
    deleteTask,
    loadTemplates,
    addTemplate,
    updateTemplate,
    deleteTemplate,
    loadRuns,
    loadRunsByTask,
    deleteRun,
    deleteRunsByTask,
  };
}
