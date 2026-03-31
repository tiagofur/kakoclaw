package bridge

import (
	"encoding/json"
)

func safeClose(ch chan Event) {
	defer func() { recover() }()
	close(ch)
}

// readLoop runs in a goroutine, reading stdout and routing events.
func (b *Bridge) readLoop() {
	defer close(b.done)

	for b.scanner.Scan() {
		line := b.scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var ev Event
		if err := json.Unmarshal(line, &ev); err != nil {
			// Malformed or unstructured log line, ignore.
			continue
		}

		rid := ev.RequestID

		b.pendingMu.Lock()
		ch, ok := b.pending[rid]
		b.pendingMu.Unlock()

		if !ok {
			continue
		}

		select {
		case ch <- ev:
		default:
		}

		if ev.IsTerminal() {
			b.pendingMu.Lock()
			delete(b.pending, rid)
			b.pendingMu.Unlock()
			safeClose(ch)
		}
	}

	// Unexpected exit tracking
	b.mu.Lock()
	isStopping := (b.State() == "stopped")
	if !isStopping {
		b.setState("dead")
	}
	cb := b.cfg.OnDeath
	b.mu.Unlock()

	if !isStopping && cb != nil {
		// Provide an error or just call the OnDeath closure
		cb(nil) 
	}

	// Cleanup all pending
	b.pendingMu.Lock()
	for rid, ch := range b.pending {
		safeClose(ch)
		delete(b.pending, rid)
	}
	b.pendingMu.Unlock()

	b.mu.Lock()
	if !isStopping {
		b.cmd = nil
	}
	b.mu.Unlock()
}
