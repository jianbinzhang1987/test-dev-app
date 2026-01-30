<script setup lang="ts">
import { DeploymentTask, RemoteServer, TaskStatus } from '../types';

defineProps<{
  tasks: DeploymentTask[];
  servers: RemoteServer[];
}>();

const stats = [
  { label: '总部署任务', value: '1', icon: 'fa-layer-group', color: 'bg-blue-500' },
  { label: '平均成功率', value: '99.2%', icon: 'fa-check-double', color: 'bg-emerald-500' },
  { label: '活跃节点数', value: '3', icon: 'fa-network-wired', color: 'bg-indigo-500' },
  { label: '最后同步时间', value: '3 分钟前', icon: 'fa-history', color: 'bg-orange-500' },
];
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

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <div class="lg:col-span-2 bg-white rounded-lg border border-slate-200 shadow-sm flex flex-col">
        <div class="px-5 py-3 border-b border-slate-100 flex items-center justify-between">
          <h3 class="text-sm font-bold text-slate-700">最近执行历史</h3>
          <button class="text-blue-600 text-xs font-bold hover:underline">查看全部</button>
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
              <tr v-for="task in tasks" :key="task.id" class="hover:bg-slate-50/80 transition-colors">
                <td class="px-6 py-4 font-semibold text-slate-700">{{ task.name }}</td>
                <td class="px-6 py-4">
                  <span :class="['inline-flex items-center px-2 py-0.5 rounded text-[10px] font-bold', 
                    task.status === TaskStatus.SUCCESS ? 'bg-emerald-100 text-emerald-700' : 'bg-blue-100 text-blue-700']">
                    {{ task.status === TaskStatus.SUCCESS ? '已完成' : '执行中' }}
                  </span>
                </td>
                <td class="px-6 py-4">
                  <div class="flex items-center space-x-3">
                     <div class="flex-1 min-w-[100px] bg-slate-100 rounded-full h-1.5">
                       <div class="bg-blue-600 h-1.5 rounded-full" :style="{ width: task.progress + '%' }"></div>
                     </div>
                     <span class="text-[10px] font-bold text-slate-500">{{ task.progress }}%</span>
                  </div>
                </td>
                <td class="px-6 py-4 text-xs text-slate-400 font-mono">2023-11-20 14:30</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <div class="bg-white rounded-lg border border-slate-200 shadow-sm flex flex-col">
        <div class="px-5 py-3 border-b border-slate-100">
          <h3 class="text-sm font-bold text-slate-700">资源占用监控</h3>
        </div>
        <div class="p-6 space-y-6">
          <div class="space-y-2">
            <div class="flex items-center justify-between text-xs">
              <span class="font-bold text-slate-600">网络同步带宽</span>
              <span class="font-black text-blue-600">1.2 GB/s</span>
            </div>
            <div class="w-full bg-slate-100 rounded-full h-2">
              <div class="bg-blue-500 h-2 rounded-full" style="width: 78%"></div>
            </div>
          </div>
          
          <div class="space-y-2">
            <div class="flex items-center justify-between text-xs">
              <span class="font-bold text-slate-600">磁盘存储空间</span>
              <span class="font-black text-indigo-600">42 / 100 TB</span>
            </div>
            <div class="w-full bg-slate-100 rounded-full h-2">
              <div class="bg-indigo-500 h-2 rounded-full" style="width: 42%"></div>
            </div>
          </div>

          <div class="mt-8 p-4 bg-blue-50/50 rounded-lg border border-blue-100 flex space-x-3">
            <i class="fa-solid fa-bolt-lightning text-blue-500 mt-0.5"></i>
            <div class="space-y-1">
              <p class="font-bold text-xs text-blue-800">性能建议</p>
              <p class="text-[10px] text-blue-600 leading-normal">
                检测到从服务器同步路径存在延迟，建议开启“并行分发”模式以提高跨机房传输速度。
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
