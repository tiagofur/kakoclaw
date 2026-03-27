package skills

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/sipeed/makoclaw/pkg/config"
)

// LoadAgentsFromSkillDir reads agents.json from a skill directory.
// Returns an empty map (not an error) if the file does not exist.
func LoadAgentsFromSkillDir(skillDir string) (map[string]config.SpecialistConfig, error) {
	agentsPath := filepath.Join(skillDir, "agents.json")
	data, err := os.ReadFile(agentsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]config.SpecialistConfig{}, nil
		}
		return nil, err
	}
	var agents map[string]config.SpecialistConfig
	if err := json.Unmarshal(data, &agents); err != nil {
		return nil, err
	}
	if agents == nil {
		return map[string]config.SpecialistConfig{}, nil
	}
	return agents, nil
}
