<script setup lang="ts">
import { ref, watch } from 'vue';
import { DeploymentTask } from '../types';

const props = defineProps<{
    tasks: DeploymentTask[];
}>();

const selectedTask = ref<DeploymentTask | null>(props.tasks[0] || null);

watch(() => props.tasks, (newTasks) => {
    if (!selectedTask.value && newTasks.length > 0) {
        selectedTask.value = newTasks[0];
    } else if (selectedTask.value) {
        // Keep the reference updated if the object itself changes (though Vue's reactivity should handle this if it's the same object)
        const updated = newTasks.find(t => t.id === selectedTask.value?.id);
        if (updated) selectedTask.value = updated;
    }
}, { deep: true });
</script>

<template>
    <div class="h-full flex flex-col space-y-3">
        <!-- Task Selector Tabs -->
        <div class="flex space-x-2 overflow-x-auto pb-2 shrink-0">
            <button v-for="task in tasks" :key="task.id" @click="selectedTask = task" :class="['px-4 py-1.5 rounded-full text-xs font-bold whitespace-nowrap transition-all border',
                selectedTask?.id === task.id
                    ? 'bg-blue-600 text-white border-blue-600 shadow-md'
                    : 'bg-white text-slate-500 border-slate-200 hover:bg-slate-50']">
                {{ task.name }}
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
                        <span>终端输出: {{ selectedTask?.name || '未选择任务' }}</span>
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
                <template v-if="selectedTask">
                    <div v-for="(log, idx) in selectedTask.logs" :key="idx"
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

                    <div v-if="selectedTask.progress < 100" class="flex space-x-4 mt-2">
                        <span class="text-slate-600 select-none w-8 text-right">{{ selectedTask.logs.length + 1
                            }}</span>
                        <span class="text-blue-400 animate-pulse italic">正在等待命令响应... _</span>
                    </div>

                    <div v-if="selectedTask.progress === 100" class="mt-8 pt-4 border-t border-white/5">
                        <p class="text-emerald-500 font-black tracking-tight flex items-center space-x-2">
                            <i class="fa-solid fa-circle-check"></i>
                            <span>工作流已完成，共计用时 12.4s</span>
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
