package prompts

import (
	"fmt"
	"regexp"
	"strings"
	"text/template"
	"bytes"
)

// TemplateEngine handles prompt template rendering with variable substitution
type TemplateEngine struct {
	// Functions available in templates
	funcMap template.FuncMap
}

// NewTemplateEngine creates a new template engine
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		funcMap: template.FuncMap{
			"upper":      strings.ToUpper,
			"lower":      strings.ToLower,
			"title":      strings.Title,
			"trim":       strings.TrimSpace,
			"join":       strings.Join,
			"default":    defaultValue,
			"contains":   strings.Contains,
			"hasPrefix":  strings.HasPrefix,
			"hasSuffix":  strings.HasSuffix,
			"repeat":     strings.Repeat,
			"replace":    strings.ReplaceAll,
		},
	}
}

// defaultValue returns the default value if the input is nil or empty
func defaultValue(def, value any) any {
	if value == nil || value == "" {
		return def
	}
	return value
}

// Render renders a prompt template with the provided variables
func (e *TemplateEngine) Render(templateStr string, variables map[string]any) (string, error) {
	// Parse template
	tmpl, err := template.New("prompt").Funcs(e.funcMap).Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("parse template: %w", err)
	}

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("execute template: %w", err)
	}

	return buf.String(), nil
}

// RenderSimple renders a simple template with {{variable}} syntax (without Go template features)
func (e *TemplateEngine) RenderSimple(templateStr string, variables map[string]any) (string, error) {
	result := templateStr

	// Simple variable replacement using regex
	re := regexp.MustCompile(`\{\{(\w+)\}\}`)
	matches := re.FindAllStringSubmatch(templateStr, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		placeholder := match[0] // e.g., "{{variable}}"
		varName := match[1]     // e.g., "variable"

		value, ok := variables[varName]
		if !ok {
			return "", fmt.Errorf("variable '%s' not provided", varName)
		}

		// Convert value to string
		strValue := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result, nil
}

// Validate validates that a template can be parsed
func (e *TemplateEngine) Validate(templateStr string) error {
	_, err := template.New("prompt").Funcs(e.funcMap).Parse(templateStr)
	if err != nil {
		return fmt.Errorf("invalid template: %w", err)
	}
	return nil
}

// ExtractVariables extracts variable names from a template
func (e *TemplateEngine) ExtractVariables(templateStr string) []string {
	// Extract variables from {{.Variable}} and {{ .Variable }} patterns
	re := regexp.MustCompile(`\{\{\s*\.?(\w+)\s*\}\}`)
	matches := re.FindAllStringSubmatch(templateStr, -1)

	varSet := make(map[string]bool)
	for _, match := range matches {
		if len(match) >= 2 {
			varName := match[1]
			// Skip template functions
			if varName != "if" && varName != "else" && varName != "end" &&
				varName != "range" && varName != "with" && varName != "define" &&
				varName != "template" && varName != "block" {
				varSet[varName] = true
			}
		}
	}

	// Convert set to slice
	vars := make([]string, 0, len(varSet))
	for v := range varSet {
		vars = append(vars, v)
	}

	return vars
}

// RenderWithValidation renders a template after validating required variables
func (e *TemplateEngine) RenderWithValidation(templateStr string, variables map[string]any, requiredVars []Variable) (string, error) {
	// Check required variables
	for _, v := range requiredVars {
		if !v.Required {
			continue
		}

		value, exists := variables[v.Name]
		if !exists || value == nil {
			// Use default value if available
			if v.DefaultValue != nil {
				variables[v.Name] = v.DefaultValue
				continue
			}
			return "", fmt.Errorf("required variable '%s' not provided", v.Name)
		}

		// Type validation
		if err := e.validateVariableType(v.Name, value, v.Type); err != nil {
			return "", err
		}
	}

	// Add default values for optional variables
	for _, v := range requiredVars {
		if v.Required {
			continue
		}
		if _, exists := variables[v.Name]; !exists && v.DefaultValue != nil {
			variables[v.Name] = v.DefaultValue
		}
	}

	// Render template
	return e.Render(templateStr, variables)
}

// validateVariableType validates that a variable matches its expected type
func (e *TemplateEngine) validateVariableType(name string, value any, expectedType string) error {
	if expectedType == "" {
		return nil // No type validation
	}

	switch expectedType {
	case "string":
		if _, ok := value.(string); !ok {
			return fmt.Errorf("variable '%s' must be a string", name)
		}
	case "number":
		switch value.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
			// Valid number type
		default:
			return fmt.Errorf("variable '%s' must be a number", name)
		}
	case "boolean":
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("variable '%s' must be a boolean", name)
		}
	case "array":
		switch value.(type) {
		case []any, []string, []int, []float64:
			// Valid array type
		default:
			return fmt.Errorf("variable '%s' must be an array", name)
		}
	case "object":
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("variable '%s' must be an object", name)
		}
	}

	return nil
}

// ComposeTemplates composes multiple templates into one
func (e *TemplateEngine) ComposeTemplates(templates ...string) string {
	return strings.Join(templates, "\n\n")
}

// ConditionalRender renders a template based on a condition
func (e *TemplateEngine) ConditionalRender(condition bool, trueTemplate, falseTemplate string, variables map[string]any) (string, error) {
	templateStr := trueTemplate
	if !condition {
		templateStr = falseTemplate
	}

	return e.Render(templateStr, variables)
}
