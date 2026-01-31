<script setup lang="ts">
import { ref } from 'vue';
import type { RemoteServer } from '../types';
import { useNodeService } from '../composables/useNodeService';
import CredentialDialog from '../components/CredentialDialog.vue';

const props = defineProps<{
    servers: RemoteServer[];
    loading?: boolean;
}>();

const emit = defineEmits(['update-list', 'delete']);

const nodeService = useNodeService();

// 对话框控制
const isCredentialModalOpen = ref(false);
const isTopologyOpen = ref(false);

// 当前操作状态
const isEdit = ref(false);
const currentNode = ref<Partial<RemoteServer>>({});

/**
 * 核心交互：添加服务器（一站式）
 */
const openAddModal = () => {
    isEdit.value = false;
    currentNode.value = {}; // 依靠 CredentialDialog 的 resetForm
    isCredentialModalOpen.value = true;
};

/**
 * 核心交互：编辑服务器（一站式）
 */
const openEditModal = (server: RemoteServer) => {
    isEdit.value = true;
    currentNode.value = { ...server };
    isCredentialModalOpen.value = true;
};

const isBatchTesting = ref(false);

const handleBatchTest = async () => {
    if (isBatchTesting.value) return;
    isBatchTesting.value = true;
    try {
        await nodeService.testAllNodes();
    } finally {
        isBatchTesting.value = false;
    }
};

/**
 * 处理合并后的提交逻辑
 */
const handleFullSubmit = async (data: any) => {
    try {
        // 1. 保存/更新节点基础数据
        let nodeId = data.id;
        if (isEdit.value) {
            if (!nodeId) {
                throw new Error('编辑模式缺少节点ID，无法更新');
            }
            await nodeService.updateNode(data);
        } else {
            nodeId = await nodeService.addNode(data);
        }

        // 2. 处理凭据持久化（私有字段传递）
        if (nodeId) {
            if (data.authMethod === 'password' && data._rememberPassword && data._password) {
                await nodeService.saveCredential(nodeId, data.username, data._password, data._rememberPassword);
            } else if (data.authMethod === 'key' && data._rememberPassphrase && data._keyPassphrase) {
                await nodeService.saveKeyPassphrase(nodeId, data._keyPassphrase, data._rememberPassphrase);
            }
        }

        // 3. 后置处理
        isCredentialModalOpen.value = false;
        emit('update-list');
    } catch (error) {
        console.error('Failed to process node submission:', error);
        throw error;
    }
};

/**
 * 连通性测试（Ping）
 * 如果没有凭据，引导去配置
 */
const handleSinglePing = async (server: RemoteServer) => {
    try {
        const hasCreds = await nodeService.hasStoredCredential(server.id, server.username || 'root');
        if (hasCreds || server.authMethod === 'agent') {
            await nodeService.testConnection(server.id, server.username || 'root');
        } else {
            // 引导去配置，但使用轻量级提示而不是直接跳转
            if (confirm(`未检测到节点 "${server.name}" 的存储凭据，是否立即前往配置？`)) {
                openEditModal(server);
            }
        }
    } catch (err: any) {
        console.error('Ping 失败:', err);
    }
};

const handleDelete = async (id: string) => {
    if (!id) return;
    const ok = confirm('确定要移除此节点吗？相关凭据也将被清理。');
    if (!ok) return;

    try {
        // 直接调用节点服务删除，减少事件链路问题
        await nodeService.deleteNode(id);
        // 通知父层刷新（保持向后兼容）
        emit('update-list');
    } catch (err) {
        console.error('[ServerManager] delete failed', err);
        alert('删除失败，请查看控制台日志');
    }
};
</script>

