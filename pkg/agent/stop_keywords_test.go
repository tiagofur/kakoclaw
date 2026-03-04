package agent

import "testing"

func TestIsStopCommand(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// English
		{"stop", true},
		{"cancel", true},
		{"abort", true},
		{"halt", true},
		// Case insensitive
		{"STOP", true},
		{"Cancel", true},
		{"ABORT", true},
		// With whitespace
		{"  stop  ", true},
		{"\tstop\n", true},
		// Spanish
		{"parar", true},
		{"cancelar", true},
		{"detener", true},
		// French
		{"arrêter", true},
		{"annuler", true},
		// German
		{"stopp", true},
		{"abbrechen", true},
		// Portuguese
		{"abortar", true},
		// Chinese
		{"停止", true},
		{"取消", true},
		// Japanese
		{"中止", true},
		{"やめて", true},
		{"キャンセル", true},
		// Arabic
		{"توقف", true},
		{"إلغاء", true},
		// Hindi
		{"रुको", true},
		{"बंद करो", true},
		// Russian
		{"стоп", true},
		{"отмена", true},
		{"остановить", true},
		// Non-stop words
		{"hello", false},
		{"continue", false},
		{"stopping", false},
		{"", false},
		{"   ", false},
		{"stop now", false},
		{"please stop", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsStopCommand(tt.input)
			if got != tt.want {
				t.Errorf("IsStopCommand(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
