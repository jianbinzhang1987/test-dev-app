<script setup lang="ts">
import { computed } from 'vue';
import { DeploymentTask, RemoteServer, TaskStatus, TaskRun } from '../types';

const props = defineProps<{
  tasks: DeploymentTask[];
  servers: RemoteServer[];
  runs: TaskRun[];
  windowed?: boolean;
}>();

const emit = defineEmits<{
  (e: 'viewAllRuns'): void;
}>();

const parseTime = (value?: string) => {
  if (!value) return null;
  const ts = new Date(value).getTime();
  return Number.isNaN(ts) ? null : ts;
};

const formatRelativeTime = (value?: string) => {
  const ts = parseTime(value);
  if (!ts) return '--';
  const diffMs = Date.now() - ts;
  if (diffMs < 60_000) return '刚刚';
  const diffMin = Math.floor(diffMs / 60_000);
  if (diffMin < 60) return `${diffMin} 分钟前`;
  const diffHour = Math.floor(diffMin / 60);
  if (diffHour < 24) return `${diffHour} 小时前`;
  const diffDay = Math.floor(diffHour / 24);
  if (diffDay < 7) return `${diffDay} 天前`;
  const d = new Date(ts);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`;
};

const connectedServers = computed(() =>
  props.servers.filter(s => s.status === 'connected').length
);

const totalServers = computed(() => props.servers.length);

const successRate = computed(() => {
  const finishedRuns = props.runs.filter(r => r.status === TaskStatus.SUCCESS || r.status === TaskStatus.FAILED);
  if (finishedRuns.length === 0) return '--';
  const successCount = finishedRuns.filter(r => r.status === TaskStatus.SUCCESS).length;
  return `${Math.round((successCount / finishedRuns.length) * 100)}%`;
});

const latestRunTime = computed(() => {
  const runTimes = props.runs
    .map(r => parseTime(r.finishedAt || r.startedAt))
    .filter((t): t is number => t !== null);
  const taskTimes = props.tasks
    .map(t => parseTime(t.lastRunAt))
    .filter((t): t is number => t !== null);
  const allTimes = [...runTimes, ...taskTimes];
  if (allTimes.length === 0) return undefined;
  return new Date(Math.max(...allTimes)).toLocaleString();
});

const stats = computed(() => [
  { label: '总部署任务', value: String(props.tasks.length), icon: 'fa-layer-group', color: 'bg-blue-500' },
  { label: '平均成功率', value: successRate.value, icon: 'fa-check-double', color: 'bg-emerald-500' },
  { label: '活跃节点数', value: totalServers.value ? `${connectedServers.value}/${totalServers.value}` : '0', icon: 'fa-network-wired', color: 'bg-indigo-500' },
  { label: '最后同步时间', value: formatRelativeTime(latestRunTime.value), icon: 'fa-history', color: 'bg-orange-500' },
]);

const isWindowed = computed(() => Boolean(props.windowed));

const recentRuns = computed(() => {
  const sorted = [...props.runs].sort((a, b) => {
    const aTime = parseTime(a.startedAt) ?? 0;
    const bTime = parseTime(b.startedAt) ?? 0;
    return bTime - aTime;
  });
  return sorted.slice(0, 6);
});

const onlineRate = computed(() => {
  if (totalServers.value === 0) return 0;
  return Math.round((connectedServers.value / totalServers.value) * 100);
});

const averageDelay = computed(() => {
  const delays = props.servers
    .map(s => s.delay ?? s.latency)
    .filter((v): v is number => typeof v === 'number' && v >= 0);
  if (delays.length === 0) return null;
  return Math.round(delays.reduce((sum, v) => sum + v, 0) / delays.length);
});

const delayPercent = computed(() => {
  if (averageDelay.value === null) return 0;
  return Math.min(Math.round((averageDelay.value / 300) * 100), 100);
});

const suggestion = computed(() => {
  const offline = totalServers.value - connectedServers.value;
  if (offline > 0) return `检测到 ${offline} 个节点离线，建议检查网络或凭据配置。`;
  if (successRate.value !== '--' && parseInt(successRate.value, 10) < 80) {
    return '近期成功率偏低，建议检查脚本与资源权限。';
  }
  if (averageDelay.value !== null && averageDelay.value > 150) {
    return '节点平均延迟较高，建议优化链路或启用并行同步。';
  }
  return '运行稳定，建议保持当前调度策略。';
});
</script>

<template>
  <div class="space-y-6">
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <div v-for="stat in stats" :key="stat.label" class="bg-white p-5 rounded-lg border border-slate-200 shadow-sm flex items-center space-x-4 hover:border-blue-200 transition-colors">
        <div :class="[stat.color, 'w-10 h-10 rounded flex items-center justify-center text-white text-lg shadow-sm']">
          <i :class="['fa-solid', stat.icon]"></i>
        </div>
        <div>
          <p class="text-[10px] text-slate-500 font-bold uppercase tracking-wider">{{ stat.label }}</p>
          <p class="text-xl font-black text-slate-800">{{ stat.value }}</p>
        </div>
      </div>
    </div>

    <div :class="['grid gap-6', isWindowed ? 'grid-cols-1' : 'grid-cols-1 lg:grid-cols-3']">
      <div :class="['bg-white rounded-lg border border-slate-200 shadow-sm flex flex-col', isWindowed ? '' : 'lg:col-span-2']">
        <div class="px-5 py-3 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-sm font-bold text-slate-700">最近执行历史</h3>
          <button class="text-blue-600 text-xs font-bold hover:underline" @click="emit('viewAllRuns')">查看全部</button>
        </div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead>
              <tr class="text-xs font-bold text-slate-400 bg-slate-50">
                <th class="px-6 py-3">任务描述</th>
                <th class="px-6 py-3">当前状态</th>
                <th class="px-6 py-3">完成进度</th>
                <th class="px-6 py-3">执行时间</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100">
              <tr v-for="run in recentRuns" :key="run.id" class="hover:bg-slate-50/80 transition-colors">
                <td class="px-6 py-4 font-semibold text-slate-700">{{ run.taskName }}</td>
                <td class="px-6 py-4">
                  <span :class="['inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold', 
                    run.status === TaskStatus.SUCCESS ? 'bg-emerald-100 text-emerald-700'
                    : run.status === TaskStatus.FAILED ? 'bg-rose-100 text-rose-700'
                    : 'bg-blue-100 text-blue-700']">
                    {{ run.status === TaskStatus.SUCCESS ? '已完成' : run.status === TaskStatus.FAILED ? '失败' : '执行中' }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <div class="flex items-center space-x-3">
                     <div class="flex-1 min-w-[100px] bg-slate-100 rounded-full h-1.5">
                       <div class="bg-blue-600 h-1.5 rounded-full" :style="{ width: run.progress + '%' }"></div>
                     </div>
                     <span class="text-[10px] font-bold text-slate-500">{{ run.progress }}%</span>
                  </div>
                </td>
                <td class="px-6 py-4 text-xs text-slate-400 font-mono">{{ run.finishedAt || run.startedAt }}</td>
              </tr>
              <tr v-if="recentRuns.length === 0">
                <td colspan="4" class="px-6 py-6 text-xs text-slate-400">暂无运行记录</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div :class="['bg-white rounded-lg border border-slate-200 shadow-sm flex flex-col', isWindowed ? 'order-last' : '']">
        <div class="px-5 py-3 border-b border-slate-100">
          <h3 class="text-sm font-bold text-slate-700">资源占用监控</h3>
        </div>
        <div class="p-6 space-y-6">
          <div class="space-y-2">
            <div class="flex items-center justify-between text-xs">
              <span class="font-bold text-slate-600">节点在线率</span>
              <span class="font-black text-blue-600">{{ totalServers ? `${connectedServers}/${totalServers} 在线` : '--' }}</span>
            </div>
            <div class="w-full bg-slate-100 rounded-full h-2">
              <div class="bg-blue-500 h-2 rounded-full" :style="{ width: onlineRate + '%' }"></div>
            </div>
          </div>
          
          <div class="space-y-2">
            <div class="flex items-center justify-between text-xs">
              <span class="font-bold text-slate-600">平均延迟</span>
              <span class="font-black text-indigo-600">{{ averageDelay === null ? '--' : `${averageDelay} ms` }}</span>
            </div>
            <div class="w-full bg-slate-100 rounded-full h-2">
              <div class="bg-indigo-500 h-2 rounded-full" :style="{ width: delayPercent + '%' }"></div>
            </div>
          </div>

          <div class="mt-8 p-4 bg-blue-50/50 rounded-lg border border-blue-100 flex space-x-3">
            <i class="fa-solid fa-bolt-lightning text-blue-500 mt-0.5"></i>
            <div class="space-y-1">
              <p class="font-bold text-xs text-blue-800">性能建议</p>
              <p class="text-[10px] text-blue-600 leading-normal">
                {{ suggestion }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
