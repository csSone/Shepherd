// Package monitor provides resource monitoring tests
package monitor

import (
	"testing"
	"time"

	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
	"github.com/shepherd-project/shepherd/Shepherd/internal/node"
)

// newTestLogger 创建一个用于测试的日志记录器
func newTestLogger(t *testing.T) *logger.Logger {
	log, err := logger.NewLogger(&config.LogConfig{
		Level:  "info",
		Output: "stdout",
	}, "test")
	if err != nil {
		t.Fatalf("创建日志记录器失败: %v", err)
	}
	return log
}

// TestMemoryResourceMonitor tests the MemoryResourceMonitor
func TestMemoryResourceMonitor(t *testing.T) {
	log := newTestLogger(t)

	// 创建监控器
	monitorConfig := &ResourceMonitorConfig{
		Interval:   100 * time.Millisecond,
		Logger:     log,
		MaxMetrics: 10,
	}

	monitor := NewMemoryResourceMonitor(monitorConfig)

	// 测试初始状态
	if monitor.IsRunning() {
		t.Error("监控器不应该在初始状态时运行")
	}

	// 启动监控器
	if err := monitor.Start(); err != nil {
		t.Fatalf("启动监控器失败: %v", err)
	}

	// 等待一段时间让监控器收集数据
	time.Sleep(300 * time.Millisecond)

	// 检查运行状态
	if !monitor.IsRunning() {
		t.Error("监控器应该正在运行")
	}

	// 获取资源快照
	resources := monitor.GetResources()
	if resources == nil {
		t.Error("资源快照不应为 nil")
	} else {
		// 验证基本资源数据
		if resources.CPUTotal <= 0 {
			t.Error("CPU 总量应该大于 0")
		}
		if resources.MemoryTotal <= 0 {
			t.Error("内存总量应该大于 0")
		}
		if resources.DiskTotal <= 0 {
			t.Error("磁盘总量应该大于 0")
		}
	}

	// 获取指标
	metrics := monitor.GetMetrics()
	if metrics == nil {
		t.Error("指标不应为 nil")
	} else {
		if metrics.Timestamp.IsZero() {
			t.Error("指标时间戳不应为零")
		}
	}

	// 获取历史指标
	history := monitor.GetMetricsHistory(5)
	if len(history) == 0 {
		t.Error("应该有历史指标记录")
	}

	// 停止监控器
	if err := monitor.Stop(); err != nil {
		t.Fatalf("停止监控器失败: %v", err)
	}

	// 检查停止后的状态
	if monitor.IsRunning() {
		t.Error("监控器不应该在停止后运行")
	}

	// 测试重复停止
	if err := monitor.Stop(); err != nil {
		t.Errorf("重复停止监控器不应该返回错误: %v", err)
	}
}

// TestMemoryResourceMonitorWatch tests the watch callback functionality
func TestMemoryResourceMonitorWatch(t *testing.T) {
	log := newTestLogger(t)

	monitorConfig := &ResourceMonitorConfig{
		Interval: 50 * time.Millisecond,
		Logger:   log,
	}

	monitor := NewMemoryResourceMonitor(monitorConfig)

	// 创建一个通道来接收回调
	callbackCount := 0
	callbackDone := make(chan bool, 1)

	// 添加回调
	monitor.Watch(func(resources *node.NodeResources) {
		callbackCount++
		if callbackCount >= 2 {
			callbackDone <- true
		}
	})

	// 启动监控器
	if err := monitor.Start(); err != nil {
		t.Fatalf("启动监控器失败: %v", err)
	}
	defer monitor.Stop()

	// 等待至少两次回调
	select {
	case <-callbackDone:
		// 成功收到回调
	case <-time.After(1 * time.Second):
		t.Error("未能在超时时间内收到足够的回调")
	}

	// 验证回调被调用
	if callbackCount < 2 {
		t.Errorf("期望至少 2 次回调，实际收到: %d", callbackCount)
	}
}

