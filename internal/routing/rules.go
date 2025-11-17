package routing

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// RoutingRule defines a routing rule
type RoutingRule struct {
	// Name of the rule
	Name string

	// Priority (higher priority rules are evaluated first)
	Priority int

	// Conditions that must be met for this rule to apply
	Conditions []Condition

	// Action to take when conditions are met
	Action RuleAction
}

// Condition represents a condition in a routing rule
type Condition interface {
	// Evaluate returns true if the condition is met
	Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool

	// Name returns the condition name
	Name() string
}

// RuleAction represents an action to take when a rule matches
type RuleAction interface {
	// Execute performs the action
	Execute(ctx context.Context, req *models.ChatRequest, providers []interfaces.LLMProvider) (interfaces.LLMProvider, error)

	// Name returns the action name
	Name() string
}

// ComplexityCondition matches on query complexity
type ComplexityCondition struct {
	Complexity QueryComplexity
}

// Evaluate evaluates the complexity condition
func (c *ComplexityCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	return analysis.Complexity == c.Complexity
}

// Name returns the condition name
func (c *ComplexityCondition) Name() string {
	return fmt.Sprintf("complexity:%s", c.Complexity)
}

// CategoryCondition matches on query category
type CategoryCondition struct {
	Category string
}

// Evaluate evaluates the category condition
func (c *CategoryCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	for _, cat := range analysis.Categories {
		if cat == c.Category {
			return true
		}
	}
	return false
}

// Name returns the condition name
func (c *CategoryCondition) Name() string {
	return fmt.Sprintf("category:%s", c.Category)
}

// ModelCondition matches on model name
type ModelCondition struct {
	ModelPattern *regexp.Regexp
}

// NewModelCondition creates a new model condition
func NewModelCondition(pattern string) (*ModelCondition, error) {
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	return &ModelCondition{ModelPattern: regex}, nil
}

// Evaluate evaluates the model condition
func (c *ModelCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	return c.ModelPattern.MatchString(req.Model)
}

// Name returns the condition name
func (c *ModelCondition) Name() string {
	return fmt.Sprintf("model:%s", c.ModelPattern.String())
}

// UserCondition matches on user ID
type UserCondition struct {
	UserID string
}

// Evaluate evaluates the user condition
func (c *UserCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	return req.Metadata.UserID == c.UserID || req.User == c.UserID
}

// Name returns the condition name
func (c *UserCondition) Name() string {
	return fmt.Sprintf("user:%s", c.UserID)
}

// TenantCondition matches on tenant ID
type TenantCondition struct {
	TenantID string
}

// Evaluate evaluates the tenant condition
func (c *TenantCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	return req.Metadata.TenantID == c.TenantID
}

// Name returns the condition name
func (c *TenantCondition) Name() string {
	return fmt.Sprintf("tenant:%s", c.TenantID)
}

// TokenLimitCondition matches on estimated token count
type TokenLimitCondition struct {
	MinTokens int
	MaxTokens int
}

// Evaluate evaluates the token limit condition
func (c *TokenLimitCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	tokens := analysis.EstimatedTokens
	if c.MinTokens > 0 && tokens < c.MinTokens {
		return false
	}
	if c.MaxTokens > 0 && tokens > c.MaxTokens {
		return false
	}
	return true
}

// Name returns the condition name
func (c *TokenLimitCondition) Name() string {
	return fmt.Sprintf("tokens:%d-%d", c.MinTokens, c.MaxTokens)
}

// TagCondition matches on request tags
type TagCondition struct {
	Tag string
}

// Evaluate evaluates the tag condition
func (c *TagCondition) Evaluate(ctx context.Context, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	for _, tag := range req.Metadata.Tags {
		if tag == c.Tag {
			return true
		}
	}
	return false
}

// Name returns the condition name
func (c *TagCondition) Name() string {
	return fmt.Sprintf("tag:%s", c.Tag)
}

// SelectProviderAction selects a specific provider
type SelectProviderAction struct {
	ProviderType models.ProviderType
}

// Execute executes the select provider action
func (a *SelectProviderAction) Execute(ctx context.Context, req *models.ChatRequest, providers []interfaces.LLMProvider) (interfaces.LLMProvider, error) {
	for _, provider := range providers {
		if provider.Name() == a.ProviderType {
			return provider, nil
		}
	}
	return nil, fmt.Errorf("provider %s not found", a.ProviderType)
}

// Name returns the action name
func (a *SelectProviderAction) Name() string {
	return fmt.Sprintf("select:%s", a.ProviderType)
}

// UseStrategyAction uses a specific routing strategy
type UseStrategyAction struct {
	Strategy RoutingStrategy
}

