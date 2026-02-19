// Package resource provides resource monitoring for client nodes.
package resource

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shepherd-project/shepherd/Shepherd/internal/cluster"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
)

// Monitor monitors system resources
type Monitor struct {
	mu             sync.RWMutex
	currentUsage   *cluster.ResourceUsage
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	updateInterval time.Duration
	log            *logger.Logger
}

// NewMonitor creates a new resource monitor
func NewMonitor(updateInterval time.Duration, log *logger.Logger) *Monitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Monitor{
		currentUsage:   &cluster.ResourceUsage{},
		ctx:           ctx,
		cancel:        cancel,
		updateInterval: updateInterval,
		log:           log,
	}
}

// Start starts the resource monitor
func (m *Monitor) Start() {
	m.wg.Add(1)
	go m.updateLoop()
}

// Stop stops the resource monitor
func (m *Monitor) Stop() {
	m.cancel()
	m.wg.Wait()
}

// GetCurrentUsage returns the current resource usage
func (m *Monitor) GetCurrentUsage() *cluster.ResourceUsage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid race conditions
	usage := *m.currentUsage
	return &usage
}

// updateLoop periodically updates resource usage
func (m *Monitor) updateLoop() {
	defer m.wg.Done()

	// Initial update
	m.updateUsage()

	ticker := time.NewTicker(m.updateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.updateUsage()
		}
	}
}

// updateUsage updates the current resource usage
func (m *Monitor) updateUsage() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get CPU usage
	cpuPercent, err := cpu.Percent(0, false)
	if err != nil {
		m.log.Error(fmt.Sprintf("获取CPU使用率失败: %v", err), nil)
		cpuPercent = []float64{0}
	}

	// Get memory usage
	vmStat, err := mem.VirtualMemory()
	if err != nil {
		m.log.Error(fmt.Sprintf("获取内存使用率失败: %v", err), nil)
		vmStat = &mem.VirtualMemoryStat{
			Used:  0,
			Total: 1,
		}
	}

	m.currentUsage = &cluster.ResourceUsage{
		CPUPercent:     cpuPercent[0],
		MemoryUsed:     int64(vmStat.Used),
		MemoryTotal:    int64(vmStat.Total),
		GPUPercent:     0, // TODO: Implement GPU monitoring
		GPUMemoryUsed:  0,
		GPUMemoryTotal: 0,
		DiskPercent:    0, // TODO: Implement disk monitoring
		Uptime:         0, // TODO: Track uptime
	}
}
