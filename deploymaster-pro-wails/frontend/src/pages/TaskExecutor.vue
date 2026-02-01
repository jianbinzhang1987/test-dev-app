<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { ExecuteTask, HasStoredCredential, ShowMessageDialog, ConfirmDialog } from '../../wailsjs/go/main/App';
import { internal } from '../../wailsjs/go/models';
import { DeploymentTask, RemoteServer, SVNResource, TaskStatus, TaskTemplate } from '../types';

const props = defineProps<{
    tasks: DeploymentTask[];
    servers: RemoteServer[];
    resources: SVNResource[];
    templates: TaskTemplate[];
    autoOpenModal?: boolean;
    windowed?: boolean;
}>();

const emit = defineEmits(['addTask', 'saveTask', 'updateTask', 'deleteTask', 'createTemplate', 'deleteTemplate', 'modalClose', 'viewLogs']);

const isCreateModalOpen = ref(false);
const isTemplateModalOpen = ref(false);
const selectedTaskDetails = ref<DeploymentTask | null>(null);
const selectedTemplateId = ref<string>('');
const editingTaskId = ref<string>('');

// New state for redesign
const currentStep = ref(1);
const activeFilter = ref('all');
const searchQuery = ref('');

const isRunning = (status: TaskStatus) => {
    return [TaskStatus.DOWNLOADING, TaskStatus.UPLOADING, TaskStatus.SYNCING, TaskStatus.EXECUTING].includes(status);
};

const filteredTasks = computed(() => {
    let list = props.tasks;
    if (activeFilter.value === 'running') {
        list = list.filter(t => isRunning(t.status));
    } else if (activeFilter.value === 'idle') {
        list = list.filter(t => t.status === TaskStatus.IDLE);
    } else if (activeFilter.value === 'failed') {
        list = list.filter(t => t.status === TaskStatus.FAILED);
    } else if (activeFilter.value === 'success') {
        list = list.filter(t => t.status === TaskStatus.SUCCESS);
    }

    if (searchQuery.value) {
        const q = searchQuery.value.toLowerCase();
        list = list.filter(t => t.name.toLowerCase().includes(q) || t.remotePath.toLowerCase().includes(q));
    }
    return list;
});

const taskStats = computed(() => {
    return {
        total: props.tasks.length,
        running: props.tasks.filter(t => isRunning(t.status)).length,
        failed: props.tasks.filter(t => t.status === TaskStatus.FAILED).length,
        success: props.tasks.filter(t => t.status === TaskStatus.SUCCESS).length
    };
});

const initialFormState = () => ({
    name: '',
    svnResourceId: props.resources[0]?.id || '',
    masterServerId: props.servers.find(s => s.isMaster)?.id || '',
    slaveServerIds: [] as string[],
    remotePath: '',
    slaveRemotePath: '',
    slaveRemotePaths: {} as Record<string, string>,
    commands: ''
});

const formData = ref(initialFormState());

watch(() => props.autoOpenModal, (newVal) => {
    if (newVal) {
        isCreateModalOpen.value = true;
        editingTaskId.value = '';
        formData.value = initialFormState();
        emit('modalClose');
    }
});

const masters = computed(() => props.servers.filter(s => s.isMaster));
const slaves = computed(() => props.servers.filter(s => !s.isMaster));
const isWindowed = computed(() => Boolean(props.windowed));

const handleCreateTask = () => {
    if (!formData.value.name || !formData.value.remotePath || !formData.value.masterServerId) {
        ShowMessageDialog('必填项缺失', '请检查：任务名称、主节点及主控远程路径为必填项', 'warning');
        return;
    }

    const newTask = {
        name: formData.value.name,
        svnResourceId: formData.value.svnResourceId,
        masterServerId: formData.value.masterServerId,
        slaveServerIds: formData.value.slaveServerIds,
        remotePath: formData.value.remotePath,
        slaveRemotePath: formData.value.slaveRemotePath,
        slaveRemotePaths: formData.value.slaveRemotePaths,
        commands: formData.value.commands.split('\n').map(c => c.trim()).filter(c => c),
    };

    if (editingTaskId.value) {
        emit('saveTask', { id: editingTaskId.value, ...newTask });
        editingTaskId.value = '';
    } else {
        emit('addTask', newTask);
    }
    isCreateModalOpen.value = false;
    formData.value = initialFormState();
};

const closeCreateModal = () => {
    isCreateModalOpen.value = false;
    editingTaskId.value = '';
    formData.value = initialFormState();
};

const runTask = async (task: DeploymentTask) => {
    if (task.status !== TaskStatus.IDLE && task.status !== TaskStatus.FAILED && task.status !== TaskStatus.SUCCESS) return;
    const targets = [
        props.servers.find(s => s.id === task.masterServerId),
        ...props.servers.filter(s => task.slaveServerIds.includes(s.id))
    ].filter(Boolean) as RemoteServer[];

    const missing: string[] = [];
    for (const node of targets) {
        const authMethod = node.authMethod || 'password';
        if (authMethod === 'key' || authMethod === 'agent') {
            continue;
        }
        const username = node.username || 'root';
        const has = await HasStoredCredential(node.id, username);
        if (!has) {
            missing.push(`${node.name || node.ip}(${username})`);
        }
    }
    if (missing.length > 0) {
        await ShowMessageDialog('无法执行任务', `以下节点未保存密码，无法执行任务：\\n${missing.join('\\n')}`, 'error');
        return;
    }

    const request = internal.TaskRunRequest.createFrom({
        taskId: task.id,
        taskName: task.name,
        svnResourceId: task.svnResourceId,
        masterServerId: task.masterServerId,
        slaveServerIds: task.slaveServerIds,
        remotePath: task.remotePath,
        slaveRemotePath: task.slaveRemotePath,
        slaveRemotePaths: task.slaveRemotePaths,
        commands: task.commands,
    });
    try {
        await ExecuteTask(request);
    } catch (err: any) {
        await ShowMessageDialog('任务启动失败', `${err?.message || err}`, 'error');
    }
};

const toggleSlaveSelection = (id: string) => {
    const index = formData.value.slaveServerIds.indexOf(id);
    if (index === -1) {
        formData.value.slaveServerIds.push(id);
    } else {
        formData.value.slaveServerIds.splice(index, 1);
        if (formData.value.slaveRemotePaths[id]) {
            delete formData.value.slaveRemotePaths[id];
        }
    }
};

const handleDeleteTask = async (task: DeploymentTask) => {
    const ok = await ConfirmDialog('确认删除', `确定要删除任务：${task.name} 吗？`);
    if (!ok) return;
    emit('deleteTask', task.id);
    if (selectedTaskDetails.value?.id === task.id) {
        selectedTaskDetails.value = null;
    }
};

