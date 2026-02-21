import { useState, useEffect } from 'react';
import { X, Cpu, HardDrive, Server, Thermometer, Zap, Activity, Terminal } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Progress } from '@/components/ui/progress';
import { useClient } from '@/features/cluster/hooks';
import type { Client } from '@/types';

interface ClientInfoDialogProps {
  client: Client | null;
  open: boolean;
  onClose: () => void;
}

/**
 * 格式化字节大小
 */
function formatSize(bytes: number): string {
  if (bytes === 0 || bytes === undefined) return '0 B';
  const units = ['B', 'KB', 'MB', 'GB', 'TB'];
  let size = bytes;
  let unitIndex = 0;

  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }

  return `${size.toFixed(2)} ${units[unitIndex]}`;
}

/**
 * 客户端详细信息对话框
 * 实时显示客户端的详细系统信息
 */
export function ClientInfoDialog({ client, open, onClose }: ClientInfoDialogProps) {
  // 使用 hook 获取实时客户端数据（每3秒刷新）
  const { data: liveClient } = useClient(client?.id || '', {
    enabled: open && !!client?.id,
  });

  // 使用实时数据或传入的客户端数据
  const displayClient = liveClient || client;
  const resources = displayClient?.resources;
  const capabilities = displayClient?.capabilities;

  // 计算使用率
  const cpuPercent = resources?.cpuPercent ?? 0;
  const memoryPercent = resources?.memoryTotal
    ? ((resources.memoryUsed || 0) / resources.memoryTotal) * 100
    : 0;
  const diskPercent = resources?.diskTotal
    ? ((resources.diskUsed || 0) / resources.diskTotal) * 100
    : 0;

  return (
    <Dialog open={open} onOpenChange={(open) => !open && onClose()}>
      <DialogContent className="max-w-4xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Server className="w-5 h-5" />
            客户端详细信息 - {displayClient?.name}
          </DialogTitle>
        </DialogHeader>

        {/* 基本信息 */}
        <div className="grid grid-cols-2 gap-4 mb-6">
          <div className="p-4 bg-muted rounded-lg">
            <h3 className="text-sm font-medium text-muted-foreground mb-2">基本信息</h3>
            <div className="space-y-1 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">ID:</span>
                <span className="font-mono">{displayClient?.id}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">地址:</span>
                <span>{displayClient?.address}:{displayClient?.port}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">状态:</span>
                <span className="capitalize">{displayClient?.status}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">内核版本:</span>
                <span className="font-mono text-xs">{resources?.kernelVersion || 'N/A'}</span>
              </div>
            </div>
          </div>

          <div className="p-4 bg-muted rounded-lg">
            <h3 className="text-sm font-medium text-muted-foreground mb-2">硬件规格</h3>
            <div className="space-y-1 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">CPU:</span>
                <span>{capabilities?.cpuCount || 0} 核心</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">内存:</span>
                <span>{formatSize(capabilities?.memory || 0)}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">GPU:</span>
                <span>{capabilities?.gpuCount || 0} 个</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">ROCm 版本:</span>
                <span className="font-mono">{resources?.rocmVersion || 'N/A'}</span>
              </div>
            </div>
          </div>
        </div>

        {/* 资源使用 */}
        <div className="space-y-6">
          <h3 className="text-lg font-semibold flex items-center gap-2">
            <Activity className="w-5 h-5" />
            实时资源使用
          </h3>

          {/* CPU */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Cpu className="w-4 h-4 text-blue-500" />
                <span className="font-medium">CPU 使用率</span>
              </div>
              <span className="text-sm font-mono">{cpuPercent.toFixed(1)}%</span>
            </div>
            <Progress
              value={cpuPercent}
              className="h-2"
            />
          </div>

          {/* 内存 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <HardDrive className="w-4 h-4 text-green-500" />
                <span className="font-medium">内存使用</span>
              </div>
              <span className="text-sm">
                {formatSize(resources?.memoryUsed || 0)} / {formatSize(resources?.memoryTotal || 0)}
                <span className="text-muted-foreground ml-2">
                  ({memoryPercent.toFixed(1)}%)
                </span>
              </span>
            </div>
            <Progress
              value={memoryPercent}
              className="h-2"
            />
          </div>

          {/* 磁盘 */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <HardDrive className="w-4 h-4 text-yellow-500" />
                <span className="font-medium">磁盘使用</span>
              </div>
              <span className="text-sm">
                {formatSize(resources?.diskUsed || 0)} / {formatSize(resources?.diskTotal || 0)}
                <span className="text-muted-foreground ml-2">
                  ({diskPercent.toFixed(1)}%)
                </span>
              </span>
            </div>
            <Progress
              value={diskPercent}
              className="h-2"
            />
          </div>

          {/* GPU 详情 */}
          {resources?.gpuInfo && resources.gpuInfo.length > 0 && (
            <div className="space-y-4">
              <h4 className="font-medium flex items-center gap-2">
                <Zap className="w-4 h-4 text-purple-500" />
                GPU 详情 ({resources.gpuInfo.length} 个)
              </h4>
              {resources.gpuInfo.map((gpu, index) => (
                <div key={index} className="p-4 bg-muted rounded-lg space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">GPU {gpu.index}: {gpu.name}</span>
                    <span className="text-sm text-muted-foreground">{gpu.vendor}</span>
                  </div>

                  {gpu.driverVersion && (
                    <div className="text-sm text-muted-foreground">
                      驱动版本: {gpu.driverVersion}
                    </div>
                  )}

                  {/* GPU 使用率 */}
                  <div className="space-y-1">
                    <div className="flex justify-between text-sm">
                      <span>GPU 使用率</span>
                      <span>{gpu.utilization.toFixed(1)}%</span>
                    </div>
                    <Progress value={gpu.utilization} className="h-1.5" />
                  </div>

                  {/* GPU 显存 */}
                  {gpu.totalMemory > 0 && (
                    <div className="space-y-1">
                      <div className="flex justify-between text-sm">
                        <span>显存使用</span>
                        <span>
                          {formatSize(gpu.usedMemory)} / {formatSize(gpu.totalMemory)}
                          {' '}
                          ({((gpu.usedMemory / gpu.totalMemory) * 100).toFixed(1)}%)
                        </span>
                      </div>
                      <Progress
                        value={(gpu.usedMemory / gpu.totalMemory) * 100}
                        className="h-1.5"
                      />
                    </div>
                  )}

                  {/* GPU 温度 */}
                  {gpu.temperature > 0 && (
                    <div className="flex items-center gap-2 text-sm">
                      <Thermometer className="w-4 h-4 text-orange-500" />
                      <span>温度: {gpu.temperature.toFixed(1)}°C</span>
                    </div>
                  )}

                  {/* GPU 功耗 */}
                  {gpu.powerUsage > 0 && (
                    <div className="flex items-center gap-2 text-sm">
                      <Zap className="w-4 h-4 text-yellow-500" />
                      <span>功耗: {gpu.powerUsage.toFixed(1)}W</span>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>

        {/* 元数据 */}
        {displayClient?.metadata && Object.keys(displayClient.metadata).length > 0 && (
          <div className="mt-6">
            <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
              <Terminal className="w-5 h-5" />
              元数据
            </h3>
            <div className="p-4 bg-muted rounded-lg">
              <table className="w-full text-sm">
                <tbody>
                  {Object.entries(displayClient.metadata).map(([key, value]) => (
                    <tr key={key} className="border-b border-border last:border-0">
                      <td className="py-2 text-muted-foreground w-1/3">{key}</td>
                      <td className="py-2 font-mono text-xs">{String(value)}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* 关闭按钮 */}
        <div className="flex justify-end mt-6">
          <Button onClick={onClose} variant="outline">
            关闭
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  );
}
