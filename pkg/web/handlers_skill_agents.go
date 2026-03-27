package web

import (
	"path/filepath"

	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/logger"
	"github.com/sipeed/makoclaw/pkg/skills"
)

// applySkillAgents reads agents.json from the installed skill directory and
// registers any specialists not already present in the user's config.
// Returns the number of specialists registered.
func (s *Server) applySkillAgents(userUUID, skillSlug, userSkillsDir string) int {
	skillDir := filepath.Join(userSkillsDir, skillSlug)
	agents, err := skills.LoadAgentsFromSkillDir(skillDir)
	if err != nil {
		logger.WarnCF("web", "failed to load agents.json for skill", map[string]any{
			"skill": skillSlug,
			"error": err.Error(),
		})
		return 0
	}
	if len(agents) == 0 {
		return 0
	}

	cfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		logger.WarnCF("web", "applySkillAgents: failed to load config for user", map[string]any{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		return 0
	}
	if cfg.Agents.Specialists == nil {
		cfg.Agents.Specialists = make(map[string]config.SpecialistConfig)
	}

	registered := 0
	for name, spec := range agents {
		if _, exists := cfg.Agents.Specialists[name]; exists {
			continue // user's existing specialist wins
		}
		spec.SourceSkill = skillSlug
		cfg.Agents.Specialists[name] = spec
		// Remove from RemovedSpecialists if present (re-install case)
		cfg.Agents.RemovedSpecialists = removeFromSlice(cfg.Agents.RemovedSpecialists, name)
		registered++
	}

	if registered == 0 {
		return 0
	}

	if err := config.SaveConfigForUser(userUUID, cfg); err != nil {
		logger.WarnCF("web", "applySkillAgents: failed to save config for user", map[string]any{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		return 0
	}

	logger.InfoCF("web", "registered agents from skill for user", map[string]any{
		"count":     registered,
		"skill":     skillSlug,
		"user_uuid": userUUID,
	})
	return registered
}

// removeSkillAgents removes specialists that were registered by a skill install.
// Must be called BEFORE the skill directory is deleted.
func (s *Server) removeSkillAgents(userUUID, skillSlug, userSkillsDir string) {
	skillDir := filepath.Join(userSkillsDir, skillSlug)
	agents, err := skills.LoadAgentsFromSkillDir(skillDir)
	if err != nil || len(agents) == 0 {
		return
	}

	cfg, err := config.LoadConfigForUser(userUUID)
	if err != nil {
		logger.WarnCF("web", "removeSkillAgents: failed to load config for user", map[string]any{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
		return
	}

	removed := 0
	for name := range agents {
		spec, exists := cfg.Agents.Specialists[name]
		if !exists || spec.SourceSkill != skillSlug {
			continue // user-created or from another skill — don't touch
		}
		delete(cfg.Agents.Specialists, name)
		if !containsStr(cfg.Agents.RemovedSpecialists, name) {
			cfg.Agents.RemovedSpecialists = append(cfg.Agents.RemovedSpecialists, name)
		}
		removed++
	}

	if removed == 0 {
		return
	}

	if err := config.SaveConfigForUser(userUUID, cfg); err != nil {
		logger.WarnCF("web", "removeSkillAgents: failed to save config for user", map[string]any{
			"user_uuid": userUUID,
			"error":     err.Error(),
		})
	}
	logger.InfoCF("web", "removed agents from skill for user", map[string]any{
		"count":     removed,
		"skill":     skillSlug,
		"user_uuid": userUUID,
	})
}

func removeFromSlice(s []string, item string) []string {
	result := s[:0]
	for _, v := range s {
		if v != item {
			result = append(result, v)
		}
	}
	return result
}

func containsStr(s []string, item string) bool {
	for _, v := range s {
		if v == item {
			return true
		}
	}
	return false
}
