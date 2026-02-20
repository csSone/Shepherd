import { useState } from 'react';
import { Search, RefreshCw, Radar, Server, CheckCircle2, XCircle, Clock } from 'lucide-react';
import { Button } from '@/components/ui/button';
import {
  useClients,
  useClusterTasks,
  useNetworkScan,
  useClusterOverview,
  useFilteredClients,
  useFilteredTasks,
} from '@/features/cluster/hooks';
import { ClientCard } from '@/components/cluster/ClientCard';
import { cn } from '@/lib/utils';
import type { ClientStatus, TaskStatus } from '@/types';
import { useAlertDialog } from '@/hooks/useAlertDialog';

/**
 * 集群管理页面
 */
export function ClusterPage() {
  const alertDialog = useAlertDialog();
  const { data: clients, isLoading: clientsLoading } = useClients();
  const { data: tasks, isLoading: tasksLoading } = useClusterTasks();
  const { data: overview } = useClusterOverview();
  const networkScan = useNetworkScan();

  // UI 状态
  const [activeTab, setActiveTab] = useState<'clients' | 'tasks'>('clients');
  const [search, setSearch] = useState('');
  const [statusFilter, setStatusFilter] = useState<ClientStatus | ''>('');

  // 过滤客户端
  const filteredClients = useFilteredClients(clients, {
    search,
    status: statusFilter || undefined,
  });

  // 处理网络扫描
  const handleScan = async () => {
    const confirmed = await alertDialog.confirm({
      title: '扫描网络',
      description: '确定要扫描网络中的客户端吗？',
    });
    if (confirmed) {
      networkScan.mutate({});
    }
  };

  return (
    <div className="space-y-6">
      {/* 标题和操作 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-white">集群管理</h1>
          <p className="text-gray-600 dark:text-gray-400">
            {overview?.totalClients || 0} 个客户端，{overview?.onlineClients || 0} 个在线
          </p>
        </div>

        <Button
          onClick={handleScan}
          disabled={networkScan.isPending}
          variant="default"
          size="sm"
        >
          <Radar className={cn('w-4 h-4', networkScan.isPending && 'animate-spin')} />
          扫描网络
        </Button>
      </div>

      {/* 统计卡片 */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-blue-100 dark:bg-blue-900/30 rounded-lg">
              <Server className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {overview?.totalClients || 0}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-300">总客户端</div>
            </div>
          </div>
        </div>

        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-green-100 dark:bg-green-900/30 rounded-lg">
              <CheckCircle2 className="w-5 h-5 text-green-600 dark:text-green-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {overview?.onlineClients || 0}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-300">在线</div>
            </div>
          </div>
        </div>

        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-yellow-100 dark:bg-yellow-900/30 rounded-lg">
              <Clock className="w-5 h-5 text-yellow-600 dark:text-yellow-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {overview?.busyClients || 0}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-300">忙碌</div>
            </div>
          </div>
        </div>

        <div className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="flex items-center gap-3">
            <div className="p-2 bg-purple-100 dark:bg-purple-900/30 rounded-lg">
              <XCircle className="w-5 h-5 text-purple-600 dark:text-purple-400" />
            </div>
            <div>
              <div className="text-2xl font-bold text-gray-900 dark:text-white">
                {overview?.runningTasks || 0}
              </div>
              <div className="text-sm text-gray-600 dark:text-gray-300">运行中任务</div>
            </div>
          </div>
        </div>
      </div>

      {/* 标签切换 */}
      <div className="border-b border-gray-200 dark:border-gray-700">
        <nav className="flex gap-6">
          <Button
            onClick={() => setActiveTab('clients')}
            variant={activeTab === 'clients' ? 'default' : 'ghost'}
            size="sm"
            className={cn(
              'rounded-t-lg rounded-b-none border-b-2',
              activeTab === 'clients'
                ? 'border-primary'
                : 'border-transparent hover:border-transparent'
            )}
          >
            客户端 ({clients?.length || 0})
          </Button>
          <Button
            onClick={() => setActiveTab('tasks')}
            variant={activeTab === 'tasks' ? 'default' : 'ghost'}
            size="sm"
            className={cn(
              'rounded-t-lg rounded-b-none border-b-2',
              activeTab === 'tasks'
                ? 'border-primary'
                : 'border-transparent hover:border-transparent'
            )}
          >
            任务 ({tasks?.length || 0})
          </Button>
        </nav>
      </div>

      {/* 搜索和过滤 */}
      {activeTab === 'clients' && (
        <div className="flex items-center gap-3 p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="搜索客户端名称或地址..."
              className="w-full pl-10 pr-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>

          <select
            value={statusFilter}
            onChange={(e) => setStatusFilter(e.target.value as ClientStatus | '')}
            className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white"
          >
            <option value="">所有状态</option>
            <option value="online">在线</option>
            <option value="busy">忙碌</option>
            <option value="offline">离线</option>
            <option value="error">错误</option>
          </select>
        </div>
      )}

      {/* 客户端列表 */}
      {activeTab === 'clients' && (
        <>
          {clientsLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
            </div>
          ) : filteredClients.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-300">
              <Server className="w-12 h-12 mb-4" />
              <p className="text-lg mb-2">暂无客户端</p>
              <p className="text-sm">点击"扫描网络"来发现局域网中的客户端</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
              {filteredClients.map((client) => (
                <ClientCard key={client.id} client={client} />
              ))}
            </div>
          )}
        </>
      )}

      {/* 任务列表 */}
      {activeTab === 'tasks' && (
        <>
          {tasksLoading ? (
            <div className="flex items-center justify-center py-12">
              <div className="w-8 h-8 border-4 border-blue-600 border-t-transparent rounded-full animate-spin" />
            </div>
          ) : !tasks || tasks.length === 0 ? (
            <div className="flex flex-col items-center justify-center py-12 text-gray-500 dark:text-gray-300">
              <Clock className="w-12 h-12 mb-4" />
              <p className="text-lg">暂无任务</p>
            </div>
          ) : (
            <div className="space-y-3">
              {tasks.map((task) => (
                <div
                  key={task.id}
                  className="p-4 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700"
                >
                  <div className="flex items-start justify-between mb-3">
                    <div>
                      <h3 className="font-medium text-gray-900 dark:text-white">{task.type}</h3>
                      <p className="text-sm text-gray-500 dark:text-gray-400">ID: {task.id}</p>
                    </div>
                    <span
                      className={cn(
                        'px-2 py-1 rounded text-xs font-medium',
                        task.status === 'completed'
                          ? 'bg-green-100 text-green-700 dark:bg-green-900/40 dark:text-green-300'
                          : task.status === 'running'
                            ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/40 dark:text-blue-300'
                            : task.status === 'failed'
                              ? 'bg-red-100 text-red-700 dark:bg-red-900/40 dark:text-red-300'
                              : 'bg-gray-100 text-gray-700 dark:bg-gray-700/50 dark:text-gray-300'
                      )}
                    >
                      {task.status}
                    </span>
                  </div>

                  {task.assignedTo && (
                    <div className="text-sm text-gray-600 dark:text-gray-300">
                      分配给: {task.assignedTo}
                    </div>
                  )}

                  {task.error && (
                    <div className="mt-2 p-2 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded text-sm text-red-700 dark:text-red-300">
                      {task.error}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  );
}