// TestMemoryResourceMonitorSetUpdateInterval tests setting update interval
func TestMemoryResourceMonitorSetUpdateInterval(t *testing.T) {
	log := newTestLogger(t)

	monitorConfig := &ResourceMonitorConfig{
		Interval: 100 * time.Millisecond,
		Logger:   log,
	}

	monitor := NewMemoryResourceMonitor(monitorConfig)

	// 设置新的更新间隔
	newInterval := 50 * time.Millisecond
	monitor.SetUpdateInterval(newInterval)

	// 启动监控器
	if err := monitor.Start(); err != nil {
		t.Fatalf("启动监控器失败: %v", err)
	}
	defer monitor.Stop()

	// 验证间隔已更新（通过检查统计信息）
	stats := monitor.GetStats()
	if stats.UpdateInterval != newInterval {
		t.Errorf("更新间隔未正确设置，期望: %v, 实际: %v", newInterval, stats.UpdateInterval)
	}
}

// TestMemoryResourceMonitorGetStats tests getting monitor statistics
func TestMemoryResourceMonitorGetStats(t *testing.T) {
	log := newTestLogger(t)

	monitorConfig := &ResourceMonitorConfig{
		Interval: 50 * time.Millisecond,
		Logger:   log,
	}

	monitor := NewMemoryResourceMonitor(monitorConfig)

	// 获取启动前的统计信息
	stats := monitor.GetStats()
	if stats.Running {
		t.Error("监控器不应在启动前运行")
	}

	// 启动监控器
	if err := monitor.Start(); err != nil {
		t.Fatalf("启动监控器失败: %v", err)
	}

	// 等待一些更新
	time.Sleep(200 * time.Millisecond)

	// 获取运行中的统计信息
	stats = monitor.GetStats()
	if !stats.Running {
		t.Error("监控器应该正在运行")
	}
	if stats.StartTime.IsZero() {
		t.Error("启动时间不应为零")
	}
	if stats.UpdateCount <= 0 {
		t.Error("更新计数应该大于 0")
	}

	monitor.Stop()
}

// TestMemoryResourceMonitorClearMetrics tests clearing metrics history
func TestMemoryResourceMonitorClearMetrics(t *testing.T) {
	log := newTestLogger(t)

	monitorConfig := &ResourceMonitorConfig{
		Interval:   50 * time.Millisecond,
		Logger:     log,
		MaxMetrics: 10,
	}

	monitor := NewMemoryResourceMonitor(monitorConfig)

	if err := monitor.Start(); err != nil {
		t.Fatalf("启动监控器失败: %v", err)
	}

	// 等待一些指标收集
	time.Sleep(200 * time.Millisecond)

	// 获取历史指标
	history := monitor.GetMetricsHistory(10)
	if len(history) == 0 {
		t.Error("应该有历史指标记录")
	}

	// 清空指标
	monitor.ClearMetrics()

	// 验证指标已清空
	history = monitor.GetMetricsHistory(10)
	if len(history) != 0 {
		t.Errorf("清空后不应该有历史指标，实际: %d", len(history))
	}

	monitor.Stop()
}

// TestMemoryResourceMonitorNilConfig tests creating monitor with nil config
func TestMemoryResourceMonitorNilConfig(t *testing.T) {
	monitor := NewMemoryResourceMonitor(nil)

	if monitor == nil {
		t.Error("使用 nil 配置创建监控器不应返回 nil")
	}

	// 验证默认值
	if monitor.interval != 5*time.Second {
		t.Errorf("默认间隔应该是 5 秒，实际: %v", monitor.interval)
	}
}

// TestMemoryResourceMonitorLlamacppDetection tests llama.cpp detection
func TestMemoryResourceMonitorLlamacppDetection(t *testing.T) {
	log := newTestLogger(t)

	monitorConfig := &ResourceMonitorConfig{
		Interval:      100 * time.Millisecond,
		Logger:        log,
		LlamacppPaths: []string{"/usr/local/bin/llama-cli", "/usr/bin/llama-cli"},
	}

	monitor := NewMemoryResourceMonitor(monitorConfig)

	if err := monitor.Start(); err != nil {
		t.Fatalf("启动监控器失败: %v", err)
	}
	defer monitor.Stop()

	// 等待初始化完成
	time.Sleep(200 * time.Millisecond)

	// 获取 llama.cpp 信息（可能为 nil）
	llamacppInfo := monitor.GetLlamacppInfo()

	// 如果系统上安装了 llama.cpp，应该能检测到
	// 如果没有安装，llamacppInfo 应该为 nil
	t.Logf("llama.cpp 检测结果: %v", llamacppInfo)
}