<template>
    <div class="space-y-4">
        <!-- 统计与操作栏 -->
        <div class="bg-white p-5 rounded-xl border border-slate-100 shadow-sm flex items-center justify-between">
            <div class="flex space-x-10">
                <div class="flex flex-col">
                    <span class="text-[10px] font-black text-slate-400 uppercase tracking-widest mb-1">集群节点总数</span>
                    <span class="text-2xl font-black text-slate-800 tabular-nums">
                        {{ servers.length }} <span class="text-xs font-bold text-slate-300 ml-1">UNITS</span>
                    </span>
                </div>
                <div class="w-[1px] h-10 bg-slate-100 self-center"></div>
            </div>

            <div class="flex items-center space-x-3">
                <button @click="handleBatchTest" :disabled="nodeService.loading.value || isBatchTesting"
                    class="px-5 py-2.5 bg-indigo-50 text-indigo-600 rounded-xl text-xs font-black flex items-center space-x-3 hover:bg-indigo-100 transition-all border border-indigo-100 active:scale-95 disabled:opacity-50">
                    <i :class="['fa-solid', isBatchTesting ? 'fa-spinner fa-spin' : 'fa-bolt-lightning']"></i>
                    <span class="uppercase tracking-widest">{{ isBatchTesting ? '全量拨测中...' : '一键全量测试' }}</span>
                </button>

                <button @click="openAddModal"
                    class="bg-slate-800 text-white px-6 py-2.5 rounded-xl text-xs font-black flex items-center space-x-3 shadow-xl shadow-slate-200 hover:bg-slate-900 transition-all active:scale-95">
                    <i class="fa-solid fa-plus-circle text-base"></i>
                    <span class="uppercase tracking-widest">添加服务器节点</span>
                </button>
            </div>
        </div>

        <!-- 节点列表表格 -->
        <div v-if="servers.length > 0"
            class="bg-white rounded-2xl border border-slate-100 shadow-xl shadow-slate-200/50 overflow-hidden">
            <table class="w-full">
                <thead>
                    <tr class="bg-slate-50/80 border-b border-slate-100">
                        <th
                            class="px-6 py-4 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center w-32 border-r border-slate-50">
                            连接状态</th>
                        <th class="px-6 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-left">
                            节点名称 & IP 地址</th>
                        <th class="px-6 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-left">
                            架构角色</th>
                        <th class="px-6 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-left">
                            通信协议</th>
                        <th
                            class="px-6 py-5 text-[10px] font-black text-slate-400 uppercase tracking-widest text-center w-40">
                            快捷操作列</th>
                    </tr>
                </thead>
                <tbody class="divide-y divide-slate-50">
                    <tr v-for="server in servers" :key="server.id" class="hover:bg-blue-50/30 transition-colors">
                        <!-- 状态列：居中 -->
                        <td class="px-6 py-5 border-r border-slate-50/50">
                            <div class="flex justify-center">
                                <div v-if="server.status === 'testing'" class="flex space-x-1">
                                    <span class="w-1.5 h-4 bg-blue-500 rounded-full animate-bounce"></span>
                                    <span
                                        class="w-1.5 h-4 bg-blue-500 rounded-full animate-bounce [animation-delay:0.2s]"></span>
                                    <span
                                        class="w-1.5 h-4 bg-blue-500 rounded-full animate-bounce [animation-delay:0.4s]"></span>
                                </div>
                                <div v-else-if="server.status === 'connected'"
                                    class="flex items-center space-x-2 px-3 py-1 bg-emerald-50 text-emerald-600 rounded-lg border border-emerald-100 shadow-sm">
                                    <div class="w-2 h-2 rounded-full bg-emerald-500 animate-pulse"></div>
                                    <span class="text-[10px] font-black font-mono tracking-tighter">{{ server.delay ?
                                        server.delay + 'ms' : 'ONLINE' }}</span>
                                </div>
                                <div v-else
                                    class="flex items-center space-x-2 px-3 py-1 bg-slate-50 text-slate-400 rounded-lg border border-slate-100">
                                    <div class="w-2 h-2 rounded-full bg-slate-300"></div>
                                    <span class="text-[10px] font-black tracking-tighter uppercase">OFFLINE</span>
                                </div>
                            </div>
                        </td>

                        <!-- 节点信息 -->
                        <td class="px-6 py-5">
                            <div class="flex flex-col">
                                <span class="text-sm font-black text-slate-700 leading-tight">{{ server.name }}</span>
                                <span
                                    class="text-[10px] font-bold text-slate-400 mt-1 uppercase tracking-tighter border-l-2 border-slate-200 pl-2 font-mono">
                                    {{ server.ip }}:{{ server.port }}
                                </span>
                            </div>
                        </td>

                        <!-- 角色 -->
                        <td class="px-6 py-5">
                            <span v-if="server.isMaster"
                                class="inline-flex items-center px-2 py-0.5 bg-indigo-600 text-white rounded text-[10px] font-black uppercase tracking-widest shadow-sm">
                                <i class="fa-solid fa-crown mr-1.5 text-[8px]"></i>Master
                            </span>
                            <span v-else
                                class="inline-flex items-center px-2 py-0.5 bg-slate-100 text-slate-500 rounded text-[10px] font-black border border-slate-200 uppercase tracking-widest">
                                <i class="fa-solid fa-server mr-1.5 text-[8px]"></i>Slave
                            </span>
                        </td>

                        <!-- 协议 -->
                        <td class="px-6 py-5">
                            <span
                                class="text-[11px] font-black text-slate-500 uppercase tracking-widest bg-slate-50 border border-slate-200 px-2 py-1 rounded-md">
                                {{ server.protocol }}
                            </span>
                        </td>

                        <!-- 操作列：居中 & 永久显示 -->
                        <td class="px-6 py-5">
                            <div class="flex items-center justify-center space-x-3">
                                <button @click="handleSinglePing(server)" title="链路联通性拨测"
                                    class="w-9 h-9 flex items-center justify-center text-blue-600 bg-blue-50/50 hover:bg-blue-600 hover:text-white rounded-xl transition-all border border-blue-100 active:scale-90 shadow-sm">
                                    <i class="fa-solid fa-plug text-sm"></i>
                                </button>
                                <button @click="openEditModal(server)" title="节点一站式综合配置"
                                    class="w-9 h-9 flex items-center justify-center text-amber-600 bg-amber-50/50 hover:bg-amber-600 hover:text-white rounded-xl transition-all border border-amber-100 active:scale-90 shadow-sm">
                                    <i class="fa-solid fa-cog text-sm"></i>
                                </button>
                                <button @click="handleDelete(server.id)" title="移除节点"
                                    class="w-9 h-9 flex items-center justify-center text-red-500 bg-red-50/50 hover:bg-red-500 hover:text-white rounded-xl transition-all border border-red-100 active:scale-90 shadow-sm">
                                    <i class="fa-solid fa-trash-can text-sm"></i>
                                </button>
                            </div>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>

        <!-- 空状态展示 -->
        <div v-else
            class="bg-white border-4 border-dashed border-slate-50 rounded-[2.5rem] py-24 flex flex-col items-center justify-center space-y-6">
            <div class="relative">
                <div class="w-32 h-32 bg-slate-50 rounded-full flex items-center justify-center">
                    <i class="fa-solid fa-server text-6xl text-slate-200"></i>
                </div>
                <div
                    class="absolute -bottom-2 -right-2 w-12 h-12 bg-white rounded-full flex items-center justify-center shadow-lg border border-slate-50">
                    <i class="fa-solid fa-question text-blue-400"></i>
                </div>
            </div>
            <div class="text-center">
                <h3 class="text-2xl font-black text-slate-800 tracking-tight">集群中尚未发现节点</h3>
                <p class="text-[11px] font-bold text-slate-400 mt-2 uppercase tracking-[0.2em]">请通过上方按钮注册您的第一个远程资源节点</p>
            </div>
            <button @click="openAddModal"
                class="px-10 py-3.5 bg-slate-800 text-white rounded-2xl text-xs font-black hover:bg-slate-900 shadow-2xl shadow-slate-200 transition-all active:scale-95 uppercase tracking-widest">
                立即极速添加
            </button>
        </div>

        <!-- 一站式综合配置对话框 -->
        <CredentialDialog :visible="isCredentialModalOpen" :node-data="currentNode" :is-edit="isEdit"
            @close="isCredentialModalOpen = false" @submit="handleFullSubmit" />
    </div>
</template>

<style scoped>
/* 确保表格行悬停时按钮列也有良好的背景一致性 */
tr:hover td {
    background-color: transparent;
}
</style>
