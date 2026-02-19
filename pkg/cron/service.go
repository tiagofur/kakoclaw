package cron

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adhocore/gronx"
)

type CronSchedule struct {
	Kind    string `json:"kind"`
	AtMS    *int64 `json:"atMs,omitempty"`
	EveryMS *int64 `json:"everyMs,omitempty"`
	Expr    string `json:"expr,omitempty"`
	TZ      string `json:"tz,omitempty"`
}

type CronPayload struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Deliver bool   `json:"deliver"`
	Channel string `json:"channel,omitempty"`
	To      string `json:"to,omitempty"`
}

type CronJobState struct {
	NextRunAtMS *int64 `json:"nextRunAtMs,omitempty"`
	LastRunAtMS *int64 `json:"lastRunAtMs,omitempty"`
	LastStatus  string `json:"lastStatus,omitempty"`
	LastError   string `json:"lastError,omitempty"`
}

type CronJob struct {
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Enabled        bool         `json:"enabled"`
	Schedule       CronSchedule `json:"schedule"`
	Payload        CronPayload  `json:"payload"`
	State          CronJobState `json:"state"`
	CreatedAtMS    int64        `json:"createdAtMs"`
	UpdatedAtMS    int64        `json:"updatedAtMs"`
	DeleteAfterRun bool         `json:"deleteAfterRun"`
}

type CronStore struct {
	Version int       `json:"version"`
	Jobs    []CronJob `json:"jobs"`
}

type JobHandler func(job *CronJob) (string, error)

type CronService struct {
	storePath string
	store     *CronStore
	onJob     JobHandler
	mu        sync.RWMutex
	running   bool
	stopChan  chan struct{}
	gronx     *gronx.Gronx
}

func NewCronService(storePath string, onJob JobHandler) *CronService {
	cs := &CronService{
		storePath: storePath,
		onJob:     onJob,
		stopChan:  make(chan struct{}),
		gronx:     gronx.New(),
	}
	// Initialize and load store on creation
	cs.loadStore()
	return cs
}

func (cs *CronService) Start() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.running {
		return nil
	}

	if err := cs.loadStore(); err != nil {
		return fmt.Errorf("failed to load store: %w", err)
	}

	cs.recomputeNextRuns()
	if err := cs.saveStoreUnsafe(); err != nil {
		return fmt.Errorf("failed to save store: %w", err)
	}

	cs.running = true
	go cs.runLoop()

	return nil
}

func (cs *CronService) Stop() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.running {
		return
	}

	cs.running = false
	close(cs.stopChan)
}

func (cs *CronService) runLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopChan:
			return
		case <-ticker.C:
			cs.checkJobs()
		}
	}
}

func (cs *CronService) checkJobs() {
	cs.mu.Lock()

	if !cs.running {
		cs.mu.Unlock()
		return
	}

	now := time.Now().UnixMilli()
	var dueJobs []*CronJob

	// Collect jobs that are due (we need to copy them to execute outside lock)
	for i := range cs.store.Jobs {
		job := &cs.store.Jobs[i]
		if job.Enabled && job.State.NextRunAtMS != nil && *job.State.NextRunAtMS <= now {
			// Create a shallow copy of the job for execution
			jobCopy := *job
			dueJobs = append(dueJobs, &jobCopy)
		}
	}

	// Update next run times for due jobs immediately (before executing)
	// Use map for O(n) lookup instead of O(n²) nested loop
	dueMap := make(map[string]bool, len(dueJobs))
	for _, job := range dueJobs {
		dueMap[job.ID] = true
	}
	for i := range cs.store.Jobs {
		if dueMap[cs.store.Jobs[i].ID] {
			// Reset NextRunAtMS temporarily so we don't re-execute
			cs.store.Jobs[i].State.NextRunAtMS = nil
		}
	}

	if err := cs.saveStoreUnsafe(); err != nil {
		log.Printf("[cron] failed to save store: %v", err)
	}

	cs.mu.Unlock()

	// Execute jobs outside the lock
	for _, job := range dueJobs {
		cs.executeJob(job)
	}
}

// ExecuteJob executes a job immediately (public wrapper for executeJob, mainly for testing/manual triggering)
func (cs *CronService) RunJob(jobID string) error {
	cs.mu.RLock()
	var targetJob *CronJob
	for i := range cs.store.Jobs {
		if cs.store.Jobs[i].ID == jobID {
			// Copy the job to execute outside the lock
			jobCopy := cs.store.Jobs[i]
			targetJob = &jobCopy
			break
		}
	}
	cs.mu.RUnlock()

	if targetJob == nil {
		return fmt.Errorf("job not found")
	}

	cs.executeJob(targetJob)
	return nil
}

