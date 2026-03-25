package agent

import (
	"context"
	"testing"

	"github.com/sipeed/makoclaw/pkg/config"
)

func TestSpawnForSpecialistIncludesUserIdentity(t *testing.T) {
	registry := NewSpecialistRegistry()
	specialist, err := NewSpecialistAgent("developer", &config.SpecialistConfig{
		Name:        "developer",
		Description: "writes code",
	}, newAgentTestConfig(t.TempDir()), nil, &staticProvider{response: "done"}, nil)
	if err != nil {
		t.Fatalf("NewSpecialistAgent failed: %v", err)
	}
	if err := registry.RegisterSpecialist(specialist); err != nil {
		t.Fatalf("RegisterSpecialist failed: %v", err)
	}

	spawner := NewSpecialistSpawner(registry, nil)
	ctx := ContextWithTeamContext(context.Background(), &TeamContext{
		UserUUID: "aaa-111",
		UserID:   88,
	})

	if _, err := spawner.SpawnForSpecialist(ctx, "developer", "implement this", "Developer", "web", "chat-1"); err != nil {
		t.Fatalf("SpawnForSpecialist failed: %v", err)
	}

	tasks := spawner.ListTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected one spawn task, got %d", len(tasks))
	}
	if tasks[0].UserUUID != "aaa-111" || tasks[0].UserID != 88 {
		t.Fatalf("expected spawn task user identity aaa-111/88, got %q/%d", tasks[0].UserUUID, tasks[0].UserID)
	}
	if specialist.userUUID != "aaa-111" || specialist.userID != 88 {
		t.Fatalf("expected specialist to inherit user identity aaa-111/88, got %q/%d", specialist.userUUID, specialist.userID)
	}
}
