<template>
    <div v-if="visible" class="fixed inset-0 bg-black/50 flex items-center justify-center z-[100] transition-opacity">
        <div
            class="bg-white rounded-xl shadow-2xl w-full max-w-lg mx-4 overflow-hidden border border-slate-200 animate-in fade-in zoom-in duration-200 flex flex-col">
            <!-- 标题栏 -->
            <div class="px-6 py-5 border-b border-slate-100 flex items-center justify-between bg-slate-50/50">
                <div>
                    <h3 class="text-lg font-black text-slate-800 tracking-tight">
                        {{ isEdit ? '服务器节点配置' : '添加新服务器节点' }}
                    </h3>
                    <p class="text-[10px] font-bold text-slate-400 uppercase tracking-widest mt-0.5">
                        基础信息与安全凭据一站式设置
                    </p>
                </div>
                <button @click="handleCancel" class="text-slate-400 hover:text-slate-600 transition-colors">
                    <i class="fa-solid fa-times text-lg"></i>
                </button>
            </div>

            <!-- 内容区 -->
            <div class="px-8 py-6 space-y-6 max-h-[70vh] overflow-y-auto custom-scrollbar">
                <!-- 测试结果状态条 (反馈增强) -->
                <div v-if="testing || success || error" class="animate-in slide-in-from-top-2 duration-200">
                    <div v-if="testing"
                        class="flex items-center space-x-3 text-blue-600 bg-blue-50/50 p-2.5 rounded-lg border border-blue-100">
                        <i class="fa-solid fa-spinner fa-spin text-sm"></i>
                        <span class="text-[11px] font-black uppercase tracking-widest">正在拨测远程链路安全性...</span>
                    </div>
                    <div v-else-if="success"
                        class="flex items-center justify-between text-emerald-700 bg-emerald-50 p-3 rounded-lg border border-emerald-200 shadow-sm shadow-emerald-100/50">
                        <div class="flex items-center space-x-3">
                            <div
                                class="w-8 h-8 bg-emerald-500 rounded-full flex items-center justify-center text-white shadow-lg shadow-emerald-200">
                                <i class="fa-solid fa-check text-sm"></i>
                            </div>
                            <div>
                                <p class="text-xs font-black">连接测试成功</p>
                                <p class="text-[10px] font-bold opacity-60 uppercase tracking-tighter">Remote Peer is
                                    Responsive & Authorized</p>
                            </div>
                        </div>
                        <span
                            class="px-3 py-1 bg-white/50 rounded-full border border-emerald-200 text-[10px] font-black font-mono">OK</span>
                    </div>
                    <div v-else-if="error"
                        class="flex flex-col space-y-2 text-red-700 bg-red-50 p-3 rounded-lg border border-red-200 shadow-sm">
                        <div class="flex items-center space-x-2">
                            <i class="fa-solid fa-triangle-exclamation text-sm"></i>
                            <span class="text-xs font-black">认证或连接失败</span>
                        </div>
                        <p
                            class="text-[10px] font-bold leading-relaxed bg-white/40 p-2 rounded border border-red-100 font-mono">
                            {{ error }}</p>
                    </div>
                </div>

                <!-- 核心基础信息 -->
                <section class="space-y-4">
                    <label
                        class="block text-[11px] font-black text-blue-600 uppercase tracking-widest mb-2 border-l-4 border-blue-600 pl-2 text-left">
                        1. 节点基础信息
                    </label>
                    <div class="grid grid-cols-2 gap-4">
                        <div class="col-span-2 text-left">
                            <label
                                class="block text-[10px] font-bold text-slate-500 uppercase mb-1.5 ml-1">节点易记名称</label>
                            <input v-model="form.name" type="text" placeholder="例如：生产环境主节点-01"
                                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        </div>
                        <div class="text-left">
                            <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1.5 ml-1">IP
                                地址</label>
                            <input v-model="form.ip" type="text" placeholder="192.168.1.1"
                                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-mono font-bold text-slate-700 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        </div>
                        <div class="text-left">
                            <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1.5 ml-1">端口</label>
                            <input v-model.number="form.port" type="number"
                                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-mono font-bold text-slate-700 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                        </div>
                        <div class="text-left">
                            <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1.5 ml-1">节点角色</label>
                            <select v-model="form.isMaster"
                                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all">
                                <option :value="false">Slave (从节点)</option>
                                <option :value="true">Master (主节点)</option>
                            </select>
                        </div>
                        <div class="text-left">
                            <label class="block text-[10px] font-bold text-slate-500 uppercase mb-1.5 ml-1">协议</label>
                            <select v-model="form.protocol"
                                class="w-full px-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all">
                                <option value="SFTP">SFTP (推荐)</option>
                                <option value="SCP">SCP (开发中)</option>
                                <option value="FTP">FTP (暂不支持)</option>
                            </select>
                        </div>
                    </div>
                </section>

                <!-- 认证与凭据配置 -->
                <section class="space-y-4 pt-2 border-t border-slate-100">
                    <label
                        class="block text-[11px] font-black text-emerald-600 uppercase tracking-widest mb-2 border-l-4 border-emerald-600 pl-2 text-left">
                        2. SSH 认证凭据
                    </label>
                    <div class="grid grid-cols-3 gap-3">
                        <button v-for="method in authMethods" :key="method.value"
                            @click="form.authMethod = method.value" type="button" :class="[
                                'flex flex-col items-center justify-center py-2.5 px-2 rounded-lg border-2 transition-all duration-200 text-center',
                                form.authMethod === method.value
                                    ? 'bg-emerald-50 border-emerald-600 text-emerald-700 shadow-sm'
                                    : 'bg-white border-slate-100 text-slate-500 hover:border-slate-200 hover:bg-slate-50'
                            ]">
                            <i :class="[method.icon, 'text-base mb-1']"></i>
                            <span class="text-[11px] font-bold">{{ method.label }}</span>
                        </button>
                    </div>

                    <div class="grid grid-cols-1 gap-5">
                        <div class="text-left">
                            <label
                                class="block text-[11px] font-black text-slate-400 uppercase tracking-widest mb-2">SSH
                                用户名</label>
                            <div class="relative group">
                                <span
                                    class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                                    <i class="fa-solid fa-user text-xs"></i>
                                </span>
                                <input v-model="form.username" type="text" placeholder="root"
                                    class="w-full pl-9 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                            </div>
                        </div>

                        <!-- Password Auth -->
                        <div v-if="form.authMethod === 'password'"
                            class="space-y-4 animate-in slide-in-from-top-2 duration-200">
                            <div class="text-left">
                                <label
                                    class="block text-[11px] font-black text-slate-400 uppercase tracking-widest mb-2">SSH
                                    登录密码</label>
                                <div class="relative group">
                                    <span
                                        class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                                        <i class="fa-solid fa-lock text-xs"></i>
                                    </span>
                                    <input v-model="form.password" :type="showPassword ? 'text' : 'password'"
                                        :placeholder="hasStoredPassword ? '●●●●●● (已加密保存)' : '请输入您的 SSH 登录密码'"
                                        class="w-full pl-9 pr-10 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all font-mono" />
                                    <button @click="showPassword = !showPassword" type="button"
                                        class="absolute right-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600 transition-colors">
                                        <i :class="showPassword ? 'fa-solid fa-eye-slash' : 'fa-solid fa-eye'"></i>
                                    </button>
                                </div>
                            </div>
                            <div
                                class="flex items-center p-3 bg-blue-50/50 rounded-lg border border-blue-100 text-left">
                                <input v-model="form.rememberPassword" type="checkbox" id="rememberPassword"
                                    class="w-4 h-4 text-blue-600 border-slate-300 rounded focus:ring-blue-500" />
                                <label for="rememberPassword"
                                    class="ml-3 text-xs font-bold text-slate-600 cursor-pointer">
                                    安全存储：记住密码并在系统密钥链中加密
                                    <p class="text-[10px] text-slate-400 font-medium mt-0.5 uppercase tracking-tighter">
                                        支持 KeyChain / Credential Manager</p>
                                </label>
                            </div>
                        </div>

                        <!-- Key Auth -->
                        <div v-if="form.authMethod === 'key'"
                            class="space-y-4 animate-in slide-in-from-top-2 duration-200">
                            <div class="text-left">
                                <label
                                    class="block text-[11px] font-black text-slate-400 uppercase tracking-widest mb-2">私钥文件路径</label>
                                <div class="flex space-x-2">
                                    <div class="relative flex-1 group">
                                        <span
                                            class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                                            <i class="fa-solid fa-file-invoice text-xs"></i>
                                        </span>
                                        <input v-model="form.keyPath" type="text" placeholder="请选择或输入私钥路径..."
                                            class="w-full pl-9 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all" />
                                    </div>
                                    <button @click="onSelectKeyFile" type="button"
                                        class="px-4 py-2.5 bg-slate-800 text-white rounded-lg text-xs font-bold hover:bg-slate-900 shadow-sm transition-all active:scale-95">
                                        <i class="fa-solid fa-folder-open mr-1.5"></i> 浏览
                                    </button>
                                </div>
                            </div>
                            <div class="text-left">
                                <label
                                    class="block text-[11px] font-black text-slate-400 uppercase tracking-widest mb-2">密钥短语
                                    (Passphrase) <span
                                        class="text-[9px] lowercase italic text-slate-300">可选</span></label>
                                <div class="relative group">
                                    <span
                                        class="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 group-focus-within:text-blue-500 transition-colors">
                                        <i class="fa-solid fa-key text-xs"></i>
                                    </span>
                                    <input v-model="form.keyPassphrase" type="password" placeholder="如果密钥受密码保护请填写"
                                        class="w-full pl-9 pr-4 py-2.5 bg-slate-50 border border-slate-200 rounded-lg text-sm font-bold text-slate-700 focus:outline-none focus:ring-2 focus:ring-blue-500/20 focus:border-blue-500 transition-all font-mono" />
                                </div>
                                <div class="flex items-center mt-3 ml-1">
                                    <input v-model="form.rememberPassphrase" type="checkbox" id="rememberPassphrase"
                                        class="w-4 h-4 text-emerald-600 border-slate-300 rounded focus:ring-emerald-500" />
                                    <label for="rememberPassphrase"
                                        class="ml-2 text-xs font-bold text-slate-500 cursor-pointer">在系统密钥链中记住短语</label>
                                </div>
                            </div>
                        </div>

                        <!-- Agent Auth Hint -->
                        <div v-if="form.authMethod === 'agent'"
                            class="p-4 bg-amber-50 border border-amber-100 rounded-lg animate-in slide-in-from-top-2 duration-200 text-left">
                            <div class="flex items-start space-x-3">
                                <i class="fa-solid fa-user-shield text-amber-500 mt-0.5"></i>
                                <div>
                                    <p class="text-[13px] font-bold text-amber-800">使用 SSH Agent 认证</p>
                                    <p class="text-[11px] text-amber-700/80 mt-1 leading-relaxed">系统将尝试连接本地运行的
                                        <strong>ssh-agent</strong> 进行无感登录。请确保 Agent 已持有有效密钥。
                                    </p>
                                </div>
                            </div>
                        </div>
                    </div>
                </section>
            </div>

            <!-- 底部控制栏 -->
            <div class="px-8 py-5 border-t border-slate-100 bg-slate-50/30 flex justify-between items-center">
                <div class="text-[10px] text-slate-400 font-bold uppercase tracking-tight">
                    <i class="fa-solid fa-shield-halved mr-1 text-green-500/60"></i> 本地端点对点加密
                </div>
                <div class="flex space-x-3">
                    <button @click="handleCancel" type="button"
                        class="px-5 py-2 text-xs font-bold text-slate-500 hover:text-slate-800 transition-colors uppercase tracking-widest">取消</button>
                    <button @click="onTestConnection" :disabled="testing || !canSubmit" type="button"
                        class="px-5 py-2 border-2 border-blue-600 text-blue-600 rounded-lg text-xs font-black hover:bg-blue-600 hover:text-white disabled:opacity-30 transition-all active:scale-95 flex items-center justify-center min-w-[100px]">
                        <i v-if="testing" class="fa-solid fa-spinner fa-spin mr-2"></i>
                        {{ testing ? '验证中...' : '测试连接' }}
                    </button>
                    <button @click="handleSubmit" :disabled="testing || !canSubmit" type="button"
                        class="px-6 py-2 bg-slate-800 text-white rounded-lg text-xs font-black hover:bg-slate-900 shadow-lg shadow-slate-200 disabled:opacity-30 transition-all active:scale-95 flex items-center justify-center min-w-[100px]">
                        <i v-if="success" class="fa-solid fa-check mr-2"></i>
                        {{ success ? '保存完成' : '提交保存' }}
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue';
import type { RemoteServer } from '../types';
import { useNodeService } from '../composables/useNodeService';

