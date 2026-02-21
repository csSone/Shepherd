import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { useModels } from '@/features/models/hooks';
import { formatBytes } from '@/lib/utils';
import { Package, Download, Network, Activity } from 'lucide-react';

/**
 * 仪表盘页面
 */
export function DashboardPage() {
  const { data: models, isLoading } = useModels();

  const stats = [
    {
      title: '总模型数',
      value: models?.length || 0,
      icon: Package,
      description: '已扫描的模型',
    },
    {
      title: '已加载',
      value: models?.filter((m) => m.isLoaded).length || 0,
      icon: Activity,
      description: '正在运行的模型',
    },
    {
      title: '下载任务',
      value: 0,
      icon: Download,
      description: '活跃的下载任务',
    },
    {
      title: '集群节点',
      value: 0,
      icon: Network,
      description: '在线的客户端',
    },
  ];

  if (isLoading) {
    return <div className="flex items-center justify-center h-full">加载中...</div>;
  }

  return (
    <div className="space-y-6">
      {/* 页面标题 */}
      <div>
        <h1 className="text-3xl font-bold text-foreground">仪表盘</h1>
        <p className="text-muted-foreground font-medium">Shepherd 模型管理系统概览</p>
      </div>

      {/* 统计卡片 */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        {stats.map((stat) => {
          const Icon = stat.icon;
          return (
            <Card key={stat.title}>
              <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                <CardTitle className="text-sm font-medium">{stat.title}</CardTitle>
                <Icon className="h-4 w-4 text-muted-foreground" />
              </CardHeader>
              <CardContent>
                <div className="text-2xl font-bold">{stat.value}</div>
                <p className="text-xs text-muted-foreground">{stat.description}</p>
              </CardContent>
            </Card>
          );
        })}
      </div>

      {/* 最近模型 */}
      <Card>
        <CardHeader>
          <CardTitle>最近模型</CardTitle>
          <CardDescription>最近扫描的模型列表</CardDescription>
        </CardHeader>
        <CardContent>
          {models && models.length > 0 ? (
            <div className="space-y-4">
              {models.slice(0, 5).map((model) => (
                <div key={model.id} className="flex items-center justify-between">
                  <div>
                    <div className="font-medium">{model.alias || model.name}</div>
                    <div className="text-sm text-muted-foreground">
                      {model.metadata.architecture} • {formatBytes(model.size)}
                    </div>
                  </div>
                  <div className="text-sm text-muted-foreground">
                    {new Date(model.scannedAt).toLocaleDateString()}
                  </div>
                </div>
              ))}
            </div>
          ) : (
            <div className="text-center text-muted-foreground py-8">
              暂无模型数据
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
