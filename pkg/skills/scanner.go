package skills

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// SecuritySeverity represents the severity level of a security finding
type SecuritySeverity string

const (
	SeverityCritical SecuritySeverity = "critical" // Immediate rejection
	SeverityHigh     SecuritySeverity = "high"     // Manual review required
	SeverityMedium   SecuritySeverity = "medium"   // Warning, auto-approve possible
	SeverityLow      SecuritySeverity = "low"      // Informational
)

// SecurityFinding represents a security issue found in a skill
type SecurityFinding struct {
	ID          string           `json:"id"`
	Severity    SecuritySeverity `json:"severity"`
	Category    string           `json:"category"`
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Location    string           `json:"location"`
	Pattern     string           `json:"pattern"`
	Suggestion  string           `json:"suggestion"`
}

// SecurityScanResult contains the full scan results
type SecurityScanResult struct {
	Score        int               `json:"score"`
	Passed       bool              `json:"passed"`
	AutoApprove  bool              `json:"auto_approve"`
	Findings     []SecurityFinding `json:"findings"`
	Summary      string            `json:"summary"`
	ScanDuration int64             `json:"scan_duration_ms"`
}

// PatternRule defines a security pattern to match
type PatternRule struct {
	ID          string
	Pattern     *regexp.Regexp
	Severity    SecuritySeverity
	Category    string
	Title       string
	Description string
	Suggestion  string
}

// AISecurityReviewer interface for AI-assisted security review
type AISecurityReviewer interface {
	ReviewSkill(ctx context.Context, content string) ([]SecurityFinding, error)
}

// SkillSecurityScanner performs security analysis on skills
type SkillSecurityScanner struct {
	dangerousPatterns       []PatternRule
	promptInjectionPatterns []PatternRule
	credentialPatterns      []PatternRule
	urlPatterns             []PatternRule
	behaviorPatterns        []PatternRule
	aiReviewer              AISecurityReviewer
}

// NewSkillSecurityScanner creates a new security scanner
func NewSkillSecurityScanner(aiReviewer AISecurityReviewer) *SkillSecurityScanner {
	return &SkillSecurityScanner{
		dangerousPatterns:       DangerousCommandPatterns,
		promptInjectionPatterns: PromptInjectionPatterns,
		credentialPatterns:      CredentialPatterns,
		urlPatterns:             URLPatterns,
		behaviorPatterns:        BehaviorPatterns,
		aiReviewer:              aiReviewer,
	}
}

// ScanSkill performs a comprehensive security scan on skill content
func (s *SkillSecurityScanner) ScanSkill(ctx context.Context, content string) (*SecurityScanResult, error) {
	startTime := time.Now()
	findings := []SecurityFinding{}

	// 1. Static pattern analysis
	findings = append(findings, s.scanPatterns(content, s.dangerousPatterns)...)
	findings = append(findings, s.scanPatterns(content, s.promptInjectionPatterns)...)
	findings = append(findings, s.scanPatterns(content, s.credentialPatterns)...)
	findings = append(findings, s.scanPatterns(content, s.urlPatterns)...)
	findings = append(findings, s.scanPatterns(content, s.behaviorPatterns)...)

	// 2. Structural validation
	findings = append(findings, s.validateStructure(content)...)

	// 3. AI-assisted review (if available)
	if s.aiReviewer != nil {
		aiFindings, err := s.aiReviewer.ReviewSkill(ctx, content)
		if err == nil && len(aiFindings) > 0 {
			findings = append(findings, aiFindings...)
		}
	}

	// Calculate score
	score := s.calculateScore(findings)

	return &SecurityScanResult{
		Score:        score,
		Passed:       score >= 60 && !s.hasCriticalFindings(findings),
		AutoApprove:  score >= 80 && !s.hasHighOrCriticalFindings(findings),
		Findings:     findings,
		Summary:      s.generateSummary(findings, score),
		ScanDuration: time.Since(startTime).Milliseconds(),
	}, nil
}

