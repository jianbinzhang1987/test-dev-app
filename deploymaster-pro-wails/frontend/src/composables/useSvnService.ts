import { ref } from 'vue';
import {
  GetSVNResources,
  AddSVNResource,
  UpdateSVNResource,
  DeleteSVNResource,
  TestSVNConnection,
  RefreshSVNResource,
  SaveSVNCredential,
  HasStoredSVNCredential,
} from '../../wailsjs/go/main/App';
import { internal } from '../../wailsjs/go/models';
import { SVNResource, SVNTestResult } from '../types';

const resources = ref<SVNResource[]>([]);
const loading = ref(false);
const error = ref<string | null>(null);

const modelToResource = (res: internal.SVNResource): SVNResource => {
  return {
    id: res.id,
    url: res.url,
    name: res.name,
    type: res.type as any,
    revision: res.revision,
    status: res.status as any,
    lastChecked: res.lastChecked,
    size: res.size,
    username: res.username,
  };
};

export function useSvnService() {
  const loadResources = async () => {
    loading.value = true;
    error.value = null;
    try {
      const list = await GetSVNResources();
      resources.value = list.map(modelToResource);
    } catch (err: any) {
      error.value = `请求加载失败: ${err.message || err}`;
      console.error('加载 SVN 资源失败:', err);
    } finally {
      loading.value = false;
    }
  };

  const addResource = async (
    resource: Partial<SVNResource>,
    creds?: { username?: string; password?: string; remember?: boolean }
  ) => {
    loading.value = true;
    try {
      if (creds?.remember && creds.password && !creds.username) {
        throw new Error('SVN 用户名不能为空');
      }
      const created = await AddSVNResource(internal.SVNResource.createFrom(resource));
      if (creds?.remember && creds.password) {
        await SaveSVNCredential(created.id, creds.username || '', creds.password, true);
      }
      await loadResources();
      return created;
    } catch (err: any) {
      error.value = `添加资源失败: ${err.message || err}`;
      console.error('添加 SVN 资源失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const updateResource = async (
    resource: Partial<SVNResource>,
    creds?: { username?: string; password?: string; remember?: boolean }
  ) => {
    loading.value = true;
    try {
      if (!resource.id) {
        throw new Error('更新资源失败：缺少资源 ID');
      }
      if (creds?.remember && creds.password && !creds.username) {
        throw new Error('SVN 用户名不能为空');
      }
      await UpdateSVNResource(internal.SVNResource.createFrom(resource));
      if (creds?.remember && creds.password) {
        await SaveSVNCredential(resource.id, creds.username || '', creds.password, true);
      }
      await loadResources();
    } catch (err: any) {
      error.value = `更新资源失败: ${err.message || err}`;
      console.error('更新 SVN 资源失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const deleteResource = async (id: string) => {
    loading.value = true;
    try {
      await DeleteSVNResource(id);
      resources.value = resources.value.filter(r => r.id !== id);
      loadResources();
    } catch (err: any) {
      error.value = `删除资源失败: ${err.message || err}`;
      console.error('删除 SVN 资源失败:', err);
      throw err;
    } finally {
      loading.value = false;
    }
  };

  const testConnection = async (url: string, username: string, password: string, resourceID: string): Promise<SVNTestResult> => {
    const result = await TestSVNConnection(url, username, password, resourceID);
    return {
      ok: result.ok,
      revision: result.revision,
      message: result.message,
      durationMs: result.durationMs,
      checkedAt: result.checkedAt ? new Date(result.checkedAt).toLocaleString() : undefined,
    };
  };

  const refreshResource = async (id: string) => {
    const updated = await RefreshSVNResource(id);
    await loadResources();
    return updated;
  };

  const refreshAll = async () => {
    const ids = resources.value.map(r => r.id);
    await Promise.all(ids.map(id => RefreshSVNResource(id)));
    await loadResources();
  };

  const hasStoredCredential = async (resourceID: string, username: string) => {
    return HasStoredSVNCredential(resourceID, username);
  };

  return {
    resources,
    loading,
    error,
    loadResources,
    addResource,
    updateResource,
    deleteResource,
    testConnection,
    refreshResource,
    refreshAll,
    hasStoredCredential,
  };
}
