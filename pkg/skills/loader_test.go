package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testSkillContent = `---
name: test-skill
description: A test skill
emoji: test
---

# Test Skill

This is the skill content.
`

const testSkillContentWithDeps = `---
name: dep-skill
description: Skill with dependencies
dependencies: dep1, dep2
---

# Dep Skill

Content here.
`

func createSkillDir(t *testing.T, baseDir, skillName, content string) {
	t.Helper()
	skillDir := filepath.Join(baseDir, skillName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("failed to create skill dir %s: %v", skillDir, err)
	}
	skillFile := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillFile, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write skill file %s: %v", skillFile, err)
	}
}

func TestNewSkillsLoader(t *testing.T) {
	sl := NewSkillsLoader("/tmp/ws", "/tmp/global", "/tmp/builtin")
	if sl == nil {
		t.Fatal("NewSkillsLoader returned nil")
	}
}

func TestListSkills_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	// Create the directories so ReadDir does not fail with "not exist"
	os.MkdirAll(filepath.Join(workspace, "skills"), 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	skills := sl.ListSkills()
	if len(skills) != 0 {
		t.Errorf("ListSkills() returned %d skills, want 0", len(skills))
	}
}

func TestListSkills_FindsSkills(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	createSkillDir(t, workspaceSkills, "my-skill", testSkillContent)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	skills := sl.ListSkills()

	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1", len(skills))
	}
	if skills[0].Name != "my-skill" {
		t.Errorf("skill Name = %q, want %q", skills[0].Name, "my-skill")
	}
	if skills[0].Source != "workspace" {
		t.Errorf("skill Source = %q, want %q", skills[0].Source, "workspace")
	}
	if skills[0].Description != "A test skill" {
		t.Errorf("skill Description = %q, want %q", skills[0].Description, "A test skill")
	}
}

func TestListSkills_Priority(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	// Create the same skill name in both workspace and global dirs
	workspaceContent := `---
name: shared-skill
description: Workspace version
---

Workspace content.
`
	globalContent := `---
name: shared-skill
description: Global version
---

Global content.
`
	createSkillDir(t, workspaceSkills, "shared-skill", workspaceContent)
	createSkillDir(t, globalDir, "shared-skill", globalContent)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	skills := sl.ListSkills()

	if len(skills) != 1 {
		t.Fatalf("ListSkills() returned %d skills, want 1 (workspace should shadow global)", len(skills))
	}
	if skills[0].Source != "workspace" {
		t.Errorf("skill Source = %q, want %q (workspace should win over global)", skills[0].Source, "workspace")
	}
	if skills[0].Description != "Workspace version" {
		t.Errorf("skill Description = %q, want %q", skills[0].Description, "Workspace version")
	}
}

func TestListSkills_AllSourcePriority(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	// Skill only in builtin
	createSkillDir(t, builtinDir, "builtin-only", `---
name: builtin-only
description: Builtin skill
---

Builtin content.
`)

	// Skill only in global
	createSkillDir(t, globalDir, "global-only", `---
name: global-only
description: Global skill
---

Global content.
`)

	// Skill only in workspace
	createSkillDir(t, workspaceSkills, "ws-only", `---
name: ws-only
description: Workspace skill
---

Workspace content.
`)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	skills := sl.ListSkills()

	if len(skills) != 3 {
		t.Fatalf("ListSkills() returned %d skills, want 3", len(skills))
	}

	sourceMap := make(map[string]string)
	for _, s := range skills {
		sourceMap[s.Name] = s.Source
	}

	if sourceMap["ws-only"] != "workspace" {
		t.Errorf("ws-only source = %q, want %q", sourceMap["ws-only"], "workspace")
	}
	if sourceMap["global-only"] != "global" {
		t.Errorf("global-only source = %q, want %q", sourceMap["global-only"], "global")
	}
	if sourceMap["builtin-only"] != "builtin" {
		t.Errorf("builtin-only source = %q, want %q", sourceMap["builtin-only"], "builtin")
	}
}

func TestLoadSkill(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	createSkillDir(t, workspaceSkills, "my-skill", testSkillContent)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	content, ok := sl.LoadSkill("my-skill")
	if !ok {
		t.Fatal("LoadSkill(my-skill) returned false, want true")
	}

	// The content should include the body of the skill
	if !strings.Contains(content, "# Test Skill") {
		t.Error("LoadSkill content does not contain expected heading")
	}
	if !strings.Contains(content, "This is the skill content.") {
		t.Error("LoadSkill content does not contain expected body text")
	}
	if content == "" {
		t.Error("LoadSkill content is empty")
	}
}

