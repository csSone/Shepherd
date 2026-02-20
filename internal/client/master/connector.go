// Package master provides master connection functionality for client nodes.
package master

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shepherd-project/shepherd/Shepherd/internal/client"
	"github.com/shepherd-project/shepherd/Shepherd/internal/cluster"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
	"github.com/shepherd-project/shepherd/Shepherd/internal/logger"
)

// Connector manages the connection to the master node
type Connector struct {
	config     *config.ClientConfig
	clientInfo *client.ClientInfo
	httpClient *http.Client
	connected  bool
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	log        *logger.Logger
}

// NewConnector creates a new master connector
func NewConnector(cfg *config.ClientConfig, log *logger.Logger) (*Connector, error) {
	// Generate client info
	clientInfo, err := generateClientInfo(cfg)
	if err != nil {
		return nil, fmt.Errorf("生成客户端信息失败: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Connector{
		config:     cfg,
		clientInfo: clientInfo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		connected: false,
		ctx:       ctx,
		cancel:    cancel,
		log:       log,
	}, nil
}

// Start starts the connector and connects to the master
func (c *Connector) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config.MasterAddress == "" {
		return fmt.Errorf("master 地址未配置")
	}

	// Register with master
	if err := c.register(); err != nil {
		return fmt.Errorf("注册到 master 失败: %w", err)
	}

	c.connected = true
	c.log.Info(fmt.Sprintf("已连接到 Master: %s", c.config.MasterAddress))

	// Start heartbeat
	c.wg.Add(1)
	go c.heartbeatLoop()

	return nil
}

// Stop stops the connector
func (c *Connector) Stop() {
	c.cancel()
	c.wg.Wait()

	c.mu.Lock()
	if c.connected {
		c.unregister()
		c.connected = false
	}
	c.mu.Unlock()
}

// IsConnected returns whether the client is connected to the master
func (c *Connector) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// GetClientInfo returns the client information
func (c *Connector) GetClientInfo() *client.ClientInfo {
	return c.clientInfo
}

// register registers this client with the master
func (c *Connector) register() error {
	url := fmt.Sprintf("%s/api/master/clients/register", c.config.MasterAddress)

	body, err := json.Marshal(c.clientInfo)
	if err != nil {
		return fmt.Errorf("序列化客户端信息失败: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("注册请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("注册失败: HTTP %d - %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// unregister unregisters this client from the master
func (c *Connector) unregister() error {
	url := fmt.Sprintf("%s/api/master/clients/%s", c.config.MasterAddress, c.clientInfo.ID)

	req, err := http.NewRequestWithContext(context.Background(), "DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// heartbeatLoop sends periodic heartbeats to the master
func (c *Connector) heartbeatLoop() {
	defer c.wg.Done()

	interval := time.Duration(c.config.Heartbeat.Interval) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Send initial heartbeat
	c.sendHeartbeat()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.sendHeartbeat()
		}
	}
}

// sendHeartbeat sends a heartbeat to the master
func (c *Connector) sendHeartbeat() error {
	heartbeat := &cluster.Heartbeat{
		ClientID:  c.clientInfo.ID,
		Timestamp: time.Now(),
		Status:    cluster.ClientStatusOnline,
		Resources: c.getResourceUsage(),
	}

	url := fmt.Sprintf("%s/api/master/heartbeat", c.config.MasterAddress)
	body, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("序列化心跳失败: %w", err)
	}

	resp, err := c.httpClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		c.mu.Lock()
		c.connected = false
		c.mu.Unlock()
		return fmt.Errorf("发送心跳失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		c.mu.Lock()
		if !c.connected {
			c.connected = true
			c.log.Info("重新连接到 Master")
		}
		c.mu.Unlock()
	}

	return nil
}

// getResourceUsage gets current resource usage
func (c *Connector) getResourceUsage() *cluster.ResourceUsage {
	vmStat, _ := mem.VirtualMemory()
	hostInfo, _ := host.Info()

	// 获取磁盘使用率
	diskPercent := c.getDiskUsage()

	// 获取 GPU 使用率
	gpuPercent, gpuMemUsed, gpuMemTotal := c.getGPUUsage()

	return &cluster.ResourceUsage{
		CPUPercent:     float64(runtime.NumCPU()), // Simplified
		MemoryUsed:     int64(vmStat.Used),
		MemoryTotal:    int64(vmStat.Total),
		GPUPercent:     float64(gpuPercent),
		GPUMemoryUsed:  gpuMemUsed,
		GPUMemoryTotal: gpuMemTotal,
		DiskPercent:    diskPercent,
		Uptime:         int64(hostInfo.Uptime),
	}
}

// getDiskUsage 获取磁盘使用率
func (c *Connector) getDiskUsage() float64 {
	// 获取根分区使用率
	if diskStat, err := disk.Usage("/"); err == nil {
		return diskStat.UsedPercent
	}
	return 0
}

// getGPUUsage 获取 GPU 使用率和内存使用情况
func (c *Connector) getGPUUsage() (percent, memUsed, memTotal int64) {
	// 尝试检测 NVIDIA GPU
	if nvidiaSMI := c.detectNvidiaGPU(); nvidiaSMI {
		return c.getNvidiaGPUUsage()
	}

	// 尝试检测 AMD GPU (ROCm)
	if radeonSMI := c.detectAmdGPU(); radeonSMI {
		return c.getAmdGPUUsage()
	}

	return 0, 0, 0
}

// detectNvidiaGPU 检测是否存在 NVIDIA GPU
func (c *Connector) detectNvidiaGPU() bool {
	_, err := exec.LookPath("nvidia-smi")
	return err == nil
}

// detectAmdGPU 检测是否存在 AMD GPU
func (c *Connector) detectAmdGPU() bool {
	// 检查 ROCm SMI 工具
	_, err := exec.LookPath("rocm-smi")
	if err == nil {
		return true
	}

	// 检查 sysfs 中的 AMD GPU
	_, err = os.Stat("/sys/class/drm/card0/device/vendor")
	if err == nil {
		// 读取 vendor ID
		data, err := os.ReadFile("/sys/class/drm/card0/device/vendor")
		if err == nil && strings.Contains(string(data), "0x1002") {
			return true // AMD vendor ID
		}
	}

	return false
}

// getNvidiaGPUUsage 获取 NVIDIA GPU 使用率
func (c *Connector) getNvidiaGPUUsage() (percent, memUsed, memTotal int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "nvidia-smi",
		"--query-gpu=utilization.gpu,memory.used,memory.total",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0
	}

	// 解析输出: "85, 1024, 8192"
	parts := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(parts) >= 3 {
		if p, err := strconv.ParseInt(strings.TrimSpace(parts[0]), 10, 64); err == nil {
			percent = p
		}
		if u, err := strconv.ParseInt(strings.TrimSpace(parts[1]), 10, 64); err == nil {
			memUsed = u
		}
		if t, err := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64); err == nil {
			memTotal = t
		}
	}

	return percent, memUsed, memTotal
}

// getAmdGPUUsage 获取 AMD GPU 使用率 (ROCm)
func (c *Connector) getAmdGPUUsage() (percent, memUsed, memTotal int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "rocm-smi",
		"--showuse",
		"--showmem",
		"--csv")

	output, err := cmd.Output()
	if err != nil {
		// 回退到基本的 GPU 检测
		return c.getAmdGPUBasic()
	}

	// 解析 ROCm SMI 输出
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "GPU use") {
			// 提取使用率百分比
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "%" && i > 0 {
					if p, err := strconv.ParseInt(strings.TrimSuffix(parts[i-1], "%"), 10, 64); err == nil {
						percent = p
					}
					break
				}
			}
		}
	}

	// 获取 VRAM 使用情况（简化版）
	return percent, 0, 0
}