// Execute executes the use strategy action
func (a *UseStrategyAction) Execute(ctx context.Context, req *models.ChatRequest, providers []interfaces.LLMProvider) (interfaces.LLMProvider, error) {
	return a.Strategy.SelectProvider(providers, req)
}

// Name returns the action name
func (a *UseStrategyAction) Name() string {
	return fmt.Sprintf("strategy:%s", a.Strategy.Name())
}

// RejectAction rejects the request
type RejectAction struct {
	Reason string
}

// Execute executes the reject action
func (a *RejectAction) Execute(ctx context.Context, req *models.ChatRequest, providers []interfaces.LLMProvider) (interfaces.LLMProvider, error) {
	return nil, fmt.Errorf("request rejected: %s", a.Reason)
}

// Name returns the action name
func (a *RejectAction) Name() string {
	return "reject"
}

// RulesEngine evaluates routing rules
type RulesEngine struct {
	mu       sync.RWMutex
	rules    []*RoutingRule
	analyzer *QueryAnalyzer
	fallback RoutingStrategy
}

// NewRulesEngine creates a new rules engine
func NewRulesEngine(analyzer *QueryAnalyzer, fallback RoutingStrategy) *RulesEngine {
	if fallback == nil {
		fallback = NewRoundRobinStrategy()
	}
	return &RulesEngine{
		rules:    make([]*RoutingRule, 0),
		analyzer: analyzer,
		fallback: fallback,
	}
}

// AddRule adds a routing rule
func (re *RulesEngine) AddRule(rule *RoutingRule) {
	re.mu.Lock()
	defer re.mu.Unlock()

	re.rules = append(re.rules, rule)

	// Sort rules by priority (higher first)
	re.sortRules()
}

// RemoveRule removes a routing rule by name
func (re *RulesEngine) RemoveRule(name string) {
	re.mu.Lock()
	defer re.mu.Unlock()

	for i, rule := range re.rules {
		if rule.Name == name {
			re.rules = append(re.rules[:i], re.rules[i+1:]...)
			break
		}
	}
}

// GetRules returns all rules
func (re *RulesEngine) GetRules() []*RoutingRule {
	re.mu.RLock()
	defer re.mu.RUnlock()

	rules := make([]*RoutingRule, len(re.rules))
	copy(rules, re.rules)
	return rules
}

// Evaluate evaluates rules and selects a provider
func (re *RulesEngine) Evaluate(ctx context.Context, req *models.ChatRequest, providers []interfaces.LLMProvider) (interfaces.LLMProvider, error) {
	// Analyze the request
	analysis := re.analyzer.Analyze(req)

	re.mu.RLock()
	rules := make([]*RoutingRule, len(re.rules))
	copy(rules, re.rules)
	re.mu.RUnlock()

	// Evaluate rules in priority order
	for _, rule := range rules {
		if re.evaluateRule(ctx, rule, req, analysis) {
			// Rule matched, execute action
			provider, err := rule.Action.Execute(ctx, req, providers)
			if err == nil {
				return provider, nil
			}
			// If action failed, continue to next rule
		}
	}

	// No rule matched, use fallback strategy
	return re.fallback.SelectProvider(providers, req)
}

// evaluateRule evaluates a single rule
func (re *RulesEngine) evaluateRule(ctx context.Context, rule *RoutingRule, req *models.ChatRequest, analysis *QueryAnalysis) bool {
	// All conditions must be met
	for _, condition := range rule.Conditions {
		if !condition.Evaluate(ctx, req, analysis) {
			return false
		}
	}
	return len(rule.Conditions) > 0
}

// sortRules sorts rules by priority (descending)
func (re *RulesEngine) sortRules() {
	// Simple bubble sort since we typically have few rules
	n := len(re.rules)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if re.rules[j].Priority < re.rules[j+1].Priority {
				re.rules[j], re.rules[j+1] = re.rules[j+1], re.rules[j]
			}
		}
	}
}

// RuleBuilder helps build routing rules
type RuleBuilder struct {
	rule *RoutingRule
}

// NewRuleBuilder creates a new rule builder
func NewRuleBuilder(name string) *RuleBuilder {
	return &RuleBuilder{
		rule: &RoutingRule{
			Name:       name,
			Priority:   0,
			Conditions: make([]Condition, 0),
		},
	}
}

// WithPriority sets the rule priority
func (rb *RuleBuilder) WithPriority(priority int) *RuleBuilder {
	rb.rule.Priority = priority
	return rb
}

