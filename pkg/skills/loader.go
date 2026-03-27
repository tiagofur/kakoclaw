package skills

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type SkillMetadata struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Dependencies []string `json:"dependencies,omitempty"`
}

type SkillInfo struct {
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Path        string `json:"path"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

// Slugify converts a name to a URL-safe slug: lowercase, non-alphanumeric
// characters replaced with hyphens, trimmed, max 64 chars.
func Slugify(name string) string {
	name = strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(name, "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 64 {
		slug = slug[:64]
	}
	return slug
}

type SkillsLoader struct {
	workspace       string // user workspace, if multiuser-enabled
	workspaceSkills string // workspace skills (项目级别)
	userSkillsPath  string // user-specific skills (~/.makoclaw/users/<uuid>/skills)
	globalSkills    string // 全局 skills (~/.makoclaw/skills)
	builtinSkills   string // 内置 skills
}

func NewSkillsLoader(workspace string, globalSkills string, builtinSkills string) *SkillsLoader {
	return &SkillsLoader{
		workspace:       workspace,
		workspaceSkills: filepath.Join(workspace, "skills"),
		userSkillsPath:  "",           // Will be set via SetUserSkillsPath if needed
		globalSkills:    globalSkills, // ~/.makoclaw/skills
		builtinSkills:   builtinSkills,
	}
}

// SetUserSkillsPath sets the user-specific skills directory for multiuser support.
// This should be ~/.makoclaw/users/<userUUID>/skills
func (sl *SkillsLoader) SetUserSkillsPath(path string) {
	sl.userSkillsPath = path
}

// scanSkillsFlat scans a directory one level deep for skills (dirs containing SKILL.md).
func (sl *SkillsLoader) scanSkillsFlat(dir, source string) []SkillInfo {
	var result []SkillInfo
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return result
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		skillFile := filepath.Join(dir, d.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err != nil {
			continue
		}
		info := SkillInfo{
			Name:   d.Name(),
			Slug:   Slugify(d.Name()),
			Path:   skillFile,
			Source: source,
		}
		if metadata := sl.getSkillMetadata(skillFile); metadata != nil {
			info.Description = metadata.Description
		}
		result = append(result, info)
	}
	return result
}

// scanSkillsNested scans a directory for skills, supporting one level of category
// subdirectories. This handles both flat skills (github/SKILL.md) and categorized
// skills (development/code-review-checklist/SKILL.md).
func (sl *SkillsLoader) scanSkillsNested(dir, source string) []SkillInfo {
	var result []SkillInfo
	dirs, err := os.ReadDir(dir)
	if err != nil {
		return result
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		skillFile := filepath.Join(dir, d.Name(), "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			// Flat skill: dir/skillName/SKILL.md
			info := SkillInfo{
				Name:   d.Name(),
				Slug:   Slugify(d.Name()),
				Path:   skillFile,
				Source: source,
			}
			if metadata := sl.getSkillMetadata(skillFile); metadata != nil {
				info.Description = metadata.Description
			}
			result = append(result, info)
			continue
		}
		// No SKILL.md here — check one level deeper (category subdirectory)
		subDirs, err := os.ReadDir(filepath.Join(dir, d.Name()))
		if err != nil {
			continue
		}
		for _, sub := range subDirs {
			if !sub.IsDir() {
				continue
			}
			subSkillFile := filepath.Join(dir, d.Name(), sub.Name(), "SKILL.md")
			if _, err := os.Stat(subSkillFile); err != nil {
				continue
			}
			info := SkillInfo{
				Name:   sub.Name(),
				Slug:   Slugify(sub.Name()),
				Path:   subSkillFile,
				Source: source,
			}
			if metadata := sl.getSkillMetadata(subSkillFile); metadata != nil {
				info.Description = metadata.Description
			}
			result = append(result, info)
		}
	}
	return result
}

// hasSkillName checks if a skill with the given name already exists in the list.
func hasSkillName(skills []SkillInfo, name string) bool {
	for _, s := range skills {
		if s.Name == name {
			return true
		}
	}
	return false
}

// ListSkills returns all skills from all tiers merged by priority.
// Used by the agent context (BuildSkillsSummary) to load all available skills.
func (sl *SkillsLoader) ListSkills() []SkillInfo {
	skills := make([]SkillInfo, 0)

	// 1. User-specific skills - highest priority
	if sl.userSkillsPath != "" {
		skills = append(skills, sl.scanSkillsFlat(sl.userSkillsPath, "user")...)
	}

	// 2. Workspace skills - skip if already in user tier
	if sl.workspaceSkills != "" {
		for _, s := range sl.scanSkillsFlat(sl.workspaceSkills, "workspace") {
			if !hasSkillName(skills, s.Name) {
				skills = append(skills, s)
			}
		}
	}

	// 3. Global skills - skip if already in higher tiers
	if sl.globalSkills != "" {
		for _, s := range sl.scanSkillsFlat(sl.globalSkills, "global") {
			if !hasSkillName(skills, s.Name) {
				skills = append(skills, s)
			}
		}
	}

	// 4. Builtin skills (with nested category support) - lowest priority
	if sl.builtinSkills != "" {
		for _, s := range sl.scanSkillsNested(sl.builtinSkills, "builtin") {
			if !hasSkillName(skills, s.Name) {
				skills = append(skills, s)
			}
		}
	}

	return skills
}

// ListUserSkills returns ONLY user-installed skills (user-specific + workspace tiers).
// Used by the web UI "Installed" tab to show only what the user explicitly installed.
func (sl *SkillsLoader) ListUserSkills() []SkillInfo {
	skills := make([]SkillInfo, 0)

	if sl.userSkillsPath != "" {
		skills = append(skills, sl.scanSkillsFlat(sl.userSkillsPath, "user")...)
	}

	if sl.workspaceSkills != "" {
		for _, s := range sl.scanSkillsFlat(sl.workspaceSkills, "workspace") {
			if !hasSkillName(skills, s.Name) {
				skills = append(skills, s)
			}
		}
	}

	return skills
}

// ListGlobalSkills returns ONLY global skills (~/.MakoClaw/skills/).
// Used by the marketplace API to include admin-placed global skills.
func (sl *SkillsLoader) ListGlobalSkills() []SkillInfo {
	if sl.globalSkills == "" {
		return nil
	}
	return sl.scanSkillsFlat(sl.globalSkills, "global")
}

// ListBuiltinSkills returns ONLY built-in skills (with nested category support).
// Used by the marketplace API to include built-in skills in the response.
func (sl *SkillsLoader) ListBuiltinSkills() []SkillInfo {
	if sl.builtinSkills == "" {
		return nil
	}
	return sl.scanSkillsNested(sl.builtinSkills, "builtin")
}

func (sl *SkillsLoader) LoadSkill(name string) (string, bool) {
	// 1. User-specific skills (~/.makoclaw/users/<uuid>/skills) - highest priority
	if sl.userSkillsPath != "" {
		skillFile := filepath.Join(sl.userSkillsPath, name, "SKILL.md")
		if content, err := os.ReadFile(skillFile); err == nil {
			return sl.stripFrontmatter(string(content)), true
		}
	}

	// 2. Workspace skills
	if sl.workspaceSkills != "" {
		skillFile := filepath.Join(sl.workspaceSkills, name, "SKILL.md")
		if content, err := os.ReadFile(skillFile); err == nil {
			return sl.stripFrontmatter(string(content)), true
		}
	}

	// 3. Global skills (~/.makoclaw/skills)
	if sl.globalSkills != "" {
		skillFile := filepath.Join(sl.globalSkills, name, "SKILL.md")
		if content, err := os.ReadFile(skillFile); err == nil {
			return sl.stripFrontmatter(string(content)), true
		}
	}

	// 4. Builtin skills (flat first, then nested categories)
	if sl.builtinSkills != "" {
		skillFile := filepath.Join(sl.builtinSkills, name, "SKILL.md")
		if content, err := os.ReadFile(skillFile); err == nil {
			return sl.stripFrontmatter(string(content)), true
		}
		// Search one level deeper: builtinSkills/*/name/SKILL.md
		if dirs, err := os.ReadDir(sl.builtinSkills); err == nil {
			for _, d := range dirs {
				if !d.IsDir() {
					continue
				}
				nestedFile := filepath.Join(sl.builtinSkills, d.Name(), name, "SKILL.md")
				if content, err := os.ReadFile(nestedFile); err == nil {
					return sl.stripFrontmatter(string(content)), true
				}
			}
		}
	}

	return "", false
}

func (sl *SkillsLoader) LoadSkillsForContext(skillNames []string) string {
	if len(skillNames) == 0 {
		return ""
	}

	var parts []string
	for _, name := range skillNames {
		content, ok := sl.LoadSkill(name)
		if ok {
			parts = append(parts, fmt.Sprintf("### Skill: %s\n\n%s", name, content))
		}
	}

	return strings.Join(parts, "\n\n---\n\n")
}

func (sl *SkillsLoader) BuildSkillsSummary() string {
	allSkills := sl.ListSkills()
	if len(allSkills) == 0 {
		return ""
	}

	var lines []string
	lines = append(lines, "<skills>")
	for _, s := range allSkills {
		escapedName := escapeXML(s.Name)
		escapedDesc := escapeXML(s.Description)
		escapedPath := escapeXML(s.Path)

		lines = append(lines, fmt.Sprintf("  <skill>"))
		lines = append(lines, fmt.Sprintf("    <name>%s</name>", escapedName))
		lines = append(lines, fmt.Sprintf("    <description>%s</description>", escapedDesc))
		lines = append(lines, fmt.Sprintf("    <location>%s</location>", escapedPath))
		lines = append(lines, fmt.Sprintf("    <source>%s</source>", s.Source))
		lines = append(lines, "  </skill>")
	}
	lines = append(lines, "</skills>")

	return strings.Join(lines, "\n")
}

// BuildSkillsSummaryForNames builds an XML summary for only the named skills.
// Returns empty string if names is empty or no matching skills are found.
func (sl *SkillsLoader) BuildSkillsSummaryForNames(names []string) string {
	if len(names) == 0 {
		return ""
	}

	nameSet := make(map[string]bool, len(names))
	for _, n := range names {
		nameSet[n] = true
	}

	allSkills := sl.ListSkills()
	if len(allSkills) == 0 {
		return ""
	}

	var lines []string
	matched := false
	lines = append(lines, "<skills>")
	for _, s := range allSkills {
		if !nameSet[s.Name] {
			continue
		}
		matched = true
		escapedName := escapeXML(s.Name)
		escapedDesc := escapeXML(s.Description)
		escapedPath := escapeXML(s.Path)

		lines = append(lines, "  <skill>")
		lines = append(lines, fmt.Sprintf("    <name>%s</name>", escapedName))
		lines = append(lines, fmt.Sprintf("    <description>%s</description>", escapedDesc))
		lines = append(lines, fmt.Sprintf("    <location>%s</location>", escapedPath))
		lines = append(lines, fmt.Sprintf("    <source>%s</source>", s.Source))
		lines = append(lines, "  </skill>")
	}
	lines = append(lines, "</skills>")

	if !matched {
		return ""
	}

	return strings.Join(lines, "\n")
}

func (sl *SkillsLoader) getSkillMetadata(skillPath string) *SkillMetadata {
	content, err := os.ReadFile(skillPath)
	if err != nil {
		return nil
	}

	frontmatter := sl.extractFrontmatter(string(content))
	if frontmatter == "" {
		return &SkillMetadata{
			Name: filepath.Base(filepath.Dir(skillPath)),
		}
	}

	// Try JSON first (for backward compatibility)
	var jsonMeta struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal([]byte(frontmatter), &jsonMeta); err == nil {
		return &SkillMetadata{
			Name:        jsonMeta.Name,
			Description: jsonMeta.Description,
		}
	}

	// Fall back to simple YAML parsing
	yamlMeta := sl.parseSimpleYAML(frontmatter)
	meta := &SkillMetadata{
		Name:        yamlMeta["name"],
		Description: yamlMeta["description"],
	}
	if depsStr, ok := yamlMeta["dependencies"]; ok && depsStr != "" {
		parts := strings.Split(depsStr, ",")
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p != "" {
				meta.Dependencies = append(meta.Dependencies, p)
			}
		}
	}
	return meta
}

// GetSkillDependencies extracts the dependencies from SKILL.md content without
// loading a full skill. It returns the slice of dependency slugs (or nil if none).
func (sl *SkillsLoader) GetSkillDependencies(content string) []string {
	frontmatter := sl.extractFrontmatter(content)
	if frontmatter == "" {
		return nil
	}
	yamlMeta := sl.parseSimpleYAML(frontmatter)
	depsStr, ok := yamlMeta["dependencies"]
	if !ok || depsStr == "" {
		return nil
	}
	var deps []string
	for _, p := range strings.Split(depsStr, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			deps = append(deps, p)
		}
	}
	return deps
}

// parseSimpleYAML parses simple key: value YAML format
// Example: name: github\n description: "..."
func (sl *SkillsLoader) parseSimpleYAML(content string) map[string]string {
	result := make(map[string]string)

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			value = strings.Trim(value, "\"'")
			result[key] = value
		}
	}

	return result
}

func (sl *SkillsLoader) extractFrontmatter(content string) string {
	// (?s) enables DOTALL mode so . matches newlines
	// Match first ---, capture everything until next --- on its own line
	re := regexp.MustCompile(`(?s)^---\n(.*)\n---`)
	match := re.FindStringSubmatch(content)
	if len(match) > 1 {
		return match[1]
	}
	return ""
}

func (sl *SkillsLoader) stripFrontmatter(content string) string {
	re := regexp.MustCompile(`^---\n.*?\n---\n`)
	return re.ReplaceAllString(content, "")
}

func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}
