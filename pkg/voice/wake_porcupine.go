//go:build porcupine

package voice

import (
	"context"
)

// WakeDetector uses Porcupine CGo bindings for wake word detection.
// This is a stub for the porcupine build tag — production implementation
// would integrate the Picovoice Porcupine SDK for low-latency keyword spotting.
//
// To build with Porcupine support:
//   go build -tags porcupine ./pkg/voice/...
type WakeDetector struct{}

// NewWakeDetector creates a new Porcupine-backed WakeDetector.
func NewWakeDetector() *WakeDetector { return &WakeDetector{} }

// Listen blocks until the context is cancelled.
// TODO: Integrate Porcupine CGo bindings for real-time wake word detection.
func (w *WakeDetector) Listen(ctx context.Context, keyword string, sensitivity float64, onDetected func()) {
	<-ctx.Done()
}
