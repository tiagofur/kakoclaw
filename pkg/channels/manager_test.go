package channels

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sipeed/makoclaw/pkg/bus"
	"github.com/sipeed/makoclaw/pkg/config"
	"github.com/sipeed/makoclaw/pkg/storage"
)

// mockChannel is a minimal Channel implementation for testing the Manager
// without starting real channel connections.
type mockChannel struct {
	*BaseChannel
	started bool
	stopped bool
}

func newMockChannel(name string) *mockChannel {
	return &mockChannel{
		BaseChannel: NewBaseChannel(name, nil, nil, nil),
	}
}

func (m *mockChannel) Start(ctx context.Context) error {
	m.started = true
	m.setRunning(true)
	return nil
}

func (m *mockChannel) Stop(ctx context.Context) error {
	m.stopped = true
	m.setRunning(false)
	return nil
}

func (m *mockChannel) Send(ctx context.Context, msg bus.OutboundMessage) error {
	return nil
}

// ---------------------------------------------------------------------------
// Manager construction with empty config (no channels enabled)
// ---------------------------------------------------------------------------

func TestNewManager_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager with empty config failed: %v", err)
	}
	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	channels := mgr.GetEnabledChannels()
	if len(channels) != 0 {
		t.Errorf("GetEnabledChannels() = %v; want empty", channels)
	}
}

// ---------------------------------------------------------------------------
// RegisterChannel / GetChannel
// ---------------------------------------------------------------------------

func TestManager_RegisterAndGetChannel(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	mock := newMockChannel("test-channel")
	mgr.RegisterChannel("test-channel", mock)

	ch, ok := mgr.GetChannel("test-channel")
	if !ok {
		t.Fatal("GetChannel returned false for registered channel")
	}
	if ch.Name() != "test-channel" {
		t.Errorf("Name() = %q; want %q", ch.Name(), "test-channel")
	}
}

func TestManager_GetChannel_NotFound(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	_, ok := mgr.GetChannel("nonexistent")
	if ok {
		t.Error("GetChannel should return false for unregistered channel")
	}
}

// ---------------------------------------------------------------------------
// UnregisterChannel
// ---------------------------------------------------------------------------

func TestManager_UnregisterChannel(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	mock := newMockChannel("to-remove")
	mgr.RegisterChannel("to-remove", mock)

	// Verify it's registered
	_, ok := mgr.GetChannel("to-remove")
	if !ok {
		t.Fatal("channel should be registered before unregister")
	}

	mgr.UnregisterChannel("to-remove")

	_, ok = mgr.GetChannel("to-remove")
	if ok {
		t.Error("GetChannel should return false after unregister")
	}
}

// ---------------------------------------------------------------------------
// GetStatus
// ---------------------------------------------------------------------------

func TestManager_GetStatus_Empty(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	status := mgr.GetStatus()
	if len(status) != 0 {
		t.Errorf("GetStatus() on empty manager = %v; want empty map", status)
	}
}

func TestManager_GetStatus_WithChannels(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	// Register two channels, one started and one not
	started := newMockChannel("started-ch")
	started.Start(context.Background())

	stopped := newMockChannel("stopped-ch")

	mgr.RegisterChannel("started-ch", started)
	mgr.RegisterChannel("stopped-ch", stopped)

	status := mgr.GetStatus()
	if len(status) != 2 {
		t.Fatalf("GetStatus() returned %d entries; want 2", len(status))
	}

	// Check started channel status
	startedStatus, ok := status["started-ch"].(map[string]interface{})
	if !ok {
		t.Fatal("started-ch status is not a map")
	}
	if startedStatus["enabled"] != true {
		t.Errorf("started-ch enabled = %v; want true", startedStatus["enabled"])
	}
	if startedStatus["running"] != true {
		t.Errorf("started-ch running = %v; want true", startedStatus["running"])
	}

	// Check stopped channel status
	stoppedStatus, ok := status["stopped-ch"].(map[string]interface{})
	if !ok {
		t.Fatal("stopped-ch status is not a map")
	}
	if stoppedStatus["running"] != false {
		t.Errorf("stopped-ch running = %v; want false", stoppedStatus["running"])
	}
}

// ---------------------------------------------------------------------------
// GetEnabledChannels
// ---------------------------------------------------------------------------

