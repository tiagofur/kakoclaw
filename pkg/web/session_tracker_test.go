package web

import (
	"sync"
	"testing"
)

func TestSessionTrackerRecordAccumulatesAndIsolatesSessions(t *testing.T) {
	tracker := NewSessionTracker(t.TempDir(), 1000)

	tracker.Record("user-1", "session-a", 500, 0.01)
	tracker.Record("user-1", "session-a", 300, 0.02)
	tracker.Record("user-1", "session-b", 200, 0.03)

	statsA := tracker.Stats("user-1", "session-a")
	if statsA.SessionID != "session-a" {
		t.Fatalf("expected session-a, got %q", statsA.SessionID)
	}
	if statsA.TokensUsed != 800 {
		t.Fatalf("expected 800 tokens, got %d", statsA.TokensUsed)
	}
	if statsA.NumTurns != 800 {
		t.Fatalf("expected 800 turns, got %d", statsA.NumTurns)
	}
	if statsA.CostUSD != 0.03 {
		t.Fatalf("expected 0.03 cost, got %v", statsA.CostUSD)
	}

	statsB := tracker.Stats("user-1", "session-b")
	if statsB.TokensUsed != 200 {
		t.Fatalf("expected isolated session tokens to remain 200, got %d", statsB.TokensUsed)
	}
}

func TestSessionTrackerShouldResetUsesStrictGreaterThan(t *testing.T) {
	tracker := NewSessionTracker(t.TempDir(), 1000)

	tracker.Record("user-1", "session-a", 900, 0)
	tracker.Record("user-1", "session-a", 100, 0)
	if tracker.ShouldReset("user-1", "session-a") {
		t.Fatal("expected exact limit to avoid reset")
	}

	tracker.Record("user-1", "session-a", 1, 0)
	if !tracker.ShouldReset("user-1", "session-a") {
		t.Fatal("expected reset after exceeding limit")
	}
}

func TestSessionTrackerZeroSessionResetsNewActiveSession(t *testing.T) {
	tracker := NewSessionTracker(t.TempDir(), 1000)

	tracker.Record("user-1", "session-a", 1050, 0.05)
	tracker.ZeroSession("user-1", "session-b")

	stats := tracker.Stats("user-1", "session-b")
	if stats.SessionID != "session-b" {
		t.Fatalf("expected new session id, got %q", stats.SessionID)
	}
	if stats.TokensUsed != 0 || stats.CostUSD != 0 || stats.NumTurns != 0 {
		t.Fatalf("expected zeroed stats, got %+v", stats)
	}

	activeStats := tracker.Stats("user-1", "")
	if activeStats.SessionID != "session-b" {
		t.Fatalf("expected active session to switch to session-b, got %q", activeStats.SessionID)
	}
}

func TestSessionTrackerPersistsAndReloadsState(t *testing.T) {
	stateDir := t.TempDir()
	tracker := NewSessionTracker(stateDir, 1000)
	tracker.Record("user-1", "session-a", 800, 0.015)

	reloaded := NewSessionTracker(stateDir, 1000)
	stats := reloaded.Stats("user-1", "session-a")
	if stats.TokensUsed != 800 {
		t.Fatalf("expected 800 persisted tokens, got %d", stats.TokensUsed)
	}
	if stats.CostUSD != 0.015 {
		t.Fatalf("expected 0.015 persisted cost, got %v", stats.CostUSD)
	}

	activeStats := reloaded.Stats("user-1", "")
	if activeStats.SessionID != "session-a" {
		t.Fatalf("expected persisted active session, got %q", activeStats.SessionID)
	}
}

func TestSessionTrackerStartsEmptyWhenStateFileMissing(t *testing.T) {
	tracker := NewSessionTracker(t.TempDir(), 1000)

	stats := tracker.Stats("user-1", "session-a")
	if stats.TokensUsed != 0 || stats.CostUSD != 0 || stats.NumTurns != 0 {
		t.Fatalf("expected zero stats for missing state file, got %+v", stats)
	}
	if stats.TokenLimit != 1000 {
		t.Fatalf("expected token limit 1000, got %d", stats.TokenLimit)
	}
}

func TestSessionTrackerRecordIsConcurrencySafe(t *testing.T) {
	tracker := NewSessionTracker(t.TempDir(), 0)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)

		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				tracker.Record("user-1", "session-a", 1, 0.001)
			}
		}()

		go func() {
			defer wg.Done()
			for j := 0; j < 10; j++ {
				tracker.Record("user-1", "session-b", 2, 0.002)
			}
		}()
	}
	wg.Wait()

	statsA := tracker.Stats("user-1", "session-a")
	if statsA.TokensUsed != 1000 {
		t.Fatalf("expected 1000 tokens in session-a, got %d", statsA.TokensUsed)
	}

	statsB := tracker.Stats("user-1", "session-b")
	if statsB.TokensUsed != 1000 {
		t.Fatalf("expected 1000 tokens in session-b, got %d", statsB.TokensUsed)
	}
}
