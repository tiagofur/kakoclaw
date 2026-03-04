package agent

import "strings"

// stopKeywords maps language codes to their stop/cancel keywords
var stopKeywords = map[string][]string{
	"en": {"stop", "cancel", "abort", "halt"},
	"es": {"parar", "cancelar", "detener"},
	"fr": {"arrêter", "annuler", "stop"},
	"de": {"stopp", "abbrechen", "halt"},
	"pt": {"parar", "cancelar", "abortar"},
	"zh": {"停止", "取消"},
	"ja": {"停止", "中止", "やめて", "キャンセル"},
	"ar": {"توقف", "إلغاء"},
	"hi": {"रुको", "बंद करो", "रद्द करो"},
	"ru": {"стоп", "отмена", "остановить"},
}

// allStopKeywords is a flattened set for O(1) lookup
var allStopKeywords map[string]struct{}

func init() {
	allStopKeywords = make(map[string]struct{})
	for _, keywords := range stopKeywords {
		for _, kw := range keywords {
			allStopKeywords[strings.ToLower(kw)] = struct{}{}
		}
	}
}

// IsStopCommand returns true if the text matches any stop keyword in any supported language
func IsStopCommand(text string) bool {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return false
	}
	_, ok := allStopKeywords[normalized]
	return ok
}