func TestManager_GetEnabledChannels(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	mgr.RegisterChannel("alpha", newMockChannel("alpha"))
	mgr.RegisterChannel("beta", newMockChannel("beta"))
	mgr.RegisterChannel("gamma", newMockChannel("gamma"))

	enabled := mgr.GetEnabledChannels()
	if len(enabled) != 3 {
		t.Fatalf("GetEnabledChannels() returned %d; want 3", len(enabled))
	}

	// Verify all names are present (order is non-deterministic from map)
	nameSet := make(map[string]bool)
	for _, name := range enabled {
		nameSet[name] = true
	}
	for _, want := range []string{"alpha", "beta", "gamma"} {
		if !nameSet[want] {
			t.Errorf("GetEnabledChannels() missing %q", want)
		}
	}
}

// ---------------------------------------------------------------------------
// SendToChannel
// ---------------------------------------------------------------------------

func TestManager_SendToChannel_NotFound(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	err = mgr.SendToChannel(context.Background(), "nonexistent", "chat1", "hello")
	if err == nil {
		t.Error("SendToChannel to nonexistent channel should return error")
	}
}

func TestManager_SendToChannel_Success(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	mock := newMockChannel("test")
	mgr.RegisterChannel("test", mock)

	err = mgr.SendToChannel(context.Background(), "test", "chat1", "hello")
	if err != nil {
		t.Errorf("SendToChannel to registered channel failed: %v", err)
	}
}

// ---------------------------------------------------------------------------
// RestartChannel for unknown channel type
// ---------------------------------------------------------------------------

func TestManager_RestartChannel_UnknownType(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	err = mgr.RestartChannel(context.Background(), "unknown-channel-type")
	if err == nil {
		t.Error("RestartChannel with unknown type should return error")
	}
}

// ---------------------------------------------------------------------------
// StopAll with registered mock channels
// ---------------------------------------------------------------------------

func TestManager_StopAll(t *testing.T) {
	cfg := &config.Config{}
	mb := bus.NewMessageBus()
	defer mb.Close()

	mgr, err := NewManager(cfg, mb, nil)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	ch1 := newMockChannel("ch1")
	ch2 := newMockChannel("ch2")
	ch1.Start(context.Background())
	ch2.Start(context.Background())

	mgr.RegisterChannel("ch1", ch1)
	mgr.RegisterChannel("ch2", ch2)

	if err := mgr.StopAll(context.Background()); err != nil {
		t.Fatalf("StopAll failed: %v", err)
	}

	if ch1.IsRunning() {
		t.Error("ch1 should not be running after StopAll")
	}
	if ch2.IsRunning() {
		t.Error("ch2 should not be running after StopAll")
	}
}

// ---------------------------------------------------------------------------
// applyDMPolicy wiring (Task 4.1)
// ---------------------------------------------------------------------------

// spyChannel wraps BaseChannel and records SetDMPolicy calls.
type spyChannel struct {
	*BaseChannel
	dmPolicySet   string
	pairingStore  *storage.PairingStore
	dmPolicyCalls int
}

func newSpyChannel(name string) *spyChannel {
	return &spyChannel{
		BaseChannel: NewBaseChannel(name, nil, nil, nil),
	}
}

func (s *spyChannel) Start(ctx context.Context) error  { return nil }
func (s *spyChannel) Stop(ctx context.Context) error   { return nil }
func (s *spyChannel) Send(ctx context.Context, msg bus.OutboundMessage) error { return nil }

func (s *spyChannel) SetDMPolicy(policy string, ps *storage.PairingStore) {
	s.dmPolicyCalls++
	s.dmPolicySet = policy
	s.pairingStore = ps
	s.BaseChannel.SetDMPolicy(policy, ps)
}

// mockDMPolicyConfig wraps a policy string and satisfies DMPolicyProvider.
type mockDMPolicyConfig struct {
	policy string
}

func (m mockDMPolicyConfig) GetDMPolicy() string { return m.policy }