func TestLoadSkill_SingleLineFrontmatterStripped(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	// Single-line frontmatter (name only on one line between --- delimiters)
	singleLineFM := "---\nname: simple\n---\n\n# Simple Skill\n\nBody here.\n"
	createSkillDir(t, workspaceSkills, "simple-skill", singleLineFM)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	content, ok := sl.LoadSkill("simple-skill")
	if !ok {
		t.Fatal("LoadSkill returned false, want true")
	}
	if !strings.Contains(content, "# Simple Skill") {
		t.Error("content does not contain expected heading")
	}
	if !strings.Contains(content, "Body here.") {
		t.Error("content does not contain expected body")
	}
}

func TestLoadSkill_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(filepath.Join(workspace, "skills"), 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	content, ok := sl.LoadSkill("nonexistent")
	if ok {
		t.Error("LoadSkill(nonexistent) returned true, want false")
	}
	if content != "" {
		t.Errorf("LoadSkill(nonexistent) content = %q, want empty", content)
	}
}

func TestLoadSkill_Priority(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	// Same skill in workspace and global, workspace should win
	createSkillDir(t, workspaceSkills, "prio-skill", `---
name: prio-skill
description: workspace version
---

Workspace body.
`)
	createSkillDir(t, globalDir, "prio-skill", `---
name: prio-skill
description: global version
---

Global body.
`)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	content, ok := sl.LoadSkill("prio-skill")
	if !ok {
		t.Fatal("LoadSkill(prio-skill) returned false, want true")
	}
	if !strings.Contains(content, "Workspace body.") {
		t.Error("LoadSkill did not return workspace version content")
	}
	if strings.Contains(content, "Global body.") {
		t.Error("LoadSkill returned global version content instead of workspace version")
	}
}

func TestGetSkillDependencies(t *testing.T) {
	sl := NewSkillsLoader("", "", "")

	deps := sl.GetSkillDependencies(testSkillContentWithDeps)
	if len(deps) != 2 {
		t.Fatalf("GetSkillDependencies() returned %d deps, want 2", len(deps))
	}
	if deps[0] != "dep1" {
		t.Errorf("deps[0] = %q, want %q", deps[0], "dep1")
	}
	if deps[1] != "dep2" {
		t.Errorf("deps[1] = %q, want %q", deps[1], "dep2")
	}
}

func TestGetSkillDependencies_None(t *testing.T) {
	sl := NewSkillsLoader("", "", "")

	deps := sl.GetSkillDependencies(testSkillContent)
	if deps != nil {
		t.Errorf("GetSkillDependencies() = %v, want nil (no dependencies)", deps)
	}
}

func TestGetSkillDependencies_NoFrontmatter(t *testing.T) {
	sl := NewSkillsLoader("", "", "")

	deps := sl.GetSkillDependencies("# Just a heading\n\nNo frontmatter here.")
	if deps != nil {
		t.Errorf("GetSkillDependencies() = %v, want nil (no frontmatter)", deps)
	}
}

func TestBuildSkillsSummary(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	workspaceSkills := filepath.Join(workspace, "skills")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(workspaceSkills, 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	createSkillDir(t, workspaceSkills, "summary-skill", testSkillContent)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	summary := sl.BuildSkillsSummary()

	if summary == "" {
		t.Fatal("BuildSkillsSummary() returned empty string, want XML output")
	}
	if !strings.Contains(summary, "<skills>") {
		t.Error("summary missing <skills> opening tag")
	}
	if !strings.Contains(summary, "</skills>") {
		t.Error("summary missing </skills> closing tag")
	}
	if !strings.Contains(summary, "<name>summary-skill</name>") {
		t.Error("summary missing skill name")
	}
	if !strings.Contains(summary, "<description>A test skill</description>") {
		t.Error("summary missing skill description")
	}
	if !strings.Contains(summary, "<source>workspace</source>") {
		t.Error("summary missing skill source")
	}
}

func TestBuildSkillsSummary_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	workspace := filepath.Join(tmpDir, "workspace")
	globalDir := filepath.Join(tmpDir, "global")
	builtinDir := filepath.Join(tmpDir, "builtin")

	os.MkdirAll(filepath.Join(workspace, "skills"), 0o755)
	os.MkdirAll(globalDir, 0o755)
	os.MkdirAll(builtinDir, 0o755)

	sl := NewSkillsLoader(workspace, globalDir, builtinDir)
	summary := sl.BuildSkillsSummary()

	if summary != "" {
		t.Errorf("BuildSkillsSummary() with no skills = %q, want empty", summary)
	}
}

func TestSlugifySkillName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"simple", "simple"},
		{"My Skill", "my-skill"},
		{"skill_with_underscores", "skill-with-underscores"},
		{"UPPERCASE", "uppercase"},
		{"--trim-dashes--", "trim-dashes"},
		{"special!@#chars", "special-chars"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Slugify(tt.name)
			if got != tt.want {
				t.Errorf("Slugify(%q) = %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