func (cs *CronService) executeJob(job *CronJob) {
	startTime := time.Now().UnixMilli()

	var err error
	if cs.onJob != nil {
		_, err = cs.onJob(job)
	}

	// Now acquire lock to update state
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Find the job in store and update it
	for i := range cs.store.Jobs {
		if cs.store.Jobs[i].ID == job.ID {
			cs.store.Jobs[i].State.LastRunAtMS = &startTime
			cs.store.Jobs[i].UpdatedAtMS = time.Now().UnixMilli()

			if err != nil {
				cs.store.Jobs[i].State.LastStatus = "error"
				cs.store.Jobs[i].State.LastError = err.Error()
			} else {
				cs.store.Jobs[i].State.LastStatus = "ok"
				cs.store.Jobs[i].State.LastError = ""
			}

			// Compute next run time
			if cs.store.Jobs[i].Schedule.Kind == "at" {
				if cs.store.Jobs[i].DeleteAfterRun {
					cs.removeJobUnsafe(job.ID)
				} else {
					cs.store.Jobs[i].Enabled = false
					cs.store.Jobs[i].State.NextRunAtMS = nil
				}
			} else {
				nextRun := cs.computeNextRun(&cs.store.Jobs[i].Schedule, time.Now().UnixMilli())
				cs.store.Jobs[i].State.NextRunAtMS = nextRun
			}
			break
		}
	}

	if err := cs.saveStoreUnsafe(); err != nil {
		log.Printf("[cron] failed to save store: %v", err)
	}
}

func (cs *CronService) computeNextRun(schedule *CronSchedule, nowMS int64) *int64 {
	if schedule.Kind == "at" {
		if schedule.AtMS != nil && *schedule.AtMS > nowMS {
			return schedule.AtMS
		}
		return nil
	}

	if schedule.Kind == "every" {
		if schedule.EveryMS == nil || *schedule.EveryMS <= 0 {
			return nil
		}
		next := nowMS + *schedule.EveryMS
		return &next
	}

	if schedule.Kind == "cron" {
		if schedule.Expr == "" {
			return nil
		}

		// Use gronx to calculate next run time
		now := time.UnixMilli(nowMS)

		// Apply timezone if specified
		if schedule.TZ != "" {
			if loc, err := time.LoadLocation(schedule.TZ); err == nil {
				now = now.In(loc)
			}
		}

		nextTime, err := gronx.NextTickAfter(schedule.Expr, now, false)
		if err != nil {
			log.Printf("[cron] failed to compute next run for expr '%s': %v", schedule.Expr, err)
			return nil
		}

		// Convert back to UTC milliseconds
		nextMS := nextTime.UTC().UnixMilli()
		return &nextMS
	}

	return nil
}

func (cs *CronService) recomputeNextRuns() {
	now := time.Now().UnixMilli()
	for i := range cs.store.Jobs {
		job := &cs.store.Jobs[i]
		if job.Enabled {
			job.State.NextRunAtMS = cs.computeNextRun(&job.Schedule, now)
		}
	}
}

func (cs *CronService) getNextWakeMS() *int64 {
	var nextWake *int64
	for _, job := range cs.store.Jobs {
		if job.Enabled && job.State.NextRunAtMS != nil {
			if nextWake == nil || *job.State.NextRunAtMS < *nextWake {
				nextWake = job.State.NextRunAtMS
			}
		}
	}
	return nextWake
}

func (cs *CronService) Load() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.loadStore()
}

func (cs *CronService) SetOnJob(handler JobHandler) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.onJob = handler
}

func (cs *CronService) loadStore() error {
	cs.store = &CronStore{
		Version: 1,
		Jobs:    []CronJob{},
	}

	data, err := os.ReadFile(cs.storePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, cs.store)
}

func (cs *CronService) saveStoreUnsafe() error {
	dir := filepath.Dir(cs.storePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cs.store, "", "  ")
	if err != nil {
		return err
	}

	// Atomic write: write to temp file then rename to avoid corruption on crash.
	tmpPath := cs.storePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return err
	}
	return os.Rename(tmpPath, cs.storePath)
}

// ValidateSchedule checks that a CronSchedule is well-formed.
// Returns an error describing the problem, or nil if valid.
func (cs *CronService) ValidateSchedule(schedule *CronSchedule) error {
	switch schedule.Kind {
	case "at":
		if schedule.AtMS == nil {
			return fmt.Errorf("atMs is required for 'at' schedule")
		}
	case "every":
		if schedule.EveryMS == nil || *schedule.EveryMS <= 0 {
			return fmt.Errorf("everyMs must be a positive integer for 'every' schedule")
		}
	case "cron":
		if schedule.Expr == "" {
			return fmt.Errorf("expr is required for 'cron' schedule")
		}
		if !cs.gronx.IsValid(schedule.Expr) {
			return fmt.Errorf("invalid cron expression: %s", schedule.Expr)
		}
	default:
		return fmt.Errorf("unknown schedule kind: %s", schedule.Kind)
	}
	return nil
}

