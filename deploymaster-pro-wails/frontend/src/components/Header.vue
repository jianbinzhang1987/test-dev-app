<script setup lang="ts">
import { computed } from 'vue';

const props = defineProps<{
  activeTab: string;
}>();

const emit = defineEmits(['newDeployment']);

const title = computed(() => {
  switch (props.activeTab) {
    case 'dashboard': return '系统状态概览';
    case 'svn': return 'SVN 资源配置';
    case 'servers': return '远程服务器拓扑';
    case 'tasks': return '自动化任务编排';
    case 'logs': return '控制台执行日志';
    default: return '部署控制台';
  }
});
</script>

<template>
  <header class="h-14 bg-white border-b border-slate-200 flex items-center justify-between px-6 shrink-0 shadow-sm z-10">
    <div class="flex items-center space-x-3">
      <h1 class="text-base font-bold text-slate-800 tracking-tight">{{ title }}</h1>
      <span class="px-2 py-0.5 bg-blue-50 text-blue-600 text-[10px] font-bold rounded border border-blue-100 uppercase">本地实例</span>
    </div>
    
    <div class="flex items-center space-x-4">
      <button 
        @click="emit('newDeployment')"
        class="bg-blue-600 text-white px-4 py-1.5 rounded text-xs font-bold hover:bg-blue-700 transition-all shadow-md active:scale-95 flex items-center space-x-2"
      >
        <i class="fa-solid fa-bolt-lightning"></i>
        <span>新建部署</span>
      </button>
    </div>
  </header>
</template>