// WithComplexity adds a complexity condition
func (rb *RuleBuilder) WithComplexity(complexity QueryComplexity) *RuleBuilder {
	rb.rule.Conditions = append(rb.rule.Conditions, &ComplexityCondition{
		Complexity: complexity,
	})
	return rb
}

// WithCategory adds a category condition
func (rb *RuleBuilder) WithCategory(category string) *RuleBuilder {
	rb.rule.Conditions = append(rb.rule.Conditions, &CategoryCondition{
		Category: category,
	})
	return rb
}

// WithUser adds a user condition
func (rb *RuleBuilder) WithUser(userID string) *RuleBuilder {
	rb.rule.Conditions = append(rb.rule.Conditions, &UserCondition{
		UserID: userID,
	})
	return rb
}

// WithTenant adds a tenant condition
func (rb *RuleBuilder) WithTenant(tenantID string) *RuleBuilder {
	rb.rule.Conditions = append(rb.rule.Conditions, &TenantCondition{
		TenantID: tenantID,
	})
	return rb
}

// WithTokenRange adds a token limit condition
func (rb *RuleBuilder) WithTokenRange(min, max int) *RuleBuilder {
	rb.rule.Conditions = append(rb.rule.Conditions, &TokenLimitCondition{
		MinTokens: min,
		MaxTokens: max,
	})
	return rb
}

// WithTag adds a tag condition
func (rb *RuleBuilder) WithTag(tag string) *RuleBuilder {
	rb.rule.Conditions = append(rb.rule.Conditions, &TagCondition{
		Tag: tag,
	})
	return rb
}

// WithModel adds a model pattern condition
func (rb *RuleBuilder) WithModel(pattern string) *RuleBuilder {
	condition, err := NewModelCondition(pattern)
	if err == nil {
		rb.rule.Conditions = append(rb.rule.Conditions, condition)
	}
	return rb
}

// SelectProvider sets the action to select a specific provider
func (rb *RuleBuilder) SelectProvider(providerType models.ProviderType) *RuleBuilder {
	rb.rule.Action = &SelectProviderAction{
		ProviderType: providerType,
	}
	return rb
}

// UseStrategy sets the action to use a routing strategy
func (rb *RuleBuilder) UseStrategy(strategy RoutingStrategy) *RuleBuilder {
	rb.rule.Action = &UseStrategyAction{
		Strategy: strategy,
	}
	return rb
}

// Reject sets the action to reject the request
func (rb *RuleBuilder) Reject(reason string) *RuleBuilder {
	rb.rule.Action = &RejectAction{
		Reason: reason,
	}
	return rb
}

// Build builds and returns the routing rule
func (rb *RuleBuilder) Build() *RoutingRule {
	return rb.rule
}

// ParseRule parses a rule from a simple DSL
// Format: "IF <condition> [AND <condition>] THEN <action>"
// Example: "IF complexity:complex AND category:code THEN provider:openai"
func ParseRule(name string, ruleString string) (*RoutingRule, error) {
	parts := strings.Split(ruleString, " THEN ")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid rule format: must contain 'THEN'")
	}

	conditionStr := strings.TrimPrefix(parts[0], "IF ")
	actionStr := parts[1]

	rule := &RoutingRule{
		Name:       name,
		Priority:   0,
		Conditions: make([]Condition, 0),
	}

	// Parse conditions
	conditionParts := strings.Split(conditionStr, " AND ")
	for _, condPart := range conditionParts {
		condPart = strings.TrimSpace(condPart)
		parts := strings.SplitN(condPart, ":", 2)
		if len(parts) != 2 {
			continue
		}

		condType := parts[0]
		condValue := parts[1]

		switch condType {
		case "complexity":
			rule.Conditions = append(rule.Conditions, &ComplexityCondition{
				Complexity: QueryComplexity(condValue),
			})
		case "category":
			rule.Conditions = append(rule.Conditions, &CategoryCondition{
				Category: condValue,
			})
		case "user":
			rule.Conditions = append(rule.Conditions, &UserCondition{
				UserID: condValue,
			})
		case "tenant":
			rule.Conditions = append(rule.Conditions, &TenantCondition{
				TenantID: condValue,
			})
		}
	}

	// Parse action
	actionParts := strings.SplitN(actionStr, ":", 2)
	if len(actionParts) != 2 {
		return nil, fmt.Errorf("invalid action format")
	}

	actionType := actionParts[0]
	actionValue := actionParts[1]

	switch actionType {
	case "provider":
		rule.Action = &SelectProviderAction{
			ProviderType: models.ProviderType(actionValue),
		}
	case "reject":
		rule.Action = &RejectAction{
			Reason: actionValue,
		}
	}

	return rule, nil
}
