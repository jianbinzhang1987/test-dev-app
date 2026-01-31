<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { HasStoredSVNCredential, ShowMessageDialog } from '../../wailsjs/go/main/App';
import { SVNResource, SVNTestResult } from '../types';

const props = defineProps<{
  resources: SVNResource[];
  loading?: boolean;
  testConnection?: (payload: { url: string; username: string; password: string; resourceId?: string }) => Promise<SVNTestResult>;
  refreshAll?: () => Promise<void>;
}>();

const emit = defineEmits(['add', 'update', 'delete']);

const viewMode = ref<'grid' | 'list'>('grid');
const isModalOpen = ref(false);
const modalMode = ref<'add' | 'edit'>('add');
const currentRes = ref<Partial<SVNResource>>({});
const isTesting = ref(false);
const credential = ref({ username: '', password: '', remember: true });
const showPasswordMask = ref(false);

const openAddModal = () => {
  modalMode.value = 'add';
  currentRes.value = { type: 'file', status: 'online', revision: 'HEAD' };
  credential.value = { username: '', password: '', remember: true };
  showPasswordMask.value = false;
  isModalOpen.value = true;
};

const openEditModal = (res: SVNResource) => {
  modalMode.value = 'edit';
  currentRes.value = { ...res };
  credential.value = { username: res.username || '', password: '', remember: true };
  showPasswordMask.value = !!savedCredentialMap.value[res.id];
  isModalOpen.value = true;
};

const handleSubmit = () => {
  if (credential.value.remember && credential.value.password && !credential.value.username) {
    ShowMessageDialog('缺少用户名', '请先填写 SVN 用户名后再保存密码。', 'warning');
    return;
  }

  if (modalMode.value === 'add') {
    const newRes = {
      ...currentRes.value,
      lastChecked: new Date().toLocaleString().slice(0, 16),
      username: credential.value.username || '',
    } as SVNResource;
    emit('add', newRes, { ...credential.value });
  } else {
    emit('update', { ...currentRes.value, username: credential.value.username || '' } as SVNResource, { ...credential.value });
  }
  isModalOpen.value = false;
};

const handleTestConnection = async () => {
  if (!currentRes.value.url) {
    await ShowMessageDialog('缺少 URL', '请先输入 SVN 仓库 URL。', 'warning');
    return;
  }
  if (!props.testConnection) {
    await ShowMessageDialog('后端未就绪', 'SVN 后端未就绪，无法测试连接。', 'error');
    return;
  }

  isTesting.value = true;
  try {
    const result = await props.testConnection({
      url: currentRes.value.url,
      username: credential.value.username || '',
      password: credential.value.password || '',
      resourceId: currentRes.value.id || '',
    });
    currentRes.value.status = result.ok ? 'online' : 'error';
    if (result.revision) {
      currentRes.value.revision = result.revision;
    }
    currentRes.value.lastChecked = new Date().toLocaleString().slice(0, 16);
    await ShowMessageDialog(
      result.ok ? '连接成功' : '连接失败',
      result.ok ? '连接测试成功！SVN 仓库响应正常。' : `连接失败：${result.message || '未知错误'}`,
      result.ok ? 'info' : 'error'
    );
  } catch (err: any) {
    await ShowMessageDialog('连接失败', `${err.message || err}`, 'error');
  } finally {
    isTesting.value = false;
  }
};

const handleRefreshAll = async () => {
  if (!props.refreshAll) return;
  await props.refreshAll();
};

const savedCredentialMap = ref<Record<string, boolean>>({});

const refreshCredentialStatus = async () => {
  const entries = await Promise.all(
    props.resources.map(async (res) => {
      if (!res.id || !res.username) return [res.id, false] as const;
      const has = await HasStoredSVNCredential(res.id, res.username);
      return [res.id, has] as const;
    })
  );
  const map: Record<string, boolean> = {};
  for (const [id, has] of entries) {
    if (id) map[id] = has;
  }
  savedCredentialMap.value = map;
};

watch(
  () => props.resources.map(r => `${r.id}:${r.username || ''}`),
  () => {
    refreshCredentialStatus();
  },
  { immediate: true }
);

watch(
  () => credential.value.username,
  (val) => {
    currentRes.value.username = val || '';
  }
);