const props = defineProps<{
    visible: boolean;
    nodeData: Partial<RemoteServer>;
    isEdit?: boolean;
}>();

const emit = defineEmits(['close', 'submit', 'success']);

const nodeService = useNodeService();

const authMethods = [
    { value: 'password', label: '密码认证', icon: 'fa-solid fa-key' },
    { value: 'key', label: '私钥文件', icon: 'fa-solid fa-id-card' },
    { value: 'agent', label: 'SSH Agent', icon: 'fa-solid fa-user-secret' },
] as const;

type AuthMethod = (typeof authMethods)[number]['value'];

const getDefaultForm = () => ({
    name: '',
    ip: '',
    port: 22,
    isMaster: false,
    protocol: 'SFTP',
    username: 'root',
    authMethod: 'password' as AuthMethod,
    password: '',
    keyPath: '',
    keyPassphrase: '',
    rememberPassword: true,
    rememberPassphrase: true,
});

const form = reactive(getDefaultForm());

const showPassword = ref(false);
const testing = ref(false);
const success = ref(false);
const error = ref('');
const hasStoredPassword = ref(false);

const checkStoredCredentials = async () => {
    if (props.isEdit && props.nodeData?.id) {
        hasStoredPassword.value = await nodeService.hasStoredCredential(props.nodeData.id, form.username || 'root');
    } else {
        hasStoredPassword.value = false;
    }
};

