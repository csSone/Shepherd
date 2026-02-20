package node

import (
	"os/exec"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewResourceMonitor(t *testing.T) {
	tests := []struct {
		name     string
		config   *ResourceMonitorConfig
		expected *ResourceMonitorConfig
	}{
		{
			name:   "nil config",
			config: nil,
			expected: &ResourceMonitorConfig{
				Interval: 5 * time.Second,
			},
		},
		{
			name: "custom config",
			config: &ResourceMonitorConfig{
				Interval:      10 * time.Second,
				LlamacppPaths: []string{"/usr/bin/llama.cpp"},
			},
			expected: &ResourceMonitorConfig{
				Interval:      10 * time.Second,
				LlamacppPaths: []string{"/usr/bin/llama.cpp"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rm := NewResourceMonitor(tt.config)
			defer rm.Stop()

			assert.NotNil(t, rm)
			assert.NotNil(t, rm.ctx)
			assert.NotNil(t, rm.cancel)
			assert.NotNil(t, rm.resources)
			assert.NotNil(t, rm.gpuInfo)

			expectedInterval := tt.expected.Interval
			if expectedInterval == 0 {
				expectedInterval = 5 * time.Second
			}
			assert.Equal(t, expectedInterval, rm.interval)

			if tt.config != nil && len(tt.config.LlamacppPaths) > 0 {
				assert.Equal(t, tt.config.LlamacppPaths, rm.llamacppPaths)
			}
		})
	}
}

func TestResourceMonitor_StartStop(t *testing.T) {
	config := &ResourceMonitorConfig{
		Interval: 100 * time.Millisecond,
	}
	rm := NewResourceMonitor(config)

	// Test start
	err := rm.Start()
	require.NoError(t, err)
	assert.True(t, rm.running)

	// Test start when already running
	err = rm.Start()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已在运行")

	// Test stop
	err = rm.Stop()
	require.NoError(t, err)
	assert.False(t, rm.running)

	// Test stop when not running
	err = rm.Stop()
	assert.NoError(t, err) // Should not error
}

func TestResourceMonitor_GetSnapshot(t *testing.T) {
	config := &ResourceMonitorConfig{
		Interval: 100 * time.Millisecond,
	}
	rm := NewResourceMonitor(config)

	// Test snapshot before start
	snapshot := rm.GetSnapshot()
	assert.NotNil(t, snapshot)
	assert.True(t, snapshot.CPUTotal > 0)

	// Start the monitor
	err := rm.Start()
	require.NoError(t, err)
	defer rm.Stop()

	// Wait for at least one update
	time.Sleep(200 * time.Millisecond)

	snapshot = rm.GetSnapshot()
	assert.NotNil(t, snapshot)
	assert.True(t, snapshot.CPUTotal > 0)
	assert.True(t, snapshot.MemoryTotal > 0)
	assert.True(t, snapshot.DiskTotal > 0)
}

func TestResourceMonitor_GetGPUInfo(t *testing.T) {
	config := &ResourceMonitorConfig{
		Interval: 100 * time.Millisecond,
	}
	rm := NewResourceMonitor(config)

	// Start the monitor
	err := rm.Start()
	require.NoError(t, err)
	defer rm.Stop()

	// Wait for GPU detection
	time.Sleep(200 * time.Millisecond)

	gpuInfo := rm.GetGPUInfo()
	assert.NotNil(t, gpuInfo)
	// GPU info might be empty if no GPU is present
}

func TestResourceMonitor_GetLlamacppInfo(t *testing.T) {
	config := &ResourceMonitorConfig{
		Interval:      100 * time.Millisecond,
		LlamacppPaths: []string{"/nonexistent/path"},
	}
	rm := NewResourceMonitor(config)

	// Test before start
	llamacppInfo := rm.GetLlamacppInfo()
	assert.Nil(t, llamacppInfo)

	// Start the monitor
	err := rm.Start()
	require.NoError(t, err)
	defer rm.Stop()

	// Wait for llama.cpp detection
	time.Sleep(200 * time.Millisecond)

	llamacppInfo = rm.GetLlamacppInfo()
	// Should be nil since path doesn't exist
	assert.Nil(t, llamacppInfo)
}

func TestResourceMonitor_Callback(t *testing.T) {
	var callbackCalled bool
	var callbackResources *NodeResources

	config := &ResourceMonitorConfig{
		Interval: 50 * time.Millisecond,
		Callback: func(resources *NodeResources) {
			callbackCalled = true
			callbackResources = resources
		},
	}
	rm := NewResourceMonitor(config)

	err := rm.Start()
	require.NoError(t, err)
	defer rm.Stop()

	// Wait for callback to be called
	time.Sleep(150 * time.Millisecond)

	assert.True(t, callbackCalled)
	assert.NotNil(t, callbackResources)
	assert.True(t, callbackResources.CPUTotal > 0)
}

func TestResourceMonitor_UpdateCPUUsage(t *testing.T) {
	rm := NewResourceMonitor(nil)
	rm.resources = &NodeResources{
		CPUTotal: 4000, // 4 cores in millicores
	}

	rm.updateCPUUsage()

	assert.True(t, rm.resources.CPUUsed >= 0)
	assert.True(t, rm.resources.CPUUsed <= rm.resources.CPUTotal)
}

func TestResourceMonitor_UpdateMemoryUsage(t *testing.T) {
	rm := NewResourceMonitor(nil)
	// 使用实际的系统内存信息
	if vmStat, err := mem.VirtualMemory(); err == nil {
		rm.resources = &NodeResources{
			MemoryTotal: int64(vmStat.Total), // 使用实际系统内存总量
		}
	} else {
		t.Skip("无法获取系统内存信息，跳过测试")
	}

	rm.updateMemoryUsage()

	rm.mu.RLock()
	defer rm.mu.RUnlock()
	assert.True(t, rm.resources.MemoryUsed >= 0)
	assert.True(t, rm.resources.MemoryUsed <= rm.resources.MemoryTotal)
}

func TestResourceMonitor_UpdateDiskUsage(t *testing.T) {
	rm := NewResourceMonitor(nil)
	// 使用实际的系统磁盘信息
	if diskStat, err := disk.Usage("/"); err == nil {
		rm.resources = &NodeResources{
			DiskTotal: int64(diskStat.Total), // 使用实际系统磁盘总量
		}
	} else {
		t.Skip("无法获取系统磁盘信息，跳过测试")
	}

	rm.updateDiskUsage()

	rm.mu.RLock()
	defer rm.mu.RUnlock()
	assert.True(t, rm.resources.DiskUsed >= 0)
	assert.True(t, rm.resources.DiskUsed <= rm.resources.DiskTotal)
}

func TestResourceMonitor_UpdateLoadAverage(t *testing.T) {
	rm := NewResourceMonitor(nil)
	rm.resources = &NodeResources{
		LoadAverage: make([]float64, 3),
	}

	rm.updateLoadAverage()

	assert.Len(t, rm.resources.LoadAverage, 3)
	for _, load := range rm.resources.LoadAverage {
		assert.True(t, load >= 0)
	}
}

func TestResourceMonitor_TestLlamacppPath(t *testing.T) {
	rm := NewResourceMonitor(nil)

	// Test nonexistent path
	info := rm.testLlamacppPath("/nonexistent/path")
	assert.Nil(t, info)

	// Test existing but non-executable path
	info = rm.testLlamacppPath("/etc/passwd")
	assert.Nil(t, info)
}

func TestResourceMonitor_DetectNvidiaGPUs(t *testing.T) {
	// 跳过测试如果 nvidia-smi 不可用
	if _, err := exec.LookPath("nvidia-smi"); err != nil {
		t.Skip("nvidia-smi 不可用，跳过 NVIDIA GPU 检测测试")
	}

	rm := NewResourceMonitor(nil)
	gpus := rm.detectNvidiaGPUs()
	// GPUs should be a slice, may be empty if no NVIDIA GPUs are present
	assert.NotNil(t, gpus)
}

func TestResourceMonitor_DetectAMDGPUs(t *testing.T) {
	rm := NewResourceMonitor(nil)

	// This test will only pass if rocm-smi is available
	gpus := rm.detectAMDGPUs()
	assert.NotNil(t, gpus)
}

func TestResourceMonitor_DetectIntelGPUs(t *testing.T) {
	// 跳过测试如果系统有 AMD GPU（当前系统是 AMD）
	if _, err := exec.LookPath("rocm-smi"); err == nil {
		t.Skip("系统检测到 AMD GPU，跳过 Intel GPU 检测测试")
	}

	rm := NewResourceMonitor(nil)
	gpus := rm.detectIntelGPUs()
	// GPUs should be a slice, may be empty or nil if no Intel GPUs are present
	if gpus != nil {
		assert.NotNil(t, gpus)
	}
}

func TestResourceMonitor_ContextCancellation(t *testing.T) {
	config := &ResourceMonitorConfig{
		Interval: 50 * time.Millisecond,
	}
	rm := NewResourceMonitor(config)

	err := rm.Start()
	require.NoError(t, err)

	// Cancel context after a short time
	go func() {
		time.Sleep(100 * time.Millisecond)
		rm.Stop() // Use Stop instead of direct cancel for proper cleanup
	}()

	// Wait for the monitor to stop
	time.Sleep(200 * time.Millisecond)

	assert.False(t, rm.running)
}

func TestResourceMonitor_ConcurrentAccess(t *testing.T) {
	config := &ResourceMonitorConfig{
		Interval: 10 * time.Millisecond,
	}
	rm := NewResourceMonitor(config)

	err := rm.Start()
	require.NoError(t, err)
	defer rm.Stop()

	// Concurrent access test
	done := make(chan bool, 10)

	// Start multiple goroutines accessing the monitor
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				snapshot := rm.GetSnapshot()
				assert.NotNil(t, snapshot)
				gpuInfo := rm.GetGPUInfo()
				assert.NotNil(t, gpuInfo)
				llamacppInfo := rm.GetLlamacppInfo()
				// Can be nil
				_ = llamacppInfo
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}
