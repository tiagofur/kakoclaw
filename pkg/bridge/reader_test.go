package bridge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/bridge"
)

func TestReader_MalformedLinesIgnored(t *testing.T) {
	dir := t.TempDir()
	mockJS := `
const readline = require('readline');
const rl = readline.createInterface({ input: process.stdin, terminal: false });

rl.on('line', (line) => {
    let req = JSON.parse(line);
    const rid = req.request_id || "";
    // Output valid JSON, then garbled JSON, then valid result
    process.stdout.write(JSON.stringify({event:"system",request_id:rid}) + "\n");
    process.stdout.write("Gargle Garble Syntax Error\n");
    process.stdout.write(JSON.stringify({event:"result",request_id:rid,content:"done"}) + "\n");
});

rl.on('close', () => { process.exit(0); });
`
	if err := os.WriteFile(filepath.Join(dir, "malformed.js"), []byte(mockJS), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := bridge.BridgeConfig{
		Backend:  "malformed.js",
		NodePath: "node",
		Cwd:      dir,
	}
	b := bridge.New(cfg)
	defer b.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ch, err := b.Execute(ctx, bridge.Request{Command: "query"})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	var events []bridge.Event
	for ev := range ch {
		events = append(events, ev)
	}

	// Should have parsed 2 events successfully, ignored the malformed string.
	if len(events) != 2 {
		t.Fatalf("Expected 2 events, got %d", len(events))
	}
	if events[0].Type != "system" {
		t.Errorf("events[0] got %q", events[0].Type)
	}
	if events[1].Type != "result" {
		t.Errorf("events[1] got %q", events[1].Type)
	}
}

func TestReader_OnDeathCalledOnEOF(t *testing.T) {
	dir := t.TempDir()
	crashJS := `
const readline = require('readline');
const rl = readline.createInterface({ input: process.stdin, terminal: false });
rl.on('line', () => {
    // exit prematurely
    process.exit(1);
});
`
	if err := os.WriteFile(filepath.Join(dir, "crash.js"), []byte(crashJS), 0644); err != nil {
		t.Fatal(err)
	}

	called := make(chan struct{})
	cfg := bridge.BridgeConfig{
		Backend:  "crash.js",
		NodePath: "node",
		Cwd:      dir,
		OnDeath: func(err error) {
			close(called)
		},
	}
	b := bridge.New(cfg)
	defer b.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = b.Start(ctx)
	b.Execute(ctx, bridge.Request{Command: "query"})

	select {
	case <-called:
		// success
	case <-time.After(2 * time.Second):
		t.Fatal("OnDeath was not invoked after process crash")
	}

	if b.State() != "dead" {
		t.Errorf("Expected bridge state dead, got %s", b.State())
	}
}
