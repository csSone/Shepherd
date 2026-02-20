// Package node provides distributed node management implementation.
package node

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/shirou/gopsutil/v3/mem"
)

// ResourceMonitor 负责收集节点的资源信息
type ResourceMonitor struct {
	// 配置参数
	interval      time.Duration        // 采样间隔
	llamacppPaths []string             // llama.cpp 可执行文件路径
	callback      func(*NodeResources) // 资源更新回调函数

	// 运行时状态
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mu         sync.RWMutex
	lastUpdate time.Time
	startTime  time.Time

	// 资源数据
	resources    *NodeResources
	gpuInfo      []GPUInfo
	llamacppInfo *LlamacppInfo

	// 日志
	log *logger.Logger
}

// ResourceMonitorConfig 资源监控器配置
type ResourceMonitorConfig struct {
	Interval      time.Duration        // 采样间隔，默认5秒
	LlamacppPaths []string             // llama.cpp 可执行文件路径
	Callback      func(*NodeResources) // 资源更新回调
	Logger        *logger.Logger
}

// NewResourceMonitor 创建新的资源监控器
func NewResourceMonitor(config *ResourceMonitorConfig) *ResourceMonitor {
	if config == nil {
		config = &ResourceMonitorConfig{}
	}

	if config.Interval == 0 {
		config.Interval = 5 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 初始化资源数据
	resources := &NodeResources{
		CPUTotal:    int64(runtime.NumCPU()) * 1000, // 转换为 millicores
		MemoryTotal: 0,                              // 将在初始化时设置
		DiskTotal:   0,                              // 将在初始化时设置
		GPUInfo:     make([]GPUInfo, 0),
		LoadAverage: make([]float64, 3),
	}

	rm := &ResourceMonitor{
		interval:      config.Interval,
		llamacppPaths: config.LlamacppPaths,
		callback:      config.Callback,
		ctx:           ctx,
		cancel:        cancel,
		resources:     resources,
		gpuInfo:       make([]GPUInfo, 0),
		llamacppInfo:  nil,
		log:           config.Logger,
		startTime:     time.Now(),
	}

	return rm
}

// Start 启动资源监控
func (rm *ResourceMonitor) Start() error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.running {
		return fmt.Errorf("资源监控器已在运行")
	}

	rm.running = true
	rm.lastUpdate = time.Time{}

	// 初始化资源信息
	if err := rm.initializeResources(); err != nil {
		rm.log.Errorf("初始化资源信息失败: %v", err)
		// 不返回错误，继续运行
	}

	rm.wg.Add(1)
	go rm.monitorLoop()

	if rm.log != nil {
		rm.log.Infof("资源监控器已启动，采样间隔: %v", rm.interval)
	}
	return nil
}

// Stop 停止资源监控
func (rm *ResourceMonitor) Stop() error {
	rm.mu.Lock()
	if !rm.running {
		rm.mu.Unlock()
		return nil
	}

	rm.running = false
	rm.mu.Unlock()

	// 先取消context，让monitorLoop能够退出
	rm.cancel()
	rm.wg.Wait()

	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.log != nil {
		rm.log.Infof("资源监控器已停止")
	}
	return nil
}

// GetSnapshot 获取当前资源快照
func (rm *ResourceMonitor) GetSnapshot() *NodeResources {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// 返回资源数据的副本
	if rm.resources == nil {
		return nil
	}

	snapshot := *rm.resources
	if len(rm.gpuInfo) > 0 {
		snapshot.GPUInfo = make([]GPUInfo, len(rm.gpuInfo))
		copy(snapshot.GPUInfo, rm.gpuInfo)
	}

	return &snapshot
}

// GetGPUInfo 获取GPU信息
func (rm *ResourceMonitor) GetGPUInfo() []GPUInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	gpuInfo := make([]GPUInfo, len(rm.gpuInfo))
	copy(gpuInfo, rm.gpuInfo)
	return gpuInfo
}

// GetLlamacppInfo 获取llama.cpp信息
func (rm *ResourceMonitor) GetLlamacppInfo() *LlamacppInfo {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	if rm.llamacppInfo == nil {
		return nil
	}

	info := *rm.llamacppInfo
	return &info
}

