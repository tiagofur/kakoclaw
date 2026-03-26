package tools

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/cron"
	"github.com/sipeed/makoclaw/pkg/utils"
)

// JobExecutor is the interface for executing cron jobs through the agent
type JobExecutor interface {
	ProcessDirectWithChannel(ctx context.Context, content, sessionKey, channel, chatID string) (string, error)
	ProcessDirectWithChannelForUser(ctx context.Context, userID int64, content, sessionKey, channel, chatID string) (string, error)
}

// CronTool provides scheduling capabilities for the agent
type CronTool struct {
	cronService *cron.CronService
	executor    JobExecutor
	msgBus      *bus.MessageBus
	userID      int64
	channel     string
	chatID      string
	mu          sync.RWMutex
}

// NewCronTool creates a new CronTool
func NewCronTool(cronService *cron.CronService, executor JobExecutor, msgBus *bus.MessageBus) *CronTool {
	return &CronTool{
		cronService: cronService,
		executor:    executor,
		msgBus:      msgBus,
		userID:      1, // Default to user 1 for backward compatibility
	}
}

// SetUserID implements the UserAwareTool interface.
func (t *CronTool) SetUserID(userID int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if userID > 0 {
		t.userID = userID
	}
}

// Name returns the tool name
func (t *CronTool) Name() string {
	return "cron"
}

// Description returns the tool description
func (t *CronTool) Description() string {
	return "Schedule reminders and tasks. IMPORTANT: When user asks to be reminded or scheduled, you MUST call this tool. Use 'at_seconds' for one-time reminders (e.g., 'remind me in 10 minutes' → at_seconds=600). Use 'every_seconds' ONLY for recurring tasks (e.g., 'every 2 hours' → every_seconds=7200). Use 'cron_expr' for complex recurring schedules (e.g., '0 9 * * *' for daily at 9am)."
}

// Parameters returns the tool parameters schema
func (t *CronTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"add", "update", "edit", "consolidate", "list", "remove", "enable", "disable"},
				"description": "Action to perform. Use 'add' when user wants to schedule a reminder or task.",
			},
			"name": map[string]interface{}{
				"type":        "string",
				"description": "Optional job name override. For update/edit, defaults to the existing name unless a new message is provided.",
			},
			"message": map[string]interface{}{
				"type":        "string",
				"description": "The reminder/task message to display when triggered (required for add, optional for update/edit)",
			},
			"at_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "One-time reminder: seconds from now when to trigger (e.g., 600 for 10 minutes later). Use this for one-time reminders like 'remind me in 10 minutes'.",
			},
			"every_seconds": map[string]interface{}{
				"type":        "integer",
				"description": "Recurring interval in seconds (e.g., 3600 for every hour). Use this ONLY for recurring tasks like 'every 2 hours' or 'daily reminder'.",
			},
			"cron_expr": map[string]interface{}{
				"type":        "string",
				"description": "Cron expression for complex recurring schedules (e.g., '0 9 * * *' for daily at 9am). Use this for complex recurring schedules.",
			},
			"job_id": map[string]interface{}{
				"type":        "string",
				"description": "Job ID (for update/edit/remove/enable/disable)",
			},
			"job_ids": map[string]interface{}{
				"type":        "array",
				"items":       map[string]interface{}{"type": "string"},
				"description": "List of source job IDs to consolidate (for consolidate action).",
			},
			"remove_originals": map[string]interface{}{
				"type":        "boolean",
				"default":     true,
				"description": "Whether to remove source jobs after creating the consolidated job.",
			},
			"confirmed": map[string]interface{}{
				"type":        "boolean",
				"description": "Set to true only after explicit user confirmation for destructive operations.",
			},
			"job_type": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"task", "reminder"},
				"default":     "task",
				"description": "Job type: 'task' (default, process through agent for complex operations) or 'reminder' (send direct message without agent processing). Use 'task' when the prompt needs processing (e.g., 'check weather and summarize'), use 'reminder' for simple notifications (e.g., 'Meeting in 10 minutes').",
			},
		},
		"required": []string{"action"},
	}
}

// SetContext sets the current session context for job creation
func (t *CronTool) SetContext(channel, chatID string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.channel = channel
	t.chatID = chatID
}

// Execute runs the tool with given arguments
func (t *CronTool) Execute(ctx context.Context, args map[string]interface{}) (string, error) {
	action, ok := args["action"].(string)
	if !ok {
		return "", fmt.Errorf("action is required")
	}

	switch action {
	case "add":
		return t.addJob(ctx, args)
	case "update", "edit":
		return t.updateJob(ctx, args)
	case "consolidate":
		return t.consolidateJobs(ctx, args)
	case "list":
		return t.listJobs()
	case "remove":
		return t.removeJob(ctx, args)
	case "enable":
		return t.enableJob(args, true)
	case "disable":
		return t.enableJob(args, false)
	default:
		return "", fmt.Errorf("unknown action: %s", action)
	}
}

