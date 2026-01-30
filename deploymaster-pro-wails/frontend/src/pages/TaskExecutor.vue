<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { DeploymentTask, RemoteServer, SVNResource, TaskStatus } from '../types';

const props = defineProps<{
    tasks: DeploymentTask[];
    servers: RemoteServer[];
    resources: SVNResource[];
    autoOpenModal?: boolean;
}>();

const emit = defineEmits(['addTask', 'updateTask', 'modalClose']);

const isCreateModalOpen = ref(false);
const selectedTaskDetails = ref<DeploymentTask | null>(null);

const initialFormState = () => ({
    name: '',
    svnResourceId: props.resources[0]?.id || '',
    masterServerId: props.servers.find(s => s.isMaster)?.id || '',
    slaveServerIds: [] as string[],
    remotePath: '',
    commands: ''
});

const formData = ref(initialFormState());

watch(() => props.autoOpenModal, (newVal) => {
    if (newVal) {
        isCreateModalOpen.value = true;
        emit('modalClose');
    }
});

const masters = computed(() => props.servers.filter(s => s.isMaster));
const slaves = computed(() => props.servers.filter(s => !s.isMaster));

const handleCreateTask = () => {
    if (!formData.value.name || !formData.value.remotePath || !formData.value.masterServerId) {
        alert('请检查：任务名称、主节点及远程路径为必填项');
        return;
    }

    const newTask: DeploymentTask = {
        id: 'task-' + Math.random().toString(36).substring(2, 9),
        name: formData.value.name,
        svnResourceId: formData.value.svnResourceId,
        masterServerId: formData.value.masterServerId,
        slaveServerIds: formData.value.slaveServerIds,
        remotePath: formData.value.remotePath,
        commands: formData.value.commands.split('\n').map(c => c.trim()).filter(c => c),
        status: TaskStatus.IDLE,
        progress: 0,
        logs: [`[${new Date().toLocaleTimeString()}] [系统] 初始化任务流成功，等待用户触发执行。`]
    };

    emit('addTask', newTask);
    isCreateModalOpen.value = false;
    formData.value = initialFormState();
};

const simulateRun = async (task: DeploymentTask) => {
    if (task.status !== TaskStatus.IDLE && task.status !== TaskStatus.FAILED && task.status !== TaskStatus.SUCCESS) return;

    const stages = [
        { status: TaskStatus.DOWNLOADING, progress: 15, log: '正在建立 SVN 安全隧道并获取修订号 R-' + (props.resources.find(r => r.id === task.svnResourceId)?.revision || 'HEAD') + ' ...' },
        { status: TaskStatus.DOWNLOADING, progress: 30, log: 'SVN 资源检出完成。MD5: ' + Math.random().toString(16).substring(2, 8) + '。' },
        { status: TaskStatus.UPLOADING, progress: 45, log: '正在通过 ' + (props.servers.find(s => s.id === task.masterServerId)?.protocol || 'SFTP') + ' 上传资源包至主控机...' },
        { status: TaskStatus.SYNCING, progress: 65, log: '主控机正在通过 P2P 协议向 ' + task.slaveServerIds.length + ' 台从机广播增量数据...' },
        { status: TaskStatus.EXECUTING, progress: 85, log: '数据一致性校验通过。正在启动远程自定义脚本执行序列...' },
        { status: TaskStatus.SUCCESS, progress: 100, log: '✓ 任务执行成功。所有节点已同步至最新状态。' }
    ];

    let currentLogs = [...task.logs, `[${new Date().toLocaleTimeString()}] [信息] 启动自动化分发流水线...`];
    emit('updateTask', { ...task, status: TaskStatus.DOWNLOADING, progress: 5, logs: currentLogs });

    for (const stage of stages) {
        await new Promise(r => setTimeout(r, 1200 + Math.random() * 800));
        currentLogs = [...currentLogs, `[${new Date().toLocaleTimeString()}] ${stage.log}`];
        emit('updateTask', {
            ...task,
            status: stage.status,
            progress: stage.progress,
            logs: currentLogs
        });
    }
};

