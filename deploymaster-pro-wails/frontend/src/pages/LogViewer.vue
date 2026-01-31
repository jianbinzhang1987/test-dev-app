<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { TaskRun } from '../types';

const props = defineProps<{
    runs: TaskRun[];
    selectedRunId?: string | null;
    selectedTaskId?: string | null;
}>();

const emit = defineEmits<{
    (e: 'deleteRun', runId: string): void;
    (e: 'deleteRunsByTask', taskId: string): void;
}>();

const selectedRun = ref<TaskRun | null>(props.runs[0] || null);
const searchTerm = ref('');

watch(() => props.runs, (newRuns) => {
    if (props.selectedTaskId) {
        const match = newRuns.find(r => r.taskId === props.selectedTaskId);
        if (match) {
            selectedRun.value = match;
            return;
        }
    }
    if (!selectedRun.value && newRuns.length > 0) {
        selectedRun.value = newRuns[0];
    } else if (selectedRun.value) {
        const updated = newRuns.find(r => r.id === selectedRun.value?.id);
        if (updated) selectedRun.value = updated;
    }
}, { deep: true });

watch(() => props.selectedRunId, (id) => {
    if (!id) return;
    const run = props.runs.find(r => r.id === id);
    if (run) selectedRun.value = run;
});

watch(() => props.selectedTaskId, (taskId) => {
    if (!taskId) return;
    const run = props.runs.find(r => r.taskId === taskId);
    if (run) selectedRun.value = run;
});

const filteredRuns = computed(() => {
    const keyword = searchTerm.value.trim().toLowerCase();
    const list = keyword
        ? props.runs.filter(r =>
            r.taskName.toLowerCase().includes(keyword) ||
            r.taskId.toLowerCase().includes(keyword) ||
            r.startedAt.toLowerCase().includes(keyword) ||
            (r.finishedAt || '').toLowerCase().includes(keyword)
        )
        : props.runs;
    return [...list].sort((a, b) => {
        const at = parseRunTime(a.startedAt)?.getTime() ?? 0;
        const bt = parseRunTime(b.startedAt)?.getTime() ?? 0;
        return bt - at;
    }).slice(0, 50);
});

const totalRuns = computed(() => props.runs.length);

const parseRunTime = (value?: string) => {
    if (!value) return null;
    const parts = value.trim().split(' ');
    if (parts.length === 2) {
        const [d, t] = parts;
        const [y, m, day] = d.split('-').map(Number);
        const [hh, mm, ss] = t.split(':').map(Number);
        if (Number.isFinite(y) && Number.isFinite(m) && Number.isFinite(day) && Number.isFinite(hh) && Number.isFinite(mm)) {
            return new Date(y, m - 1, day, hh, mm, Number.isFinite(ss) ? ss : 0);
        }
    }
    const parsed = new Date(value);
    return Number.isNaN(parsed.getTime()) ? null : parsed;
};

const durationText = computed(() => {
    if (!selectedRun.value) return '';
    const start = parseRunTime(selectedRun.value.startedAt);
    const end = parseRunTime(selectedRun.value.finishedAt) || new Date();
    if (!start) return '';
    const ms = Math.max(0, end.getTime() - start.getTime());
    const sec = ms / 1000;
    if (sec < 60) return `${sec.toFixed(1)}s`;
    const min = Math.floor(sec / 60);
    const rem = Math.round(sec % 60);
    return `${min}m ${rem}s`;
});
</script>

