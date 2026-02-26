import { useState, useRef, useEffect } from 'react';
import { X, HelpCircle, Loader2, ChevronDown, Info, RotateCcw } from 'lucide-react';
import { cn } from '@/lib/utils';
import type { LoadModelParams, ModelCapabilities } from '@/types';
import { useGPUs, useModelCapabilities, useSetModelCapabilities, useLlamacppBackends, useEstimateVRAM, useModelLoadConfig, useSaveModelLoadConfig, useDeleteModelLoadConfig, type SystemGPUInfo, type LlamacppBackend } from '@/features/models/hooks';
import { useOnlineNodes } from '@/features/cluster/hooks';
import type { UnifiedNode } from '@/types';

// NumberInput 组件 - 数字输入框
interface NumberInputProps {
  value: number | undefined;
  onChange: (value: number) => void;
  disabled?: boolean;
  min?: number;
  max?: number;
  step?: number;
  placeholder?: string;
  className?: string;
  allowNegative?: boolean;
  allowMinusOne?: boolean;
}

const NumberInput = ({
  value,
  onChange,
  disabled,
  min,
  max,
  step = 1,
  placeholder,
  className = '',
  allowNegative = false,
  allowMinusOne = false,
}: NumberInputProps) => {
  const [inputValue, setInputValue] = useState(String(value ?? ''));
  const [error, setError] = useState('');

  // 同步外部 value 变化
  useEffect(() => {
    if (value !== undefined && String(value) !== inputValue) {
      setInputValue(String(value));
      setError('');
    }
  }, [value]);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setInputValue(newValue);

    // 空值处理
    if (newValue === '') {
      setError('');
      return;
    }

    // 验证数字
    const num = Number(newValue);
    if (isNaN(num)) {
      setError('请输入有效数字');
      return;
    }

    // 验证范围
    if (min !== undefined && num < min && !(allowMinusOne && num === -1) && !(allowNegative && num < 0)) {
      setError(`最小值为 ${min}`);
      return;
    }
    if (max !== undefined && num > max) {
      setError(`最大值为 ${max}`);
      return;
    }

    // 特殊值验证
    if (allowMinusOne && num === -1) {
      setError('');
      onChange(-1);
      return;
    }

    if (allowNegative && num < 0) {
      setError('');
      onChange(num);
      return;
    }

    if (!allowNegative && !allowMinusOne && num < 0) {
      setError('不允许负值');
      return;
    }

    setError('');
    onChange(num);
  };

  const handleBlur = () => {
    if (inputValue === '' && value !== undefined) {
      setInputValue(String(value));
      setError('');
    }
  };

  return (
    <div>
      <input
        type="number"
        value={inputValue}
        onChange={handleChange}
        onBlur={handleBlur}
        disabled={disabled}
        min={allowMinusOne ? -1 : min}
        max={max}
        step={step}
        placeholder={placeholder}
        className={cn(
          "w-full px-2 py-1.5 text-sm",
          "border-2 rounded-md",
          error ? "border-red-500 dark:border-red-500" : "border-border",
          "bg-input",
          "text-foreground",
          "focus:outline-none focus:ring-2",
          error ? "focus:ring-red-500 focus:border-red-500" : "focus:ring-blue-500 focus:border-blue-500",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          "transition-colors",
          className
        )}
      />
      {error && (
        <p className="mt-1 text-xs text-red-600 dark:text-red-400">{error}</p>
      )}
    </div>
  );
};

interface LoadModelDialogProps {
  isOpen: boolean;
  onClose: () => void;
  onConfirm: (params: LoadModelParams) => void;
  modelId: string;
  modelName: string;
  modelPath?: string;
  isLoading?: boolean;
}

// 预设配置
const PRESETS = {
  fast: {
    name: '快速加载',
    description: '最小化内存占用，快速启动',
    params: {
      ctxSize: 4096,
      batchSize: 512,
      gpuLayers: 20,
      flashAttention: true,
      kvCacheUnified: false,
    } as Partial<LoadModelParams>
  },
  balanced: {
    name: '均衡模式',
    description: '性能与内存的平衡',
    params: {
      ctxSize: 8192,
      batchSize: 1024,
      gpuLayers: 35,
      flashAttention: true,
      kvCacheUnified: true,
    } as Partial<LoadModelParams>
  },
  performance: {
    name: '性能优先',
    description: '最大化性能，需要更多内存',
    params: {
      ctxSize: 16384,
      batchSize: 2048,
      gpuLayers: 99,
      flashAttention: true,
      kvCacheUnified: true,
      uBatchSize: 512,
    } as Partial<LoadModelParams>
  },
  max: {
    name: '最大配置',
    description: '最高性能，适合高端硬件',
    params: {
      ctxSize: 131072,
      batchSize: 4096,
      gpuLayers: 999,
      flashAttention: true,
      noMmap: true,
      lockMemory: true,
    } as Partial<LoadModelParams>
  }
};

