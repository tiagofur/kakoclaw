package workflow

import (
	"encoding/json"
	"testing"
)

func TestNewEngine(t *testing.T) {
	e := NewEngine(nil, nil, nil)
	if e == nil {
		t.Fatal("NewEngine returned nil")
	}
}

// --- interpolate tests ---

func TestInterpolate_StepOutput(t *testing.T) {
	results := []StepResult{
		{Output: "hello world"},
	}
	got := interpolate("Result: {{step.1.output}}", results, nil)
	want := "Result: hello world"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolate_StepError(t *testing.T) {
	results := []StepResult{
		{Output: "ok", Error: "something failed"},
	}
	got := interpolate("Error was: {{step.1.error}}", results, nil)
	want := "Error was: something failed"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolate_Params(t *testing.T) {
	params := map[string]string{
		"name":  "Alice",
		"color": "blue",
	}
	got := interpolate("Hello {{param.name}}, your color is {{param.color}}", nil, params)
	want := "Hello Alice, your color is blue"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolate_MissingIndex(t *testing.T) {
	results := []StepResult{
		{Output: "only one"},
	}
	text := "Value: {{step.99.output}}"
	got := interpolate(text, results, nil)
	if got != text {
		t.Errorf("expected unchanged text %q, got %q", text, got)
	}
}

func TestInterpolate_MultipleReplacements(t *testing.T) {
	results := []StepResult{
		{Output: "first", Error: "err1"},
		{Output: "second", Error: "err2"},
	}
	params := map[string]string{"key": "val"}
	got := interpolate("{{step.1.output}} and {{step.2.error}} with {{param.key}}", results, params)
	want := "first and err2 with val"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestInterpolate_NoPlaceholders(t *testing.T) {
	text := "plain text without any templates"
	got := interpolate(text, nil, nil)
	if got != text {
		t.Errorf("got %q, want %q", got, text)
	}
}

func TestInterpolate_ZeroIndex(t *testing.T) {
	// {{step.0.output}} — 1-based indexing means 0 becomes -1, out of bounds
	results := []StepResult{
		{Output: "first"},
	}
	text := "{{step.0.output}}"
	got := interpolate(text, results, nil)
	if got != text {
		t.Errorf("step.0 should be out of bounds (1-based indexing); got %q, want %q", got, text)
	}
}

func TestInterpolate_MissingParam(t *testing.T) {
	params := map[string]string{"exists": "yes"}
	text := "{{param.missing}}"
	got := interpolate(text, nil, params)
	if got != text {
		t.Errorf("missing param should be left unchanged; got %q, want %q", got, text)
	}
}

func TestInterpolate_NilParams(t *testing.T) {
	text := "{{param.name}}"
	got := interpolate(text, nil, nil)
	if got != text {
		t.Errorf("nil params should leave placeholder unchanged; got %q, want %q", got, text)
	}
}

// --- evaluateCondition tests ---

// newTestEngine creates a minimal Engine for testing condition evaluation.
func newTestEngine() *Engine {
	return NewEngine(nil, nil, nil)
}

// makeConditionStep builds a Step with ConditionConfig serialized into Config.
func makeConditionStep(operator, value, reference string) Step {
	cfg := ConditionConfig{
		Operator:  operator,
		Value:     value,
		Reference: reference,
	}
	raw, _ := json.Marshal(cfg)
	return Step{
		ID:     "cond-1",
		Type:   StepCondition,
		Label:  "test condition",
		Config: raw,
	}
}

func TestEvaluateCondition_Contains(t *testing.T) {
	e := newTestEngine()
	results := []StepResult{
		{Output: "The quick brown fox"},
	}

	tests := []struct {
		name      string
		value     string
		wantMatch bool
	}{
		{"substring present", "quick brown", true},
		{"case insensitive", "QUICK", true},
		{"substring absent", "lazy dog", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := makeConditionStep("contains", tt.value, "{{step.1.output}}")
			got, err := e.evaluateCondition(step, results, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("contains(%q) = %v, want %v", tt.value, got, tt.wantMatch)
			}
		})
	}
}

func TestEvaluateCondition_Equals(t *testing.T) {
	e := newTestEngine()
	results := []StepResult{
		{Output: "  exact match  "},
	}

	tests := []struct {
		name      string
		value     string
		wantMatch bool
	}{
		{"exact match with whitespace trimming", "exact match", true},
		{"mismatch", "different", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := makeConditionStep("equals", tt.value, "{{step.1.output}}")
			got, err := e.evaluateCondition(step, results, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("equals(%q) = %v, want %v", tt.value, got, tt.wantMatch)
			}
		})
	}
}

func TestEvaluateCondition_NotEmpty(t *testing.T) {
	e := newTestEngine()

	tests := []struct {
		name      string
		output    string
		wantMatch bool
	}{
		{"non-empty output", "some content", true},
		{"empty output", "", false},
		{"whitespace only", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := []StepResult{{Output: tt.output}}
			step := makeConditionStep("not_empty", "", "{{step.1.output}}")
			got, err := e.evaluateCondition(step, results, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("not_empty(%q) = %v, want %v", tt.output, got, tt.wantMatch)
			}
		})
	}
}

func TestEvaluateCondition_Regex(t *testing.T) {
	e := newTestEngine()
	results := []StepResult{
		{Output: "error code: 404"},
	}

	tests := []struct {
		name      string
		pattern   string
		wantMatch bool
	}{
		{"matching regex", `\d{3}`, true},
		{"non-matching regex", `^success`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := makeConditionStep("regex", tt.pattern, "{{step.1.output}}")
			got, err := e.evaluateCondition(step, results, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantMatch {
				t.Errorf("regex(%q) = %v, want %v", tt.pattern, got, tt.wantMatch)
			}
		})
	}
}

func TestEvaluateCondition_InvalidRegex(t *testing.T) {
	e := newTestEngine()
	results := []StepResult{{Output: "test"}}
	step := makeConditionStep("regex", "[invalid", "{{step.1.output}}")
	_, err := e.evaluateCondition(step, results, nil)
	if err == nil {
		t.Fatal("expected error for invalid regex, got nil")
	}
}

func TestEvaluateCondition_UnknownOperator(t *testing.T) {
	e := newTestEngine()
	results := []StepResult{{Output: "test"}}
	step := makeConditionStep("greater_than", "5", "{{step.1.output}}")
	_, err := e.evaluateCondition(step, results, nil)
	if err == nil {
		t.Fatal("expected error for unknown operator, got nil")
	}
}

func TestEvaluateCondition_OutOfBounds(t *testing.T) {
	e := newTestEngine()
	results := []StepResult{{Output: "only one"}}
	// Reference step.99 which doesn't exist. Interpolation leaves it unchanged,
	// so the subject becomes the literal "{{step.99.output}}" which is non-empty.
	step := makeConditionStep("not_empty", "", "{{step.99.output}}")
	got, err := e.evaluateCondition(step, results, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The unresolved placeholder is non-empty text, so not_empty returns true.
	// The key behavior being tested is that no panic/crash occurs.
	if !got {
		t.Error("expected true since unresolved placeholder is non-empty text")
	}
}
