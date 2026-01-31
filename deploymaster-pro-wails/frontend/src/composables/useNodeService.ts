import { ref } from 'vue';
import {
    GetNodes,
    AddNode,
    UpdateNode,
    DeleteNode,
    TestNodeConnection,
    GetTopology,
    SaveCredential,
    SaveKeyPassphrase,
    DeleteCredential,
    HasStoredCredential,
    TestConnectionWithCredentials,
    SelectKeyFile
} from '../../wailsjs/go/main/App';
import { internal } from '../../wailsjs/go/models';
import { RemoteServer } from '../types';

// 全局共享状态
const servers = ref<RemoteServer[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

/**
 * 将 Go 后端模型转换为前端使用的 RemoteServer 类型
 */
const nodeToServer = (node: internal.Node): RemoteServer => {
    return {
        id: node.id,
        name: node.name,
        ip: node.ip,
        port: node.port,
        protocol: node.protocol as any,
        isMaster: node.isMaster,
        username: node.username,
        authMethod: node.authMethod as any,
        keyPath: node.keyPath,
    };
};

export function useNodeService() {
    /**
     * 加载所有节点 - 修正：保留当前已有的状态
     */
    const loadNodes = async () => {
        loading.value = true;
        error.value = null;
        try {
            const nodes = await GetNodes();
            const newServers = nodes.map(nodeToServer);

            // 合并现有状态，防止刷新后丢失“在线/延迟”标志
            servers.value = newServers.map(newS => {
                const existing = servers.value.find(s => s.id === newS.id);
                if (existing) {
                    return {
                        ...newS,
                        status: existing.status,
                        delay: existing.delay,
                        latency: existing.latency,
                        lastChecked: existing.lastChecked
                    };
                }
                return newS;
            });
        } catch (err: any) {
            error.value = `请求加载失败: ${err.message || err}`;
            console.error('加载节点失败:', err);
        } finally {
            loading.value = false;
        }
    };

    /**
     * 添加新节点
     */
    const addNode = async (server: Partial<RemoteServer>) => {
        loading.value = true;
        try {
            // 如果节点没有 ID（新添加），为其生成一个唯一 ID
            if (!server.id) {
                server.id = crypto.randomUUID();
            }
            const newNode = internal.Node.createFrom(server);
            await AddNode(newNode);
            await loadNodes();
            return server.id; // 返回生成的 ID
        } catch (err: any) {
            error.value = `添加节点失败: ${err.message || err}`;
            console.error('添加节点失败:', err);
            throw err;
        } finally {
            loading.value = false;
        }
    };

    /**
     * 更新现有节点
     */
    const updateNode = async (server: Partial<RemoteServer>) => {
        loading.value = true;
        try {
            if (!server.id) {
                throw new Error('更新节点失败：缺少节点 ID');
            }
            const node = internal.Node.createFrom(server);
            await UpdateNode(node);
            await loadNodes();
        } catch (err: any) {
            error.value = `更新节点失败: ${err.message || err}`;
            console.error('更新节点失败:', err);
            throw err;
        } finally {
            loading.value = false;
        }
    };

    /**
     * 删除节点
     */
    const deleteNode = async (id: string) => {
        loading.value = true;
        try {
            const target = servers.value.find(s => s.id === id);
            if (!target) {
                throw new Error('未找到要删除的节点');
            }

            await DeleteNode(id);

            // 本地立即移除，避免 UI 残留
            servers.value = servers.value.filter(s => s.id !== id);

            // 清理凭据：用户名可能为空，后端会忽略不存在的记录
            try {
                await DeleteCredential(id, target.username || '');
            } catch (credErr) {
                console.warn('删除节点成功，但清理凭据时出错:', credErr);
            }

            // 再从后端拉取一次，确保与磁盘同步（解耦加载态）
            // 注意：这里不再等待 loading 结束，以免阻塞 UI
            loadNodes();
        } catch (err: any) {
            error.value = `删除节点失败: ${err.message || err}`;
            console.error('删除节点失败:', err);
            throw err;
        } finally {
            loading.value = false;
        }
    };

    /**
     * 测试节点连接
     */
    const testConnection = async (nodeId: string, username: string = 'root', password: string = '') => {
        const serverIdx = servers.value.findIndex(s => s.id === nodeId);
        if (serverIdx !== -1) {
            servers.value[serverIdx].status = 'testing';
        }

        try {
            const status = await TestNodeConnection(nodeId, username, password);
            if (serverIdx !== -1) {
                // 使用 reactive 方式确保更新生效
                const updatedServer = {
                    ...servers.value[serverIdx],
                    latency: status.latency,
                    delay: status.latency,
                    status: (status.status === 'connected') ? 'connected' : 'disconnected' as any,
                    lastChecked: status.lastChecked ? new Date(status.lastChecked).toLocaleTimeString() : '--',
                };
                servers.value[serverIdx] = updatedServer;
            }
            return status;
        } catch (err: any) {
            if (serverIdx !== -1) {
                servers.value[serverIdx].status = 'disconnected';
            }
            console.error('测试连接失败:', err);
            throw err;
        }
    };

    /**
     * 批量测试所有节点
     */
    const testAllNodes = async () => {
        const promises = servers.value.map(async (server) => {
            try {
                // 仅对有基本信息的节点进行自动测试
                const hasCreds = await hasStoredCredential(server.id, server.username || 'root');
                if (hasCreds || server.authMethod === 'agent') {
                    return await testConnection(server.id, server.username || 'root');
                }
            } catch (e) {
                console.warn(`节点 ${server.id} 自动拨测跳过:`, e);
            }
        });
        return Promise.all(promises);
    };

    /**
     * 获取拓扑数据
     */
    const getTopology = async () => {
        try {
            const topology = await GetTopology();
            return {
                master: topology.master ? nodeToServer(topology.master) : null,
                slaves: topology.slaves ? topology.slaves.map(nodeToServer) : [],
                total: topology.total,
            };
        } catch (err: any) {
            console.error('获取拓扑数据失败:', err);
            throw err;
        }
    };

    /**
     * 保存节点凭据
     */
    const saveCredential = async (nodeId: string, username: string, password: string, rememberPassword: boolean = true) => {
        try {
            await SaveCredential(nodeId, username, password, rememberPassword);
        } catch (err: any) {
            console.error('保存凭据失败:', err);
            throw err;
        }
    };

    /**
     * 保存密钥密码短语
     */
    const saveKeyPassphrase = async (nodeId: string, passphrase: string, rememberPassphrase: boolean = true) => {
        try {
            await SaveKeyPassphrase(nodeId, passphrase, rememberPassphrase);
        } catch (err: any) {
            console.error('保存密钥密码短语失败:', err);
            throw err;
        }
    };

    /**
     * 检查是否已存储凭据
     */
    const hasStoredCredential = async (nodeId: string, username: string = 'root') => {
        try {
            return await HasStoredCredential(nodeId, username);
        } catch (err) {
            return false;
        }
    };

    /**
     * 测试连接并可选保存凭据
     */
    const testWithCredentials = async (
        nodeId: string,
        username: string,
        password: string,
        keyPassphrase: string,
        saveCredentials: boolean
    ) => {
        try {
            return await TestConnectionWithCredentials(
                nodeId,
                username,
                password,
                keyPassphrase,
                saveCredentials
            );
        } catch (err: any) {
            console.error('连接测试失败:', err);
            throw err;
        }
    };

    /**
     * 选择密钥文件
     */
    const selectKeyFile = async () => {
        try {
            return await SelectKeyFile();
        } catch (err: any) {
            console.error('选择文件失败:', err);
            return '';
        }
    };

    return {
        servers,
        loading,
        error,
        loadNodes,
        addNode,
        updateNode,
        deleteNode,
        testConnection,
        testAllNodes,
        getTopology,
        saveCredential,
        saveKeyPassphrase,
        hasStoredCredential,
        testWithCredentials,
        selectKeyFile
    };
}