const handleSaveTemplate = async () => {
    if (!selectedTaskDetails.value) return;
    emit('createTemplate', {
        name: selectedTaskDetails.value.name,
        svnResourceId: selectedTaskDetails.value.svnResourceId,
        masterServerId: selectedTaskDetails.value.masterServerId,
        slaveServerIds: selectedTaskDetails.value.slaveServerIds,
        remotePath: selectedTaskDetails.value.remotePath,
        slaveRemotePath: selectedTaskDetails.value.slaveRemotePath,
        slaveRemotePaths: selectedTaskDetails.value.slaveRemotePaths,
        commands: selectedTaskDetails.value.commands,
        sourceTaskId: selectedTaskDetails.value.id,
    });
};

const handleCloneFromTemplate = () => {
    const tpl = props.templates.find(t => t.id === selectedTemplateId.value);
    if (!tpl) return;
    emit('addTask', {
        name: `${tpl.name}-克隆`,
        svnResourceId: tpl.svnResourceId,
        masterServerId: tpl.masterServerId,
        slaveServerIds: tpl.slaveServerIds,
        remotePath: tpl.remotePath,
        slaveRemotePath: tpl.slaveRemotePath,
        slaveRemotePaths: tpl.slaveRemotePaths,
        commands: tpl.commands,
        templateId: tpl.id,
    });
    isTemplateModalOpen.value = false;
    selectedTemplateId.value = '';
};

const handleEditTask = (task: DeploymentTask) => {
    editingTaskId.value = task.id;
    formData.value = {
        name: task.name,
        svnResourceId: task.svnResourceId,
        masterServerId: task.masterServerId,
        slaveServerIds: [...task.slaveServerIds],
        remotePath: task.remotePath,
        slaveRemotePath: task.slaveRemotePath || '',
        slaveRemotePaths: { ...(task.slaveRemotePaths || {}) },
        commands: task.commands.join('\n'),
    };
    isCreateModalOpen.value = true;
    selectedTaskDetails.value = null;
};
</script>