func getStringSliceArg(args map[string]interface{}, key string) ([]string, bool) {
	value, ok := args[key]
	if !ok {
		return nil, false
	}

	if values, ok := value.([]string); ok {
		if len(values) == 0 {
			return nil, false
		}
		return values, true
	}

	rawValues, ok := value.([]interface{})
	if !ok || len(rawValues) == 0 {
		return nil, false
	}

	result := make([]string, 0, len(rawValues))
	for _, item := range rawValues {
		asString, ok := item.(string)
		if !ok {
			return nil, false
		}
		asString = strings.TrimSpace(asString)
		if asString == "" {
			return nil, false
		}
		result = append(result, asString)
	}

	return result, true
}

func getSecondsArg(args map[string]interface{}, key string) (int64, bool) {
	value, ok := args[key]
	if !ok {
		return 0, false
	}

	switch numeric := value.(type) {
	case float64:
		return int64(numeric), true
	case int:
		return int64(numeric), true
	case int64:
		return numeric, true
	default:
		return 0, false
	}
}

func buildScheduleFromArgs(args map[string]interface{}, existing *cron.CronSchedule) (*cron.CronSchedule, error) {
	if atSeconds, ok := getSecondsArg(args, "at_seconds"); ok {
		atMS := time.Now().UnixMilli() + atSeconds*1000
		schedule := cron.CronSchedule{Kind: "at", AtMS: &atMS}
		return &schedule, nil
	}

	if everySeconds, ok := getSecondsArg(args, "every_seconds"); ok {
		everyMS := everySeconds * 1000
		schedule := cron.CronSchedule{Kind: "every", EveryMS: &everyMS}
		return &schedule, nil
	}

	if cronExpr, ok := args["cron_expr"].(string); ok && strings.TrimSpace(cronExpr) != "" {
		schedule := cron.CronSchedule{Kind: "cron", Expr: strings.TrimSpace(cronExpr)}
		return &schedule, nil
	}

	if existing != nil {
		schedule := *existing
		return &schedule, nil
	}

	return nil, fmt.Errorf("one of at_seconds, every_seconds, or cron_expr is required")
}

func (t *CronTool) addJob(ctx context.Context, args map[string]interface{}) (string, error) {
	t.mu.RLock()
	channel := t.channel
	chatID := t.chatID
	userID := t.userID
	t.mu.RUnlock()

	if channel == "" || chatID == "" {
		return "Error: no session context (channel/chat_id not set). Use this tool in an active conversation.", nil
	}

	message, ok := args["message"].(string)
	if !ok || message == "" {
		return "Error: message is required for add", nil
	}

	schedule, err := buildScheduleFromArgs(args, nil)
	if err != nil {
		return "Error: one of at_seconds, every_seconds, or cron_expr is required", nil
	}

	// Read job_type parameter, default to "task"
	jobType := "task" // Default to agent processing
	if jt, ok := args["job_type"].(string); ok && jt != "" {
		jobType = jt
	} else if d, ok := args["deliver"].(bool); ok {
		// Backward compatibility: if deliver is set, convert it
		if d {
			jobType = "reminder"
		} else {
			jobType = "task"
		}
	}

	// Truncate message for job name (max 30 chars)
	messagePreview := utils.Truncate(message, 30)

	// Extract Agent Name from the context
	agentName := utils.AgentNameFrom(ctx)

	job, err := t.cronService.AddJob(
		userID,
		messagePreview,
		*schedule,
		message,
		jobType,
		channel,
		chatID,
		agentName,
	)
	if err != nil {
		return fmt.Sprintf("Error adding job: %v", err), nil
	}

	return fmt.Sprintf("Created job '%s' (id: %s)", job.Name, job.ID), nil
}