func TestApplyDMPolicy_OpenPolicy(t *testing.T) {
	mb := bus.NewMessageBus()
	defer mb.Close()
	mgr := &Manager{channels: make(map[string]Channel), bus: mb}

	ch := newSpyChannel("test")
	mgr.applyDMPolicy("test", ch, mockDMPolicyConfig{policy: "open"})

	if ch.dmPolicyCalls != 1 {
		t.Fatalf("SetDMPolicy call count = %d; want 1", ch.dmPolicyCalls)
	}
	if ch.dmPolicySet != "open" {
		t.Errorf("SetDMPolicy policy = %q; want %q", ch.dmPolicySet, "open")
	}
	if ch.pairingStore != nil {
		t.Error("SetDMPolicy pairingStore should be nil for open policy")
	}
}

func TestApplyDMPolicy_DisabledPolicy(t *testing.T) {
	mb := bus.NewMessageBus()
	defer mb.Close()
	mgr := &Manager{channels: make(map[string]Channel), bus: mb}

	ch := newSpyChannel("test")
	mgr.applyDMPolicy("test", ch, mockDMPolicyConfig{policy: "disabled"})

	if ch.dmPolicyCalls != 1 {
		t.Fatalf("SetDMPolicy call count = %d; want 1", ch.dmPolicyCalls)
	}
	if ch.dmPolicySet != "disabled" {
		t.Errorf("SetDMPolicy policy = %q; want %q", ch.dmPolicySet, "disabled")
	}
	if ch.pairingStore != nil {
		t.Error("SetDMPolicy pairingStore should be nil for disabled policy")
	}
}

func TestApplyDMPolicy_PairingWithStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.db")
	store, err := storage.NewUserStorage(path)
	if err != nil {
		t.Fatalf("NewUserStorage: %v", err)
	}
	defer store.Close()

	mb := bus.NewMessageBus()
	defer mb.Close()
	mgr := &Manager{channels: make(map[string]Channel), bus: mb, storage: store}

	ch := newSpyChannel("test")
	mgr.applyDMPolicy("test", ch, mockDMPolicyConfig{policy: "pairing"})

	if ch.dmPolicyCalls != 1 {
		t.Fatalf("SetDMPolicy call count = %d; want 1", ch.dmPolicyCalls)
	}
	if ch.dmPolicySet != "pairing" {
		t.Errorf("SetDMPolicy policy = %q; want %q", ch.dmPolicySet, "pairing")
	}
	if ch.pairingStore == nil {
		t.Error("SetDMPolicy pairingStore should be non-nil when storage is available")
	}
}

func TestApplyDMPolicy_PairingWithoutStorage(t *testing.T) {
	mb := bus.NewMessageBus()
	defer mb.Close()
	// Manager has no storage — pairing falls through to SetDMPolicy("pairing", nil)
	mgr := &Manager{channels: make(map[string]Channel), bus: mb, storage: nil}

	ch := newSpyChannel("test")
	mgr.applyDMPolicy("test", ch, mockDMPolicyConfig{policy: "pairing"})

	if ch.dmPolicyCalls != 1 {
		t.Fatalf("SetDMPolicy call count = %d; want 1", ch.dmPolicyCalls)
	}
	if ch.dmPolicySet != "pairing" {
		t.Errorf("SetDMPolicy policy = %q; want %q", ch.dmPolicySet, "pairing")
	}
	if ch.pairingStore != nil {
		t.Error("SetDMPolicy pairingStore should be nil when storage is unavailable")
	}
}

func TestApplyDMPolicy_EmptyPolicy(t *testing.T) {
	mb := bus.NewMessageBus()
	defer mb.Close()
	mgr := &Manager{channels: make(map[string]Channel), bus: mb}

	ch := newSpyChannel("test")
	mgr.applyDMPolicy("test", ch, mockDMPolicyConfig{policy: ""})

	if ch.dmPolicyCalls != 0 {
		t.Errorf("SetDMPolicy should not be called for empty policy; got %d calls", ch.dmPolicyCalls)
	}
}

func TestApplyDMPolicy_NonProvider(t *testing.T) {
	mb := bus.NewMessageBus()
	defer mb.Close()
	mgr := &Manager{channels: make(map[string]Channel), bus: mb}

	ch := newSpyChannel("test")
	// Pass a plain struct that does NOT implement DMPolicyProvider
	mgr.applyDMPolicy("test", ch, struct{ Foo string }{Foo: "bar"})

	if ch.dmPolicyCalls != 0 {
		t.Errorf("SetDMPolicy should not be called when config is not a DMPolicyProvider; got %d calls", ch.dmPolicyCalls)
	}
}
