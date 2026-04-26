package marketing

import "testing"

func TestRender(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]string
		want     string
	}{
		{
			name:     "Simple Replacement",
			template: "Hello {{ name }}!",
			data:     map[string]string{"name": "World"},
			want:     "Hello World!",
		},
		{
			name:     "HTML Escaping",
			template: "Message: {{ msg }}",
			data:     map[string]string{"msg": "<b>Hello</b> & Welcome"},
			want:     "Message: &lt;b&gt;Hello&lt;/b&gt; &amp; Welcome",
		},
		{
			name:     "Raw Output",
			template: "HTML: {{{ html }}}",
			data:     map[string]string{"html": "<b>Hello</b>"},
			want:     "HTML: <b>Hello</b>",
		},
		{
			name:     "Whitespace in Tags",
			template: "Values: {{key1}} and {{{  key2  }}}",
			data:     map[string]string{"key1": "v1", "key2": "v2"},
			want:     "Values: v1 and v2",
		},
		{
			name:     "Missing Keys",
			template: "Hello {{ name }}, welcome to {{ city }}.",
			data:     map[string]string{"name": "Alice"},
			want:     "Hello Alice, welcome to .",
		},
		{
			name:     "Complex Keys",
			template: "Data: {{ user.profile_name }}",
			data:     map[string]string{"user.profile_name": "jdoe"},
			want:     "Data: jdoe",
		},
		{
			name:     "Multiple Occurrences",
			template: "{{ name }} {{ name }} {{ name }}",
			data:     map[string]string{"name": "Echo"},
			want:     "Echo Echo Echo",
		},
		{
			name:     "No Tags",
			template: "Plain text with no tags.",
			data:     map[string]string{"foo": "bar"},
			want:     "Plain text with no tags.",
		},
		{
			name:     "Mixed Escaped and Raw",
			template: "{{ val }} vs {{{ val }}}",
			data:     map[string]string{"val": "<br>"},
			want:     "&lt;br&gt; vs <br>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Render(tt.template, tt.data)
			if got != tt.want {
				t.Errorf("Render() = %q, want %q", got, tt.want)
			}
		})
	}
}