watch(
  () => currentRes.value.id,
  (id) => {
    if (!id) {
      showPasswordMask.value = false;
      return;
    }
    if (credential.value.password) {
      showPasswordMask.value = false;
      return;
    }
    showPasswordMask.value = !!savedCredentialMap.value[id];
  }
);

const passwordDisplay = computed(() => {
  if (showPasswordMask.value && !credential.value.password) {
    return '********';
  }
  return credential.value.password;
});

const handlePasswordInput = (evt: Event) => {
  const target = evt.target as HTMLInputElement;
  credential.value.password = target.value;
  showPasswordMask.value = false;
};

const getStatusClass = (status: string) => {
  switch (status) {
    case 'online': return 'text-emerald-600 bg-emerald-50 border-emerald-100';
    case 'syncing': return 'text-blue-600 bg-blue-50 border-blue-100 animate-pulse';
    case 'error': return 'text-red-600 bg-red-50 border-red-100';
    default: return '';
  }
};

const getStatusText = (status: string) => {
  switch (status) {
    case 'online': return '连接正常';
    case 'syncing': return '同步中';
    case 'error': return '鉴权失败';
    default: return '';
  }
};
</script>

<template>
  <div class="space-y-6">
    <!-- Stats & Action Bar -->
    <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
      <div class="bg-white border border-slate-200 rounded-lg p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">资源总数</p>
          <p class="text-xl font-black text-slate-800">{{ resources.length }}</p>
        </div>
        <div class="w-10 h-10 bg-blue-50 text-blue-500 rounded flex items-center justify-center">
          <i class="fa-solid fa-boxes-stacked"></i>
        </div>
      </div>
      <div class="bg-white border border-slate-200 rounded-lg p-4 shadow-sm flex items-center justify-between">
        <div>
          <p class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">离线/错误</p>
          <p class="text-xl font-black text-red-500">{{ resources.filter(r => r.status === 'error').length }}</p>
        </div>
        <div class="w-10 h-10 bg-red-50 text-red-500 rounded flex items-center justify-center">
          <i class="fa-solid fa-plug-circle-exclamation"></i>
        </div>
      </div>
      <div class="md:col-span-2 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-lg p-4 shadow-md text-white flex items-center justify-between relative overflow-hidden">
        <div class="relative z-10">
          <h4 class="font-bold text-sm mb-1">快速添加 SVN 库</h4>
          <p class="text-[10px] opacity-80 max-w-[200px]">支持直接导入 SVN URL，自动识别资源类型并进行连接测试。</p>
        </div>
        <button 
          @click="openAddModal"
          class="relative z-10 bg-white text-blue-700 px-4 py-1.5 rounded text-[11px] font-bold hover:bg-slate-50 transition-colors shadow-lg"
        >
          立即配置 <i class="fa-solid fa-arrow-right ml-1"></i>
        </button>
        <i class="fa-solid fa-code-commit absolute -right-4 -bottom-4 text-6xl text-white/10 -rotate-12"></i>
      </div>
    </div>

    <!-- Search & View Control -->
    <div class="flex justify-between items-center bg-white p-3 rounded-lg border border-slate-200 shadow-sm">
      <div class="relative w-96">
        <input 
          type="text" 
          placeholder="输入名称或 URL 进行筛选..." 
          class="w-full pl-9 pr-4 py-1.5 border border-slate-200 rounded text-xs focus:border-blue-500 focus:ring-1 focus:ring-blue-500 outline-none transition-all"
        />
        <i class="fa-solid fa-search absolute left-3 top-2.5 text-slate-400 text-xs"></i>
      </div>
      <div class="flex items-center space-x-4">
        <div class="flex border border-slate-200 rounded p-0.5 bg-slate-50">
          <button 
            @click="viewMode = 'grid'"
            class="p-1.5 rounded px-2.5 text-[10px] transition-all"
            :class="viewMode === 'grid' ? 'bg-white shadow-sm text-blue-600' : 'text-slate-400'"
          >
            <i class="fa-solid fa-table-cells-large"></i>
          </button>
          <button 
            @click="viewMode = 'list'"
            class="p-1.5 rounded px-2.5 text-[10px] transition-all"
            :class="viewMode === 'list' ? 'bg-white shadow-sm text-blue-600' : 'text-slate-400'"
          >
            <i class="fa-solid fa-list-ul"></i>
          </button>
        </div>
        <div class="h-6 w-[1px] bg-slate-100"></div>
        <button @click="handleRefreshAll" :disabled="props.loading" class="text-slate-500 hover:text-blue-600 p-1 transition-colors disabled:opacity-40" title="强制刷新全库">
          <i class="fa-solid fa-rotate"></i>
        </button>
      </div>
    </div>

    <!-- Grid View -->
    <div v-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2 xl:grid-cols-3 gap-5">
      <div v-for="res in resources" :key="res.id" :class="['bg-white border', res.status === 'error' ? 'border-red-100' : 'border-slate-200', 'rounded-xl hover:shadow-lg transition-all flex flex-col group']">
        <div class="p-5 flex-1">
          <div class="flex items-start justify-between mb-4">
            <div :class="['w-12 h-12 rounded-lg border flex items-center justify-center transition-colors', 
              res.status === 'error' ? 'bg-red-50 border-red-100 text-red-300' : 'bg-slate-50 border-slate-100 text-slate-400 group-hover:bg-blue-50 group-hover:border-blue-100 group-hover:text-blue-500']">
              <i :class="['fa-solid', res.type === 'file' ? 'fa-file-shield' : 'fa-folder-open', 'text-2xl']"></i>
            </div>
            <div class="flex flex-col items-end space-y-2">
              <span :class="['flex items-center space-x-1 text-[10px] font-bold px-1.5 py-0.5 rounded border', getStatusClass(res.status)]">
                <i :class="['fa-solid', res.status === 'error' ? 'fa-exclamation-triangle' : (res.status === 'syncing' ? 'fa-sync fa-spin' : 'fa-check-circle'), 'scale-75']"></i>
                <span>{{ getStatusText(res.status) }}</span>
              </span>
              <span v-if="savedCredentialMap[res.id]" class="text-[9px] font-bold px-1.5 py-0.5 rounded border border-emerald-100 bg-emerald-50 text-emerald-600">
                已保存密码
              </span>
              <span class="text-[9px] font-mono text-slate-400">ID: {{ res.id }}</span>
            </div>
          </div>
          
          <h4 class="font-black text-slate-800 text-sm mb-1 group-hover:text-blue-700 transition-colors">{{ res.name }}</h4>
          <p class="text-[10px] text-slate-400 font-mono mb-4 break-all line-clamp-2 leading-relaxed bg-slate-50/50 p-2 rounded border border-slate-100/50">
            {{ res.url }}
          </p>

          <div class="grid grid-cols-2 gap-4 text-[10px]">
            <div class="space-y-1">
              <p class="text-slate-400 font-bold uppercase tracking-tighter">最新修订号</p>
              <p class="font-mono font-black text-slate-700">R-{{ res.revision }}</p>
            </div>
            <div class="space-y-1">
              <p class="text-slate-400 font-bold uppercase tracking-tighter">{{ res.type === 'file' ? '文件大小' : '包含子项' }}</p>
              <p class="font-black text-slate-700">{{ res.size || (res.type === 'folder' ? '--' : '未知') }}</p>
            </div>
          </div>
        </div>

        <div class="px-5 py-3 border-t border-slate-50 flex items-center justify-between bg-slate-50/30 rounded-b-xl">
          <span class="text-[9px] text-slate-400">
            <i class="fa-regular fa-clock mr-1"></i>
            {{ res.lastChecked }}
          </span>
          <div class="flex items-center space-x-1">
             <button class="w-8 h-8 rounded-full flex items-center justify-center text-slate-400 hover:bg-white hover:text-blue-600 hover:shadow-sm transition-all border border-transparent hover:border-slate-100" title="查看变更日志">
              <i class="fa-solid fa-history text-xs"></i>
             </button>
             <button @click="openEditModal(res)" class="w-8 h-8 rounded-full flex items-center justify-center text-slate-400 hover:bg-white hover:text-blue-600 hover:shadow-sm transition-all border border-transparent hover:border-slate-100" title="资源设置">
              <i class="fa-solid fa-sliders text-xs"></i>
             </button>
             <button class="bg-white border border-slate-200 text-slate-700 px-3 py-1 rounded text-[10px] font-bold hover:border-blue-500 hover:text-blue-600 transition-all shadow-sm">
              预览目录
             </button>
          </div>
        </div>
      </div>

      <!-- Add New Resource Placeholder -->
      <button 
        @click="openAddModal"
        class="border-2 border-dashed border-slate-200 rounded-xl p-6 flex flex-col items-center justify-center text-slate-400 hover:border-blue-300 hover:text-blue-400 transition-all hover:bg-blue-50/30 min-h-[220px]"
      >
         <div class="w-12 h-12 rounded-full bg-slate-100 flex items-center justify-center mb-3">
           <i class="fa-solid fa-plus text-xl"></i>
         </div>
         <p class="font-bold text-xs uppercase tracking-widest">添加新资源</p>
         <p class="text-[10px] opacity-60 mt-1">从当前已连通的仓库库浏览</p>
      </button>
    </div>

    <!-- List View -->
    <div v-else class="bg-white border border-slate-200 rounded-lg shadow-sm overflow-hidden">
      <table class="w-full text-left text-xs">
        <thead>
          <tr class="bg-slate-50 border-b border-slate-200 text-[10px] font-black text-slate-500 uppercase tracking-tighter">
            <th class="px-6 py-3 w-10">类型</th>
            <th class="px-6 py-3">资源名称</th>
            <th class="px-6 py-3">SVN URL</th>
            <th class="px-6 py-3">修订号</th>
            <th class="px-6 py-3">检查状态</th>
            <th class="px-6 py-3">凭据</th>
            <th class="px-6 py-3 text-right">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-slate-100">
          <tr v-for="res in resources" :key="res.id" class="hover:bg-slate-50/50 transition-colors group">
            <td class="px-6 py-4">
              <i :class="['fa-solid', res.type === 'file' ? 'fa-file-zipper text-indigo-400' : 'fa-folder text-amber-400', 'text-lg']"></i>
            </td>
            <td class="px-6 py-4 font-bold text-slate-700">{{ res.name }}</td>
            <td class="px-6 py-4">
              <code class="bg-slate-100 px-1.5 py-0.5 rounded text-[10px] text-slate-500 font-mono truncate max-w-xs inline-block">{{ res.url }}</code>
            </td>
            <td class="px-6 py-4">
              <span class="font-black text-slate-500">R-{{ res.revision }}</span>
            </td>
            <td class="px-6 py-4">
              <span :class="['flex items-center space-x-1 text-[10px] font-bold px-1.5 py-0.5 rounded border inline-flex', getStatusClass(res.status)]">
                <i :class="['fa-solid', res.status === 'error' ? 'fa-exclamation-triangle' : (res.status === 'syncing' ? 'fa-sync fa-spin' : 'fa-check-circle'), 'scale-75']"></i>
                <span>{{ getStatusText(res.status) }}</span>
              </span>
            </td>
            <td class="px-6 py-4">
              <span v-if="savedCredentialMap[res.id]" class="text-[10px] font-bold px-2 py-0.5 rounded border border-emerald-100 bg-emerald-50 text-emerald-600">
                已保存
              </span>
              <span v-else class="text-[10px] font-bold px-2 py-0.5 rounded border border-slate-200 text-slate-400">
                未保存
              </span>
            </td>
            <td class="px-6 py-4 text-right">
              <div class="flex justify-end space-x-2">
                 <button @click="openEditModal(res)" class="text-blue-600 font-bold hover:underline">配置</button>
                 <button @click="emit('delete', res.id)" class="text-slate-400 hover:text-red-500"><i class="fa-solid fa-trash"></i></button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="isModalOpen" class="fixed inset-0 bg-slate-900/60 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <div class="bg-white w-full max-w-lg rounded-xl shadow-2xl overflow-hidden border border-slate-200">
          <div class="px-6 py-4 border-b border-slate-100 flex justify-between items-center bg-slate-50">
            <h3 class="text-sm font-black text-slate-800 uppercase tracking-widest">
              {{ modalMode === 'add' ? '添加 SVN 资源' : '配置资源属性' }}
            </h3>
            <button @click="isModalOpen = false" class="text-slate-400 hover:text-slate-600"><i class="fa-solid fa-xmark"></i></button>
          </div>
          
          <form @submit.prevent="handleSubmit" class="p-6 space-y-4">
            <div class="space-y-1">
              <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">资源显示名称</label>
              <input 
                required
                type="text" 
                v-model="currentRes.name"
                class="w-full px-3 py-2 border border-slate-200 rounded text-xs focus:border-blue-500 outline-none" 
                placeholder="例如: 财务结算系统-正式版"
              />
            </div>

            <div class="space-y-1">
              <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">SVN 仓库 URL</label>
              <div class="flex space-x-2">
                <input 
                  required
                  type="text" 
                  v-model="currentRes.url"
                  class="flex-1 px-3 py-2 border border-slate-200 rounded text-xs font-mono focus:border-blue-500 outline-none" 
                  placeholder="svn://..."
                />
                <button 
                  type="button"
                  @click="handleTestConnection"
                  :disabled="isTesting"
                  class="px-3 bg-slate-100 border border-slate-200 rounded text-[10px] font-bold hover:bg-slate-200 transition-colors"
                >
                  <i v-if="isTesting" class="fa-solid fa-sync fa-spin"></i>
                  <span v-else>测试</span>
                </button>
              </div>
            </div>

            <div class="grid grid-cols-1 gap-4">
              <div class="space-y-1">
                <label class="text-[10px] font-bold text-slate-500 uppercase tracking-wider">资源类型</label>
                <select 
                  v-model="currentRes.type"
                  class="w-full px-2 py-2 border border-slate-200 rounded text-xs bg-white outline-none"
                >
                  <option value="file">压缩包/文件 (File)</option>
                  <option value="folder">源码目录 (Folder)</option>
                </select>
              </div>
            </div>

            <div class="p-4 bg-slate-50 rounded border border-slate-100">
              <h5 class="text-[10px] font-bold text-slate-600 mb-2 uppercase tracking-widest border-b border-slate-200 pb-1">本地鉴权凭据</h5>
              <div class="space-y-2 mt-2">
                 <div class="flex items-center space-x-2">
                    <i class="fa-solid fa-user text-slate-300 text-[10px]"></i>
                    <input v-model="credential.username" type="text" placeholder="SVN 用户名" class="flex-1 bg-transparent border-none p-0 text-xs focus:ring-0 outline-none placeholder:text-slate-300" />
                 </div>
                 <div class="flex items-center space-x-2">
                    <i class="fa-solid fa-lock text-slate-300 text-[10px]"></i>
                    <input :value="passwordDisplay" @input="handlePasswordInput" type="password" placeholder="••••••••" class="flex-1 bg-transparent border-none p-0 text-xs focus:ring-0 outline-none placeholder:text-slate-300" />
                 </div>
                 <label class="flex items-center space-x-2 text-[10px] text-slate-400 pt-1">
                   <input v-model="credential.remember" type="checkbox" class="rounded border-slate-300 text-blue-600 focus:ring-blue-500" />
                   <span>记住密码（本地加密保存）</span>
                 </label>
              </div>
            </div>

            <div class="flex justify-end space-x-3 pt-4">
              <button 
                type="button"
                @click="isModalOpen = false"
                class="px-4 py-2 text-xs font-bold text-slate-500 hover:text-slate-700"
              >
                取消
              </button>
              <button 
                type="submit"
                class="px-6 py-2 bg-blue-600 text-white rounded text-xs font-bold shadow-lg hover:bg-blue-700 transition-all"
              >
                {{ modalMode === 'add' ? '创建资源' : '保存更改' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </Teleport>

    <!-- Help & Info -->
    <div class="mt-8 flex items-start space-x-4 p-5 bg-slate-100/50 border border-slate-200 rounded-lg">
      <div class="w-10 h-10 bg-white rounded shadow-sm border border-slate-200 flex items-center justify-center text-slate-400 shrink-0">
        <i class="fa-solid fa-circle-info"></i>
      </div>
      <div class="space-y-1">
        <h5 class="text-xs font-black text-slate-700 uppercase">关于修订号 (Revision)</h5>
        <p class="text-[10px] text-slate-500 leading-relaxed">
          系统默认拉取 SVN 路径对应的最新内容（HEAD）。
          若资源状态显示为“鉴权失败”，请前往<span class="text-blue-600 font-bold cursor-pointer">系统设置 - 凭据管理</span>更新 SVN 账号信息。
        </p>
      </div>
    </div>
  </div>
</template>