func (cs *CronService) AddJob(name string, schedule CronSchedule, message string, deliver bool, channel, to string) (*CronJob, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Validate schedule
	if err := cs.ValidateSchedule(&schedule); err != nil {
		return nil, fmt.Errorf("invalid schedule: %w", err)
	}

	now := time.Now().UnixMilli()

	// One-time tasks (at) should be deleted after execution
	deleteAfterRun := (schedule.Kind == "at")

	job := CronJob{
		ID:       generateID(),
		Name:     name,
		Enabled:  true,
		Schedule: schedule,
		Payload: CronPayload{
			Kind:    "agent_turn",
			Message: message,
			Deliver: deliver,
			Channel: channel,
			To:      to,
		},
		State: CronJobState{
			NextRunAtMS: cs.computeNextRun(&schedule, now),
		},
		CreatedAtMS:    now,
		UpdatedAtMS:    now,
		DeleteAfterRun: deleteAfterRun,
	}

	cs.store.Jobs = append(cs.store.Jobs, job)
	if err := cs.saveStoreUnsafe(); err != nil {
		return nil, err
	}

	return &job, nil
}

// UpdateJob updates an existing job's name, schedule, and payload fields.
// Returns the updated job, or nil if not found.
func (cs *CronService) UpdateJob(jobID string, name string, schedule CronSchedule, message string, deliver bool, channel, to string) (*CronJob, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Validate schedule
	if err := cs.ValidateSchedule(&schedule); err != nil {
		return nil, fmt.Errorf("invalid schedule: %w", err)
	}

	for i := range cs.store.Jobs {
		if cs.store.Jobs[i].ID == jobID {
			now := time.Now().UnixMilli()

			cs.store.Jobs[i].Name = name
			cs.store.Jobs[i].Schedule = schedule
			cs.store.Jobs[i].Payload.Message = message
			cs.store.Jobs[i].Payload.Deliver = deliver
			cs.store.Jobs[i].Payload.Channel = channel
			cs.store.Jobs[i].Payload.To = to
			cs.store.Jobs[i].UpdatedAtMS = now
			cs.store.Jobs[i].DeleteAfterRun = (schedule.Kind == "at")

			// Recompute next run if enabled
			if cs.store.Jobs[i].Enabled {
				cs.store.Jobs[i].State.NextRunAtMS = cs.computeNextRun(&schedule, now)
			}

			if err := cs.saveStoreUnsafe(); err != nil {
				return nil, err
			}

			job := cs.store.Jobs[i]
			return &job, nil
		}
	}

	return nil, nil
}

func (cs *CronService) RemoveJob(jobID string) bool {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	return cs.removeJobUnsafe(jobID)
}

func (cs *CronService) removeJobUnsafe(jobID string) bool {
	before := len(cs.store.Jobs)
	var jobs []CronJob
	for _, job := range cs.store.Jobs {
		if job.ID != jobID {
			jobs = append(jobs, job)
		}
	}
	cs.store.Jobs = jobs
	removed := len(cs.store.Jobs) < before

	if removed {
		if err := cs.saveStoreUnsafe(); err != nil {
			log.Printf("[cron] failed to save store after remove: %v", err)
		}
	}

	return removed
}

func (cs *CronService) EnableJob(jobID string, enabled bool) *CronJob {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i := range cs.store.Jobs {
		job := &cs.store.Jobs[i]
		if job.ID == jobID {
			job.Enabled = enabled
			job.UpdatedAtMS = time.Now().UnixMilli()

			if enabled {
				job.State.NextRunAtMS = cs.computeNextRun(&job.Schedule, time.Now().UnixMilli())
			} else {
				job.State.NextRunAtMS = nil
			}

			if err := cs.saveStoreUnsafe(); err != nil {
				log.Printf("[cron] failed to save store after enable: %v", err)
			}
			return job
		}
	}

	return nil
}

func (cs *CronService) ListJobs(includeDisabled bool) []CronJob {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if includeDisabled {
		return cs.store.Jobs
	}

	var enabled []CronJob
	for _, job := range cs.store.Jobs {
		if job.Enabled {
			enabled = append(enabled, job)
		}
	}

	return enabled
}

func (cs *CronService) Status() map[string]interface{} {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var enabledCount int
	for _, job := range cs.store.Jobs {
		if job.Enabled {
			enabledCount++
		}
	}

	return map[string]interface{}{
		"enabled":      cs.running,
		"jobs":         len(cs.store.Jobs),
		"nextWakeAtMS": cs.getNextWakeMS(),
	}
}

func generateID() string {
	// Use crypto/rand for better uniqueness under concurrent access
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fallback to time-based if crypto/rand fails
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
