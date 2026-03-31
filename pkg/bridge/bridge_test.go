package bridge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sipeed/makoclaw/pkg/bridge"
)

const longLivedMockJS = `
const readline = require('readline');
const rl = readline.createInterface({ input: process.stdin, terminal: false });

rl.on('line', (line) => {
    let req;
    try {
        req = JSON.parse(line);
    } catch(e) {
        process.stdout.write(JSON.stringify({event:"error",message:"invalid json"}) + "\n");
        return;
    }

    const rid = req.request_id || "";

    if (req.command === "ping") {
        process.stdout.write(JSON.stringify({event:"pong",request_id:rid}) + "\n");
    } else if (req.command === "query") {
        process.stdout.write(JSON.stringify({event:"system",request_id:rid,session_id:"test-123",tools:["Read"],model:"claude-3"}) + "\n");
        process.stdout.write(JSON.stringify({event:"assistant",request_id:rid,text:"hello world"}) + "\n");
        process.stdout.write(JSON.stringify({event:"result",request_id:rid,content:"done",cost_usd:0.01,session_id:"test-123",duration_ms:100,num_turns:1}) + "\n");
    } else {
        process.stdout.write(JSON.stringify({event:"error",request_id:rid,message:"unknown command: " + req.command}) + "\n");
    }
});

rl.on('close', () => {
    process.exit(0);
});
`

func newMockBridge(t *testing.T, dir string, jsBody string) *bridge.Bridge {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "mock.js"), []byte(jsBody), 0644); err != nil {
		t.Fatal(err)
	}
	cfg := bridge.BridgeConfig{
		Backend:  "mock.js",
		NodePath: "node",
		Cwd:      dir,
	}
	b := bridge.New(cfg)
	t.Cleanup(func() { b.Stop() })
	return b
}

func TestBridge_Execute_ParsesEvents(t *testing.T) {
	dir := t.TempDir()
	b := newMockBridge(t, dir, longLivedMockJS)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ch, err := b.Execute(ctx, bridge.Request{Command: "query", Prompt: "test"})
	if err != nil {
		t.Fatalf("Execute() error: %v", err)
	}

	var events []bridge.Event
	for ev := range ch {
		events = append(events, ev)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d: %+v", len(events), events)
	}

	if events[0].Type != "system" {
		t.Errorf("event[0].Type = %q, want %q", events[0].Type, "system")
	}
	if events[1].Type != "assistant" || events[1].Text != "hello world" {
		t.Errorf("event[1].Text = %q", events[1].Text)
	}
	if events[2].Type != "result" {
		t.Errorf("event[2].Type = %q", events[2].Type)
	}
}

func TestBridge_Lifecycle_StateTransitions(t *testing.T) {
	dir := t.TempDir()
	b := newMockBridge(t, dir, longLivedMockJS)

	if b.State() != "idle" {
		t.Errorf("Expected idle state, got %v", b.State())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = b.Start(ctx)
	if b.State() != "running" {
		t.Errorf("Expected running state, got %v", b.State())
	}

	_ = b.Stop()
	if b.State() != "stopped" {
		t.Errorf("Expected stopped state, got %v", b.State())
	}
}

func TestBridge_Ping(t *testing.T) {
	dir := t.TempDir()
	b := newMockBridge(t, dir, longLivedMockJS)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := b.Ping(ctx); err != nil {
		t.Fatalf("Ping error: %v", err)
	}
}