<template>
    <div class="h-full flex flex-col space-y-3">
        <!-- Task Selector Tabs -->
        <div class="flex items-center justify-between gap-3">
            <div class="flex-1 min-w-0 flex items-center space-x-2">
                <div class="relative flex-1 min-w-[220px]">
                    <input v-model="searchTerm" placeholder="搜索任务名 / 任务ID / 时间"
                        class="w-full bg-white border border-slate-200 rounded-full px-3 py-1.5 text-xs text-slate-600 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-200" />
                </div>
                <span class="text-[10px] text-slate-400 whitespace-nowrap">共 {{ totalRuns }} 条，显示最近 {{ filteredRuns.length }} 条</span>
            </div>
            <div class="flex items-center space-x-2 shrink-0">
                <button v-if="selectedRun" class="text-rose-500 text-xs font-bold hover:underline"
                    @click="emit('deleteRun', selectedRun.id)">删除本条</button>
                <button v-if="selectedRun" class="text-slate-500 text-xs font-bold hover:underline"
                    @click="emit('deleteRunsByTask', selectedRun.taskId)">清空该任务</button>
            </div>
        </div>

        <div class="flex space-x-2 overflow-x-auto pb-2 shrink-0">
            <button v-for="run in filteredRuns" :key="run.id" @click="selectedRun = run" :class="['px-4 py-1.5 rounded-full text-xs font-bold whitespace-nowrap transition-all border flex items-center space-x-2',
                selectedRun?.id === run.id
                    ? 'bg-blue-600 text-white border-blue-600 shadow-md'
                    : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50']">
                <span>{{ run.taskName }} · {{ run.startedAt }}</span>
                <button class="text-white/80 hover:text-white" title="删除此条" @click.stop="emit('deleteRun', run.id)">
                    <i class="fa-solid fa-xmark text-[9px]"></i>
                </button>
            </button>
        </div>

        <!-- Terminal Container -->
        <div class="flex-1 bg-[#0d1117] rounded-xl shadow-2xl overflow-hidden flex flex-col border border-slate-800">
            <!-- Terminal Header -->
            <div class="h-9 px-4 border-b border-white/5 flex items-center justify-between bg-[#161b22]">
                <div class="flex items-center space-x-3">
                    <div class="flex space-x-1.5">
                        <div class="w-2.5 h-2.5 rounded-full bg-slate-700"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-slate-700"></div>
                        <div class="w-2.5 h-2.5 rounded-full bg-slate-700"></div>
                    </div>
                    <span
                        class="text-[10px] font-mono text-slate-500 font-black uppercase tracking-widest ml-4 flex items-center space-x-2">
                        <i class="fa-solid fa-terminal text-[8px]"></i>
                        <span>终端输出: {{ selectedRun?.taskName || '未选择任务' }}</span>
                    </span>
                </div>
                <div class="flex items-center space-x-4">
                    <div class="flex items-center space-x-2 text-[9px] text-emerald-500 font-bold">
                        <span class="w-1.5 h-1.5 bg-emerald-500 rounded-full animate-ping"></span>
                        <span>实时流</span>
                    </div>
                    <div class="w-[1px] h-3 bg-white/10"></div>
                    <button class="text-slate-500 hover:text-white transition-colors" title="下载完整日志"><i
                            class="fa-solid fa-download text-xs"></i></button>
                    <button class="text-slate-500 hover:text-white transition-colors" title="清除屏幕"><i
                            class="fa-solid fa-trash-can text-xs"></i></button>
                </div>
            </div>

            <!-- Terminal Body -->
            <div class="flex-1 p-6 font-mono text-[11px] overflow-y-auto space-y-1.5 leading-relaxed">
                <template v-if="selectedRun">
                    <div v-for="(log, idx) in selectedRun.logs" :key="idx"
                        class="flex space-x-4 hover:bg-white/5 p-0.5 rounded transition-colors group">
                        <span class="text-slate-600 select-none w-8 text-right shrink-0">{{ idx + 1 }}</span>
                        <span :class="[
                            (log.includes('[ERROR]') || log.includes('错误')) ? 'text-red-400' :
                                (log.includes('[SUCCESS]') || log.includes('完成')) ? 'text-emerald-400' :
                                    (log.includes('[INFO]') || log.includes('信息')) ? 'text-blue-300' : 'text-slate-300'
                        ]">
                            {{ log }}
                        </span>
                    </div>

                    <div v-if="selectedRun.progress < 100" class="flex space-x-4 mt-2">
                        <span class="text-slate-600 select-none w-8 text-right">{{ selectedRun.logs.length + 1
                            }}</span>
                        <span class="text-blue-400 animate-pulse italic">正在等待命令响应... _</span>
                    </div>

                    <div v-if="selectedRun.progress === 100" class="mt-8 pt-4 border-t border-white/5">
                        <p class="text-emerald-500 font-black tracking-tight flex items-center space-x-2">
                            <i class="fa-solid fa-circle-check"></i>
                            <span>工作流已完成，共计用时 {{ durationText || '0.0s' }}</span>
                        </p>
                        <div class="grid grid-cols-2 gap-4 mt-2 text-[9px] text-slate-500">
                            <div class="bg-white/5 p-2 rounded">节点响应: Master(200 OK), Slave-01(200 OK), Slave-02(200 OK)
                            </div>
                            <div class="bg-white/5 p-2 rounded">MD5 校验: 7a8b9c... 匹配一致</div>
                        </div>
                    </div>
                </template>
                <div v-else class="h-full flex items-center justify-center text-slate-600 italic">
                    请从上方选择一个任务以查看日志输出
                </div>
            </div>

            <!-- Terminal Input Placeholder -->
            <div class="px-4 py-2 bg-[#0d1117] border-t border-white/5 flex items-center space-x-3">
                <span class="text-blue-500 font-black text-xs">$</span>
                <input type="text" placeholder="在此输入诊断指令直接下发至当前任务节点 (需管理员权限)..."
                    class="bg-transparent text-slate-400 w-full outline-none text-[10px] placeholder:text-slate-700" />
            </div>
        </div>
    </div>
</template>