// initializeResources 初始化资源信息
func (rm *ResourceMonitor) initializeResources() error {
	// 获取内存总量
	if vmStat, err := mem.VirtualMemory(); err == nil {
		rm.resources.MemoryTotal = int64(vmStat.Total)
	} else {
		if rm.log != nil {
			rm.log.Errorf("获取内存总量失败: %v", err)
		}
	}

	// 获取磁盘总量（根目录）
	if diskStat, err := disk.Usage("/"); err == nil {
		rm.resources.DiskTotal = int64(diskStat.Total)
	} else {
		if rm.log != nil {
			rm.log.Errorf("获取磁盘总量失败: %v", err)
		}
	}

	// 检测GPU
	rm.detectGPUs()

	// 检测llama.cpp
	rm.detectLlamacpp()

	return nil
}

// monitorLoop 监控循环
func (rm *ResourceMonitor) monitorLoop() {
	defer rm.wg.Done()

	// 立即执行一次更新
	rm.updateResources()

	ticker := time.NewTicker(rm.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.updateResources()
		}
	}
}

// updateResources 更新资源信息
func (rm *ResourceMonitor) updateResources() {
	// 检查context是否已取消
	select {
	case <-rm.ctx.Done():
		return
	default:
	}

	rm.mu.Lock()
	defer rm.mu.Unlock()

	rm.lastUpdate = time.Now()

	// 更新CPU使用率
	rm.updateCPUUsage()

	// 更新内存使用情况
	rm.updateMemoryUsage()

	// 更新磁盘使用情况
	rm.updateDiskUsage()

	// 更新GPU信息
	rm.updateGPUInfo()

	// 更新系统负载
	rm.updateLoadAverage()

	// 更新运行时间
	rm.resources.Uptime = int64(time.Since(rm.startTime).Seconds())

	// 调用回调函数（在锁外执行以避免死锁）
	var resourcesCopy *NodeResources
	if rm.callback != nil {
		// 创建副本避免并发访问问题
		resourcesCopy = &NodeResources{}
		*resourcesCopy = *rm.resources
		if len(rm.gpuInfo) > 0 {
			resourcesCopy.GPUInfo = make([]GPUInfo, len(rm.gpuInfo))
			copy(resourcesCopy.GPUInfo, rm.gpuInfo)
		}
	}
	rm.mu.Unlock()

	if rm.callback != nil && resourcesCopy != nil {
		rm.callback(resourcesCopy)
	}

	// 重新获取锁以保持函数语义
	rm.mu.Lock()
}

// updateCPUUsage 更新CPU使用率
func (rm *ResourceMonitor) updateCPUUsage() {
	if cpuPercent, err := cpu.Percent(0, false); err == nil && len(cpuPercent) > 0 {
		// 转换为 millicores
		rm.resources.CPUUsed = int64(cpuPercent[0] * float64(rm.resources.CPUTotal) / 100.0)
	} else {
		if rm.log != nil {
			rm.log.Errorf("获取CPU使用率失败: %v", err)
		}
		rm.resources.CPUUsed = 0
	}
}

// updateMemoryUsage 更新内存使用情况
func (rm *ResourceMonitor) updateMemoryUsage() {
	if vmStat, err := mem.VirtualMemory(); err == nil {
		rm.resources.MemoryUsed = int64(vmStat.Used)
	} else {
		if rm.log != nil {
			rm.log.Errorf("获取内存使用情况失败: %v", err)
		}
		rm.resources.MemoryUsed = 0
	}
}

// updateDiskUsage 更新磁盘使用情况
func (rm *ResourceMonitor) updateDiskUsage() {
	if diskStat, err := disk.Usage("/"); err == nil {
		rm.resources.DiskUsed = int64(diskStat.Used)
	} else {
		if rm.log != nil {
			rm.log.Errorf("获取磁盘使用情况失败: %v", err)
		}
		rm.resources.DiskUsed = 0
	}
}

// updateGPUInfo 更新GPU信息
func (rm *ResourceMonitor) updateGPUInfo() {
	// 定期重新检测GPU（每分钟）
	if len(rm.gpuInfo) == 0 || time.Since(rm.lastUpdate) > time.Minute {
		rm.detectGPUs()
	}

	// 更新现有GPU的使用情况
	for i := range rm.gpuInfo {
		switch rm.gpuInfo[i].Vendor {
		case "NVIDIA":
			rm.updateNvidiaGPU(&rm.gpuInfo[i])
		case "AMD":
			rm.updateAMDGPU(&rm.gpuInfo[i])
		case "Intel":
			rm.updateIntelGPU(&rm.gpuInfo[i])
		}
	}
}

// updateLoadAverage 更新系统负载
func (rm *ResourceMonitor) updateLoadAverage() {
	if loadStat, err := load.Avg(); err == nil {
		rm.resources.LoadAverage = []float64{loadStat.Load1, loadStat.Load5, loadStat.Load15}
	} else {
		if rm.log != nil {
			rm.log.Errorf("获取系统负载失败: %v", err)
		}
		rm.resources.LoadAverage = []float64{0, 0, 0}
	}
}

