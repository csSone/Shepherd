// Package scheduler provides distributed task scheduling functionality.
package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/shepherd-project/shepherd/Shepherd/internal/cluster"
	"github.com/shepherd-project/shepherd/Shepherd/internal/config"
)

// Scheduler manages task distribution across clients
type Scheduler struct {
	strategy  SchedulingStrategy
	clientMgr ClientManager
	tasks     map[string]*cluster.Task
	queue     chan *cluster.Task
	mu        sync.RWMutex
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// ClientManager interface for managing clients
type ClientManager interface {
	GetOnlineClients() []*cluster.Client
	GetClient(clientID string) (*cluster.Client, bool)
	SendCommand(clientID string, command *cluster.Command) (map[string]interface{}, error)
}

// SchedulingStrategy defines the scheduling strategy
type SchedulingStrategy string

const (
	RoundRobinStrategy  SchedulingStrategy = "round_robin"
	LeastLoadedStrategy SchedulingStrategy = "least_loaded"
	ResourceAwareStrategy SchedulingStrategy = "resource_aware"
)

// NewScheduler creates a new task scheduler
func NewScheduler(cfg *config.SchedulerConfig, clientMgr ClientManager) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())

	var strategy SchedulingStrategy
	switch cfg.Strategy {
	case "round_robin":
		strategy = RoundRobinStrategy
	case "least_loaded":
		strategy = LeastLoadedStrategy
	case "resource_aware":
		strategy = ResourceAwareStrategy
	default:
		strategy = RoundRobinStrategy
	}

	return &Scheduler{
		strategy:  strategy,
		clientMgr: clientMgr,
		tasks:     make(map[string]*cluster.Task),
		queue:     make(chan *cluster.Task, cfg.MaxQueueSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	s.wg.Add(1)
	go s.dispatchLoop()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.cancel()
	s.wg.Wait()
}

// SubmitTask submits a new task for scheduling
func (s *Scheduler) SubmitTask(taskType cluster.TaskType, payload map[string]interface{}) (*cluster.Task, error) {
	task := &cluster.Task{
		ID:        uuid.New().String(),
		Type:      taskType,
		Payload:   payload,
		Status:    cluster.TaskStatusPending,
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	select {
	case s.queue <- task:
		return task, nil
	case <-time.After(5 * time.Second):
		return nil, fmt.Errorf("任务队列已满")
	}
}

// GetTask retrieves a task by ID
func (s *Scheduler) GetTask(taskID string) (*cluster.Task, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[taskID]
	return task, exists
}

// ListTasks returns all tasks
func (s *Scheduler) ListTasks() []*cluster.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*cluster.Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// CancelTask cancels a task
func (s *Scheduler) CancelTask(taskID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[taskID]
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}

	if task.Status == cluster.TaskStatusRunning {
		return fmt.Errorf("无法取消正在运行的任务: %s", taskID)
	}

	task.Status = cluster.TaskStatusCancelled

	return nil
}

// dispatchLoop dispatches tasks from the queue
func (s *Scheduler) dispatchLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case task := <-s.queue:
			s.dispatchTask(task)
		}
	}
}

// dispatchTask dispatches a task to a client
func (s *Scheduler) dispatchTask(task *cluster.Task) {
	// Get online clients
	clients := s.clientMgr.GetOnlineClients()
	if len(clients) == 0 {
		s.updateTaskStatus(task, cluster.TaskStatusPending, "", "没有可用的客户端")
		return
	}

	// Select client based on strategy
	var selectedClient *cluster.Client
	var err error

	switch s.strategy {
	case RoundRobinStrategy:
		selectedClient, err = s.selectRoundRobin(clients)
	case LeastLoadedStrategy:
		selectedClient, err = s.selectLeastLoaded(clients)
	case ResourceAwareStrategy:
		selectedClient, err = s.selectResourceAware(clients)
	}

	if err != nil {
		s.updateTaskStatus(task, cluster.TaskStatusPending, "", err.Error())
		return
	}

	// Send task to client
	command := &cluster.Command{
		ID:      uuid.New().String(),
		Type:    string(task.Type),
		Payload: task.Payload,
	}

	s.updateTaskStatus(task, cluster.TaskStatusRunning, selectedClient.ID, "")

	result, err := s.clientMgr.SendCommand(selectedClient.ID, command)
	if err != nil {
		s.updateTaskStatus(task, cluster.TaskStatusFailed, "", err.Error())
		return
	}

	s.updateTaskStatus(task, cluster.TaskStatusCompleted, "", "")
	task.Result = result

	now := time.Now()
	task.CompletedAt = &now
}

// selectRoundRobin selects a client using round-robin
func (s *Scheduler) selectRoundRobin(clients []*cluster.Client) (*cluster.Client, error) {
	// Simple round-robin based on task count
	minTasks := int64(1<<63 - 1)
	var selected *cluster.Client

	for _, client := range clients {
		taskCount := int64(0)
		for _, task := range s.tasks {
			if task.AssignedTo == client.ID && task.Status == cluster.TaskStatusRunning {
				taskCount++
			}
		}

		if taskCount < minTasks {
			minTasks = taskCount
			selected = client
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("没有可用的客户端")
	}

	return selected, nil
}

// selectLeastLoaded selects the least loaded client
func (s *Scheduler) selectLeastLoaded(clients []*cluster.Client) (*cluster.Client, error) {
	// Select client with fewest running tasks
	return s.selectRoundRobin(clients)
}

// selectResourceAware selects a client based on resource availability
func (s *Scheduler) selectResourceAware(clients []*cluster.Client) (*cluster.Client, error) {
	// Select client with most available resources
	// This is a simplified implementation
	maxMemory := int64(0)
	var selected *cluster.Client

	for _, client := range clients {
		if client.Capabilities != nil {
			availableMemory := client.Capabilities.Memory
			if availableMemory > maxMemory {
				maxMemory = availableMemory
				selected = client
			}
		}
	}

	if selected == nil {
		return nil, fmt.Errorf("没有可用的客户端")
	}

	return selected, nil
}

// updateTaskStatus updates the status of a task
func (s *Scheduler) updateTaskStatus(task *cluster.Task, status cluster.TaskStatus, assignedTo string, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.Status = status
	if assignedTo != "" {
		task.AssignedTo = assignedTo
	}
	if errorMsg != "" {
		task.Error = errorMsg
	}

	if status == cluster.TaskStatusRunning {
		now := time.Now()
		task.StartedAt = &now
	}
}