func (t *CronTool) updateJob(ctx context.Context, args map[string]interface{}) (string, error) {
	t.mu.RLock()
	userID := t.userID
	t.mu.RUnlock()

	jobID, ok := args["job_id"].(string)
	if !ok || strings.TrimSpace(jobID) == "" {
		return "Error: job_id is required for update/edit", nil
	}

	jobs := t.cronService.ListJobsForUser(userID, true)
	var existing *cron.CronJob
	for index := range jobs {
		if jobs[index].ID == jobID {
			existing = &jobs[index]
			break
		}
	}
	if existing == nil {
		return fmt.Sprintf("Job %s not found", jobID), nil
	}

	schedule, err := buildScheduleFromArgs(args, &existing.Schedule)
	if err != nil {
		return fmt.Sprintf("Error updating job: %v", err), nil
	}

	message := existing.Payload.Message
	if updatedMessage, ok := args["message"].(string); ok && strings.TrimSpace(updatedMessage) != "" {
		message = updatedMessage
	}

	name := existing.Name
	if updatedName, ok := args["name"].(string); ok && strings.TrimSpace(updatedName) != "" {
		name = strings.TrimSpace(updatedName)
	} else if message != existing.Payload.Message {
		name = utils.Truncate(message, 30)
	}

	jobType := existing.Payload.JobType
	if updatedType, ok := args["job_type"].(string); ok && updatedType != "" {
		jobType = updatedType
	}
	if jobType == "" {
		jobType = "task"
	}

	updatedJob, err := t.cronService.UpdateJobForUser(
		userID,
		jobID,
		name,
		*schedule,
		message,
		jobType,
		existing.Payload.Channel,
		existing.Payload.To,
		existing.Payload.Agent,
	)
	if err != nil {
		return fmt.Sprintf("Error updating job: %v", err), nil
	}
	if updatedJob == nil {
		return fmt.Sprintf("Job %s not found", jobID), nil
	}

	return fmt.Sprintf("Updated job '%s' (id: %s)", updatedJob.Name, updatedJob.ID), nil
}

func (t *CronTool) consolidateJobs(ctx context.Context, args map[string]interface{}) (string, error) {
	t.mu.RLock()
	channel := t.channel
	chatID := t.chatID
	userID := t.userID
	t.mu.RUnlock()

	jobIDs, ok := getStringSliceArg(args, "job_ids")
	if !ok || len(jobIDs) < 2 {
		return "Error: job_ids must include at least two job IDs for consolidate", nil
	}

	jobs := t.cronService.ListJobsForUser(userID, true)
	jobsByID := make(map[string]cron.CronJob, len(jobs))
	for _, job := range jobs {
		jobsByID[job.ID] = job
	}

	sourceJobs := make([]cron.CronJob, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		source, exists := jobsByID[jobID]
		if !exists {
			return fmt.Sprintf("Error: source job %s not found", jobID), nil
		}
		sourceJobs = append(sourceJobs, source)
	}

	first := sourceJobs[0]
	schedule, err := buildScheduleFromArgs(args, &first.Schedule)
	if err != nil {
		return fmt.Sprintf("Error consolidating jobs: %v", err), nil
	}

	message := first.Payload.Message
	if updatedMessage, ok := args["message"].(string); ok && strings.TrimSpace(updatedMessage) != "" {
		message = strings.TrimSpace(updatedMessage)
	}

	name := utils.Truncate(message, 30)
	if customName, ok := args["name"].(string); ok && strings.TrimSpace(customName) != "" {
		name = strings.TrimSpace(customName)
	}

	jobType := first.Payload.JobType
	if updatedType, ok := args["job_type"].(string); ok && strings.TrimSpace(updatedType) != "" {
		jobType = strings.TrimSpace(updatedType)
	}
	if jobType == "" {
		jobType = "task"
	}

	jobChannel := first.Payload.Channel
	if jobChannel == "" {
		jobChannel = channel
	}
	jobTo := first.Payload.To
	if jobTo == "" {
		jobTo = chatID
	}

	newJob, err := t.cronService.AddJob(
		userID,
		name,
		*schedule,
		message,
		jobType,
		jobChannel,
		jobTo,
		utils.AgentNameFrom(ctx),
	)
	if err != nil {
		return fmt.Sprintf("Error consolidating jobs: %v", err), nil
	}

	removeOriginals := true
	if explicit, ok := args["remove_originals"].(bool); ok {
		removeOriginals = explicit
	}
	if IsSensitiveConfirmationRequired(ctx) && removeOriginals && !isConfirmed(args) {
		return "Confirmation required: consolidate with remove_originals=true is destructive. Re-run with confirmed=true after user approval.", nil
	}

	removed := 0
	failedRemovals := make([]string, 0)
	if removeOriginals {
		for _, source := range sourceJobs {
			if t.cronService.RemoveJobForUser(userID, source.ID) {
				removed++
			} else {
				failedRemovals = append(failedRemovals, source.ID)
			}
		}
	}

	if len(failedRemovals) > 0 {
		return fmt.Sprintf(
			"Consolidated into job '%s' (id: %s). Removed %d/%d originals, failed removals: %s",
			newJob.Name,
			newJob.ID,
			removed,
			len(sourceJobs),
			strings.Join(failedRemovals, ", "),
		), nil
	}

	if !removeOriginals {
		return fmt.Sprintf(
			"Consolidated into job '%s' (id: %s). Original jobs kept (%d source jobs).",
			newJob.Name,
			newJob.ID,
			len(sourceJobs),
		), nil
	}

	return fmt.Sprintf(
		"Consolidated %d jobs into '%s' (id: %s) and removed all originals.",
		len(sourceJobs),
		newJob.Name,
		newJob.ID,
	), nil
}