// 参数帮助说明
const PARAM_HELP = {
  ctxSize: '模型一次能处理的文本最大长度，单位token，值越大内存占用越高',
  batchSize: '每次推理处理的样本数量，影响吞吐量',
  threads: '模型运行的线程数，-1表示自动选择',
  gpuLayers: '卸载到GPU的模型层数，-1表示全部，0表示仅CPU',
  temperature: '控制生成文本的随机性，值越高越多样',
  topP: '核采样阈值，值越低越保守',
  topK: '保留前K个最高概率的token',
  minP: '过滤概率低于此值的token',
  repeatPenalty: '惩罚重复token，值越高越抑制重复',
  presencePenalty: '鼓励模型使用新token，避免重复',
  frequencyPenalty: '惩罚高频token，值越高越抑制常见词',
  flashAttention: '启用Flash Attention加速，提升推理速度',
  noMmap: '禁止使用内存映射加载模型',
  lockMemory: '锁定模型到物理内存，防止被系统回收',
  uBatchSize: '微批大小，用于优化内存使用',
  parallelSlots: '并发处理的槽位数',
  kvCacheSize: 'KV缓存的最大token数量',
  kvCacheUnified: '启用共享KV缓存，提升多任务效率',
  kvCacheType: 'KV缓存的数据类型，f16精度较高，f32精度最高',
};