// scanPatterns scans content against a set of pattern rules
func (s *SkillSecurityScanner) scanPatterns(content string, patterns []PatternRule) []SecurityFinding {
	findings := []SecurityFinding{}
	lines := strings.Split(content, "\n")

	for _, pattern := range patterns {
		for lineNum, line := range lines {
			if pattern.Pattern.MatchString(line) {
				match := pattern.Pattern.FindString(line)
				findings = append(findings, SecurityFinding{
					ID:          pattern.ID,
					Severity:    pattern.Severity,
					Category:    pattern.Category,
					Title:       pattern.Title,
					Description: pattern.Description,
					Location:    fmt.Sprintf("Line %d", lineNum+1),
					Pattern:     truncateString(match, 100),
					Suggestion:  pattern.Suggestion,
				})
			}
		}
	}

	return findings
}

// validateStructure checks for structural issues in the skill
func (s *SkillSecurityScanner) validateStructure(content string) []SecurityFinding {
	findings := []SecurityFinding{}

	// Check for missing frontmatter
	if !strings.HasPrefix(strings.TrimSpace(content), "---") {
		findings = append(findings, SecurityFinding{
			ID:          "STRUCT001",
			Severity:    SeverityMedium,
			Category:    "structure",
			Title:       "Missing frontmatter",
			Description: "Skill should have YAML frontmatter with name and description",
			Location:    "Line 1",
			Suggestion:  "Add frontmatter: ---\\nname: skill-name\\ndescription: ...\\n---",
		})
	}

	// Check for missing description in frontmatter
	if strings.Contains(content, "---") && !strings.Contains(content, "description:") {
		findings = append(findings, SecurityFinding{
			ID:          "STRUCT002",
			Severity:    SeverityMedium,
			Category:    "structure",
			Title:       "Missing description",
			Description: "Frontmatter should include a description field",
			Location:    "Frontmatter",
			Suggestion:  "Add 'description: your skill description' to frontmatter",
		})
	}

	// Check for excessively large content
	if len(content) > 100*1024 { // 100KB
		findings = append(findings, SecurityFinding{
			ID:          "STRUCT003",
			Severity:    SeverityMedium,
			Category:    "structure",
			Title:       "Content too large",
			Description: fmt.Sprintf("Skill content is %d bytes, recommended max is 100KB", len(content)),
			Location:    "Entire file",
			Suggestion:  "Reduce content size or split into multiple files in references/",
		})
	}

	// Check for missing H1 title
	if !regexp.MustCompile(`(?m)^#\s+.+`).MatchString(content) {
		findings = append(findings, SecurityFinding{
			ID:          "STRUCT004",
			Severity:    SeverityLow,
			Category:    "structure",
			Title:       "Missing H1 title",
			Description: "Skill should have a main title (# Title)",
			Location:    "Content",
			Suggestion:  "Add a main title after the frontmatter: # Skill Name",
		})
	}

	return findings
}

// calculateScore computes the security score based on findings
func (s *SkillSecurityScanner) calculateScore(findings []SecurityFinding) int {
	score := 100

	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			score -= 40
		case SeverityHigh:
			score -= 20
		case SeverityMedium:
			score -= 10
		case SeverityLow:
			score -= 2
		}
	}

	if score < 0 {
		score = 0
	}

	return score
}