// getAmdGPUBasic 获取基本的 AMD GPU 信息（回退方法）
func (c *Connector) getAmdGPUBasic() (percent, memUsed, memTotal int64) {
	// 尝试从 sysfs 读取基本信息
	// 这只返回有限的信息，但至少能检测到 GPU
	return 0, 0, 0
}

// generateClientInfo generates client information
func generateClientInfo(cfg *config.ClientConfig) (*client.ClientInfo, error) {
	// Get hostname
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	// Generate client ID if not provided
	clientID := cfg.ClientInfo.ID
	if clientID == "" {
		// Use MAC address or hostname as ID
		interfaces, _ := net.Interfaces()
		for _, iface := range interfaces {
			if iface.HardwareAddr != nil && len(iface.HardwareAddr) > 0 {
				clientID = iface.HardwareAddr.String()
				break
			}
		}
		if clientID == "" {
			clientID = hostname
		}
	}

	// Get client name
	clientName := cfg.ClientInfo.Name
	if clientName == "" {
		clientName = hostname
	}

	// Get local IP address
	localIP := "127.0.0.1"
	addrs, _ := net.InterfaceAddrs()
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIP = ipnet.IP.String()
				break
			}
		}
	}

	// Get system info
	hostInfo, _ := host.Info()
	vmStat, _ := mem.VirtualMemory()

	// Build capabilities
	capabilities := &cluster.Capabilities{
		CPUCount:       runtime.NumCPU(),
		Memory:         int64(vmStat.Total),
		SupportsLlama:  true,
		SupportsPython: cfg.CondaEnv.Enabled,
	}

	if cfg.CondaEnv.Enabled {
		envs := make([]string, 0, len(cfg.CondaEnv.Environments))
		for name := range cfg.CondaEnv.Environments {
			envs = append(envs, name)
		}
		capabilities.CondaEnvs = envs
	}

	// Build metadata
	metadata := make(map[string]string)
	for k, v := range cfg.ClientInfo.Metadata {
		metadata[k] = v
	}
	metadata["os"] = hostInfo.OS
	metadata["platform"] = hostInfo.Platform
	metadata["platformFamily"] = hostInfo.PlatformFamily
	metadata["platformVersion"] = hostInfo.PlatformVersion
	metadata["kernelVersion"] = hostInfo.KernelVersion
	metadata["kernelArch"] = hostInfo.KernelArch

	return &client.ClientInfo{
		ID:      clientID,
		Name:    clientName,
		Address: localIP,
		Port:    9191, // Default client port
		Tags:    cfg.ClientInfo.Tags,
		Capabilities: capabilities,
		Version: "0.1.0-alpha",
		Metadata: metadata,
	}, nil
}