export function LoadModelDialog({
  isOpen,
  onClose,
  onConfirm,
  modelId,
  modelName,
  modelPath,
  isLoading = false,
}: LoadModelDialogProps) {
  // 获取在线节点列表（用于节点选择）
  const { data: onlineNodes = [] } = useOnlineNodes();

  // 获取 llama.cpp 后端列表
  const { data: llamacppBackends = [] } = useLlamacppBackends();

  // 模型加载配置相关 hooks
  const { data: loadConfigData, isLoading: isLoadingConfig } = useModelLoadConfig(isOpen ? modelId : '');
  const saveModelLoadConfig = useSaveModelLoadConfig();
  const deleteModelLoadConfig = useDeleteModelLoadConfig();

  // 初始化参数状态
  const [params, setParams] = useState<LoadModelParams>({
    modelId,
    ctxSize: 8192,
    batchSize: 4096,
    threads: 4,
    gpuLayers: 99,
    temperature: 0.7,
    topP: 0.95,
    topK: 40,
    repeatPenalty: 1.1,
    seed: -1,
    nPredict: -1,
    llamaCppPath: '/usr/local/bin',
    mainGpu: 'default',
    capabilities: {
      thinking: false,
      tools: false,
      translation: false,
      embedding: false,
    },
    flashAttention: true,
    noMmap: false,
    lockMemory: false,
    logitsAll: false,
    reranking: false,
    minP: 0.05,
    presencePenalty: 0.0,
    frequencyPenalty: 0.0,
    uBatchSize: 512,
    parallelSlots: 4,
    kvCacheSize: 8192,
    kvCacheUnified: true,
    kvCacheTypeK: 'f16',
    kvCacheTypeV: 'f16',
    directIo: 'default',
    disableJinja: false,
    chatTemplate: '',
    contextShift: false,
    extraArgs: '',
  });

  const [estimateResult, setEstimateResult] = useState<string | null>(null);
  const [loadingStatus, setLoadingStatus] = useState<string>('就绪');

  // Tooltip 状态
  const [activeTooltip, setActiveTooltip] = useState<string | null>(null);

  // 获取 GPU 列表（依赖 params.llamaCppPath）
  const { data: gpuData } = useGPUs(params.llamaCppPath);
  const gpus = gpuData?.gpus || [];
  const gpuDevices = gpuData?.devices || [];

  // 获取模型能力配置
  const { data: savedCapabilities } = useModelCapabilities(isOpen ? modelId : '');

  // 设置模型能力的 mutation
  const setModelCapabilities = useSetModelCapabilities();

  // 显存估算 mutation
  const estimateVRAM = useEstimateVRAM();

  // 当对话框打开时，加载已保存的能力配置
  useEffect(() => {
    if (isOpen && savedCapabilities) {
      setParams(prev => ({
        ...prev,
        capabilities: {
          thinking: savedCapabilities.thinking || false,
          tools: savedCapabilities.tools || false,
          translation: prev.capabilities?.translation || false,
          embedding: savedCapabilities.embedding || false,
          reranking: savedCapabilities.rerank || false,
        },
      }));
    }
  }, [isOpen, savedCapabilities]);

  // 当对话框打开时，加载已保存的模型加载配置
  useEffect(() => {
    if (isOpen && loadConfigData && !isLoadingConfig) {
      if (loadConfigData.exists && loadConfigData.config) {
        // 从保存的配置中恢复参数
        const savedConfig = loadConfigData.config.config;
        setParams(prev => ({
          ...prev,
          ...(savedConfig as Partial<LoadModelParams>),
        }));
      }
      // 如果不存在保存的配置，使用默认值（不改变当前状态）
    }
  }, [isOpen, loadConfigData, isLoadingConfig]);

  // 处理能力变化，应用约束规则并保存
  const handleCapabilityChange = (key: string, value: boolean) => {
    setParams(prev => {
      const currentCaps = prev.capabilities || {};

      // 应用约束规则
      let newCaps = { ...currentCaps, [key]: value };
      let newReranking = prev.reranking || false;

      // 规则 1: embedding 和 reranking 互斥（embedding 是非聊天能力）
      if (key === 'embedding' && value) {
        newReranking = false;
        newCaps.thinking = false;
        newCaps.tools = false;
      } else if (key === 'reranking' && value) {
        newCaps.embedding = false;
        newCaps.thinking = false;
        newCaps.tools = false;
      }

      // 规则 2: thinking 或 tools 启用时，禁用 embedding 和 reranking
      if ((key === 'thinking' || key === 'tools') && value) {
        newCaps.embedding = false;
        newReranking = false;
      }

      // 保存到服务器（只保存聊天相关的能力）
      const capabilitiesToSave = {
        thinking: newCaps.thinking || false,
        tools: newCaps.tools || false,
        rerank: newReranking,
        embedding: newCaps.embedding || false,
      };

      setModelCapabilities.mutate({
        modelId,
        capabilities: capabilitiesToSave,
      });

      return {
        ...prev,
        capabilities: newCaps,
        reranking: newReranking,
      };
    });
  };

  if (!isOpen) return null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoadingStatus('加载中...');

    // 保存当前配置
    try {
      await saveModelLoadConfig.mutateAsync({
        modelId,
        config: params,
      });
    } catch (error) {
      console.error('保存模型加载配置失败:', error);
      // 保存失败不影响加载操作
    }

    onConfirm(params);
  };

  const applyPreset = (presetParams: Partial<LoadModelParams>) => {
    setParams(prev => ({ ...prev, ...presetParams }));
  };

  // 重置为默认配置
  const handleResetConfig = async () => {
    // 删除保存的配置
    try {
      await deleteModelLoadConfig.mutateAsync(modelId);
    } catch (error) {
      console.error('删除模型加载配置失败:', error);
    }

    // 重置为默认参数
    setParams({
      modelId,
      ctxSize: 8192,
      batchSize: 4096,
      threads: 4,
      gpuLayers: 99,
      temperature: 0.7,
      topP: 0.95,
      topK: 40,
      repeatPenalty: 1.1,
      seed: -1,
      nPredict: -1,
      llamaCppPath: '/usr/local/bin',
      mainGpu: 'default',
      capabilities: {
        thinking: false,
        tools: false,
        translation: false,
        embedding: false,
      },
      flashAttention: true,
      noMmap: false,
      lockMemory: false,
      logitsAll: false,
      reranking: false,
      minP: 0.05,
      presencePenalty: 0.0,
      frequencyPenalty: 0.0,
      uBatchSize: 512,
      parallelSlots: 4,
      kvCacheSize: 8192,
      kvCacheUnified: true,
      kvCacheTypeK: 'f16',
      kvCacheTypeV: 'f16',
      directIo: 'default',
      disableJinja: false,
      chatTemplate: '',
      contextShift: false,
      extraArgs: '',
    });
  };

  // Tooltip 交互
  const handleTooltipEnter = (key: string) => {
    setActiveTooltip(key);
  };

  const handleTooltipLeave = () => {
    setActiveTooltip(null);
  };

  // 帮助图标按钮 - 苹果风格
  const HelpIconButton = ({ paramKey }: { paramKey: string }) => {
    const helpText = PARAM_HELP[paramKey as keyof typeof PARAM_HELP];
    const buttonRef = useRef<HTMLButtonElement>(null);
    const tooltipRef = useRef<HTMLDivElement>(null);
    const [position, setPosition] = useState({ top: 0, left: 0 });

    const updatePosition = () => {
      if (buttonRef.current && activeTooltip === paramKey) {
        const rect = buttonRef.current.getBoundingClientRect();
        // 计算位置：在按钮上方，tooltip 的高度约 60px，箭头约 6px，间距 8px
        setPosition({
          top: rect.top - 8, // 向上偏移
          left: rect.left + rect.width / 2,
        });
      }
    };

    useEffect(() => {
      if (activeTooltip === paramKey) {
        updatePosition();
        const handleScroll = () => updatePosition();
        window.addEventListener('scroll', handleScroll, true);
        window.addEventListener('resize', updatePosition);
        return () => {
          window.removeEventListener('scroll', handleScroll, true);
          window.removeEventListener('resize', updatePosition);
        };
      }
      // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [activeTooltip, paramKey]);

    const handleMouseEnter = () => {
      handleTooltipEnter(paramKey);
    };

    const handleMouseLeave = (e: React.MouseEvent) => {
      // 检查鼠标是否移动到了 tooltip 上
      const relatedTarget = e.relatedTarget as HTMLElement;
      if (relatedTarget &&
          relatedTarget.nodeType === Node.ELEMENT_NODE &&
          tooltipRef.current &&
          tooltipRef.current.contains(relatedTarget)) {
        return; // 鼠标移到了 tooltip 上，不隐藏
      }
      handleTooltipLeave();
    };

    return (
      <div className="relative inline-block">
        <button
          ref={buttonRef}
          type="button"
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
          onFocus={() => handleTooltipEnter(paramKey)}
          onBlur={(e) => {
            // 检查焦点是否移动到了 tooltip 上
            const relatedTarget = e.relatedTarget as HTMLElement;
            if (relatedTarget &&
                relatedTarget.nodeType === Node.ELEMENT_NODE &&
                tooltipRef.current &&
                tooltipRef.current.contains(relatedTarget)) {
              return; // 焦点移到了 tooltip 上，不隐藏
            }
            handleTooltipLeave();
          }}
          className="ml-1 w-2.5 h-2.5 rounded-full bg-gradient-to-br from-gray-100 to-gray-200 dark:from-gray-700 dark:to-gray-600
                     text-muted-foreground text-[10px] font-medium flex items-center justify-center
                     hover:from-blue-50 hover:to-blue-100 dark:hover:from-blue-900/40 dark:hover:to-blue-800/40
                     hover:text-blue-600 dark:hover:text-blue-400
                     focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-1
                     transition-all duration-200 cursor-help shadow-sm hover:shadow"
          aria-label={`查看 ${paramKey} 的帮助说明`}
          aria-describedby={`tooltip-${paramKey}`}
        >
          ?
        </button>

        {/* Tooltip */}
        {activeTooltip === paramKey && (
          <div
            ref={tooltipRef}
            id={`tooltip-${paramKey}`}
            role="tooltip"
            className="fixed z-[100]"
            style={{
              top: `${position.top}px`,
              left: `${position.left}px`,
              transform: 'translateX(-50%) translateY(-100%)',
              animation: 'tooltipFadeIn 0.2s ease-out forwards',
            }}
            onMouseEnter={() => handleTooltipEnter(paramKey)}
            onMouseLeave={handleTooltipLeave}
          >
            <style>{`
              @keyframes tooltipFadeIn {
                from {
                  opacity: 0;
                  transform: 'translateX(-50%) translateY(-10%)';
                }
                to {
                  opacity: 1;
                  transform: 'translateX(-50%) translateY(-100%)';
                }
              }
            `}</style>
            <div className="relative mb-1.5">
              {/* 提示框内容 */}
              <div className="max-w-xs px-4 py-3 bg-background/95 backdrop-blur-xl rounded-xl shadow-2xl border border-white/10">
                <div className="flex items-start gap-3">
                  <Info className="w-4 h-4 text-blue-400 mt-0.5 flex-shrink-0" />
                  <p className="text-sm text-foreground leading-relaxed">
                    {helpText || '暂无说明'}
                  </p>
                </div>
              </div>

              {/* 下方箭头 */}
              <div className="absolute -bottom-1.5 left-1/2 -translate-x-1/2">
                <div className="w-0 h-0 border-l-[6px] border-l-transparent border-r-[6px] border-r-transparent border-t-[6px] border-t-gray-900/95 backdrop-blur-xl" />
              </div>
            </div>
          </div>
        )}
      </div>
    );
  };

  const renderHelpButton = (paramKey: string) => <HelpIconButton paramKey={paramKey} />;

  // 下拉选择组件 - 用于布尔值和固定选项
  const SelectInput = ({
    value,
    onChange,
    disabled,
    children,
    className = '',
  }: {
    value: string | number | undefined;
    onChange: (e: React.ChangeEvent<HTMLSelectElement>) => void;
    disabled?: boolean;
    children: React.ReactNode;
    className?: string;
  }) => (
    <div className="relative">
      <select
        value={value ?? ''}
        onChange={onChange}
        disabled={disabled}
        className={cn(
          "w-full px-2 py-1.5 pr-8 text-sm",
          "border-2 border-border",
          "rounded-md bg-input",
          "text-foreground",
          "appearance-none cursor-pointer",
          "hover:border-blue-400 dark:hover:border-blue-500",
          "focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500",
          "disabled:opacity-50 disabled:cursor-not-allowed",
          "transition-colors",
          className
        )}
      >
        {children}
      </select>
      <ChevronDown className="absolute right-2 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground pointer-events-none" />
    </div>
  );

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/50">
      <div className="bg-card rounded-lg shadow-xl max-w-6xl w-full max-h-[90vh] flex flex-col">
        {/* 标题栏 */}
        <div className="flex items-center justify-between p-4 border-b border-border flex-shrink-0">
          <h2 className="text-lg font-semibold text-foreground">
            加载模型配置
          </h2>
          <button
            onClick={onClose}
            disabled={isLoading}
            className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-50"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 预设配置按钮 */}
        <div className="px-4 py-3 border-b border-border bg-muted/50 flex-shrink-0">
          <div className="flex items-center justify-between gap-2">
            <div className="flex items-center gap-2 flex-wrap">
              <span className="text-sm font-medium text-foreground mr-2">预设配置:</span>
              {Object.entries(PRESETS).map(([key, preset]) => (
                <button
                  key={key}
                  type="button"
                  onClick={() => applyPreset(preset.params)}
                  disabled={isLoading}
                  className={cn(
                    "px-3 py-1.5 text-sm font-medium rounded-md transition-all duration-200",
                    "border shadow-sm",
                    "hover:shadow-md hover:-translate-y-px active:translate-y-0",
                    "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
                    // 根据预设类型使用不同的样式
                    key === 'fast' && "bg-secondary text-secondary-foreground border-border hover:bg-secondary/80",
                    key === 'balanced' && "bg-primary/10 text-primary border-primary/30 hover:bg-primary/15",
                    key === 'performance' && "bg-accent text-accent-foreground border-border hover:bg-accent/80",
                    key === 'max' && "bg-destructive/10 text-destructive border-destructive/30 hover:bg-destructive/15",
                    "disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-sm"
                  )}
                  title={preset.description}
                >
                  {preset.name}
                </button>
              ))}
            </div>

            {/* 重置配置按钮 */}
            <button
              type="button"
              onClick={handleResetConfig}
              disabled={isLoading}
              className={cn(
                "flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium rounded-md transition-all duration-200",
                "border shadow-sm",
                "hover:shadow-md hover:-translate-y-px active:translate-y-0",
                "focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background",
                "bg-muted text-muted-foreground border-border hover:bg-muted/80",
                "disabled:opacity-40 disabled:cursor-not-allowed disabled:hover:translate-y-0 disabled:hover:shadow-sm"
              )}
              title="重置为默认配置"
            >
              <RotateCcw className="w-3.5 h-3.5" />
              重置
            </button>
          </div>
        </div>

        {/* 表单内容 */}
        <form onSubmit={handleSubmit} className="flex flex-col flex-1 min-h-0">
          <div className="flex-1 overflow-y-auto p-4 min-h-0">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* 左列：基础配置 */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-foreground pb-2 border-b border-border">
                基础配置
              </h3>

              {/* 模型信息 */}
              <div className="space-y-3">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    模型
                  </label>
                  <div className="px-3 py-2 bg-muted rounded-md text-foreground text-sm">
                    {modelName}
                  </div>
                  {modelPath && (
                    <div className="mt-1 text-xs text-muted-foreground truncate">
                      {modelPath}
                    </div>
                  )}
                </div>

                {/* Llama.cpp 版本选择 */}
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Llama.cpp 版本
                  </label>
                  <SelectInput
                    value={params.llamaCppPath || ''}
                    onChange={(e) => setParams({ ...params, llamaCppPath: e.target.value })}
                    disabled={isLoading}
                    className="w-full"
                  >
                    {llamacppBackends.length > 0 ? (
                      llamacppBackends.map((backend: LlamacppBackend) => (
                        <option
                          key={backend.path}
                          value={backend.path}
                          disabled={!backend.available}
                        >
                          {backend.name}
                          {backend.description && ` (${backend.description})`}
                          {!backend.available && ' - 不可用'}
                        </option>
                      ))
                    ) : (
                      <option value="" disabled>
                        未配置 llama.cpp 后端
                      </option>
                    )}
                  </SelectInput>
                  {llamacppBackends.length === 0 && (
                    <p className="mt-1 text-xs text-yellow-600 dark:text-yellow-400">
                      请在服务器配置中添加 llama.cpp 路径
                    </p>
                  )}
                </div>

                {/* 能力开关 */}
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    能力
                  </label>
                  <div className="border border-border rounded-lg p-3 bg-card">
                    <div className="space-y-2">
                      {/* 聊天能力 */}
                      <div className="text-xs text-muted-foreground uppercase tracking-wide mb-1">聊天能力</div>
                      {[
                        { key: 'thinking', label: '思考能力' },
                        {key: 'tools', label: '工具使用' },
                      ].map(({ key, label }) => (
                        <label key={key} className="flex items-center gap-2 text-sm text-foreground cursor-pointer hover:bg-accent p-1 rounded hover:bg-accent">
                          <input
                            type="checkbox"
                            checked={params.capabilities?.[key as keyof NonNullable<typeof params.capabilities>] || false}
                            onChange={(e) => handleCapabilityChange(key, e.target.checked)}
                            disabled={isLoading || (params.reranking || params.capabilities?.embedding)}
                            className="rounded border-border text-blue-600 focus:ring-blue-500 w-4 h-4"
                          />
                          <span>{label}</span>
                        </label>
                      ))}

                      {/* 非聊天能力 */}
                      <div className="text-xs text-muted-foreground uppercase tracking-wide mb-1 mt-2">非聊天能力</div>
                      {[
                        {key: 'translation', label: '直译' },
                        {key: 'embedding', label: '嵌入' },
                      ].map(({ key, label }) => (
                        <label key={key} className="flex items-center gap-2 text-sm text-foreground cursor-pointer hover:bg-accent p-1 rounded hover:bg-accent">
                          <input
                            type="checkbox"
                            checked={params.capabilities?.[key as keyof NonNullable<typeof params.capabilities>] || false}
                            onChange={(e) => handleCapabilityChange(key, e.target.checked)}
                            disabled={isLoading || ((params.capabilities?.thinking || params.capabilities?.tools) && key === 'embedding')}
                            className="rounded border-border text-blue-600 focus:ring-blue-500 w-4 h-4"
                          />
                          <span>{label}</span>
                        </label>
                      ))}
                    </div>
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    选择模型支持的功能能力（互斥：思考/工具与嵌入）
                  </p>
                </div>

                {/* 主GPU选择 */}
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    主GPU
                  </label>
                  <SelectInput
                    value={params.mainGpu || 'default'}
                    onChange={(e) => setParams({ ...params, mainGpu: e.target.value })}
                    disabled={isLoading || gpus.length === 0}
                    className="w-full"
                  >
                    <option value="default">默认</option>
                    {gpus.map((gpu: SystemGPUInfo) => (
                      <option key={gpu.id} value={gpu.id}>
                        {gpu.name}
                      </option>
                    ))}
                  </SelectInput>
                  {gpus.length === 0 && (
                    <p className="mt-1 text-xs text-yellow-600 dark:text-yellow-400">
                      未检测到GPU，请确保ROCm正确安装
                    </p>
                  )}
                </div>

                {/* 设备选择 */}
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    设备
                  </label>
                  <div className="border border-border rounded-lg p-3 bg-card">
                    {gpus.length > 0 ? (
                      <div className="space-y-1">
                        {gpus.map((gpu: SystemGPUInfo) => (
                          <label
                            key={gpu.id}
                            className="flex items-center gap-2 text-sm text-foreground cursor-pointer hover:bg-accent p-1 rounded"
                          >
                            <input
                              type="checkbox"
                              checked={gpu.available}
                              disabled={true}
                              className="rounded border-border text-blue-600 focus:ring-blue-500 w-4 h-4"
                            />
                            <span className="flex-1">
                              <span className="font-medium">{gpu.id}</span>
                              <span className="text-muted-foreground mx-1">·</span>
                              <span className="text-muted-foreground">{gpu.name}</span>
                              {gpu.totalMemory && (
                                <span className="ml-2 text-muted-foreground">
                                  总计 {gpu.totalMemory}
                                </span>
                              )}
                              {gpu.freeMemory && (
                                <span className="ml-2 text-green-600 dark:text-green-400">
                                  可用 {gpu.freeMemory}
                                </span>
                              )}
                            </span>
                            {gpu.available ? (
                              <span className="text-xs text-green-600 dark:text-green-400">就绪</span>
                            ) : (
                              <span className="text-xs text-red-600 dark:text-red-400">不可用</span>
                            )}
                          </label>
                        ))}
                      </div>
                    ) : (
                      <div className="text-sm text-muted-foreground text-center py-2">
                        未检测到GPU设备
                      </div>
                    )}
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    选择用于模型加载的设备
                  </p>
                </div>

                {/* 节点选择（仅多节点环境显示） */}
                {onlineNodes.length > 0 && (
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">
                      运行节点
                    </label>
                    <SelectInput
                      value={params.nodeId || 'auto'}
                      onChange={(e) => setParams({ 
                        ...params, 
                        nodeId: e.target.value === 'auto' ? undefined : e.target.value 
                      })}
                      disabled={isLoading}
                      className="w-full"
                    >
                      <option value="auto">
                        🎯 自动调度（推荐）- 系统选择最佳节点
                      </option>
                      <optgroup label="指定节点">
                        {onlineNodes.map((node: UnifiedNode) => (
                          <option key={node.id} value={node.id}>
                            {node.name} ({node.address}:{node.port})
                            {node.capabilities?.gpuCount && ` · ${node.capabilities.gpuCount} GPU`}
                            {node.resources?.gpuInfo?.[0] && 
                              ` · 显存 ${Math.round((node.resources.gpuInfo[0].totalMemory - node.resources.gpuInfo[0].usedMemory) / 1024 / 1024 / 1024)}GB 可用`
                            }
                          </option>
                        ))}
                      </optgroup>
                    </SelectInput>
                    <p className="mt-1 text-xs text-muted-foreground">
                      自动调度会根据 GPU 显存和负载选择最佳节点
                    </p>
                  </div>
                )}

                {/* 设备状态 */}
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    加载状态
                  </label>
                  <div className="px-3 py-2 bg-muted rounded-md min-h-[40px] text-sm">
                    {isLoading && loadingStatus === '加载中...' ? (
                      <span className="flex items-center gap-2">
                        <Loader2 className="w-4 h-4 animate-spin" />
                        {loadingStatus}
                      </span>
                    ) : (
                      <span className="text-muted-foreground">{loadingStatus}</span>
                    )}
                  </div>
                </div>
              </div>
            </div>

            {/* 右列：高级参数 */}
            <div className="space-y-4">
              <h3 className="text-sm font-semibold text-foreground pb-2 border-b border-border">
                高级参数
              </h3>

              {/* 上下文与加速 */}
              <div className="space-y-3">
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  上下文与加速
                </h4>
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      上下文窗口
                      {renderHelpButton('ctxSize')}
                    </label>
                    <NumberInput
                      value={params.ctxSize}
                      onChange={(v) => setParams({ ...params, ctxSize: v })}
                      disabled={isLoading}
                      min={512}
                      max={131072}
                      step={512}
                      placeholder="8192"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      Flash Attention
                      {renderHelpButton('flashAttention')}
                    </label>
                    <SelectInput
                      value={params.flashAttention ? 'on' : 'off'}
                      onChange={(e) => setParams({ ...params, flashAttention: e.target.value === 'on' })}
                      disabled={isLoading}
                    >
                      <option value="on">on</option>
                      <option value="off">off</option>
                    </SelectInput>
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      禁用内存映射
                      {renderHelpButton('noMmap')}
                    </label>
                    <SelectInput
                      value={params.noMmap ? 'true' : 'false'}
                      onChange={(e) => setParams({ ...params, noMmap: e.target.value === 'true' })}
                      disabled={isLoading}
                    >
                      <option value="true">true</option>
                      <option value="false">false</option>
                    </SelectInput>
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      锁定物理内存
                      {renderHelpButton('lockMemory')}
                    </label>
                    <SelectInput
                      value={params.lockMemory ? 'true' : 'false'}
                      onChange={(e) => setParams({ ...params, lockMemory: e.target.value === 'true' })}
                      disabled={isLoading}
                    >
                      <option value="true">true</option>
                      <option value="false">false</option>
                    </SelectInput>
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      GPU层数
                      {renderHelpButton('gpuLayers')}
                    </label>
                    <NumberInput
                      value={params.gpuLayers}
                      onChange={(v) => setParams({ ...params, gpuLayers: v })}
                      disabled={isLoading}
                      min={-1}
                      max={999}
                      step={1}
                      placeholder="-1 表示全部"
                      allowMinusOne={true}
                    />
                  </div>
                </div>
              </div>

              {/* 采样参数 */}
              <div className="space-y-3">
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  采样参数
                </h4>
                <div className="grid grid-cols-3 gap-2">
                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      温度
                      {renderHelpButton('temperature')}
                    </label>
                    <NumberInput
                      value={params.temperature}
                      onChange={(v) => setParams({ ...params, temperature: v })}
                      disabled={isLoading}
                      min={0}
                      max={2}
                      step={0.1}
                      placeholder="0.7"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      Top-P
                      {renderHelpButton('topP')}
                    </label>
                    <NumberInput
                      value={params.topP}
                      onChange={(v) => setParams({ ...params, topP: v })}
                      disabled={isLoading}
                      min={0}
                      max={1}
                      step={0.05}
                      placeholder="0.95"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      Top-K
                      {renderHelpButton('topK')}
                    </label>
                    <NumberInput
                      value={params.topK}
                      onChange={(v) => setParams({ ...params, topK: v })}
                      disabled={isLoading}
                      min={1}
                      max={1000}
                      step={1}
                      placeholder="40"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      Min-P
                      {renderHelpButton('minP')}
                    </label>
                    <NumberInput
                      value={params.minP}
                      onChange={(v) => setParams({ ...params, minP: v })}
                      disabled={isLoading}
                      min={0}
                      max={1}
                      step={0.01}
                      placeholder="0.05"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      重复惩罚
                      {renderHelpButton('repeatPenalty')}
                    </label>
                    <NumberInput
                      value={params.repeatPenalty}
                      onChange={(v) => setParams({ ...params, repeatPenalty: v })}
                      disabled={isLoading}
                      min={0}
                      max={2}
                      step={0.05}
                      placeholder="1.1"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      存在惩罚
                      {renderHelpButton('presencePenalty')}
                    </label>
                    <NumberInput
                      value={params.presencePenalty}
                      onChange={(v) => setParams({ ...params, presencePenalty: v })}
                      disabled={isLoading}
                      min={0}
                      max={2}
                      step={0.1}
                      placeholder="0.0"
                    />
                  </div>
                </div>
              </div>

              {/* 批处理与并发 */}
              <div className="space-y-3">
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  批处理与并发
                </h4>
                <div className="grid grid-cols-4 gap-2">
                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      批次大小
                      {renderHelpButton('batchSize')}
                    </label>
                    <NumberInput
                      value={params.batchSize}
                      onChange={(v) => setParams({ ...params, batchSize: v })}
                      disabled={isLoading}
                      min={64}
                      max={16384}
                      step={64}
                      placeholder="4096"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      微批大小
                      {renderHelpButton('uBatchSize')}
                    </label>
                    <NumberInput
                      value={params.uBatchSize}
                      onChange={(v) => setParams({ ...params, uBatchSize: v })}
                      disabled={isLoading}
                      min={64}
                      max={8192}
                      step={64}
                      placeholder="512"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      并发槽位
                      {renderHelpButton('parallelSlots')}
                    </label>
                    <NumberInput
                      value={params.parallelSlots}
                      onChange={(v) => setParams({ ...params, parallelSlots: v })}
                      disabled={isLoading}
                      min={1}
                      max={128}
                      step={1}
                      placeholder="4"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      线程数
                      {renderHelpButton('threads')}
                    </label>
                    <NumberInput
                      value={params.threads}
                      onChange={(v) => setParams({ ...params, threads: v })}
                      disabled={isLoading}
                      min={-1}
                      max={256}
                      step={1}
                      placeholder="-1 表示自动"
                      allowMinusOne={true}
                    />
                  </div>
                </div>
              </div>

              {/* KV缓存 */}
              <div className="space-y-3">
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  KV缓存
                </h4>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      缓存大小
                      {renderHelpButton('kvCacheSize')}
                    </label>
                    <NumberInput
                      value={params.kvCacheSize}
                      onChange={(v) => setParams({ ...params, kvCacheSize: v })}
                      disabled={isLoading}
                      min={512}
                      max={131072}
                      step={512}
                      placeholder="8192"
                    />
                  </div>

                  <div>
                    <label className="flex items-center text-xs font-medium text-foreground mb-1">
                      统一缓存
                      {renderHelpButton('kvCacheUnified')}
                    </label>
                    <SelectInput
                      value={params.kvCacheUnified ? 'true' : 'false'}
                      onChange={(e) => setParams({ ...params, kvCacheUnified: e.target.value === 'true' })}
                      disabled={isLoading}
                    >
                      <option value="true">true</option>
                      <option value="false">false</option>
                    </SelectInput>
                  </div>

                  <div>
                    <label className="text-xs font-medium text-foreground mb-1">
                      KV类型K
                    </label>
                    <SelectInput
                      value={params.kvCacheTypeK || 'f16'}
                      onChange={(e) => setParams({ ...params, kvCacheTypeK: e.target.value })}
                      disabled={isLoading}
                    >
                      <option value="f16">f16</option>
                      <option value="f32">f32</option>
                      <option value="q8_0">q8_0</option>
                      <option value="q4_0">q4_0</option>
                    </SelectInput>
                  </div>

                  <div>
                    <label className="text-xs font-medium text-foreground mb-1">
                      KV类型V
                    </label>
                    <SelectInput
                      value={params.kvCacheTypeV || 'f16'}
                      onChange={(e) => setParams({ ...params, kvCacheTypeV: e.target.value })}
                      disabled={isLoading}
                    >
                      <option value="f16">f16</option>
                      <option value="f32">f32</option>
                      <option value="q8_0">q8_0</option>
                      <option value="q4_0">q4_0</option>
                    </SelectInput>
                  </div>
                </div>
              </div>

              {/* 其他参数 */}
              <div className="space-y-3">
                <h4 className="text-xs font-semibold text-muted-foreground uppercase">
                  其他参数
                </h4>
                <div className="grid grid-cols-2 gap-2">
                  <div>
                    <label className="text-xs font-medium text-foreground mb-1">
                      随机种子
                    </label>
                    <NumberInput
                      value={params.seed}
                      onChange={(v) => setParams({ ...params, seed: v })}
                      disabled={isLoading}
                      min={-1}
                      max={4294967295}
                      step={1}
                      placeholder="-1 表示随机"
                      allowMinusOne={true}
                    />
                  </div>

                  <div>
                    <label className="text-xs font-medium text-foreground mb-1">
                      Max Tokens
                    </label>
                    <NumberInput
                      value={params.nPredict}
                      onChange={(v) => setParams({ ...params, nPredict: v })}
                      disabled={isLoading}
                      min={-1}
                      max={65536}
                      step={64}
                      placeholder="-1 表示无限"
                      allowMinusOne={true}
                    />
                  </div>

                  <div>
                    <label className="text-xs font-medium text-foreground mb-1">
                      DirectIO
                    </label>
                    <SelectInput
                      value={params.directIo || 'default'}
                      onChange={(e) => setParams({ ...params, directIo: e.target.value })}
                      disabled={isLoading}
                    >
                      <option value="default">default</option>
                      <option value="true">true</option>
                      <option value="false">false</option>
                    </SelectInput>
                  </div>

                  <div className="col-span-2">
                    <label className="text-xs font-medium text-foreground mb-1">
                      额外参数
                    </label>
                    <textarea
                      value={params.extraArgs || ''}
                      onChange={(e) => setParams({ ...params, extraArgs: e.target.value })}
                      disabled={isLoading}
                      rows={2}
                      placeholder="例如: --timeout 30 --grp-attn-n 8"
                      className="w-full px-2 py-1.5 text-sm border-2 border-border rounded-md bg-input text-foreground focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-y"
                    />
                    <p className="text-xs text-muted-foreground mt-1">
                      输入额外的命令行参数，用空格分隔
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
          </div>

          {/* 按钮区域 - 固定在底部 */}
          <div className="flex justify-end items-center gap-3 px-4 py-3 border-t border-border bg-card flex-shrink-0">
            <button
              type="button"
              onClick={onClose}
              disabled={isLoading}
              className="px-4 py-2 text-foreground hover:bg-accent rounded transition-colors disabled:opacity-50"
            >
              取消
            </button>

            <button
              type="button"
              onClick={async () => {
                setEstimateResult('计算中...');

                try {
                  const result = await estimateVRAM.mutateAsync({
                    modelId,
                    llamaBinPath: params.llamaCppPath || '/home/user/workspace/llama.cpp/build-rocm/bin',
                    ctxSize: params.ctxSize,
                    batchSize: params.batchSize,
                    uBatchSize: params.uBatchSize,
                    parallel: params.parallelSlots,
                    flashAttention: params.flashAttention,
                    kvUnified: params.kvCacheUnified,
                    cacheTypeK: params.kvCacheTypeK,
                    cacheTypeV: params.kvCacheTypeV,
                  });

                  if (result.vramGB) {
                    setEstimateResult(`约需 ${result.vramGB} GB 显存`);
                  } else if (result.error) {
                    setEstimateResult(`估算失败: ${result.error}`);
                  } else {
                    setEstimateResult('估算失败');
                  }
                } catch (error) {
                  setEstimateResult(`估算出错: ${error instanceof Error ? error.message : '未知错误'}`);
                }
              }}
              disabled={isLoading || estimateVRAM.isPending}
              className="px-4 py-2 text-sm border border-border rounded hover:bg-accent disabled:opacity-50"
            >
              {estimateVRAM.isPending ? '计算中...' : '估算显存'}
            </button>
            {estimateResult && (
              <span className="text-sm text-muted-foreground">{estimateResult}</span>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className={cn(
                'px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700 dark:bg-blue-500 dark:hover:bg-blue-600 transition-colors',
                isLoading && 'opacity-50 cursor-not-allowed'
              )}
            >
              {isLoading ? (
                <span className="flex items-center gap-2">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  加载中...
                </span>
              ) : (
                '开始加载'
              )}
            </button>
          </div>
      </form>
    </div>
  </div>
  );
}