// hasCriticalFindings checks if there are any critical severity findings
func (s *SkillSecurityScanner) hasCriticalFindings(findings []SecurityFinding) bool {
	for _, f := range findings {
		if f.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

// hasHighOrCriticalFindings checks for high or critical findings
func (s *SkillSecurityScanner) hasHighOrCriticalFindings(findings []SecurityFinding) bool {
	for _, f := range findings {
		if f.Severity == SeverityCritical || f.Severity == SeverityHigh {
			return true
		}
	}
	return false
}

// generateSummary creates a human-readable summary of the scan results
func (s *SkillSecurityScanner) generateSummary(findings []SecurityFinding, score int) string {
	if len(findings) == 0 {
		return "No security issues detected. Skill passed all checks."
	}

	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, f := range findings {
		switch f.Severity {
		case SeverityCritical:
			criticalCount++
		case SeverityHigh:
			highCount++
		case SeverityMedium:
			mediumCount++
		case SeverityLow:
			lowCount++
		}
	}

	var parts []string
	if criticalCount > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", criticalCount))
	}
	if highCount > 0 {
		parts = append(parts, fmt.Sprintf("%d high", highCount))
	}
	if mediumCount > 0 {
		parts = append(parts, fmt.Sprintf("%d medium", mediumCount))
	}
	if lowCount > 0 {
		parts = append(parts, fmt.Sprintf("%d low", lowCount))
	}

	return fmt.Sprintf("Found %d issues (%s). Score: %d/100", len(findings), strings.Join(parts, ", "), score)
}

// truncateString truncates a string to a maximum length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// LLMSecurityReviewer implements AISecurityReviewer using an LLM
type LLMSecurityReviewer struct {
	agentLoop interface {
		ProcessDirect(ctx context.Context, content, sessionKey string) (string, error)
	}
}

// NewLLMSecurityReviewer creates a new LLM-based security reviewer
func NewLLMSecurityReviewer(agentLoop interface {
	ProcessDirect(ctx context.Context, content, sessionKey string) (string, error)
}) *LLMSecurityReviewer {
	return &LLMSecurityReviewer{agentLoop: agentLoop}
}

// ReviewSkill uses an LLM to review the skill for security issues
func (r *LLMSecurityReviewer) ReviewSkill(ctx context.Context, content string) ([]SecurityFinding, error) {
	if r.agentLoop == nil {
		return nil, nil
	}

	prompt := fmt.Sprintf(`You are a security expert reviewing an AI agent skill for potential security issues.

Analyze this skill content and identify security concerns NOT covered by simple pattern matching:

1. Logical security issues (e.g., workflows that could be exploited)
2. Social engineering vectors (e.g., instructions that could manipulate users)
3. Privilege escalation paths
4. Data leakage risks
5. Dependency risks (e.g., suggesting unverified packages)
6. Subtle prompt injection attempts

SKILL CONTENT:
---
%s
---

Respond ONLY with a JSON object in this exact format (no markdown, no explanation):
{
  "findings": [
    {
      "severity": "critical|high|medium|low",
      "category": "category_name",
      "title": "brief title",
      "description": "detailed description",
      "suggestion": "how to fix"
    }
  ]
}

If no issues found, return: {"findings": []}
`, content)

	result, err := r.agentLoop.ProcessDirect(ctx, prompt, "security_review")
	if err != nil {
		return nil, err
	}

	// Try to extract JSON from the response (may have surrounding text)
	jsonStart := strings.Index(result, "{")
	jsonEnd := strings.LastIndex(result, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonEnd <= jsonStart {
		return nil, fmt.Errorf("no valid JSON in response")
	}

	jsonStr := result[jsonStart : jsonEnd+1]

	var response struct {
		Findings []struct {
			Severity    string `json:"severity"`
			Category    string `json:"category"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Suggestion  string `json:"suggestion"`
		} `json:"findings"`
	}

	if err := json.Unmarshal([]byte(jsonStr), &response); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	findings := make([]SecurityFinding, 0, len(response.Findings))
	for i, f := range response.Findings {
		severity := SeverityMedium
		switch strings.ToLower(f.Severity) {
		case "critical":
			severity = SeverityCritical
		case "high":
			severity = SeverityHigh
		case "medium":
			severity = SeverityMedium
		case "low":
			severity = SeverityLow
		}

		findings = append(findings, SecurityFinding{
			ID:          fmt.Sprintf("AI%03d", i+1),
			Severity:    severity,
			Category:    f.Category,
			Title:       f.Title,
			Description: f.Description,
			Location:    "AI Analysis",
			Suggestion:  f.Suggestion,
		})
	}

	return findings, nil
}