const resetForm = () => {
    Object.assign(form, getDefaultForm());
    error.value = '';
    success.value = false;
    testing.value = false;
    showPassword.value = false; // 强制隐藏
};

// 监听 visible 变化，开启时若非编辑模式则重置
watch(() => props.visible, (isVisible) => {
    if (isVisible) {
        error.value = '';
        success.value = false;
        showPassword.value = false; // 每次打开都默认隐藏
        if (!props.isEdit) {
            resetForm();
        }
    }
});

// 监听 nodeData 变化用于编辑模式
watch(() => props.nodeData, (val) => {
    if (val && props.isEdit) {
        if (val.name) form.name = val.name;
        if (val.ip) form.ip = val.ip;
        if (val.port) form.port = val.port;
        if (val.isMaster !== undefined) form.isMaster = val.isMaster;
        if (val.protocol) form.protocol = val.protocol;
        if (val.username) form.username = val.username;
        if (val.authMethod) form.authMethod = val.authMethod as AuthMethod;
        if (val.keyPath) form.keyPath = val.keyPath;
    }
}, { deep: true });

const canSubmit = computed(() => {
    if (!form.name || !form.ip || !form.port || !form.username) return false;
    if (form.authMethod === 'key' && !form.keyPath) return false;
    return true;
});

