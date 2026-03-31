package tools

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/providers"
)

type SubagentTask struct {
	ID            string `json:"id"`
	Task          string `json:"task"`
	Label         string `json:"label"`
	OriginChannel string `json:"origin_channel"`
	OriginChatID  string `json:"origin_chat_id"`
	Status        string `json:"status"`  // "running", "completed", "failed", "stalled", "cancelled"
	Result        string `json:"result"`
	Created       int64  `json:"created"`
	LastActivity  int64  `json:"last_activity"` // unix millis of last output
	ElapsedMs     int64  `json:"elapsed_ms"`
	cancel        context.CancelFunc
}

type SubagentManager struct {
	tasks          map[string]*SubagentTask
	mu             sync.RWMutex
	provider       providers.LLMProvider
	bus            *bus.MessageBus
	workspace      string
	nextID         int
	maxLifetime    time.Duration // max time per task (default 30 min)
	stallTimeout   time.Duration // stall detection threshold (default 60s)
}

func NewSubagentManager(provider providers.LLMProvider, workspace string, bus *bus.MessageBus) *SubagentManager {
	return &SubagentManager{
		tasks:        make(map[string]*SubagentTask),
		provider:     provider,
		bus:          bus,
		workspace:    workspace,
		nextID:       1,
		maxLifetime:  30 * time.Minute,
		stallTimeout: 60 * time.Second,
	}
}

func (sm *SubagentManager) Spawn(ctx context.Context, task, label, originChannel, originChatID string) (string, error) {
	// Check if provider is available (degraded mode check)
	if sm.provider == nil {
		return "", fmt.Errorf("cannot spawn subagent: LLM provider not configured. Please configure a provider to use subagent features")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	taskID := fmt.Sprintf("subagent-%d", sm.nextID)
	sm.nextID++

	now := time.Now().UnixMilli()
	taskCtx, taskCancel := context.WithTimeout(context.Background(), sm.maxLifetime)
	subagentTask := &SubagentTask{
		ID:            taskID,
		Task:          task,
		Label:         label,
		OriginChannel: originChannel,
		OriginChatID:  originChatID,
		Status:        "running",
		Created:       now,
		LastActivity:  now,
		cancel:        taskCancel,
	}
	sm.tasks[taskID] = subagentTask

	go func() {
		defer taskCancel()
		sm.runTask(taskCtx, subagentTask)
	}()

	// Start stall detector
	go sm.detectStall(subagentTask)

	if label != "" {
		return fmt.Sprintf("Spawned subagent '%s' for task: %s", label, task), nil
	}
	return fmt.Sprintf("Spawned subagent for task: %s", task), nil
}

func (sm *SubagentManager) runTask(ctx context.Context, task *SubagentTask) {
	startTime := time.Now()
	messages := []providers.Message{
		{
			Role:    "system",
			Content: "You are a subagent. Complete the given task independently and report the result.",
		},
		{
			Role:    "user",
			Content: task.Task,
		},
	}

	response, err := sm.provider.Chat(ctx, messages, nil, sm.provider.GetDefaultModel(), map[string]interface{}{
		"max_tokens": 4096,
	})

	sm.mu.Lock()
	task.ElapsedMs = time.Since(startTime).Milliseconds()
	task.LastActivity = time.Now().UnixMilli()

	if err != nil {
		if task.Status != "cancelled" { // Don't overwrite if already cancelled
			task.Status = "failed"
			task.Result = fmt.Sprintf("Error: %v", err)
		}
	} else {
		task.Status = "completed"
		task.Result = response.Content
	}
	sm.mu.Unlock()

	// Send announce message back to main agent
	if sm.bus != nil {
		label := task.Label
		if label == "" {
			label = task.ID
		}
		announceContent := fmt.Sprintf("Task '%s' %s (%.1fs).\n\nResult:\n%s",
			label, task.Status, float64(task.ElapsedMs)/1000, task.Result)
		sm.bus.PublishInbound(bus.InboundMessage{
			Channel:  "system",
			SenderID: fmt.Sprintf("subagent:%s", task.ID),
			ChatID:   fmt.Sprintf("%s:%s", task.OriginChannel, task.OriginChatID),
			Content:  announceContent,
		})
	}
}

// detectStall monitors a task for stalls (no activity for stallTimeout)
func (sm *SubagentManager) detectStall(task *SubagentTask) {
	ticker := time.NewTicker(sm.stallTimeout / 2)
	defer ticker.Stop()

	for range ticker.C {
		sm.mu.RLock()
		status := task.Status
		lastAct := task.LastActivity
		sm.mu.RUnlock()

		if status != "running" {
			return // Task finished, stop monitoring
		}

		sinceActivity := time.Since(time.UnixMilli(lastAct))
		if sinceActivity > sm.stallTimeout {
			sm.mu.Lock()
			if task.Status == "running" {
				task.Status = "stalled"
			}
			sm.mu.Unlock()

			// Emit stall warning via bus
			if sm.bus != nil {
				label := task.Label
				if label == "" {
					label = task.ID
				}
				sm.bus.PublishInbound(bus.InboundMessage{
					Channel:  "system",
					SenderID: fmt.Sprintf("subagent:%s", task.ID),
					ChatID:   fmt.Sprintf("%s:%s", task.OriginChannel, task.OriginChatID),
					Content:  fmt.Sprintf("⚠️ Subagent '%s' appears stalled (no activity for %s)", label, sinceActivity.Round(time.Second)),
				})
			}
			return
		}
	}
}

// CancelTask cancels a running subagent task
func (sm *SubagentManager) CancelTask(taskID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	task, ok := sm.tasks[taskID]
	if !ok {
		return fmt.Errorf("task %s not found", taskID)
	}
	if task.Status != "running" && task.Status != "stalled" {
		return fmt.Errorf("task %s is not running (status: %s)", taskID, task.Status)
	}

	task.Status = "cancelled"
	if task.cancel != nil {
		task.cancel()
	}
	return nil
}

func (sm *SubagentManager) GetTask(taskID string) (*SubagentTask, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	task, ok := sm.tasks[taskID]
	return task, ok
}

func (sm *SubagentManager) ListTasks() []*SubagentTask {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	tasks := make([]*SubagentTask, 0, len(sm.tasks))
	for _, task := range sm.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}