const toggleSlaveSelection = (id: string) => {
    const index = formData.value.slaveServerIds.indexOf(id);
    if (index === -1) {
        formData.value.slaveServerIds.push(id);
    } else {
        formData.value.slaveServerIds.splice(index, 1);
    }
};
</script>

<template>
    <div class="space-y-6 pb-10">
        <!-- Banner -->
        <div class="bg-slate-900 rounded-3xl p-10 text-white relative overflow-hidden border border-white/5 shadow-2xl">
            <div class="relative z-10 max-w-2xl">
                <div class="flex items-center space-x-4 mb-6">
                    <div
                        class="w-14 h-14 bg-gradient-to-tr from-blue-600 to-indigo-600 rounded-2xl flex items-center justify-center shadow-[0_0_30px_rgba(37,99,235,0.4)]">
                        <i class="fa-solid fa-wand-magic-sparkles text-2xl"></i>
                    </div>
                    <div>
                        <h2 class="text-2xl font-black tracking-tight">任务编排中心</h2>
                        <p class="text-[10px] font-black text-blue-400 uppercase tracking-[0.3em] mt-1">Full-Cycle
                            Automation Orchestrator</p>
                    </div>
                </div>
                <p class="text-slate-400 text-sm leading-relaxed mb-8 font-medium">
                    在这里定义复杂的部署逻辑。您可以选择 SVN 源码库，指定主控分发节点及目标从机群组，并编写在远程服务器上运行的 Bash 指令序列。系统将自动处理并发分发和日志采集。
                </p>
                <div class="flex space-x-4">
                    <button @click="isCreateModalOpen = true"
                        class="bg-blue-600 text-white px-8 py-3 rounded-xl font-black text-xs shadow-xl hover:bg-blue-500 hover:-translate-y-0.5 active:translate-y-0 transition-all flex items-center space-x-3">
                        <i class="fa-solid fa-plus-circle"></i>
                        <span>构建新工作流</span>
                    </button>
                    <button
                        class="bg-white/5 border border-white/10 hover:bg-white/10 px-8 py-3 rounded-xl font-bold text-xs transition-all text-slate-300 flex items-center space-x-2">
                        <i class="fa-solid fa-clone"></i>
                        <span>从模板克隆</span>
                    </button>
                </div>
            </div>
            <i
                class="fa-solid fa-code-branch absolute -right-12 -bottom-12 text-[260px] text-white/5 -rotate-12 pointer-events-none"></i>
        </div>

        <!-- List Header -->
        <div class="flex items-center justify-between px-2">
            <div class="flex items-center space-x-3">
                <h3 class="text-sm font-black text-slate-800 uppercase tracking-widest">当前队列任务</h3>
                <span class="bg-slate-200 text-slate-600 text-[9px] font-black px-2 py-0.5 rounded-full">{{ tasks.length
                    }}</span>
            </div>
            <div class="flex items-center space-x-4">
                <div class="flex items-center space-x-2 text-[10px] font-bold text-slate-400">
                    <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                    <span>就绪</span>
                    <span class="w-2 h-2 rounded-full bg-blue-500 ml-2"></span>
                    <span>运行中</span>
                </div>
            </div>
        </div>

        <!-- Task List -->
        <div class="grid grid-cols-1 gap-4">
            <div v-if="tasks.length === 0"
                class="bg-white border-2 border-dashed border-slate-200 rounded-2xl p-20 text-center flex flex-col items-center justify-center space-y-4">
                <div class="w-16 h-16 bg-slate-50 rounded-full flex items-center justify-center text-slate-200">
                    <i class="fa-solid fa-wind text-3xl"></i>
                </div>
                <div>
                    <p class="text-sm font-black text-slate-400 uppercase tracking-widest">暂无编排任务</p>
                    <p class="text-xs text-slate-300 mt-1">点击上方“构建新工作流”开始您的第一次分发</p>
                </div>
            </div>
            <div v-for="task in tasks" :key="task.id"
                class="bg-white rounded-2xl border border-slate-200 p-6 shadow-sm hover:shadow-xl hover:border-blue-200 transition-all group flex items-center space-x-8">
                <div :class="['w-14 h-14 rounded-2xl flex items-center justify-center text-2xl font-black shrink-0 shadow-inner border-2 transition-all',
                    task.status === TaskStatus.SUCCESS ? 'bg-emerald-50 border-emerald-100 text-emerald-500' :
                        task.status === TaskStatus.IDLE ? 'bg-slate-50 border-slate-100 text-slate-300' :
                            'bg-blue-50 border-blue-200 text-blue-500 animate-pulse']">
                    <i v-if="task.status === TaskStatus.SUCCESS" class="fa-solid fa-check-double"></i>
                    <i v-else class="fa-solid fa-cube"></i>
                </div>

                <div class="flex-1 min-w-0">
                    <div class="flex items-center space-x-3 mb-2">
                        <h4 class="font-black text-slate-800 text-base truncate">{{ task.name }}</h4>
                        <span :class="['text-[9px] font-black px-2 py-0.5 rounded-full border uppercase tracking-tighter shadow-sm',
                            task.status === TaskStatus.SUCCESS ? 'bg-emerald-500 border-emerald-600 text-white' :
                                task.status === TaskStatus.IDLE ? 'bg-white border-slate-200 text-slate-400' :
                                    'bg-blue-600 border-blue-700 text-white']">
                            {{ task.status === TaskStatus.IDLE ? '已就绪' : task.status }}
                        </span>
                    </div>
                    <div class="flex items-center space-x-6 text-[11px] text-slate-400 font-bold">
                        <span class="flex items-center space-x-2"><i class="fa-solid fa-folder-tree text-blue-400"></i>
                            <span class="text-slate-600">{{ task.remotePath }}</span></span>
                        <span class="flex items-center space-x-2"><i class="fa-solid fa-server text-indigo-400"></i>
                            <span class="text-slate-600">{{ task.slaveServerIds.length }} 台从机</span></span>
                    </div>
                </div>

                <div class="w-64 shrink-0 px-4">
                    <div class="flex justify-between mb-2 items-end">
                        <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest">总体进度</span>
                        <span class="text-xs font-mono font-black text-slate-800">{{ task.progress }}%</span>
                    </div>
                    <div class="w-full bg-slate-100 rounded-full h-2.5 overflow-hidden shadow-inner">
                        <div :class="['h-full transition-all duration-700 ease-out',
                            task.status === TaskStatus.SUCCESS ? 'bg-emerald-500' : 'bg-gradient-to-r from-blue-600 to-indigo-600 shadow-[0_0_10px_rgba(79,70,229,0.4)]']"
                            :style="{ width: task.progress + '%' }"></div>
                    </div>
                </div>

                <div class="flex items-center space-x-3 shrink-0">
                    <button @click="selectedTaskDetails = task"
                        class="w-11 h-11 flex items-center justify-center rounded-xl bg-slate-50 border border-slate-200 text-slate-400 hover:text-blue-600 hover:bg-blue-50 hover:border-blue-200 transition-all shadow-sm"
                        title="查看任务参数详情">
                        <i class="fa-solid fa-sliders text-sm"></i>
                    </button>
                    <button @click="simulateRun(task)"
                        :disabled="task.status !== TaskStatus.IDLE && task.status !== TaskStatus.SUCCESS" :class="['px-8 py-2.5 rounded-xl text-xs font-black uppercase tracking-widest transition-all shadow-lg flex items-center space-x-3',
                            (task.status !== TaskStatus.IDLE && task.status !== TaskStatus.SUCCESS)
                                ? 'bg-slate-100 text-slate-300 cursor-not-allowed shadow-none'
                                : 'bg-slate-900 text-white hover:bg-blue-600 hover:shadow-blue-200 active:scale-95']">
                        <i class="fa-solid fa-bolt-lightning text-[10px]"></i>
                        <span>立即执行</span>
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
                    <!-- Header -->
                    <div class="px-10 py-8 border-b border-slate-100 flex justify-between items-center bg-slate-50/50">
                        <div class="flex items-center space-x-4">
                            <div
                                class="w-14 h-14 bg-blue-600 rounded-2xl flex items-center justify-center text-white shadow-2xl">
                                <i class="fa-solid fa-plus-circle text-2xl"></i>
                            </div>
                            <div>
                                <h2 class="text-2xl font-black text-slate-800 tracking-tight">配置新部署工作流</h2>
                                <p class="text-[10px] font-bold text-slate-400 uppercase tracking-[0.2em] mt-1">Design
                                    your automated deployment pipeline</p>
                            </div>
                        </div>
                        <button @click="isCreateModalOpen = false"
                            class="w-12 h-12 rounded-full hover:bg-slate-200 flex items-center justify-center text-slate-400 transition-colors">
                            <i class="fa-solid fa-xmark text-xl"></i>
                        </button>
                    </div>

                    <!-- Content Area -->
                    <div class="flex-1 overflow-hidden flex">
                        <!-- Left: Config -->
                        <div class="flex-1 overflow-y-auto p-10 space-y-8 border-r border-slate-100">
                            <div class="grid grid-cols-1 gap-8">
                                <div class="space-y-2">
                                    <label
                                        class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">1.
                                        任务流基本标识</label>
                                    <input type="text" v-model="formData.name" placeholder="例如: 支付网关-生产环境-V2.5 滚动更新"
                                        class="w-full px-6 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold focus:ring-4 focus:ring-blue-500/10 focus:bg-white focus:border-blue-500 outline-none transition-all" />
                                </div>

                                <div class="grid grid-cols-2 gap-6">
                                    <div class="space-y-2">
                                        <label
                                            class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">2.
                                            源码资源 (SVN)</label>
                                        <div class="relative">
                                            <select v-model="formData.svnResourceId"
                                                class="w-full px-6 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold appearance-none outline-none focus:border-blue-500">
                                                <option v-for="r in resources" :key="r.id" :value="r.id">{{ r.name }}
                                                    (R-{{ r.revision }})</option>
                                            </select>
                                            <i
                                                class="fa-solid fa-chevron-down absolute right-6 top-5 text-slate-300 pointer-events-none"></i>
                                        </div>
                                    </div>
                                    <div class="space-y-2">
                                        <label
                                            class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">3.
                                            主控分发机 (Master)</label>
                                        <div class="relative">
                                            <select v-model="formData.masterServerId"
                                                class="w-full px-6 py-4 bg-slate-50 border border-slate-200 rounded-2xl text-sm font-bold appearance-none outline-none focus:border-blue-500">
                                                <option v-for="s in masters" :key="s.id" :value="s.id">{{ s.name }} ({{
                                                    s.ip }})</option>
                                            </select>
                                            <i
                                                class="fa-solid fa-chevron-down absolute right-6 top-5 text-slate-300 pointer-events-none"></i>
                                        </div>
                                    </div>
                                </div>

                                <div class="space-y-2">
                                    <label
                                        class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">4.
                                        远程部署根路径</label>
                                    <div class="relative group">
                                        <i class="fa-solid fa-folder absolute left-6 top-4.5 text-blue-400 mt-0.5"></i>
                                        <input type="text" v-model="formData.remotePath"
                                            placeholder="/var/www/html/project_root"
                                            class="w-full pl-14 pr-6 py-4 bg-blue-50/30 border border-blue-100 rounded-2xl font-mono text-xs font-bold text-blue-700 outline-none focus:bg-white focus:border-blue-500 transition-all" />
                                    </div>
                                </div>

                                <div class="space-y-3">
                                    <label
                                        class="text-[10px] font-black text-slate-500 uppercase tracking-widest ml-1">5.
                                        分发目标节点 (Slaves)</label>
                                    <div class="grid grid-cols-2 gap-3 max-h-48 overflow-y-auto pr-2 custom-scrollbar">
                                        <div v-for="s in slaves" :key="s.id" @click="toggleSlaveSelection(s.id)"
                                            :class="['p-4 rounded-2xl border-2 cursor-pointer transition-all flex items-center justify-between group',
                                                formData.slaveServerIds.includes(s.id) ? 'bg-blue-600 border-blue-600 shadow-lg shadow-blue-500/20' : 'bg-slate-50 border-slate-200 hover:border-blue-200 hover:bg-white']">
                                            <div class="flex items-center space-x-3">
                                                <div
                                                    :class="['w-8 h-8 rounded-lg flex items-center justify-center text-xs',
                                                        formData.slaveServerIds.includes(s.id) ? 'bg-white/20 text-white' : 'bg-white border border-slate-100 text-slate-400']">
                                                    <i class="fa-solid fa-server"></i>
                                                </div>
                                                <div>
                                                    <p
                                                        :class="['text-[11px] font-black', formData.slaveServerIds.includes(s.id) ? 'text-white' : 'text-slate-700']">
                                                        {{ s.name }}</p>
                                                    <p
                                                        :class="['text-[9px] font-mono', formData.slaveServerIds.includes(s.id) ? 'text-blue-100' : 'text-slate-400']">
                                                        {{ s.ip }}</p>
                                                </div>
                                            </div>
                                            <i v-if="formData.slaveServerIds.includes(s.id)"
                                                class="fa-solid fa-circle-check text-white text-sm"></i>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <!-- Right: Commands -->
                        <div class="flex-1 bg-slate-900 flex flex-col relative">
                            <div class="px-8 py-10 flex flex-col h-full">
                                <div class="flex items-center justify-between mb-4">
                                    <label class="text-[10px] font-black text-slate-400 uppercase tracking-[0.2em]">6.
                                        后置自定义脚本 (Bash)</label>
                                    <span
                                        class="text-[9px] font-mono text-emerald-500 flex items-center space-x-2 bg-emerald-500/10 px-2 py-1 rounded">
                                        <span class="w-1 h-1 bg-emerald-500 rounded-full animate-pulse"></span>
                                        <span>AUTO-SAVE ON</span>
                                    </span>
                                </div>
                                <div
                                    class="flex-1 bg-[#0d1117] rounded-3xl border border-white/5 shadow-inner p-6 font-mono text-[11px] text-emerald-400 relative overflow-hidden group">
                                    <div
                                        class="absolute top-0 left-0 w-1 bg-blue-600 h-full opacity-0 group-focus-within:opacity-100 transition-opacity">
                                    </div>
                                    <textarea v-model="formData.commands" spellcheck="false"
                                        class="w-full h-full bg-transparent outline-none resize-none leading-relaxed placeholder:text-slate-700 custom-scrollbar"
                                        placeholder="# 每行一个指令..."></textarea>
                                </div>
                                <div
                                    class="mt-6 p-5 bg-blue-500/10 border border-blue-500/20 rounded-2xl flex items-start space-x-4">
                                    <i class="fa-solid fa-shield-halved text-blue-400 text-lg"></i>
                                    <p class="text-[10px] text-blue-300 leading-normal font-medium italic">
                                        指令执行安全声明：所有 Shell 命令将以部署用户权限执行。建议在执行脚本末尾添加清理缓存的逻辑。
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>

                    <!-- Footer -->
                    <div class="px-10 py-8 border-t border-slate-100 bg-slate-50/80 flex justify-end space-x-6">
                        <button @click="isCreateModalOpen = false"
                            class="px-8 py-3 text-xs font-black text-slate-400 hover:text-slate-600 uppercase tracking-widest transition-colors">
                            放弃此次构建
                        </button>
                        <button @click="handleCreateTask"
                            class="px-14 py-3 bg-slate-900 text-white rounded-2xl text-xs font-black shadow-2xl hover:bg-blue-600 hover:-translate-y-1 active:translate-y-0 transition-all uppercase tracking-[0.15em]">
                            生成自动化工作流
                        </button>
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
                        <button
                            class="text-xs font-black text-red-500 hover:bg-red-50 px-6 py-3 rounded-xl transition-all uppercase tracking-widest">删除工作流</button>
                        <button @click="simulateRun(selectedTaskDetails!); selectedTaskDetails = null"
                            class="bg-slate-900 text-white px-12 py-3 rounded-2xl text-xs font-black shadow-xl hover:bg-blue-600 transition-all active:scale-95 uppercase tracking-widest">
                            确认并立即触发
                        </button>
                    </div>
                </div>
            </div>
        </Teleport>
    </div>
</template>
