<script setup lang="ts">
defineProps<{
  activeTab: string;
}>();

const emit = defineEmits(['update:activeTab']);

const menuItems = [
  { id: 'dashboard', icon: 'fa-chart-pie', label: '运行概览' },
  { id: 'svn', icon: 'fa-code-branch', label: 'SVN 资源库' },
  { id: 'servers', icon: 'fa-server', label: '节点拓扑' },
  { id: 'tasks', icon: 'fa-wand-magic-sparkles', label: '任务编排' },
  { id: 'logs', icon: 'fa-terminal', label: '实时日志' },
];

const onTabChange = (id: string) => {
  emit('update:activeTab', id);
};
</script>

<template>
  <aside class="w-56 bg-slate-800 text-slate-300 flex flex-col border-r border-slate-700/30">
    <div class="p-5 flex items-center space-x-3 border-b border-slate-700/50 bg-slate-800/80">
      <div class="w-8 h-8 bg-blue-500 rounded flex items-center justify-center shadow-lg">
        <i class="fa-solid fa-rocket text-white"></i>
      </div>
      <div class="leading-tight">
        <h2 class="text-sm font-bold text-white tracking-tight">DeployMaster</h2>
        <p class="text-[10px] text-slate-500">本地控制台</p>
      </div>
    </div>
    
    <nav class="flex-1 mt-4 px-2 space-y-0.5">
      <p class="px-3 py-2 text-[10px] font-bold text-slate-500 uppercase tracking-widest">主菜单</p>
      <button
        v-for="item in menuItems"
        :key="item.id"
        @click="onTabChange(item.id)"
        class="w-full flex items-center space-x-3 px-3 py-2.5 rounded transition-all duration-150 text-sm"
        :class="activeTab === item.id 
          ? 'bg-blue-600 text-white shadow-md font-medium' 
          : 'hover:bg-slate-700/50 hover:text-white'"
      >
        <i :class="['fa-solid', item.icon, 'w-5 opacity-80']"></i>
        <span>{{ item.label }}</span>
      </button>
    </nav>

    <div class="p-4 mt-auto">
      <button class="w-full py-2 px-3 bg-slate-700/50 hover:bg-slate-700 rounded text-xs font-bold flex items-center justify-center space-x-2 transition-colors border border-slate-600/50 text-slate-400 hover:text-white">
        <i class="fa-solid fa-gear"></i>
        <span>系统设置</span>
      </button>
    </div>
  </aside>
</template>
