//go:build !porcupine

package voice

import (
	"context"
)

// WakeDetector listens for a wake word using energy-threshold keyword detection.
// This is a pure-Go stub that blocks until context cancellation.
// Production implementations should integrate platform-specific mic capture.
type WakeDetector struct{}

// NewWakeDetector creates a new WakeDetector.
func NewWakeDetector() *WakeDetector { return &WakeDetector{} }

// Listen blocks until the context is cancelled.
// In production, this would capture audio from the microphone and detect the wake word.
func (w *WakeDetector) Listen(ctx context.Context, keyword string, sensitivity float64, onDetected func()) {
	<-ctx.Done()
}
