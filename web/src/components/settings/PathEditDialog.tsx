import { useState, useEffect } from 'react';
import { FolderOpen } from 'lucide-react';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
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
              <input
                id="path"
                type="text"
                value={formData.path}
                onChange={(e) =>
                  setFormData({ ...formData, path: e.target.value })
                }
                placeholder={type === 'llamacpp' ? '/usr/local/bin/llama.cpp' : '~/.cache/huggingface/hub'}
                className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent transition-all"
                required
              />
              <p className="text-[11px] text-muted-foreground">
                {type === 'llamacpp'
                  ? 'llama.cpp 可执行文件所在目录的绝对路径'
                  : '包含 GGUF 模型文件的目录绝对路径'}
              </p>
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