func (t *CronTool) listJobs() (string, error) {
	t.mu.RLock()
	userID := t.userID
	t.mu.RUnlock()

	jobs := t.cronService.ListJobsForUser(userID, false)

	if len(jobs) == 0 {
		return "No scheduled jobs.", nil
	}

	result := "Scheduled jobs:\n"
	for _, j := range jobs {
		var scheduleInfo string
		if j.Schedule.Kind == "every" && j.Schedule.EveryMS != nil {
			scheduleInfo = fmt.Sprintf("every %ds", *j.Schedule.EveryMS/1000)
		} else if j.Schedule.Kind == "cron" {
			scheduleInfo = j.Schedule.Expr
		} else if j.Schedule.Kind == "at" {
			scheduleInfo = "one-time"
		} else {
			scheduleInfo = "unknown"
		}
		result += fmt.Sprintf("- %s (id: %s, %s)\n", j.Name, j.ID, scheduleInfo)
	}

	return result, nil
}

func (t *CronTool) removeJob(ctx context.Context, args map[string]interface{}) (string, error) {
	t.mu.RLock()
	userID := t.userID
	t.mu.RUnlock()

	if IsSensitiveConfirmationRequired(ctx) && !isConfirmed(args) {
		return "Confirmation required: remove is destructive. Re-run with confirmed=true after user approval.", nil
	}

	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return "Error: job_id is required for remove", nil
	}

	if t.cronService.RemoveJobForUser(userID, jobID) {
		return fmt.Sprintf("Removed job %s", jobID), nil
	}
	return fmt.Sprintf("Job %s not found", jobID), nil
}

func (t *CronTool) enableJob(args map[string]interface{}, enable bool) (string, error) {
	t.mu.RLock()
	userID := t.userID
	t.mu.RUnlock()

	jobID, ok := args["job_id"].(string)
	if !ok || jobID == "" {
		return "Error: job_id is required for enable/disable", nil
	}

	job := t.cronService.EnableJobForUser(userID, jobID, enable)
	if job == nil {
		return fmt.Sprintf("Job %s not found", jobID), nil
	}

	status := "enabled"
	if !enable {
		status = "disabled"
	}
	return fmt.Sprintf("Job '%s' %s", job.Name, status), nil
}

// ExecuteJob executes a cron job through the agent
func (t *CronTool) ExecuteJob(ctx context.Context, job *cron.CronJob) string {
	// Get channel/chatID from job payload
	channel := job.Payload.Channel
	chatID := job.Payload.To

	// Default values if not set
	if channel == "" {
		channel = "cli"
	}
	if chatID == "" {
		chatID = "direct"
	}

	// Migrate old format if needed (backward compatibility)
	jobType := job.Payload.JobType
	if jobType == "" {
		if job.Payload.Deliver {
			jobType = "reminder"
		} else {
			jobType = "task"
		}
	}

	// Execute based on job type
	switch jobType {
	case "reminder":
		// Direct channel delivery (simple notification)
		msgContent := job.Payload.Message
		if job.Payload.Agent != "" {
			msgContent = fmt.Sprintf("%s\n\n*(Scheduled by agent: %s)*", msgContent, job.Payload.Agent)
		}
		
		t.msgBus.PublishOutbound(bus.OutboundMessage{
			UserID:  job.UserID,
			Channel: channel,
			ChatID:  chatID,
			Content: msgContent,
		})
		return "delivered reminder"

	case "task":
		fallthrough
	default:
		// Process through agent (for complex tasks)
		sessionKey := fmt.Sprintf("cron-%s", job.ID)

		// Call agent with the job's message
		response, err := t.executor.ProcessDirectWithChannelForUser(
			ctx,
			job.UserID,
			job.Payload.Message,
			sessionKey,
			channel,
			chatID,
		)

		if err != nil {
			return fmt.Sprintf("Error: %v", err)
		}

		// Response is automatically sent via MessageBus by AgentLoop
		_ = response // Will be sent by AgentLoop
		return "task completed"
	}
}
