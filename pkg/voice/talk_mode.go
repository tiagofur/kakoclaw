package voice

// TalkModeSession manages a continuous voice conversation loop:
// audio capture → STT → agent processing → TTS playback.
type TalkModeSession struct {
	stt    STTProvider
	tts    TTSProvider
	stopCh chan struct{}
}

// NewTalkModeSession creates a new TalkModeSession with the given STT and TTS providers.
func NewTalkModeSession(stt STTProvider, tts TTSProvider) *TalkModeSession {
	return &TalkModeSession{stt: stt, tts: tts, stopCh: make(chan struct{})}
}

// IsActive returns true if the session has not been stopped.
func (s *TalkModeSession) IsActive() bool {
	select {
	case <-s.stopCh:
		return false
	default:
		return true
	}
}

// Stop terminates the talk mode session. Safe to call multiple times.
func (s *TalkModeSession) Stop() {
	select {
	case <-s.stopCh:
	default:
		close(s.stopCh)
	}
}
