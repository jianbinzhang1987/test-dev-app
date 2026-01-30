<script setup lang="ts">
import { ref, computed } from 'vue';
import { RemoteServer } from '../types';

const props = defineProps<{
  servers: RemoteServer[];
}>();

const emit = defineEmits(['add', 'update', 'delete']);

const isModalOpen = ref(false);
const isTopologyOpen = ref(false);
const modalMode = ref<'add' | 'edit'>('add');
const currentSrv = ref<Partial<RemoteServer>>({});
const testingId = ref<string | null>(null);
const hoveredNode = ref<string | null>(null);

const openAddModal = () => {
  modalMode.value = 'add';
  currentSrv.value = { port: 22, protocol: 'SFTP', isMaster: false };
  isModalOpen.value = true;
};

const openEditModal = (srv: RemoteServer) => {
  modalMode.value = 'edit';
  currentSrv.value = { ...srv };
  isModalOpen.value = true;
};

const handleSubmit = () => {
  if (modalMode.value === 'add') {
    const newSrv = {
      ...currentSrv.value,
      id: 'srv-' + Math.random().toString(36).substring(2, 9),
    } as RemoteServer;
    emit('add', newSrv);
  } else {
    emit('update', currentSrv.value as RemoteServer);
  }
  isModalOpen.value = false;
};

const handleSinglePing = (server: RemoteServer) => {
  testingId.value = server.id;
  setTimeout(() => {
    const mockLatency = Math.floor(Math.random() * 50) + 10;
    emit('update', {
      ...server,
      latency: mockLatency,
      lastChecked: new Date().toLocaleTimeString(),
    });
    testingId.value = null;
  }, 1200);
};

const masterNode = computed(() => props.servers.find(s => s.isMaster));
const slaveNodes = computed(() => props.servers.filter(s => !s.isMaster));
</script>

