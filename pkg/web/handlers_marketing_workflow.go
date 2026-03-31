package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/marketing/workflow"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// mktWorkflow is the shared campaign workflow engine (stateless, thread-safe).
var mktWorkflow = workflow.NewEngine()

// handleMarketingWorkflowRouter dispatches all workflow endpoints:
//
//	/api/v1/marketing/campaigns/{account}/{campaign}/workflow
//	/api/v1/marketing/campaigns/{account}/{campaign}/workflow/start
//	/api/v1/marketing/campaigns/{account}/{campaign}/workflow/generate
//	/api/v1/marketing/campaigns/{account}/{campaign}/workflow/approve
//	/api/v1/marketing/campaigns/{account}/{campaign}/workflow/reject
//	/api/v1/marketing/campaigns/{account}/{campaign}/workflow/skip
func (s *Server) handleMarketingWorkflowRouter(w http.ResponseWriter, r *http.Request) {
	// Expect path: /api/v1/marketing/campaigns/{account}/{campaign}/workflow[/{action}]
	suffix := strings.TrimPrefix(r.URL.Path, "/api/v1/marketing/campaigns/")
	parts := strings.SplitN(suffix, "/", 4)
	// parts[0] = account, parts[1] = campaign, parts[2] = "workflow", parts[3] = action (optional)
	if len(parts) < 3 || parts[2] != "workflow" {
		http.Error(w, "invalid workflow path", http.StatusBadRequest)
		return
	}

	account := parts[0]
	campaign := parts[1]
	action := ""
	if len(parts) >= 4 {
		action = parts[3]
	}

	if isUnsafeMarketingSegment(account) || isUnsafeMarketingSegment(campaign) {
		http.Error(w, "access denied", http.StatusForbidden)
		return
	}

	userStore, userUUID, ok := s.getUserStorage(r)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	userWorkspace, err := config.EnsureUserWorkspace(userUUID)
	if err != nil {
		http.Error(w, "failed to access workspace", http.StatusInternalServerError)
		return
	}

	campaignPath := filepath.Join(userWorkspace, "marketing", account, campaign)
	if _, err := os.Stat(campaignPath); os.IsNotExist(err) {
		http.Error(w, "campaign not found", http.StatusNotFound)
		return
	}

	wfEngine := mktWorkflow

	switch {
	case action == "" && r.Method == http.MethodGet:
		s.handleWorkflowGetStatus(w, wfEngine, account, campaign, campaignPath)
	case action == "start" && r.Method == http.MethodPost:
		s.handleWorkflowStart(w, wfEngine, account, campaign, campaignPath)
	case action == "generate" && r.Method == http.MethodPost:
		s.handleWorkflowGenerate(w, r, wfEngine, userStore, userUUID, account, campaign, campaignPath)
	case action == "approve" && r.Method == http.MethodPost:
		s.handleWorkflowApprove(w, wfEngine, campaignPath)
	case action == "reject" && r.Method == http.MethodPost:
		s.handleWorkflowReject(w, r, wfEngine, campaignPath)
	case action == "skip" && r.Method == http.MethodPost:
		s.handleWorkflowSkip(w, wfEngine, campaignPath)
	case action == "reset" && r.Method == http.MethodPost:
		s.handleWorkflowReset(w, r, wfEngine, campaignPath)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleWorkflowGetStatus returns the current workflow state.
func (s *Server) handleWorkflowGetStatus(w http.ResponseWriter, engine *workflow.Engine, account, campaign, campaignPath string) {
	state, err := engine.Load(campaignPath)
	if err != nil {
		http.Error(w, "failed to load workflow: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if state == nil {
		writeJSONResponse(w, map[string]interface{}{
			"status":   "not_started",
			"account":  account,
			"campaign": campaign,
		})
		return
	}

	completed, total, pct := state.Progress()

	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"started_at":    state.StartedAt,
		"completed_at":  state.CompletedAt,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// handleWorkflowStart initializes the workflow for a campaign.
func (s *Server) handleWorkflowStart(w http.ResponseWriter, engine *workflow.Engine, account, campaign, campaignPath string) {
	state, err := engine.Start(account, campaign, campaignPath)
	if err != nil {
		http.Error(w, "failed to start workflow: "+err.Error(), http.StatusInternalServerError)
		return
	}

	completed, total, pct := state.Progress()
	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"started_at":    state.StartedAt,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// handleWorkflowGenerate invokes AI to generate content for the current stage.
func (s *Server) handleWorkflowGenerate(w http.ResponseWriter, r *http.Request, engine *workflow.Engine, userStore *storage.Storage, userUUID, account, campaign, campaignPath string) {
	var req workflow.GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is fine — use defaults
		req = workflow.GenerateRequest{}
	}

	// Create the AI invoker using the existing agent loop
	invoker := &agentLoopInvoker{
		server:    s,
		userUUID:  userUUID,
		userStore: userStore,
	}

	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()

	state, result, err := engine.Generate(ctx, campaignPath, invoker, req)
	if err != nil {
		http.Error(w, "generation failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	completed, total, pct := state.Progress()
	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"ai_output":     result,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// handleWorkflowApprove approves the current stage.
func (s *Server) handleWorkflowApprove(w http.ResponseWriter, engine *workflow.Engine, campaignPath string) {
	state, err := engine.Approve(campaignPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed, total, pct := state.Progress()
	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"completed_at":  state.CompletedAt,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// handleWorkflowReject rejects the current stage with feedback.
func (s *Server) handleWorkflowReject(w http.ResponseWriter, r *http.Request, engine *workflow.Engine, campaignPath string) {
	var req workflow.RejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Feedback) == "" {
		http.Error(w, "feedback is required", http.StatusBadRequest)
		return
	}

	state, err := engine.Reject(campaignPath, req.Feedback)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed, total, pct := state.Progress()
	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// handleWorkflowSkip skips the current stage.
func (s *Server) handleWorkflowSkip(w http.ResponseWriter, engine *workflow.Engine, campaignPath string) {
	state, err := engine.Skip(campaignPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed, total, pct := state.Progress()
	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// handleWorkflowReset resets a specific stage back to pending.
func (s *Server) handleWorkflowReset(w http.ResponseWriter, r *http.Request, engine *workflow.Engine, campaignPath string) {
	var req struct {
		Stage workflow.Stage `json:"stage"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	state, err := engine.ResetStage(campaignPath, req.Stage)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	completed, total, pct := state.Progress()
	writeJSONResponse(w, map[string]interface{}{
		"status":        "active",
		"account":       state.Account,
		"campaign":      state.Campaign,
		"current_stage": state.CurrentStage,
		"stages":        state.Stages,
		"progress": map[string]interface{}{
			"completed": completed,
			"total":     total,
			"pct":       pct,
		},
	})
}

// agentLoopInvoker implements workflow.AIInvoker using the server's agent loop.
type agentLoopInvoker struct {
	server    *Server
	userUUID  string
	userStore *storage.Storage
}

func (a *agentLoopInvoker) Invoke(ctx context.Context, prompt, sessionKey string) (string, error) {
	agentLoop, err := a.server.newAgentLoopForUser(a.userUUID, a.userStore)
	if err != nil {
		return "", fmt.Errorf("create agent loop: %w", err)
	}
	result, err := agentLoop.ProcessDirect(ctx, prompt, sessionKey)
	if err != nil {
		return "", fmt.Errorf("agent process: %w", err)
	}
	return result, nil
}