const onSelectKeyFile = async () => {
    const path = await nodeService.selectKeyFile();
    if (path) {
        form.keyPath = path;
    }
};

const onTestConnection = async () => {
    error.value = '';
    testing.value = true;
    success.value = false;

    try {
        const status = await nodeService.testWithCredentials(
            props.nodeData.id || '',
            form.username,
            form.password,
            form.keyPassphrase,
            false
        );

        if (status.status === 'connected') {
            success.value = true;
            setTimeout(() => { success.value = false; }, 3000);
        } else {
            error.value = status.errorMsg || '连接失败，请检查配置或网络';
        }
    } catch (err: any) {
        error.value = err.message || '连接测试异常：系统内部错误';
    } finally {
        testing.value = false;
    }
};

const handleSubmit = async () => {
    error.value = '';
    success.value = false;

    try {
        const nodeUpdate = {
            id: props.nodeData.id, // 确保编辑模式保留ID
            ...props.nodeData,
            name: form.name,
            ip: form.ip,
            port: form.port,
            isMaster: form.isMaster,
            protocol: form.protocol,
            username: form.username,
            authMethod: form.authMethod,
            keyPath: form.keyPath,
            // 仅当用户输入密码/短语时才传递，用于决定是否保存凭据
            _password: form.password?.trim() ? form.password : undefined,
            _keyPassphrase: form.keyPassphrase?.trim() ? form.keyPassphrase : undefined,
            _rememberPassword: form.rememberPassword,
            _rememberPassphrase: form.rememberPassphrase,
        };

        emit('submit', nodeUpdate);
        success.value = true;
        // 不再自动关闭，仅在 3 秒后清空成功状态以允许下一次操作
        setTimeout(() => {
            success.value = false;
        }, 3000);
    } catch (err: any) {
        error.value = err.message || '保存失败';
    }
};

const handleCancel = () => {
    emit('close');
};
</script>

<style scoped>
.animate-in {
    animation: fadeIn 0.2s ease-out;
}

@keyframes fadeIn {
    from {
        opacity: 0;
        transform: scale(0.98);
    }

    to {
        opacity: 1;
        transform: scale(1);
    }
}

.shake {
    animation: shake 0.4s cubic-bezier(.36, .07, .19, .97) both;
}

@keyframes shake {

    10%,
    90% {
        transform: translate3d(-1px, 0, 0);
    }

    20%,
    80% {
        transform: translate3d(2px, 0, 0);
    }

    30%,
    50%,
    70% {
        transform: translate3d(-4px, 0, 0);
    }

    40%,
    60% {
        transform: translate3d(4px, 0, 0);
    }
}

.custom-scrollbar::-webkit-scrollbar {
    width: 4px;
}

.custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
    background: #e2e8f0;
    border-radius: 10px;
}
</style>