<template>
  <div class="space-y-4">
    <!-- Statistics Bar -->
    <div class="bg-white p-4 rounded border border-slate-200 shadow-sm flex items-center justify-between">
      <div class="flex space-x-8">
        <div class="flex flex-col">
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-0.5">已注册节点</span>
          <span class="text-xl font-black text-slate-800">{{ servers.length }} <span class="text-xs font-medium text-slate-400">个</span></span>
        </div>
        <div class="w-[1px] h-8 bg-slate-100 self-center"></div>
        <div class="flex flex-col">
          <span class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mb-0.5">网络拓扑状态</span>
          <button 
            @click="isTopologyOpen = true"
            class="text-xs font-bold text-blue-600 flex items-center space-x-1 hover:underline mt-1"
          >
            <i class="fa-solid fa-diagram-project"></i>
            <span>进入可视化分发中心</span>
          </button>
        </div>
      </div>
      <div class="flex space-x-2">
         <button class="px-4 py-1.5 border border-slate-200 rounded text-xs font-bold text-slate-600 hover:bg-slate-50 transition-colors">批量导入</button>
         <button 
          @click="openAddModal"
          class="bg-slate-800 text-white px-5 py-1.5 rounded text-xs font-bold flex items-center space-x-2 shadow hover:bg-slate-900 transition-colors"
         >
          <i class="fa-solid fa-plus-circle"></i>
          <span>添加远程服务器</span>
        </button>
      </div>
    </div>

    <!-- Servers Table -->
    <div class="bg-white rounded border border-slate-200 shadow-sm overflow-hidden font-sans">
      <table class="w-full text-left text-sm">
        <thead>
          <tr class="bg-slate-50 text-[10px] font-bold text-slate-500 uppercase tracking-tighter border-b border-slate-200">
            <th class="px-6 py-3 w-24">角色类型</th>
            <th class="px-6 py-3">节点名称</th>
            <th class="px-6 py-3">地址/端口</th>
            <th class="px-6 py-3">通信协议</th>
            <th class="px-6 py-3">响应延迟</th>
            <th class="px-6 py-3 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="server in servers" :key="server.id" class="hover:bg-slate-50/50 transition-colors group">
            <td class="px-6 py-4">
              <span v-if="server.isMaster" class="bg-blue-600 text-white text-[9px] font-black px-1.5 py-0.5 rounded uppercase shadow-sm">Master</span>
              <span v-else class="bg-slate-200 text-slate-600 text-[9px] font-black px-1.5 py-0.5 rounded uppercase">Slave</span>
            </td>
            <td class="px-6 py-4 font-bold text-slate-700">{{ server.name }}</td>
            <td class="px-6 py-4 font-mono text-xs">{{ server.ip }}:{{ server.port }}</td>
            <td class="px-6 py-4">
              <span class="text-[10px] font-black text-slate-500 border border-slate-200 px-2 py-0.5 rounded bg-slate-50 uppercase tracking-tighter">{{ server.protocol }}</span>
            </td>
            <td class="px-6 py-4">
               <div class="flex items-center space-x-2">
                <div v-if="testingId === server.id" class="flex items-center space-x-1.5 text-blue-500 animate-pulse text-[10px] font-bold">
                  <i class="fa-solid fa-spinner fa-spin"></i>
                  <span>检测中...</span>
                </div>
                <div v-else-if="server.latency" class="flex items-center space-x-1.5">
                  <div :class="['w-1.5 h-1.5 rounded-full', server.latency < 30 ? 'bg-green-500' : 'bg-yellow-500']"></div>
                  <span class="text-xs font-mono font-bold text-slate-600">{{ server.latency }}ms</span>
                </div>
                <span v-else class="text-xs text-slate-300 italic">未检测</span>
              </div>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex justify-end space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                <button @click="handleSinglePing(server)" class="w-7 h-7 flex items-center justify-center text-blue-600 hover:bg-blue-50 rounded">
                  <i class="fa-solid fa-network-wired text-xs"></i>
                </button>
                <button @click="openEditModal(server)" class="w-7 h-7 flex items-center justify-center text-slate-500 hover:bg-slate-100 rounded">
                  <i class="fa-solid fa-cog text-xs"></i>
                </button>
                <button @click="emit('delete', server.id)" class="w-7 h-7 flex items-center justify-center text-red-400 hover:bg-red-50 rounded">
                  <i class="fa-solid fa-trash-can text-xs"></i>
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Topology Modal -->
    <Teleport to="body">
      <div v-if="isTopologyOpen" class="fixed inset-0 bg-slate-950/90 backdrop-blur-xl z-[60] flex items-center justify-center p-6 animate-in fade-in duration-300">
        <div class="bg-[#0f172a] w-full h-full rounded-3xl shadow-2xl flex flex-col border border-white/10 relative overflow-hidden group/modal">
          <!-- Background -->
          <div class="absolute inset-0 pointer-events-none overflow-hidden">
             <div class="absolute top-0 left-0 w-full h-full opacity-[0.05]" style="background-image: linear-gradient(#4f46e5 1px, transparent 1px), linear-gradient(90deg, #4f46e5 1px, transparent 1px); background-size: 40px 40px;"></div>
             <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[800px] h-[800px] bg-indigo-500/10 rounded-full blur-[120px]"></div>
          </div>

          <!-- Header -->
          <div class="px-8 py-6 flex justify-between items-center relative z-10 border-b border-white/5 bg-white/5">
            <div class="flex items-center space-x-4">
              <div class="w-12 h-12 bg-indigo-600 rounded-2xl flex items-center justify-center shadow-[0_0_20px_rgba(79,70,229,0.5)]">
                <i class="fa-solid fa-diagram-project text-white text-xl"></i>
              </div>
              <div>
                <h3 class="text-xl font-black text-white tracking-tight">集群实时分发链路</h3>
                <p class="text-[10px] font-black text-indigo-400 uppercase tracking-[0.2em] mt-0.5">Real-time Cluster Distribution Topology</p>
              </div>
            </div>
            <div class="flex items-center space-x-6">
              <div class="flex items-center space-x-8 text-[10px] font-bold text-slate-400">
                <div class="flex items-center space-x-2">
                  <span class="w-3 h-3 rounded-full bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]"></span>
                  <span>活跃通道</span>
                </div>
                <div class="flex items-center space-x-2">
                  <span class="w-3 h-3 rounded-full border border-indigo-500/50 border-dashed"></span>
                  <span>待机备选</span>
                </div>
              </div>
              <button 
                @click="isTopologyOpen = false"
                class="w-10 h-10 rounded-xl bg-white/5 hover:bg-white/10 flex items-center justify-center text-slate-400 transition-all border border-white/10"
              >
                <i class="fa-solid fa-xmark"></i>
              </button>
            </div>
          </div>

          <!-- SVG Topology Area -->
          <div class="flex-1 relative flex flex-col items-center justify-center p-10 overflow-hidden">
             <div class="w-full max-w-4xl h-full relative">
                <svg class="absolute inset-0 w-full h-full pointer-events-none z-0">
                  <defs>
                    <linearGradient id="flowGradient" x1="0%" y1="0%" x2="0%" y2="100%">
                      <stop offset="0%" stopColor="#6366f1" stopOpacity="0" />
                      <stop offset="50%" stopColor="#6366f1" stopOpacity="1" />
                      <stop offset="100%" stopColor="#6366f1" stopOpacity="0" />
                    </linearGradient>
                    <filter id="glow">
                      <feGaussianBlur stdDeviation="3" result="coloredBlur" />
                      <feMerge>
                        <feMergeNode in="coloredBlur" />
                        <feMergeNode in="SourceGraphic" />
                      </feMerge>
                    </filter>
                  </defs>

                  <!-- SVN -> Master -->
                  <g v-if="masterNode">
                    <path d="M 50% 80 L 50% 210" stroke="#1e293b" stroke-width="4" fill="none" />
                    <path d="M 50% 80 L 50% 210" stroke="url(#flowGradient)" stroke-width="2" fill="none" stroke-dasharray="15, 35">
                       <animate attributeName="stroke-dashoffset" from="100" to="0" dur="2s" repeatCount="indefinite" />
                    </path>
                  </g>

                  <!-- Master -> Slaves -->
                  <g v-if="masterNode">
                    <template v-for="(slave, idx) in slaveNodes" :key="slave.id">
                      <path 
                        :d="`M 50% 290 C 50% 370, ${(100 / (slaveNodes.length + 1)) * (idx + 1)}% 390, ${(100 / (slaveNodes.length + 1)) * (idx + 1)}% 470`" 
                        stroke="#1e293b" stroke-width="3" fill="none" stroke-dasharray="5,5" 
                      />
                      <path 
                        :d="`M 50% 290 C 50% 370, ${(100 / (slaveNodes.length + 1)) * (idx + 1)}% 390, ${(100 / (slaveNodes.length + 1)) * (idx + 1)}% 470`" 
                        stroke="#4f46e5" stroke-width="2" fill="none" filter="url(#glow)" stroke-dasharray="10, 100" stroke-opacity="0.6"
                      >
                         <animate attributeName="stroke-dashoffset" from="220" to="0" :dur="`${1.5 + idx * 0.2}s`" repeatCount="indefinite" />
                      </path>
                    </template>
                  </g>
                </svg>

                <div class="relative z-10 flex flex-col items-center h-full">
                  <!-- Level 1: SVN Root -->
                  <div class="h-40 flex items-start pt-4">
                     <div class="bg-[#1e293b] p-1 rounded-2xl border border-white/5 shadow-2xl">
                        <div class="bg-indigo-950/50 px-6 py-4 rounded-xl border border-indigo-500/20 flex items-center space-x-4">
                          <div class="w-12 h-12 bg-indigo-500/20 rounded-lg flex items-center justify-center text-indigo-400 border border-indigo-500/30">
                            <i class="fa-solid fa-code-branch text-2xl"></i>
                          </div>
                          <div>
                             <p class="text-[10px] font-black text-indigo-400 uppercase tracking-widest">Source Root</p>
                             <p class="text-white font-bold text-sm">SVN 官方资源库</p>
                          </div>
                        </div>
                     </div>
                  </div>

                  <!-- Level 2: Master Node -->
                  <div class="h-56 flex items-center">
                     <div v-if="masterNode" 
                        @mouseenter="hoveredNode = masterNode.id"
                        @mouseleave="hoveredNode = null"
                        class="relative transition-all duration-300 transform"
                        :class="hoveredNode === masterNode.id ? 'scale-110' : ''"
                     >
                        <div class="p-6 bg-[#1e293b] rounded-[2.5rem] border-4 shadow-2xl w-64 flex flex-col items-center transition-all"
                           :class="hoveredNode === masterNode.id ? 'border-indigo-400 shadow-[0_0_40px_rgba(79,70,229,0.3)]' : 'border-indigo-600'">
                           <div class="w-16 h-16 bg-indigo-600 rounded-2xl flex items-center justify-center text-white text-3xl mb-4 shadow-xl">
                              <i class="fa-solid fa-crown"></i>
                           </div>
                           <h4 class="text-white font-black uppercase tracking-tighter text-base">{{ masterNode.name }}</h4>
                           <p class="text-indigo-400 font-mono text-[11px] font-bold mt-1 tracking-widest">{{ masterNode.ip }}</p>
                           <div class="mt-4 flex space-x-2">
                              <span class="bg-indigo-500/20 text-indigo-400 text-[9px] font-black px-2 py-0.5 rounded border border-indigo-500/30">SYNC MASTER</span>
                           </div>
                        </div>
                        <!-- Tooltip -->
                        <div v-if="hoveredNode === masterNode.id" class="absolute top-full mt-4 left-1/2 -translate-x-1/2 w-48 bg-slate-800 border border-white/10 rounded-xl p-3 shadow-2xl z-50 text-[10px] text-slate-300 space-y-2 animate-in slide-in-from-top-2">
                          <p class="flex justify-between"><span>协议:</span> <span class="text-indigo-400 font-bold">{{ masterNode.protocol }}</span></p>
                          <p class="flex justify-between"><span>端口:</span> <span class="text-indigo-400 font-bold">{{ masterNode.port }}</span></p>
                          <p class="flex justify-between"><span>最后延迟:</span> <span class="text-emerald-400 font-bold">{{ masterNode.latency || '--' }} ms</span></p>
                        </div>
                     </div>
                     <div v-else class="border-2 border-dashed border-red-500/30 bg-red-500/5 rounded-[2.5rem] p-10 text-red-400 text-xs font-bold uppercase tracking-widest flex flex-col items-center">
                        <i class="fa-solid fa-triangle-exclamation text-3xl mb-3 animate-pulse"></i>
                        <span>未指定 Master 节点</span>
                     </div>
                  </div>

                  <!-- Level 3: Slave Nodes -->
                  <div class="flex-1 w-full flex justify-around items-end pb-12">
                     <template v-if="slaveNodes.length > 0">
                       <div v-for="slave in slaveNodes" :key="slave.id"
                          @mouseenter="hoveredNode = slave.id"
                          @mouseleave="hoveredNode = null"
                          class="relative transition-all duration-300 transform"
                          :class="hoveredNode === slave.id ? 'scale-110 -translate-y-4' : ''"
                       >
                          <div class="p-4 bg-[#1e293b] rounded-3xl border shadow-xl w-40 flex flex-col items-center transition-all"
                             :class="hoveredNode === slave.id ? 'border-indigo-400 shadow-[0_0_30px_rgba(79,70,229,0.2)]' : 'border-white/10'">
                             <div class="w-12 h-12 rounded-xl flex items-center justify-center text-xl mb-3 transition-colors"
                                :class="hoveredNode === slave.id ? 'bg-indigo-500 text-white' : 'bg-slate-800 text-slate-500'">
                                <i class="fa-solid fa-server"></i>
                             </div>
                             <h5 class="text-slate-200 font-bold text-xs truncate w-full text-center px-2">{{ slave.name }}</h5>
                             <p class="text-slate-500 font-mono text-[9px] mt-1">{{ slave.ip }}</p>
                             <div class="mt-2 flex items-center space-x-1.5 bg-slate-900/50 px-2 py-0.5 rounded-full border border-white/5">
                                <div class="w-1.5 h-1.5 rounded-full bg-emerald-500"></div>
                                <span class="text-[8px] font-bold text-slate-400 uppercase tracking-widest">Ready</span>
                             </div>
                          </div>
                          <!-- Tooltip -->
                          <div v-if="hoveredNode === slave.id" class="absolute bottom-full mb-4 left-1/2 -translate-x-1/2 w-40 bg-slate-800 border border-white/10 rounded-xl p-3 shadow-2xl z-50 text-[10px] text-slate-300 space-y-1.5 animate-in slide-in-from-bottom-2">
                            <p class="flex justify-between"><span>协议:</span> <span class="text-indigo-400">{{ slave.protocol }}</span></p>
                            <p class="flex justify-between"><span>响应:</span> <span class="text-emerald-400">{{ slave.latency || 'N/A' }} ms</span></p>
                          </div>
                       </div>
                     </template>
                     <div v-else class="text-slate-600 font-black text-xs uppercase tracking-widest opacity-30">No Slaves Configured</div>
                  </div>
                </div>
             </div>
          </div>

          <!-- Footer Info -->
          <div class="px-8 py-4 bg-white/5 border-t border-white/5 flex items-center justify-between text-slate-500 text-[10px] font-bold uppercase tracking-widest relative z-10">
             <div class="flex items-center space-x-6">
               <div class="flex items-center space-x-2">
                 <i class="fa-solid fa-microchip text-indigo-500"></i>
                 <span>分发算法: P2P 并行广播</span>
               </div>
               <div class="flex items-center space-x-2">
                 <i class="fa-solid fa-bolt text-amber-500"></i>
                 <span>链路效率: 99.8%</span>
               </div>
             </div>
             <div class="flex items-center space-x-3">
               <span class="animate-pulse">实时拓扑更新中</span>
               <i class="fa-solid fa-sync fa-spin text-indigo-400"></i>
             </div>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Config Modal -->
    <Teleport to="body">
      <div v-if="isModalOpen" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <div class="bg-white w-full max-w-md rounded-xl shadow-2xl border border-slate-200">
           <div class="px-6 py-4 border-b flex justify-between items-center bg-slate-50 rounded-t-xl">
              <h3 class="text-sm font-black text-slate-800 uppercase tracking-widest">{{ modalMode === 'add' ? '注册新节点' : '节点参数修正' }}</h3>
              <button @click="isModalOpen = false" class="text-slate-400 hover:text-slate-600"><i class="fa-solid fa-xmark"></i></button>
           </div>
           <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
              <div class="space-y-1">
                <label class="text-[10px] font-bold text-slate-500 uppercase">节点易记名称</label>
                <input required type="text" v-model="currentSrv.name" class="w-full px-3 py-2 border rounded text-xs outline-none focus:border-indigo-500" />
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-1">
                  <label class="text-[10px] font-bold text-slate-500 uppercase">IP 地址</label>
                  <input required type="text" v-model="currentSrv.ip" class="w-full px-3 py-2 border rounded text-xs font-mono outline-none focus:border-indigo-500" />
                </div>
                <div class="space-y-1">
                  <label class="text-[10px] font-bold text-slate-500 uppercase">端口</label>
                  <input required type="number" v-model.number="currentSrv.port" class="w-full px-3 py-2 border rounded text-xs font-mono outline-none focus:border-indigo-500" />
                </div>
              </div>
              <div class="grid grid-cols-2 gap-4">
                <div class="space-y-1">
                  <label class="text-[10px] font-bold text-slate-500 uppercase">角色</label>
                  <select v-model="currentSrv.isMaster" class="w-full px-2 py-2 border rounded text-xs outline-none">
                    <option :value="false">Slave (从节点)</option>
                    <option :value="true">Master (主节点)</option>
                  </select>
                </div>
                <div class="space-y-1">
                  <label class="text-[10px] font-bold text-slate-500 uppercase">协议</label>
                  <select v-model="currentSrv.protocol" class="w-full px-2 py-2 border rounded text-xs outline-none">
                    <option value="SFTP">SFTP</option>
                    <option value="SCP">SCP</option>
                    <option value="FTP">FTP</option>
                  </select>
                </div>
              </div>
              <div class="flex justify-end space-x-3 pt-4">
                <button type="button" @click="isModalOpen = false" class="px-4 py-2 text-xs font-bold text-slate-400">取消</button>
                <button type="submit" class="px-6 py-2 bg-indigo-600 text-white rounded text-xs font-bold shadow-lg">确定保存</button>
              </div>
           </form>
        </div>
      </div>
    </Teleport>
  </div>
</template>
