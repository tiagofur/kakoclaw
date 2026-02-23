package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
)

// SpecialistSpawnTask represents a task spawned for a specialist
type SpecialistSpawnTask struct {
	ID             string
	SpecialistName string
	TaskContent    string
	Label          string
	OriginChannel  string
	OriginChatID   string
	Status         string
	Result         string
	Error          string
	CreatedAt      time.Time
	CompletedAt    time.Time
}

// SpecialistSpawner manages specialization-aware task spawning
type SpecialistSpawner struct {
	registry   *SpecialistRegistry
	messageBus *bus.MessageBus
	tasks      map[string]*SpecialistSpawnTask
	mu         sync.RWMutex
	nextID     int
}

// NewSpecialistSpawner creates a new specialist spawner
func NewSpecialistSpawner(registry *SpecialistRegistry, messageBus *bus.MessageBus) *SpecialistSpawner {
	return &SpecialistSpawner{
		registry:   registry,
		messageBus: messageBus,
		tasks:      make(map[string]*SpecialistSpawnTask),
		nextID:     1,
	}
}

// SpawnForSpecialist spawns a task for a specific specialist
func (ss *SpecialistSpawner) SpawnForSpecialist(
	ctx context.Context,
	specialistName string,
	taskContent string,
	label string,
	originChannel string,
	originChatID string,
) (string, error) {
	// Get the specialist
	specialist, err := ss.registry.GetSpecialist(specialistName)
	if err != nil {
		return "", fmt.Errorf("specialist not found: %w", err)
	}

	ss.mu.Lock()
	taskID := fmt.Sprintf("spawn-%s-%d", specialistName, ss.nextID)
	ss.nextID++

	spawnTask := &SpecialistSpawnTask{
		ID:             taskID,
		SpecialistName: specialistName,
		TaskContent:    taskContent,
		Label:          label,
		OriginChannel:  originChannel,
		OriginChatID:   originChatID,
		Status:         "running",
		CreatedAt:      time.Now(),
	}

	ss.tasks[taskID] = spawnTask
	ss.mu.Unlock()

	// Run the task asynchronously
	go ss.runSpecialistTask(ctx, specialist, spawnTask)

	displayLabel := label
	if displayLabel == "" {
		displayLabel = specialistName
	}
	return fmt.Sprintf("Spawned %s specialist for task: %s", displayLabel, taskContent), nil
}

// runSpecialistTask executes a task through a specialist with full tool access
func (ss *SpecialistSpawner) runSpecialistTask(
	ctx context.Context,
	specialist *SpecialistAgent,
	task *SpecialistSpawnTask,
) {
	task.Status = "running"

	// Execute through specialist with full tool support
	result, err := specialist.ProcessWithSpeciality(ctx, task.TaskContent)

	ss.mu.Lock()
	defer ss.mu.Unlock()

	task.CompletedAt = time.Now()

	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		task.Result = ""
	} else {
		task.Status = "completed"
		task.Result = result
		task.Error = ""
	}

	// Announce completion via message bus
	if ss.messageBus != nil {
		ss.messageBus.PublishOutbound(bus.OutboundMessage{
			UserID:  0, // Will be set by router if needed
			Channel: task.OriginChannel,
			ChatID:  task.OriginChatID,
			Content: fmt.Sprintf(
				"🔬 Specialist %s completed. Status: %s\nResult: %s",
				task.SpecialistName,
				task.Status,
				task.Result,
			),
		})
	}
}

// GetTask retrieves task information
func (ss *SpecialistSpawner) GetTask(taskID string) (*SpecialistSpawnTask, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	task, exists := ss.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	return task, nil
}

// ListTasks returns all spawn tasks
func (ss *SpecialistSpawner) ListTasks() []*SpecialistSpawnTask {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	tasks := make([]*SpecialistSpawnTask, 0, len(ss.tasks))
	for _, task := range ss.tasks {
		tasks = append(tasks, task)
	}
	return tasks
}

// ClearTask removes a completed task
func (ss *SpecialistSpawner) ClearTask(taskID string) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	task, exists := ss.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status != "completed" && task.Status != "failed" {
		return fmt.Errorf("cannot clear running task: %s", taskID)
	}

	delete(ss.tasks, taskID)
	return nil
}
