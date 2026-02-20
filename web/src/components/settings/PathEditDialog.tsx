import { useState, useEffect } from 'react';
import { FolderOpen, CheckCircle2, XCircle, Loader2 } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { DirectoryBrowser } from '@/components/settings/DirectoryBrowser';
import { cn } from '@/lib/utils';
import type { LlamaCppPathConfig, ModelPathConfig } from '@/lib/configTypes';

interface PathEditDialogProps {
  open: boolean;
  type: 'llamacpp' | 'models';
  path?: LlamaCppPathConfig | ModelPathConfig;
  onSave: (path: LlamaCppPathConfig | ModelPathConfig) => Promise<void>;
  onClose: () => void;
}

export function PathEditDialog({
  open,
  type,
  path,
  onSave,
  onClose,
}: PathEditDialogProps) {
  const isEdit = !!path;
  const typeLabel = type === 'llamacpp' ? 'Llama.cpp' : '模型';

  const [formData, setFormData] = useState({
    name: '',
    path: '',
    description: '',
  });
  const [isSaving, setIsSaving] = useState(false);
  const [isBrowserOpen, setIsBrowserOpen] = useState(false);
  const [pathValidation, setPathValidation] = useState<{
    valid: boolean;
    checking: boolean;
    message?: string;
  }>({ valid: false, checking: false });

  // 当 path 变化时更新表单数据
  useEffect(() => {
    if (path) {
      setFormData({
        name: path.name || '',
        path: path.path || '',
        description: path.description || '',
      });
    } else {
      setFormData({
        name: '',
        path: '',
        description: '',
      });
    }
  }, [path, open]);

  // 路径验证
  useEffect(() => {
    if (!formData.path) {
      setPathValidation({ valid: false, checking: false });
      return;
    }

    const validatePath = async () => {
      setPathValidation({ valid: false, checking: true });
      // 简单验证:检查路径是否为绝对路径
      const isUnixPath = formData.path.startsWith('/');
      const isWindowsPath = /^[a-zA-Z]:\\/.test(formData.path);
      const isValid = isUnixPath || isWindowsPath;

      setPathValidation({
        valid: isValid,
        checking: false,
        message: isValid ? '路径格式有效' : '请输入绝对路径',
      });
    };

    const timeoutId = setTimeout(validatePath, 500);
    return () => clearTimeout(timeoutId);
  }, [formData.path]);

  // 处理目录选择
  const handleDirectorySelect = (selectedPath: string) => {
    setFormData({ ...formData, path: selectedPath });
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // 验证
    if (!formData.path.trim()) {
      return;
    }

    setIsSaving(true);
    try {
      const dataToSave = {
        path: formData.path.trim(),
        name: formData.name.trim() || undefined,
        description: formData.description.trim() || undefined,
      };

      await onSave(dataToSave);
      onClose();
    } catch (error) {
      console.error('保存路径失败:', error);
    } finally {
      setIsSaving(false);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onClose}>
      <DialogContent className="sm:max-w-[480px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle className="text-base">
              {isEdit ? '编辑' : '添加'} {typeLabel} 路径
            </DialogTitle>
          </DialogHeader>

          <div className="space-y-4 p-6">
            {/* 路径输入（必填） */}
            <div className="space-y-2">
              <label htmlFor="path" className="text-xs font-medium flex items-center gap-1.5">
                <FolderOpen size={12} className="text-muted-foreground" />
                路径 <span className="text-destructive">*</span>
              </label>
              <div className="flex gap-2">
                <div className="flex-1 relative">
                  <input
                    id="path"
                    type="text"
                    value={formData.path}
                    onChange={(e) =>
                      setFormData({ ...formData, path: e.target.value })
                    }
                    placeholder={type === 'llamacpp' ? '/usr/local/bin/llama.cpp' : '~/.cache/huggingface/hub'}
                    className={cn(
                      "w-full rounded-md border border-input bg-background px-3 py-2 pr-8 text-sm",
                      "focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all",
                      pathValidation.checking && "opacity-50"
                    )}
                    required
                  />
                  {/* 验证状态图标 */}
                  <div className="absolute right-2 top-1/2 -translate-y-1/2">
                    {pathValidation.checking ? (
                      <Loader2 className="w-4 h-4 animate-spin text-muted-foreground" />
                    ) : formData.path && (
                      pathValidation.valid ? (
                        <CheckCircle2 className="w-4 h-4 text-green-500" />
                      ) : (
                        <XCircle className="w-4 h-4 text-red-500" />
                      )
                    )}
                  </div>
                </div>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setIsBrowserOpen(true)}
                  className="h-9 px-3"
                >
                  <FolderOpen className="w-4 h-4 mr-1" />
                  浏览
                </Button>
              </div>
              <div className="flex items-center justify-between">
                <p className="text-[11px] text-muted-foreground">
                  {type === 'llamacpp'
                    ? 'llama.cpp 可执行文件所在目录的绝对路径'
                    : '包含 GGUF 模型文件的目录绝对路径'}
                </p>
                {pathValidation.message && (
                  <p className={cn(
                    "text-[11px]",
                    pathValidation.valid ? "text-green-600" : "text-red-600"
                  )}>
                    {pathValidation.message}
                  </p>
                )}
              </div>
            </div>

            {/* 名称输入（可选） */}
            <div className="space-y-2">
              <label htmlFor="name" className="text-xs font-medium">
                名称 <span className="text-muted-foreground">（可选）</span>
              </label>
              <input
                id="name"
                type="text"
                value={formData.name}
                onChange={(e) =>
                  setFormData({ ...formData, name: e.target.value })
                }
                placeholder={
                  type === 'llamacpp'
                    ? '例如：主构建目录'
                    : '例如：HuggingFace 缓存'
                }
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all"
              />
            </div>

            {/* 描述输入（可选） */}
            <div className="space-y-2">
              <label htmlFor="description" className="text-xs font-medium">
                描述 <span className="text-muted-foreground">（可选）</span>
              </label>
              <textarea
                id="description"
                value={formData.description}
                onChange={(e) =>
                  setFormData({ ...formData, description: e.target.value })
                }
                placeholder="添加备注信息..."
                rows={2}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent resize-none transition-all"
              />
            </div>
          </div>

          {/* 目录浏览器 */}
          <DirectoryBrowser
            open={isBrowserOpen}
            initialPath={formData.path}
            onSelect={handleDirectorySelect}
            onClose={() => setIsBrowserOpen(false)}
          />

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={onClose}
              disabled={isSaving}
              className="h-8 px-3 text-xs"
            >
              取消
            </Button>
            <Button
              type="submit"
              disabled={isSaving || !formData.path.trim()}
              className="h-8 px-3 text-xs"
            >
              {isSaving ? '保存中...' : '保存'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
