package agent

import (
	"sync"
	"time"
)

// CostMetrics tracks token and API usage for agents
type CostMetrics struct {
	TotalTokensUsed     int64
	InputTokensUsed     int64
	OutputTokensUsed    int64
	APICallsCount       int64
	EstimatedCostUSD    float64
	LastUpdated         time.Time
	AvgTokensPerCall    float64
	EstimatedCallsToday int64
	EstimatedCostToday  float64
}

// AgentCostTracker tracks costs for all agents and specialists
type AgentCostTracker struct {
	agentMetrics   map[string]*CostMetrics
	mu             sync.RWMutex
	pricePerKToken map[string]float64 // Model -> price per 1K tokens
}

// NewAgentCostTracker creates a new cost tracker
func NewAgentCostTracker() *AgentCostTracker {
	return &AgentCostTracker{
		agentMetrics: make(map[string]*CostMetrics),
		pricePerKToken: map[string]float64{
			// OpenAI models
			"gpt-4":   0.03,
			"gpt-3.5": 0.0005,
			// Anthropic models
			"claude-opus":   0.015,
			"claude-sonnet": 0.003,
			"claude-haiku":  0.00025,
			// Open source (typically free)
			"ollama": 0.0,
			"vllm":   0.0,
		},
	}
}

// RecordAPICall records an API call and token usage for an agent
func (act *AgentCostTracker) RecordAPICall(
	agentName string,
	inputTokens int64,
	outputTokens int64,
	model string,
) {
	act.mu.Lock()
	defer act.mu.Unlock()

	// Get or create metrics
	if _, exists := act.agentMetrics[agentName]; !exists {
		act.agentMetrics[agentName] = &CostMetrics{}
	}

	metrics := act.agentMetrics[agentName]
	totalTokens := inputTokens + outputTokens

	// Update counters
	metrics.TotalTokensUsed += totalTokens
	metrics.InputTokensUsed += inputTokens
	metrics.OutputTokensUsed += outputTokens
	metrics.APICallsCount++
	metrics.LastUpdated = time.Now()

	// Calculate average
	if metrics.APICallsCount > 0 {
		metrics.AvgTokensPerCall = float64(metrics.TotalTokensUsed) / float64(metrics.APICallsCount)
	}

	// Estimate cost
	pricePerKToken, exists := act.pricePerKToken[model]
	if exists {
		costForThisCall := float64(totalTokens) / 1000.0 * pricePerKToken
		metrics.EstimatedCostUSD += costForThisCall
	}
}

// GetMetrics returns metrics for an agent
func (act *AgentCostTracker) GetMetrics(agentName string) *CostMetrics {
	act.mu.RLock()
	defer act.mu.RUnlock()

	if metrics, exists := act.agentMetrics[agentName]; exists {
		// Return a copy to avoid external modifications
		copy := *metrics
		return &copy
	}

	return &CostMetrics{}
}

// GetAllMetrics returns metrics for all agents
func (act *AgentCostTracker) GetAllMetrics() map[string]*CostMetrics {
	act.mu.RLock()
	defer act.mu.RUnlock()

	result := make(map[string]*CostMetrics, len(act.agentMetrics))
	for agentName, metrics := range act.agentMetrics {
		copy := *metrics
		result[agentName] = &copy
	}
	return result
}

// GetTotalCost returns total estimated cost across all agents
func (act *AgentCostTracker) GetTotalCost() float64 {
	act.mu.RLock()
	defer act.mu.RUnlock()

	total := 0.0
	for _, metrics := range act.agentMetrics {
		total += metrics.EstimatedCostUSD
	}
	return total
}

// GetCostSummary returns a summary of costs by agent
type CostSummary struct {
	TotalCost        float64            `json:"total_cost"`
	AgentCosts       map[string]float64 `json:"agent_costs"`
	AgentTokens      map[string]int64   `json:"agent_tokens"`
	MostExpensive    string             `json:"most_expensive"`
	MostUsed         string             `json:"most_used"`
	CheapestPerToken map[string]float64 `json:"cheapest_per_token"`
}

func (act *AgentCostTracker) GetCostSummary() *CostSummary {
	act.mu.RLock()
	defer act.mu.RUnlock()

	summary := &CostSummary{
		AgentCosts:       make(map[string]float64),
		AgentTokens:      make(map[string]int64),
		CheapestPerToken: make(map[string]float64),
		TotalCost:        0,
	}

	var maxCost float64
	var maxTokens int64
	var mostExpensiveAgent string
	var mostUsedAgent string

	for agentName, metrics := range act.agentMetrics {
		summary.AgentCosts[agentName] = metrics.EstimatedCostUSD
		summary.AgentTokens[agentName] = metrics.TotalTokensUsed
		summary.TotalCost += metrics.EstimatedCostUSD

		// Calculate cost per token
		if metrics.TotalTokensUsed > 0 {
			costPerToken := metrics.EstimatedCostUSD / float64(metrics.TotalTokensUsed)
			summary.CheapestPerToken[agentName] = costPerToken
		}

		// Track most expensive and most used
		if metrics.EstimatedCostUSD > maxCost {
			maxCost = metrics.EstimatedCostUSD
			mostExpensiveAgent = agentName
		}
		if metrics.TotalTokensUsed > maxTokens {
			maxTokens = metrics.TotalTokensUsed
			mostUsedAgent = agentName
		}
	}

	summary.MostExpensive = mostExpensiveAgent
	summary.MostUsed = mostUsedAgent

	return summary
}

// Reset clears all metrics (for testing or daily reset)
func (act *AgentCostTracker) Reset() {
	act.mu.Lock()
	defer act.mu.Unlock()

	act.agentMetrics = make(map[string]*CostMetrics)
}

// ResetAgent clears metrics for a specific agent
func (act *AgentCostTracker) ResetAgent(agentName string) {
	act.mu.Lock()
	defer act.mu.Unlock()

	delete(act.agentMetrics, agentName)
}