<template>
    <div :class="[isWindowed ? 'space-y-4 pb-6' : 'space-y-6 pb-10']">
        <!-- Compact Header & Summary Cards -->
        <div :class="[isWindowed ? 'flex flex-col space-y-4' : 'flex flex-col space-y-6']">
            <div :class="[isWindowed ? 'flex flex-col gap-3' : 'flex items-center justify-between']">
                <div class="flex items-center space-x-4">
                    <div
                        :class="[isWindowed ? 'w-10 h-10' : 'w-12 h-12', 'bg-blue-600 rounded-xl flex items-center justify-center text-white shadow-lg shadow-blue-500/20']">
                        <i class="fa-solid fa-wand-magic-sparkles text-xl"></i>
                    </div>
                    <div>
                        <h2 :class="[isWindowed ? 'text-lg' : 'text-xl', 'font-black text-slate-800 tracking-tight']">任务编排中心</h2>
                        <div
                            class="flex items-center space-x-2 text-[10px] font-bold text-slate-400 uppercase tracking-widest">
                            <span class="text-blue-500">Orchestrator v2.0</span>
                            <span>•</span>
                            <span>{{ tasks.length }} 个活跃任务</span>
                        </div>
                    </div>
                </div>
                <div :class="[isWindowed ? 'flex flex-wrap gap-2' : 'flex items-center space-x-3']">
                    <button @click="isTemplateModalOpen = true"
                        class="px-4 py-2 rounded-xl bg-white border border-slate-200 text-slate-600 font-bold text-xs hover:bg-slate-50 transition-all flex items-center space-x-2">
                        <i class="fa-solid fa-clone text-[10px]"></i>
                        <span>模板库</span>
                    </button>
                    <button @click="isCreateModalOpen = true; currentStep = 1"
                        class="px-6 py-2.5 rounded-xl bg-slate-900 text-white font-black text-xs hover:bg-blue-600 shadow-lg transition-all flex items-center space-x-2">
                        <i class="fa-solid fa-plus-circle"></i>
                        <span>构建新工作流</span>
                    </button>
                </div>
            </div>

            <!-- Dashboard Stats Matrix -->
            <div :class="[isWindowed ? 'grid grid-cols-2 gap-3' : 'grid grid-cols-4 gap-4']">
                <div class="bg-white p-4 rounded-2xl border border-slate-100 shadow-sm flex items-center space-x-4">
                    <div class="w-10 h-10 rounded-xl bg-slate-50 flex items-center justify-center text-slate-400">
                        <i class="fa-solid fa-list-check"></i>
                    </div>
                    <div>
                        <p class="text-[10px] font-bold text-slate-400 uppercase">总任务</p>
                        <p class="text-lg font-black text-slate-800">{{ taskStats.total }}</p>
                    </div>
                </div>
                <div class="bg-white p-4 rounded-2xl border border-slate-100 shadow-sm flex items-center space-x-4">
                    <div class="w-10 h-10 rounded-xl bg-blue-50 flex items-center justify-center text-blue-500">
                        <i :class="['fa-solid fa-spinner', taskStats.running > 0 ? 'animate-spin' : '']"></i>
                    </div>
                    <div>
                        <p class="text-[10px] font-bold text-slate-400 uppercase">运行中</p>
                        <p class="text-lg font-black text-slate-800">{{ taskStats.running }}</p>
                    </div>
                </div>
                <div class="bg-white p-4 rounded-2xl border border-slate-100 shadow-sm flex items-center space-x-4">
                    <div class="w-10 h-10 rounded-xl bg-emerald-50 flex items-center justify-center text-emerald-500">
                        <i class="fa-solid fa-circle-check"></i>
                    </div>
                    <div>
                        <p class="text-[10px] font-bold text-slate-400 uppercase">今日成功</p>
                        <p class="text-lg font-black text-slate-800">{{ taskStats.success }}</p>
                    </div>
                </div>
                <div class="bg-white p-4 rounded-2xl border border-slate-100 shadow-sm flex items-center space-x-4">
                    <div class="w-10 h-10 rounded-xl bg-red-50 flex items-center justify-center text-red-500">
                        <i class="fa-solid fa-triangle-exclamation"></i>
                    </div>
                    <div>
                        <p class="text-[10px] font-bold text-slate-400 uppercase">异常拦截</p>
                        <p class="text-lg font-black text-slate-800">{{ taskStats.failed }}</p>
                    </div>
                </div>
            </div>
        </div>

        <!-- Task Control & Filters -->
        <div :class="[isWindowed ? 'flex flex-col gap-3 bg-white p-3 rounded-2xl border border-slate-100 shadow-sm' : 'flex items-center justify-between bg-white p-2 rounded-2xl border border-slate-100 shadow-sm']">
            <div class="flex items-center p-1 bg-slate-50 rounded-xl">
                <button
                    v-for="tab in [{ id: 'all', label: '全部' }, { id: 'running', label: '运行中' }, { id: 'idle', label: '就绪' }, { id: 'failed', label: '异常' }]"
                    :key="tab.id" @click="activeFilter = tab.id"
                    :class="['px-6 py-2 rounded-lg text-xs font-black transition-all',
                        activeFilter === tab.id ? 'bg-white text-slate-900 shadow-md shadow-slate-200' : 'text-slate-400 hover:text-slate-600']">
                    {{ tab.label }}
                </button>
            </div>
            <div :class="[isWindowed ? 'flex items-center w-full' : 'flex items-center space-x-4 pr-2']">
                <div class="relative group">
                    <i
                        class="fa-solid fa-magnifying-glass absolute left-4 top-1/2 -translate-y-1/2 text-slate-300 group-focus-within:text-blue-500 transition-colors"></i>
                    <input type="text" v-model="searchQuery" placeholder="搜索任务名称或路径..."
                        :class="[isWindowed ? 'w-full' : 'w-64', 'pl-10 pr-4 py-2 bg-slate-50 border border-transparent rounded-xl text-xs font-bold outline-none focus:bg-white focus:border-blue-500/20 focus:ring-4 focus:ring-blue-500/5 transition-all']" />
                </div>
            </div>
        </div>

        <!-- Task List -->
        <div class="grid grid-cols-1 gap-4">
            <div v-if="filteredTasks.length === 0"
                :class="[isWindowed ? 'bg-white border-2 border-dashed border-slate-100 rounded-[2rem] p-12 text-center flex flex-col items-center justify-center space-y-3' : 'bg-white border-2 border-dashed border-slate-100 rounded-[2rem] p-20 text-center flex flex-col items-center justify-center space-y-4']">
                <div class="w-20 h-20 bg-slate-50 rounded-full flex items-center justify-center text-slate-200">
                    <i class="fa-solid fa-wind text-4xl"></i>
                </div>
                <div>
                    <p class="text-sm font-black text-slate-400 uppercase tracking-[0.2em]">未找到相关任务</p>
                    <p class="text-xs text-slate-300 mt-2">试试调整筛选条件或搜索关键词</p>
                </div>
            </div>
            <div v-for="task in filteredTasks" :key="task.id" :class="['bg-white rounded-[1.5rem] border border-slate-100 shadow-sm hover:shadow-xl hover:translate-x-1 transition-all group overflow-hidden relative',
                isWindowed ? 'p-4' : 'p-5',
                isWindowed ? 'flex flex-col items-start space-y-4' : 'flex items-center space-x-8']">

                <!-- Status indicator line -->
                <div :class="['absolute left-0 top-0 bottom-0 w-1.5',
                    task.status === TaskStatus.SUCCESS ? 'bg-emerald-500' :
                        task.status === TaskStatus.FAILED ? 'bg-red-500' :
                            isRunning(task.status) ? 'bg-blue-500' : 'bg-slate-200']"></div>

                <div class="flex-1 min-w-0">
                    <div class="flex items-center space-x-3 mb-1.5">
                        <div :class="['rounded-2xl flex items-center justify-center font-black shrink-0 shadow-inner border transition-all',
                            isWindowed ? 'w-9 h-9 text-base' : 'w-12 h-12 text-xl',
                            task.status === TaskStatus.SUCCESS ? 'bg-emerald-50 border-emerald-100 text-emerald-500' :
                                task.status === TaskStatus.FAILED ? 'bg-red-50 border-red-100 text-red-500' :
                                    isRunning(task.status) ? 'bg-blue-50 border-blue-200 text-blue-500 shadow-blue-500/10' :
                                        'bg-slate-50 border-slate-100 text-slate-300']">
                            <i v-if="task.status === TaskStatus.SUCCESS"
                                class="fa-solid fa-check-double rotate-0 group-hover:rotate-12 transition-transform"></i>
                            <i v-else-if="task.status === TaskStatus.FAILED" class="fa-solid fa-triangle-exclamation"></i>
                            <i v-else-if="isRunning(task.status)" class="fa-solid fa-spinner animate-spin"></i>
                            <i v-else class="fa-solid fa-cube"></i>
                        </div>
                        <h4 class="font-black text-slate-800 text-base truncate">{{ task.name }}</h4>
                        <!-- More compact status badge -->
                        <span :class="['text-[8px] font-black px-1.5 py-0.5 rounded-md border uppercase tracking-tighter shadow-sm',
                            task.status === TaskStatus.SUCCESS ? 'bg-emerald-500 border-emerald-600 text-white' :
                                task.status === TaskStatus.FAILED ? 'bg-red-500 border-red-600 text-white' :
                                    isRunning(task.status) ? 'bg-blue-600 border-blue-700 text-white' :
                                        'bg-white border-slate-200 text-slate-400']">
                            {{ task.status }}
                        </span>
                    </div>
                    <div class="flex items-center space-x-5 text-[10px] text-slate-400 font-bold overflow-hidden">
                        <span class="flex items-center space-x-2 shrink-0">
                            <i class="fa-solid fa-folder-tree text-blue-400/60"></i>
                            <span class="text-slate-600 max-w-[200px] truncate select-all">{{ task.remotePath }}</span>
                        </span>
                        <span class="flex items-center space-x-2 shrink-0">
                            <i class="fa-solid fa-server text-indigo-400/60"></i>
                            <span class="text-slate-500">{{ task.slaveServerIds.length }} 台从机</span>
                        </span>
                        <span v-if="task.lastRunAt" class="flex items-center space-x-2 shrink-0 text-slate-300">
                            <i class="fa-solid fa-clock opacity-50"></i>
                            <span>上次运行: {{ task.lastRunAt }}</span>
                        </span>
                    </div>
                </div>

                <div :class="[isWindowed ? 'w-full' : 'w-48 shrink-0 px-4']">
                    <div class="flex justify-between mb-1.5 items-end">
                        <span class="text-[9px] font-black text-slate-300 uppercase tracking-widest">Progress</span>
                        <span class="text-[11px] font-mono font-black text-slate-700">{{ task.progress }}%</span>
                    </div>
                    <div class="w-full bg-slate-100 rounded-full h-1.5 overflow-hidden shadow-inner relative">
                        <div :class="['h-full transition-all duration-700 ease-out relative z-10',
                            task.status === TaskStatus.SUCCESS ? 'bg-emerald-500' :
                                task.status === TaskStatus.FAILED ? 'bg-red-400' :
                                    'bg-gradient-to-r from-blue-600 to-indigo-600 shadow-[0_0_8px_rgba(79,70,229,0.3)]']"
                            :style="{ width: task.progress + '%' }">
                            <!-- Shimmer effect for running tasks -->
                            <div v-if="isRunning(task.status)"
                                class="absolute inset-0 bg-white/20 skew-x-[-20deg] animate-[shimmer_2s_infinite] w-full">
                            </div>
                        </div>
                    </div>
                </div>

                <div :class="['flex items-center shrink-0', isWindowed ? 'flex-wrap gap-2' : 'space-x-2']">
                    <button @click="emit('viewLogs', task.id)"
                        class="w-10 h-10 flex items-center justify-center rounded-xl bg-white border border-slate-100 text-slate-400 hover:text-indigo-600 hover:bg-indigo-50 hover:border-indigo-200 transition-all hover:shadow-md"
                        title="查看运行日志">
                        <i class="fa-solid fa-terminal text-[10px]"></i>
                    </button>
                    <button @click="selectedTaskDetails = task"
                        class="w-10 h-10 flex items-center justify-center rounded-xl bg-white border border-slate-100 text-slate-400 hover:text-blue-600 hover:bg-blue-50 hover:border-blue-200 transition-all hover:shadow-md"
                        title="任务参数审计">
                        <i class="fa-solid fa-sliders text-[10px]"></i>
                    </button>
                    <div class="h-6 w-px bg-slate-100 mx-1"></div>
                    <button @click="runTask(task)"
                        :disabled="!isRunning(task.status) && task.status !== TaskStatus.IDLE && task.status !== TaskStatus.SUCCESS && task.status !== TaskStatus.FAILED"
                        :class="['px-6 py-2 rounded-xl text-[10px] font-black uppercase tracking-widest transition-all shadow-sm flex items-center space-x-2',
                            isRunning(task.status)
                                ? 'bg-blue-50 text-blue-400 cursor-not-allowed border border-blue-100'
                                : 'bg-slate-900 text-white hover:bg-blue-600 hover:shadow-blue-500/20 active:scale-95']">
                        <i v-if="isRunning(task.status)" class="fa-solid fa-spinner animate-spin"></i>
                        <i v-else class="fa-solid fa-bolt-lightning"></i>
                        <span>{{ isRunning(task.status) ? '执行中' : '触发' }}</span>
                    </button>
                </div>
            </div>
        </div>

        <!-- Create Modal -->
        <Teleport to="body">
            <div v-if="isCreateModalOpen"
                class="fixed inset-0 bg-slate-950/80 backdrop-blur-xl z-[70] flex items-center justify-center p-6">
                <div
                    class="bg-white w-full max-w-5xl h-[90vh] rounded-[2.5rem] shadow-2xl overflow-hidden border border-white/10 flex flex-col animate-in zoom-in-95 duration-300">
                    <!-- Header with Stepper -->
                    <div class="px-10 py-8 border-b border-slate-100 bg-slate-50/30 flex flex-col space-y-6">
                        <div class="flex justify-between items-center">
                            <div class="flex items-center space-x-4">
                                <div
                                    class="w-12 h-12 bg-blue-600 rounded-2xl flex items-center justify-center text-white shadow-lg">
                                    <i class="fa-solid fa-plus-circle text-xl"></i>
                                </div>
                                <div>
                                    <h2 class="text-xl font-black text-slate-800 tracking-tight">{{ editingTaskId ?
                                        '编辑部署工作流' : '构建新部署工作流' }}</h2>
                                    <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
                                        Automated Pipeline Designer</p>
                                </div>
                            </div>
                            <button @click="closeCreateModal"
                                class="w-10 h-10 rounded-full hover:bg-slate-200 flex items-center justify-center text-slate-400 transition-colors">
                                <i class="fa-solid fa-xmark text-lg"></i>
                            </button>
                        </div>

                        <!-- Stepper Progress Indicator -->
                        <div class="flex items-center justify-center space-x-4">
                            <div v-for="step in [{ id: 1, label: '基础定义' }, { id: 2, label: '拓扑节点' }, { id: 3, label: '路径与脚本' }]"
                                :key="step.id" class="flex items-center space-x-3">
                                <div
                                    :class="['w-8 h-8 rounded-full flex items-center justify-center text-xs font-black transition-all border-2',
                                        currentStep === step.id ? 'bg-blue-600 border-blue-600 text-white shadow-md' :
                                            currentStep > step.id ? 'bg-emerald-500 border-emerald-500 text-white' : 'bg-white border-slate-200 text-slate-300']">
                                    <i v-if="currentStep > step.id" class="fa-solid fa-check"></i>
                                    <span v-else>{{ step.id }}</span>
                                </div>
                                <span :class="['text-[10px] font-black uppercase tracking-widest',
                                    currentStep >= step.id ? 'text-slate-800' : 'text-slate-300']">
                                    {{ step.label }}
                                </span>
                                <div v-if="step.id < 3" class="w-12 h-0.5 bg-slate-100 rounded-full overflow-hidden">
                                    <div
                                        :class="['h-full bg-blue-500 transition-all duration-500', currentStep > step.id ? 'w-full' : 'w-0']">
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Content Area (Step-based) -->
                    <div class="flex-1 overflow-hidden flex flex-col">

                        <!-- Step 1: Basic Definition -->
                        <div v-if="currentStep === 1"
                            class="flex-1 overflow-y-auto p-12 space-y-10 animate-in fade-in duration-500">
                            <div class="max-w-3xl mx-auto space-y-10">
                                <div class="space-y-3">
                                    <label
                                        class="text-xs font-black text-slate-400 uppercase tracking-widest flex items-center space-x-2">
                                        <span
                                            class="w-5 h-5 bg-slate-100 rounded flex items-center justify-center text-[10px]">1</span>
                                        <span>任务流标识名称</span>
                                    </label>
                                    <input type="text" v-model="formData.name" placeholder="例如: 核心业务-API网关-生产发布"
                                        class="w-full px-8 py-5 bg-slate-50 border border-slate-100 rounded-2xl text-base font-bold focus:bg-white focus:border-blue-500 focus:ring-4 focus:ring-blue-500/5 outline-none transition-all shadow-inner" />
                                </div>

                                <div class="grid grid-cols-2 gap-8">
                                    <div class="space-y-3">
                                        <label
                                            class="text-xs font-black text-slate-400 uppercase tracking-widest flex items-center space-x-2">
                                            <span
                                                class="w-5 h-5 bg-slate-100 rounded flex items-center justify-center text-[10px]">2</span>
                                            <span>源码资源 (SVN)</span>
                                        </label>
                                        <div class="relative group">
                                            <select v-model="formData.svnResourceId"
                                                class="w-full pl-8 pr-12 py-5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold appearance-none outline-none focus:bg-white focus:border-blue-500 transition-all shadow-inner">
                                                <option v-for="r in resources" :key="r.id" :value="r.id">{{ r.name }}
                                                    (R-{{ r.revision }})</option>
                                            </select>
                                            <i
                                                class="fa-solid fa-chevron-down absolute right-6 top-1/2 -translate-y-1/2 text-slate-300 pointer-events-none group-focus-within:text-blue-500"></i>
                                        </div>
                                    </div>
                                    <div class="space-y-3">
                                        <label
                                            class="text-xs font-black text-slate-400 uppercase tracking-widest flex items-center space-x-2">
                                            <span
                                                class="w-5 h-5 bg-slate-100 rounded flex items-center justify-center text-[10px]">3</span>
                                            <span>主控分发节点</span>
                                        </label>
                                        <div class="relative group">
                                            <select v-model="formData.masterServerId"
                                                class="w-full pl-8 pr-12 py-5 bg-slate-50 border border-slate-100 rounded-2xl text-sm font-bold appearance-none outline-none focus:bg-white focus:border-blue-500 transition-all shadow-inner">
                                                <option v-for="s in masters" :key="s.id" :value="s.id">{{ s.name }} ({{
                                                    s.ip }})</option>
                                            </select>
                                            <i
                                                class="fa-solid fa-server absolute right-6 top-1/2 -translate-y-1/2 text-slate-200 pointer-events-none"></i>
                                        </div>
                                    </div>
                                </div>

                                <div
                                    class="p-6 bg-blue-50/50 rounded-3xl border border-blue-100 flex items-start space-x-4">
                                    <div
                                        class="w-10 h-10 bg-white rounded-xl flex items-center justify-center text-blue-500 shadow-sm shrink-0">
                                        <i class="fa-solid fa-lightbulb"></i>
                                    </div>
                                    <div>
                                        <p class="text-xs font-black text-blue-800">小提示</p>
                                        <p class="text-[11px] text-blue-600/70 mt-1 leading-relaxed">
                                            主控节点将作为文件分发的中转站。系统会先将代码同步至主控机，再由主控机分发至各个从机节点。
                                        </p>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Step 2: Topology & Nodes -->
                        <div v-else-if="currentStep === 2"
                            class="flex-1 overflow-hidden flex animate-in fade-in duration-500">
                            <!-- Left: Slaves Selection -->
                            <div class="w-1/2 overflow-y-auto p-12 border-r border-slate-100 space-y-8">
                                <div class="space-y-2">
                                    <label class="text-xs font-black text-slate-400 uppercase tracking-widest">选择分发目标
                                        (Slaves)</label>
                                    <p class="text-[10px] text-slate-400">点击服务器卡片进行勾选，可为特定机器指定独立部署路径</p>
                                </div>
                                <div class="grid grid-cols-1 gap-3">
                                    <div v-for="s in slaves" :key="s.id" @click="toggleSlaveSelection(s.id)"
                                        :class="['p-5 rounded-2xl border-2 cursor-pointer transition-all flex flex-col space-y-3 group',
                                            formData.slaveServerIds.includes(s.id) ? 'bg-blue-600 border-blue-600 shadow-xl shadow-blue-500/20' : 'bg-slate-50 border-slate-50 hover:border-blue-200 hover:bg-white']">
                                        <div class="flex items-center justify-between">
                                            <div class="flex items-center space-x-3">
                                                <div
                                                    :class="['w-10 h-10 rounded-xl flex items-center justify-center text-sm',
                                                        formData.slaveServerIds.includes(s.id) ? 'bg-white/20 text-white' : 'bg-white border border-slate-100 text-slate-400']">
                                                    <i class="fa-solid fa-server"></i>
                                                </div>
                                                <div>
                                                    <p
                                                        :class="['text-xs font-black', formData.slaveServerIds.includes(s.id) ? 'text-white' : 'text-slate-700']">
                                                        {{ s.name }}</p>
                                                    <p
                                                        :class="['text-[10px] font-mono', formData.slaveServerIds.includes(s.id) ? 'text-blue-100' : 'text-slate-400']">
                                                        {{ s.ip }}</p>
                                                </div>
                                            </div>
                                            <i v-if="formData.slaveServerIds.includes(s.id)"
                                                class="fa-solid fa-circle-check text-white text-lg"></i>
                                        </div>
                                        <div v-if="formData.slaveServerIds.includes(s.id)" class="pt-2">
                                            <input type="text" :value="formData.slaveRemotePaths[s.id] || ''"
                                                @click.stop
                                                @input="(e) => formData.slaveRemotePaths[s.id] = (e.target as HTMLInputElement).value"
                                                placeholder="自定义从机路径 (留空则继承默认)"
                                                class="w-full px-4 py-2 rounded-xl bg-white/10 text-[10px] font-mono text-white placeholder:text-white/40 border border-white/20 outline-none focus:bg-white/20" />
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <!-- Right: Visual Preview Hint -->
                            <div
                                class="w-1/2 bg-slate-50 p-12 flex flex-col justify-center items-center relative overflow-hidden">
                                <div class="relative z-10 text-center space-y-8">
                                    <div class="flex items-center justify-center">
                                        <!-- Visual Topology representation -->
                                        <div class="flex items-center">
                                            <div
                                                class="w-20 h-20 bg-slate-900 rounded-3xl flex items-center justify-center text-white shadow-2xl relative">
                                                <i class="fa-solid fa-crown text-2xl text-blue-400"></i>
                                                <span
                                                    class="absolute -top-3 px-2 py-1 bg-blue-600 text-[8px] font-black rounded-lg">MASTER</span>
                                            </div>
                                            <div
                                                class="w-24 h-0.5 bg-dashed border-b-2 border-slate-200 border-dashed relative">
                                                <i
                                                    class="fa-solid fa-angles-right absolute right-0 -top-1.5 text-slate-300"></i>
                                            </div>
                                            <div class="flex flex-col space-y-2">
                                                <div v-for="i in Math.min(3, formData.slaveServerIds.length || 1)"
                                                    :key="i"
                                                    class="w-12 h-12 bg-white border border-slate-200 rounded-xl flex items-center justify-center text-slate-400 shadow-sm">
                                                    <i class="fa-solid fa-server text-xs"></i>
                                                </div>
                                            </div>
                                        </div>
                                    </div>
                                    <div class="space-y-4 max-w-xs">
                                        <h4 class="text-sm font-black text-slate-800 uppercase tracking-widest">拓扑预览
                                        </h4>
                                        <p class="text-[11px] text-slate-500 leading-relaxed">
                                            您已选择 <span class="font-black text-blue-600">{{
                                                formData.slaveServerIds.length }}</span> 台目标从机。
                                            部署路径将自动从主控机同步至所有选中的节点。
                                        </p>
                                    </div>
                                </div>
                                <i
                                    class="fa-solid fa-network-wired absolute -right-20 -bottom-20 text-[300px] text-slate-200/50 -rotate-12"></i>
                            </div>
                        </div>

                        <!-- Step 3: Paths & Scripting -->
                        <div v-else-if="currentStep === 3"
                            class="flex-1 overflow-hidden flex animate-in fade-in duration-500">
                            <!-- Left: Path mapping -->
                            <div class="w-1/3 overflow-y-auto p-12 border-r border-slate-100 space-y-10">
                                <div class="space-y-6">
                                    <div class="space-y-3">
                                        <label
                                            class="text-[10px] font-black text-slate-400 uppercase tracking-widest">主控机部署绝对路径</label>
                                        <div class="relative">
                                            <i
                                                class="fa-solid fa-folder absolute left-5 top-1/2 -translate-y-1/2 text-blue-400"></i>
                                            <input type="text" v-model="formData.remotePath"
                                                placeholder="/var/www/my-project"
                                                class="w-full pl-12 pr-6 py-4 bg-slate-50 border border-slate-100 rounded-2xl font-mono text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all shadow-inner" />
                                        </div>
                                    </div>
                                    <div class="space-y-3">
                                        <label
                                            class="text-[10px] font-black text-slate-400 uppercase tracking-widest">从机默认部署路径</label>
                                        <div class="relative">
                                            <i
                                                class="fa-solid fa-folder-tree absolute left-5 top-1/2 -translate-y-1/2 text-slate-300"></i>
                                            <input type="text" v-model="formData.slaveRemotePath" placeholder="留空则与主控一致"
                                                class="w-full pl-12 pr-6 py-4 bg-slate-50 border border-slate-100 rounded-2xl font-mono text-xs font-bold text-slate-700 outline-none focus:bg-white focus:border-blue-500 transition-all shadow-inner" />
                                        </div>
                                    </div>
                                </div>

                                <div class="p-6 bg-slate-900 rounded-3xl space-y-4">
                                    <p class="text-[9px] font-black text-blue-400 uppercase tracking-widest">可用环境变量</p>
                                    <div class="space-y-2">
                                        <div class="flex justify-between text-[10px] font-mono">
                                            <span class="text-slate-500">$REMOTE_PATH</span>
                                            <span class="text-slate-400">部署根路径</span>
                                        </div>
                                        <div class="flex justify-between text-[10px] font-mono">
                                            <span class="text-slate-500">$SVN_REV</span>
                                            <span class="text-slate-400">SVN版本号</span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                            <!-- Right: Enhanced Script Editor -->
                            <div class="w-2/3 bg-[#0d1117] flex flex-col relative group">
                                <div
                                    class="px-10 py-8 border-b border-white/5 flex justify-between items-center bg-white/5">
                                    <div class="flex items-center space-x-2">
                                        <i class="fa-solid fa-code text-emerald-500"></i>
                                        <label
                                            class="text-[10px] font-black text-slate-400 uppercase tracking-widest">后置处理脚本
                                            (BASH)</label>
                                    </div>
                                    <span
                                        class="text-[9px] font-mono text-emerald-500/50 bg-emerald-500/10 px-2 py-1 rounded">Syntax
                                        Highlighting On</span>
                                </div>
                                <div
                                    class="flex-1 p-10 font-mono text-[11px] text-emerald-400 relative overflow-hidden">
                                    <div
                                        class="absolute left-0 top-0 bottom-0 w-1 bg-blue-500 opacity-0 group-focus-within:opacity-100 transition-opacity">
                                    </div>
                                    <textarea v-model="formData.commands" spellcheck="false"
                                        class="w-full h-full bg-transparent outline-none resize-none leading-relaxed placeholder:text-slate-700 custom-scrollbar"
                                        placeholder="# 编写发布后的自动化指令，例如：\nsync_config.sh\npm2 restart app\nrm -rf /tmp/build"></textarea>
                                </div>
                                <div class="p-6 bg-blue-500/5 border-t border-white/5">
                                    <p class="text-[9px] text-slate-500 italic leading-normal">
                                        安全提示：所有指令将以部署用户身份由主控机通过 SSH 广播至从机执行。
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Footer with Navigation -->
                    <div class="px-10 py-8 border-t border-slate-100 bg-slate-50/50 flex justify-between items-center">
                        <div>
                            <button v-if="currentStep > 1" @click="currentStep--"
                                class="px-8 py-3 text-xs font-black text-slate-500 hover:text-slate-800 uppercase tracking-widest transition-all transition-colors flex items-center space-x-2">
                                <i class="fa-solid fa-arrow-left"></i>
                                <span>上一步</span>
                            </button>
                            <button v-else @click="closeCreateModal"
                                class="px-8 py-3 text-xs font-black text-slate-400 hover:text-red-500 uppercase tracking-widest transition-colors">
                                放弃构建
                            </button>
                        </div>
                        <div class="flex items-center space-x-4">
                            <span class="text-[10px] font-black text-slate-300 uppercase hidden md:inline">Step {{
                                currentStep }} / 3</span>
                            <button v-if="currentStep < 3" @click="currentStep++"
                                class="px-12 py-3 bg-blue-600 text-white rounded-2xl text-xs font-black shadow-lg hover:bg-blue-500 hover:-translate-y-1 transition-all uppercase tracking-widest flex items-center space-x-2">
                                <span>下一步</span>
                                <i class="fa-solid fa-arrow-right"></i>
                            </button>
                            <button v-else @click="handleCreateTask"
                                class="px-14 py-3 bg-slate-900 text-white rounded-2xl text-xs font-black shadow-xl hover:bg-emerald-600 hover:-translate-y-1 active:translate-y-0 transition-all uppercase tracking-widest flex items-center space-x-2">
                                <i class="fa-solid fa-rocket"></i>
                                <span>{{ editingTaskId ? '确认并保存修改' : '生成自动化流水线' }}</span>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
        </Teleport>

        <!-- Details Modal -->
        <Teleport to="body">
            <div v-if="selectedTaskDetails" class="fixed inset-0 z-[80] flex justify-end">
                <div class="absolute inset-0 bg-slate-950/40 backdrop-blur-[2px]" @click="selectedTaskDetails = null">
                </div>
                <div
                    class="relative w-full max-w-xl bg-white h-full shadow-2xl flex flex-col animate-in slide-in-from-right duration-500 ease-out border-l border-slate-200">
                    <div
                        class="p-10 border-b border-slate-100 flex justify-between items-center shrink-0 bg-slate-50/50">
                        <div class="flex items-center space-x-4">
                            <div
                                class="w-12 h-12 bg-slate-100 rounded-2xl flex items-center justify-center text-slate-400 border border-slate-200 shadow-inner">
                                <i class="fa-solid fa-magnifying-glass-chart text-xl"></i>
                            </div>
                            <div>
                                <h3 class="text-xl font-black text-slate-800 tracking-tight">编排配置审计</h3>
                                <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">Static
                                    Configuration Analysis</p>
                            </div>
                        </div>
                        <button @click="selectedTaskDetails = null"
                            class="w-10 h-10 rounded-full hover:bg-slate-200 flex items-center justify-center text-slate-400 transition-colors">
                            <i class="fa-solid fa-arrow-right"></i>
                        </button>
                    </div>

                    <div v-if="selectedTaskDetails" class="flex-1 overflow-y-auto p-10 space-y-10 custom-scrollbar">
                        <section class="space-y-5">
                            <h5
                                class="text-[10px] font-black text-slate-400 uppercase tracking-widest border-b border-slate-100 pb-3 flex items-center space-x-2">
                                <i class="fa-solid fa-circle-info text-blue-500"></i>
                                <span>流水线元数据</span>
                            </h5>
                            <div class="grid grid-cols-2 gap-8">
                                <div class="space-y-1">
                                    <p class="text-[9px] text-slate-400 font-bold uppercase">任务全称</p>
                                    <p class="text-sm font-black text-slate-800">{{ selectedTaskDetails.name }}</p>
                                </div>
                                <div class="space-y-1">
                                    <p class="text-[9px] text-slate-400 font-bold uppercase">最后执行状态</p>
                                    <span
                                        :class="['text-[9px] font-black uppercase px-2 py-0.5 rounded border',
                                            selectedTaskDetails.status === TaskStatus.SUCCESS ? 'bg-emerald-50 text-emerald-600 border-emerald-100' : 'bg-blue-50 text-blue-600 border-blue-100']">
                                        {{ selectedTaskDetails.status }}
                                    </span>
                                </div>
                                <div
                                    class="col-span-2 space-y-1 p-4 bg-blue-50/30 rounded-2xl border border-blue-100/50">
                                    <p class="text-[9px] text-blue-400 font-bold uppercase mb-2">部署绝对路径映射</p>
                                    <p class="text-xs font-mono font-black text-blue-700 truncate select-all">{{
                                        selectedTaskDetails.remotePath }}</p>
                                    <p v-if="selectedTaskDetails.slaveRemotePath"
                                        class="text-[9px] text-blue-400 font-bold uppercase mt-3 mb-1">从机部署路径</p>
                                    <p v-if="selectedTaskDetails.slaveRemotePath"
                                        class="text-[10px] font-mono font-black text-blue-600 truncate select-all">
                                        {{ selectedTaskDetails.slaveRemotePath }}</p>
                                </div>
                            </div>
                        </section>
                        <section class="space-y-5">
                            <h5
                                class="text-[10px] font-black text-slate-400 uppercase tracking-widest border-b border-slate-100 pb-3 flex items-center space-x-2">
                                <i class="fa-solid fa-network-wired text-indigo-500"></i>
                                <span>集群拓扑分布</span>
                            </h5>
                            <div class="space-y-3">
                                <div
                                    class="p-4 bg-slate-900 rounded-2xl border border-slate-800 flex items-center justify-between shadow-lg">
                                    <div class="flex items-center space-x-4">
                                        <div
                                            class="w-10 h-10 bg-blue-600 rounded-xl flex items-center justify-center text-white shadow-[0_0_15px_rgba(37,99,235,0.4)]">
                                            <i class="fa-solid fa-crown text-sm"></i>
                                        </div>
                                        <div>
                                            <p class="text-[9px] text-blue-400 font-bold uppercase">中转主控节点</p>
                                            <p class="text-xs font-black text-white">{{
                                                servers.find(s => s.id ===
                                                    selectedTaskDetails!.masterServerId)?.name}}</p>
                                        </div>
                                    </div>
                                    <p class="text-[10px] font-mono font-bold text-slate-500">{{
                                        servers.find(s => s.id
                                            === selectedTaskDetails!.masterServerId)?.ip}}</p>
                                </div>
                                <div class="p-5 bg-slate-50 rounded-2xl border border-slate-200">
                                    <p class="text-[9px] text-slate-400 font-bold uppercase mb-4">广播目标从机组</p>
                                    <div class="flex flex-wrap gap-2">
                                        <div v-for="sid in selectedTaskDetails.slaveServerIds" :key="sid"
                                            class="px-3 py-1.5 bg-white border border-slate-200 rounded-xl flex items-center space-x-2 shadow-sm">
                                            <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                            <span class="text-[10px] font-black text-slate-700">{{servers.find(s =>
                                                s.id === sid)?.name}}</span>
                                        </div>
                                        <div v-if="selectedTaskDetails.slaveServerIds.length === 0"
                                            class="text-[10px] text-slate-400 italic">未配置任何从节点</div>
                                    </div>
                                </div>
                                <div v-if="selectedTaskDetails.slaveServerIds.length > 0"
                                    class="p-5 bg-white rounded-2xl border border-slate-200">
                                    <p class="text-[9px] text-slate-400 font-bold uppercase mb-3">从机部署路径映射</p>
                                    <div class="space-y-2">
                                        <div v-for="sid in selectedTaskDetails.slaveServerIds" :key="sid"
                                            class="flex items-center justify-between text-[10px] font-mono">
                                            <span class="text-slate-500">
                                                {{servers.find(s => s.id === sid)?.name || sid}}
                                            </span>
                                            <span class="text-slate-700">
                                                {{ (selectedTaskDetails.slaveRemotePaths &&
                                                    selectedTaskDetails.slaveRemotePaths[sid]) ||
                                                    selectedTaskDetails.slaveRemotePath || selectedTaskDetails.remotePath }}
                                            </span>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </section>

                        <section class="space-y-5">
                            <h5
                                class="text-[10px] font-black text-slate-400 uppercase tracking-widest border-b border-slate-100 pb-3 flex items-center space-x-2">
                                <i class="fa-solid fa-code text-emerald-500"></i>
                                <span>执行指令序列预览</span>
                            </h5>
                            <div
                                class="bg-[#0d1117] rounded-[2rem] p-8 font-mono text-[11px] text-emerald-400 leading-relaxed shadow-2xl border border-white/5 relative">
                                <template v-if="selectedTaskDetails.commands.length > 0">
                                    <div v-for="(cmd, i) in selectedTaskDetails.commands" :key="i"
                                        class="flex space-x-6 py-0.5 group">
                                        <span
                                            class="text-slate-700 w-4 text-right select-none font-bold group-hover:text-blue-500">{{
                                                i + 1 }}</span>
                                        <span class="whitespace-pre-wrap">{{ cmd }}</span>
                                    </div>
                                </template>
                                <p v-else class="text-slate-600 italic">无自定义 Shell 指令</p>
                            </div>
                        </section>
                    </div>

                    <div
                        class="p-10 border-t border-slate-100 bg-slate-50/80 flex items-center justify-between shrink-0">
                        <button @click="handleDeleteTask(selectedTaskDetails!)"
                            class="text-xs font-black text-red-500 hover:bg-red-50 px-6 py-3 rounded-xl transition-all uppercase tracking-widest">删除工作流</button>
                        <button @click="handleSaveTemplate"
                            class="text-xs font-black text-slate-600 hover:bg-slate-100 px-6 py-3 rounded-xl transition-all uppercase tracking-widest">保存为模板</button>
                        <button @click="handleEditTask(selectedTaskDetails!)"
                            class="text-xs font-black text-slate-600 hover:bg-slate-100 px-6 py-3 rounded-xl transition-all uppercase tracking-widest">编辑配置</button>
                        <button @click="runTask(selectedTaskDetails!); selectedTaskDetails = null"
                            class="bg-slate-900 text-white px-12 py-3 rounded-2xl text-xs font-black shadow-xl hover:bg-blue-600 transition-all active:scale-95 uppercase tracking-widest">
                            确认并立即触发
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>

        <!-- Template Modal -->
        <Teleport to="body">
            <div v-if="isTemplateModalOpen"
                class="fixed inset-0 bg-slate-950/80 backdrop-blur-xl z-[70] flex items-center justify-center p-6">
                <div
                    class="bg-white w-full max-w-2xl rounded-[2rem] shadow-2xl overflow-hidden border border-white/10 flex flex-col animate-in zoom-in-95 duration-300">
                    <div class="px-8 py-6 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
                        <div class="flex items-center space-x-3">
                            <div class="w-10 h-10 bg-indigo-600 rounded-xl flex items-center justify-center text-white">
                                <i class="fa-solid fa-clone text-base"></i>
                            </div>
                            <div>
                                <h3 class="text-lg font-black text-slate-800 tracking-tight">选择模板</h3>
                                <p class="text-[10px] font-bold text-slate-400 uppercase tracking-[0.2em] mt-1">Clone
                                    from
                                    saved templates</p>
                            </div>
                        </div>
                        <button @click="isTemplateModalOpen = false"
                            class="w-10 h-10 rounded-full hover:bg-slate-200 flex items-center justify-center text-slate-400 transition-colors">
                            <i class="fa-solid fa-xmark text-lg"></i>
                        </button>
                    </div>
                    <div class="p-8 space-y-4 max-h-[60vh] overflow-y-auto custom-scrollbar">
                        <div v-if="templates.length === 0"
                            class="p-8 border border-dashed border-slate-200 rounded-2xl text-center text-slate-400 text-sm">
                            暂无模板，请先从任务详情中保存模板。
                        </div>
                        <div v-for="tpl in templates" :key="tpl.id"
                            class="border border-slate-200 rounded-2xl p-4 flex items-center justify-between hover:border-indigo-200 hover:bg-indigo-50/40 transition-all">
                            <label class="flex items-center space-x-3 cursor-pointer">
                                <input type="radio" class="accent-indigo-600" :value="tpl.id"
                                    v-model="selectedTemplateId" />
                                <div>
                                    <p class="text-sm font-black text-slate-800">{{ tpl.name }}</p>
                                    <p class="text-[10px] font-mono text-slate-400 mt-1">{{ tpl.remotePath }}</p>
                                    <p class="text-[9px] font-mono text-slate-300 mt-1">
                                        {{
                                            tpl.slaveRemotePaths && Object.keys(tpl.slaveRemotePaths).length > 0
                                                ? '从机路径：自定义'
                                                : (tpl.slaveRemotePath ? tpl.slaveRemotePath : '从机路径同主控')
                                        }}
                                    </p>
                                </div>
                            </label>
                            <button @click="emit('deleteTemplate', tpl.id)"
                                class="w-9 h-9 rounded-xl border border-slate-200 text-slate-400 hover:text-red-500 hover:border-red-200 hover:bg-red-50 transition-all">
                                <i class="fa-solid fa-trash-can text-xs"></i>
                            </button>
                        </div>
                    </div>
                    <div class="px-8 py-6 border-t border-slate-100 bg-slate-50/80 flex justify-end space-x-4">
                        <button @click="isTemplateModalOpen = false"
                            class="px-6 py-2 text-xs font-black text-slate-400 hover:text-slate-600 uppercase tracking-widest transition-colors">
                            取消
                        </button>
                        <button @click="handleCloneFromTemplate" :disabled="!selectedTemplateId"
                            class="px-8 py-2 bg-slate-900 text-white rounded-2xl text-xs font-black shadow-xl hover:bg-indigo-600 transition-all disabled:opacity-40 disabled:cursor-not-allowed uppercase tracking-widest">
                            使用模板创建
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>

<style scoped>
@keyframes shimmer {
    0% {
        transform: translateX(-100%) skewX(-20deg);
    }

    100% {
        transform: translateX(200%) skewX(-20deg);
    }
}

.custom-scrollbar::-webkit-scrollbar {
    width: 6px;
    height: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
    background: #e2e8f0;
    border-radius: 10px;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: #cbd5e1;
}

/* Ensure pulse animation is smooth */
@keyframes pulse-subtle {

    0%,
    100% {
        opacity: 1;
    }

    50% {
        opacity: 0.5;
    }
}

.animate-pulse-subtle {
    animation: pulse-subtle 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style>