// detectGPUs 检测GPU
func (rm *ResourceMonitor) detectGPUs() {
	rm.gpuInfo = rm.gpuInfo[:0] // 清空现有数据

	// 检测NVIDIA GPU
	if nvidiaGPUs := rm.detectNvidiaGPUs(); len(nvidiaGPUs) > 0 {
		rm.gpuInfo = append(rm.gpuInfo, nvidiaGPUs...)
	}

	// 检测AMD GPU
	if amdGPUs := rm.detectAMDGPUs(); len(amdGPUs) > 0 {
		rm.gpuInfo = append(rm.gpuInfo, amdGPUs...)
	}

	// 检测Intel GPU
	if intelGPUs := rm.detectIntelGPUs(); len(intelGPUs) > 0 {
		rm.gpuInfo = append(rm.gpuInfo, intelGPUs...)
	}

	if len(rm.gpuInfo) > 0 && rm.log != nil {
		rm.log.Infof("检测到 %d 个GPU", len(rm.gpuInfo))
	}
}

// detectNvidiaGPUs 检测NVIDIA GPU
func (rm *ResourceMonitor) detectNvidiaGPUs() []GPUInfo {
	var gpus []GPUInfo

	// 尝试使用 nvidia-smi
	cmd := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,memory.used,temperature.gpu,utilization.gpu,power.draw,driver_version", "--format=csv,noheader,nounits")
	output, err := cmd.Output()
	if err != nil {
		if rm.log != nil {
			rm.log.Debugf("未检测到NVIDIA GPU或nvidia-smi不可用: %v", err)
		}
		return gpus
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		fields := strings.Split(line, ",")
		if len(fields) < 8 {
			continue
		}

		// 解析字段
		index, _ := strconv.Atoi(strings.TrimSpace(fields[0]))
		name := strings.TrimSpace(fields[1])
		totalMemory, _ := strconv.ParseInt(strings.TrimSpace(fields[2]), 10, 64)
		usedMemory, _ := strconv.ParseInt(strings.TrimSpace(fields[3]), 10, 64)
		temperature, _ := strconv.ParseFloat(strings.TrimSpace(fields[4]), 64)
		utilization, _ := strconv.ParseFloat(strings.TrimSpace(fields[5]), 64)
		powerUsage, _ := strconv.ParseFloat(strings.TrimSpace(fields[6]), 64)
		driverVersion := strings.TrimSpace(fields[7])

		gpus = append(gpus, GPUInfo{
			Index:         index,
			Name:          name,
			Vendor:        "NVIDIA",
			TotalMemory:   totalMemory * 1024 * 1024, // MB to bytes
			UsedMemory:    usedMemory * 1024 * 1024,  // MB to bytes
			Temperature:   temperature,
			Utilization:   utilization,
			PowerUsage:    powerUsage,
			DriverVersion: driverVersion,
		})

		if rm.log != nil {
			rm.log.Debugf("检测到NVIDIA GPU[%d]: %s", index, name)
		}
	}

	return gpus
}

// detectAMDGPUs 检测AMD GPU
func (rm *ResourceMonitor) detectAMDGPUs() []GPUInfo {
	var gpus []GPUInfo

	// 尝试使用 rocm-smi
	cmd := exec.Command("rocm-smi", "--showproductname")
	output, err := cmd.Output()
	if err != nil {
		if rm.log != nil {
			rm.log.Debugf("未检测到AMD GPU或rocm-smi不可用: %v", err)
		}
		return gpus
	}

	// 解析AMD GPU信息（简化实现）
	lines := strings.Split(string(output), "\n")
	index := 0
	for _, line := range lines {
		if strings.Contains(line, "Card series") || strings.Contains(line, "GPU") {
			name := strings.TrimSpace(line)
			if name != "" {
				gpus = append(gpus, GPUInfo{
					Index:       index,
					Name:        name,
					Vendor:      "AMD",
					TotalMemory: 0, // ROCM-SMI 不提供这些信息
					UsedMemory:  0,
					Temperature: 0,
					Utilization: 0,
				})
				if rm.log != nil {
					rm.log.Debugf("检测到AMD GPU[%d]: %s", index, name)
				}
				index++
			}
		}
	}

	return gpus
}

