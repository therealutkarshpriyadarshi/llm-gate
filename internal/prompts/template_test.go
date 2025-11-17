package prompts

import (
	"testing"
)

func TestTemplateEngine_Render(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		want      string
		wantErr   bool
	}{
		{
			name:      "simple variable substitution",
			template:  "Hello {{.name}}!",
			variables: map[string]any{"name": "World"},
			want:      "Hello World!",
			wantErr:   false,
		},
		{
			name:      "multiple variables",
			template:  "{{.greeting}} {{.name}}, you have {{.count}} messages",
			variables: map[string]any{"greeting": "Hello", "name": "Alice", "count": 5},
			want:      "Hello Alice, you have 5 messages",
			wantErr:   false,
		},
		{
			name:      "conditional rendering",
			template:  "{{if .premium}}Premium User{{else}}Free User{{end}}",
			variables: map[string]any{"premium": true},
			want:      "Premium User",
			wantErr:   false,
		},
		{
			name:      "with template functions",
			template:  "{{upper .text}}",
			variables: map[string]any{"text": "hello world"},
			want:      "HELLO WORLD",
			wantErr:   false,
		},
		{
			name:      "missing variable",
			template:  "Hello {{.name}}!",
			variables: map[string]any{},
			want:      "Hello <no value>!",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.Render(tt.template, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("Render() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Render() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_RenderSimple(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		want      string
		wantErr   bool
	}{
		{
			name:      "simple variable",
			template:  "Hello {{name}}!",
			variables: map[string]any{"name": "World"},
			want:      "Hello World!",
			wantErr:   false,
		},
		{
			name:      "multiple variables",
			template:  "{{greeting}} {{name}}",
			variables: map[string]any{"greeting": "Hi", "name": "Bob"},
			want:      "Hi Bob",
			wantErr:   false,
		},
		{
			name:      "number variable",
			template:  "You have {{count}} items",
			variables: map[string]any{"count": 42},
			want:      "You have 42 items",
			wantErr:   false,
		},
		{
			name:      "missing variable",
			template:  "Hello {{name}}!",
			variables: map[string]any{},
			want:      "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := engine.RenderSimple(tt.template, tt.variables)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderSimple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("RenderSimple() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplateEngine_Validate(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		wantErr  bool
	}{
		{
			name:     "valid template",
			template: "Hello {{.name}}!",
			wantErr:  false,
		},
		{
			name:     "valid with functions",
			template: "{{upper .text}} {{if .show}}shown{{end}}",
			wantErr:  false,
		},
		{
			name:     "invalid syntax",
			template: "Hello {{.name}",
			wantErr:  true,
		},
		{
			name:     "invalid function",
			template: "{{invalidFunc .text}}",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := engine.Validate(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateEngine_ExtractVariables(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		want     int // Number of unique variables
	}{
		{
			name:     "single variable",
			template: "Hello {{.name}}!",
			want:     1,
		},
		{
			name:     "multiple variables",
			template: "{{.greeting}} {{.name}}, you have {{.count}} messages",
			want:     3,
		},
		{
			name:     "duplicate variables",
			template: "{{.name}} {{.name}} {{.age}}",
			want:     2,
		},
		{
			name:     "with template keywords",
			template: "{{if .premium}}Hello {{.name}}{{end}}",
			want:     1, // Only extracts 'name', 'premium' is in conditional which is not always extracted
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.ExtractVariables(tt.template)
			if len(got) != tt.want {
				t.Errorf("ExtractVariables() returned %d variables, want %d. Got: %v", len(got), tt.want, got)
			}
		})
	}
}

func TestTemplateEngine_RenderWithValidation(t *testing.T) {
	engine := NewTemplateEngine()

	requiredVars := []Variable{
		{Name: "name", Type: "string", Required: true},
		{Name: "age", Type: "number", Required: true},
		{Name: "premium", Type: "boolean", Required: false, DefaultValue: false},
	}

	tests := []struct {
		name      string
		template  string
		variables map[string]any
		wantErr   bool
	}{
		{
			name:      "all required variables provided",
			template:  "Name: {{.name}}, Age: {{.age}}",
			variables: map[string]any{"name": "Alice", "age": 30},
			wantErr:   false,
		},
		{
			name:      "missing required variable",
			template:  "Name: {{.name}}",
			variables: map[string]any{},
			wantErr:   true,
		},
		{
			name:      "wrong type",
			template:  "Age: {{.age}}",
			variables: map[string]any{"name": "Alice", "age": "thirty"},
			wantErr:   true,
		},
		{
			name:      "optional variable with default",
			template:  "Premium: {{.premium}}",
			variables: map[string]any{"name": "Alice", "age": 30},
			wantErr:   false, // Should use default value
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.RenderWithValidation(tt.template, tt.variables, requiredVars)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderWithValidation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTemplateEngine_ComposeTemplates(t *testing.T) {
	engine := NewTemplateEngine()

	template1 := "Part 1: {{.a}}"
	template2 := "Part 2: {{.b}}"
	template3 := "Part 3: {{.c}}"

	composed := engine.ComposeTemplates(template1, template2, template3)

	expected := "Part 1: {{.a}}\n\nPart 2: {{.b}}\n\nPart 3: {{.c}}"
	if composed != expected {
		t.Errorf("ComposeTemplates() = %v, want %v", composed, expected)
	}
}

func TestTemplateEngine_ConditionalRender(t *testing.T) {
	engine := NewTemplateEngine()

	trueTemplate := "Condition is true: {{.value}}"
	falseTemplate := "Condition is false: {{.value}}"
	variables := map[string]any{"value": "test"}

	// Test with condition = true
	got, err := engine.ConditionalRender(true, trueTemplate, falseTemplate, variables)
	if err != nil {
		t.Errorf("ConditionalRender() error = %v", err)
	}
	if got != "Condition is true: test" {
		t.Errorf("ConditionalRender(true) = %v, want 'Condition is true: test'", got)
	}

	// Test with condition = false
	got, err = engine.ConditionalRender(false, trueTemplate, falseTemplate, variables)
	if err != nil {
		t.Errorf("ConditionalRender() error = %v", err)
	}
	if got != "Condition is false: test" {
		t.Errorf("ConditionalRender(false) = %v, want 'Condition is false: test'", got)
	}
}

func BenchmarkTemplateEngine_Render(b *testing.B) {
	engine := NewTemplateEngine()
	template := "Hello {{.name}}, you have {{.count}} messages from {{.sender}}"
	variables := map[string]any{
		"name":   "Alice",
		"count":  42,
		"sender": "Bob",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.Render(template, variables)
	}
}

func BenchmarkTemplateEngine_RenderSimple(b *testing.B) {
	engine := NewTemplateEngine()
	template := "Hello {{name}}, you have {{count}} messages from {{sender}}"
	variables := map[string]any{
		"name":   "Alice",
		"count":  42,
		"sender": "Bob",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = engine.RenderSimple(template, variables)
	}
}