// detectIntelGPUs 检测Intel GPU
func (rm *ResourceMonitor) detectIntelGPUs() []GPUInfo {
	var gpus []GPUInfo

	// Intel GPU 检测通常通过系统信息
	// 这里提供一个简化实现
	if hostStat, err := host.Info(); err == nil {
		// 检查是否包含Intel集成显卡
		if strings.Contains(strings.ToLower(hostStat.KernelVersion), "i915") ||
			strings.Contains(strings.ToLower(hostStat.KernelVersion), "intel") {
			gpus = append(gpus, GPUInfo{
				Index:       0,
				Name:        "Intel Integrated Graphics",
				Vendor:      "Intel",
				TotalMemory: 0, // 集成显卡使用系统内存
				UsedMemory:  0,
				Temperature: 0,
				Utilization: 0,
			})
			if rm.log != nil {
				rm.log.Debugf("检测到Intel集成显卡")
			}
		}
	}

	return gpus
}

// updateNvidiaGPU 更新NVIDIA GPU信息
func (rm *ResourceMonitor) updateNvidiaGPU(gpu *GPUInfo) {
	cmd := exec.Command("nvidia-smi", "--query-gpu=memory.used,temperature.gpu,utilization.gpu,power.draw", "--format=csv,noheader,nounits", fmt.Sprintf("--id=%d", gpu.Index))
	output, err := cmd.Output()
	if err != nil {
		return
	}

	fields := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(fields) < 4 {
		return
	}

	if usedMemory, err := strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 64); err == nil {
		gpu.UsedMemory = usedMemory * 1024 * 1024 // MB to bytes
	}
	if temperature, err := strconv.ParseFloat(strings.TrimSpace(fields[1]), 64); err == nil {
		gpu.Temperature = temperature
	}
	if utilization, err := strconv.ParseFloat(strings.TrimSpace(fields[2]), 64); err == nil {
		gpu.Utilization = utilization
	}
	if powerUsage, err := strconv.ParseFloat(strings.TrimSpace(fields[3]), 64); err == nil {
		gpu.PowerUsage = powerUsage
	}
}

// updateAMDGPU 更新AMD GPU信息
func (rm *ResourceMonitor) updateAMDGPU(gpu *GPUInfo) {
	// AMD GPU 监控的具体实现取决于ROCm版本
	// 这里提供一个占位符实现
}

// updateIntelGPU 更新Intel GPU信息
func (rm *ResourceMonitor) updateIntelGPU(gpu *GPUInfo) {
	// Intel GPU 监控的具体实现
	// 这里提供一个占位符实现
}

// detectLlamacpp 检测llama.cpp
func (rm *ResourceMonitor) detectLlamacpp() {
	rm.llamacppInfo = nil

	for _, path := range rm.llamacppPaths {
		if info := rm.testLlamacppPath(path); info != nil {
			rm.llamacppInfo = info
			if rm.log != nil {
				rm.log.Infof("检测到llama.cpp: %s (版本: %s)", path, info.Version)
			}
			return
		}
	}

	if rm.log != nil {
		rm.log.Debugf("未检测到有效的llama.cpp安装")
	}
}

// testLlamacppPath 测试llama.cpp路径
func (rm *ResourceMonitor) testLlamacppPath(path string) *LlamacppInfo {
	// 检查文件是否存在且可执行
	cmd := exec.Command("test", "-x", path)
	if err := cmd.Run(); err != nil {
		return nil
	}

	info := &LlamacppInfo{
		Path:     path,
		Binaries: make(map[string]string),
	}

	// 获取版本信息
	cmd = exec.Command(path, "--version")
	if output, err := cmd.Output(); err == nil {
		versionStr := strings.TrimSpace(string(output))
		// 尝试解析版本字符串
		if strings.Contains(versionStr, "version") {
			parts := strings.Fields(versionStr)
			for i, part := range parts {
				if strings.ToLower(part) == "version" && i+1 < len(parts) {
					info.Version = parts[i+1]
					break
				}
			}
		} else {
			info.Version = versionStr
		}
	}

	// 检测GPU后端支持
	cmd = exec.Command(path, "--help")
	if output, err := cmd.Output(); err == nil {
		helpStr := strings.ToLower(string(output))
		if strings.Contains(helpStr, "cuda") {
			info.GPUBackend = "cuda"
			info.SupportsGPU = true
		} else if strings.Contains(helpStr, "metal") {
			info.GPUBackend = "metal"
			info.SupportsGPU = true
		} else if strings.Contains(helpStr, "opencl") {
			info.GPUBackend = "opencl"
			info.SupportsGPU = true
		}

		// 检测支持的格式
		if strings.Contains(helpStr, "gguf") {
			info.SupportedFormats = append(info.SupportedFormats, "gguf")
		}
	}

	info.Available = true
	info.Binaries["main"] = path

	return info
}
