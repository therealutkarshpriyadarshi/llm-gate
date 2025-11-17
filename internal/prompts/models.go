package prompts

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Prompt represents a prompt template with metadata
type Prompt struct {
	ID          uuid.UUID       `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Template    string          `json:"template"`
	Variables   []Variable      `json:"variables"`
	Model       string          `json:"model,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	CreatedBy   string          `json:"created_by,omitempty"`
	IsActive    bool            `json:"is_active"`
}

// Variable represents a template variable with metadata
type Variable struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // string, number, boolean, array, object
	Description  string `json:"description,omitempty"`
	Required     bool   `json:"required"`
	DefaultValue any    `json:"default_value,omitempty"`
}

// PromptVersion represents a specific version of a prompt
type PromptVersion struct {
	ID          uuid.UUID       `json:"id"`
	PromptID    uuid.UUID       `json:"prompt_id"`
	Version     int             `json:"version"`
	Template    string          `json:"template"`
	Variables   []Variable      `json:"variables"`
	Model       string          `json:"model,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	ChangeLog   string          `json:"change_log,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	CreatedBy   string          `json:"created_by,omitempty"`
}

// PromptDeployment represents a deployment of a prompt version to an environment
type PromptDeployment struct {
	ID                uuid.UUID  `json:"id"`
	PromptID          uuid.UUID  `json:"prompt_id"`
	VersionID         uuid.UUID  `json:"version_id"`
	Environment       string     `json:"environment"` // dev, staging, prod
	DeployedAt        time.Time  `json:"deployed_at"`
	DeployedBy        string     `json:"deployed_by,omitempty"`
	Status            string     `json:"status"` // active, inactive, rolled_back
	RollbackVersionID *uuid.UUID `json:"rollback_version_id,omitempty"`
}

// Experiment represents an A/B testing experiment
type Experiment struct {
	ID                uuid.UUID       `json:"id"`
	Name              string          `json:"name"`
	Description       string          `json:"description,omitempty"`
	PromptID          uuid.UUID       `json:"prompt_id"`
	Status            string          `json:"status"` // draft, running, paused, completed
	StartDate         *time.Time      `json:"start_date,omitempty"`
	EndDate           *time.Time      `json:"end_date,omitempty"`
	TrafficPercentage int             `json:"traffic_percentage"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	CreatedBy         string          `json:"created_by,omitempty"`
	WinnerVariantID   *uuid.UUID      `json:"winner_variant_id,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
	Variants          []Variant       `json:"variants,omitempty"`
}

// Variant represents a variant in an A/B test
type Variant struct {
	ID           uuid.UUID `json:"id"`
	ExperimentID uuid.UUID `json:"experiment_id"`
	Name         string    `json:"name"`
	VersionID    uuid.UUID `json:"version_id"`
	TrafficSplit int       `json:"traffic_split"`
	CreatedAt    time.Time `json:"created_at"`
}

// ExperimentMetrics represents performance metrics for a variant
type ExperimentMetrics struct {
	ID                    uuid.UUID       `json:"id"`
	ExperimentID          uuid.UUID       `json:"experiment_id"`
	VariantID             uuid.UUID       `json:"variant_id"`
	RequestCount          int             `json:"request_count"`
	SuccessCount          int             `json:"success_count"`
	ErrorCount            int             `json:"error_count"`
	TotalTokens           int             `json:"total_tokens"`
	TotalCost             float64         `json:"total_cost"`
	AvgLatencyMs          float64         `json:"avg_latency_ms"`
	UserFeedbackPositive  int             `json:"user_feedback_positive"`
	UserFeedbackNegative  int             `json:"user_feedback_negative"`
	RecordedAt            time.Time       `json:"recorded_at"`
	Metadata              json.RawMessage `json:"metadata,omitempty"`
}

// CreatePromptRequest represents a request to create a new prompt
type CreatePromptRequest struct {
	Name        string          `json:"name" binding:"required"`
	Description string          `json:"description,omitempty"`
	Template    string          `json:"template" binding:"required"`
	Variables   []Variable      `json:"variables,omitempty"`
	Model       string          `json:"model,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	CreatedBy   string          `json:"created_by,omitempty"`
}

// UpdatePromptRequest represents a request to update a prompt
type UpdatePromptRequest struct {
	Description string          `json:"description,omitempty"`
	Template    string          `json:"template,omitempty"`
	Variables   []Variable      `json:"variables,omitempty"`
	Model       string          `json:"model,omitempty"`
	Temperature *float64        `json:"temperature,omitempty"`
	MaxTokens   *int            `json:"max_tokens,omitempty"`
	TopP        *float64        `json:"top_p,omitempty"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	ChangeLog   string          `json:"change_log,omitempty"`
	UpdatedBy   string          `json:"updated_by,omitempty"`
}

// RenderPromptRequest represents a request to render a prompt with variables
type RenderPromptRequest struct {
	Variables map[string]any `json:"variables" binding:"required"`
}

// DeployPromptRequest represents a request to deploy a prompt
type DeployPromptRequest struct {
	PromptName  string `json:"prompt_name" binding:"required"`
	Version     int    `json:"version,omitempty"` // 0 or omitted = latest version
	Environment string `json:"environment" binding:"required"`
	DeployedBy  string `json:"deployed_by,omitempty"`
}

// CreateExperimentRequest represents a request to create an experiment
type CreateExperimentRequest struct {
	Name              string          `json:"name" binding:"required"`
	Description       string          `json:"description,omitempty"`
	PromptID          uuid.UUID       `json:"prompt_id" binding:"required"`
	TrafficPercentage int             `json:"traffic_percentage"`
	Variants          []VariantConfig `json:"variants" binding:"required,min=2"`
	StartDate         *time.Time      `json:"start_date,omitempty"`
	EndDate           *time.Time      `json:"end_date,omitempty"`
	CreatedBy         string          `json:"created_by,omitempty"`
	Metadata          json.RawMessage `json:"metadata,omitempty"`
}

// VariantConfig represents a variant configuration in an experiment
type VariantConfig struct {
	Name         string `json:"name" binding:"required"`
	Version      int    `json:"version" binding:"required"`
	TrafficSplit int    `json:"traffic_split" binding:"required"`
}

// PromptFilter represents filters for querying prompts
type PromptFilter struct {
	Name     string   `json:"name,omitempty"`
	Tags     []string `json:"tags,omitempty"`
	IsActive *bool    `json:"is_active,omitempty"`
	Limit    int      `json:"limit,omitempty"`
	Offset   int      `json:"offset,omitempty"`
}

// ExperimentFilter represents filters for querying experiments
type ExperimentFilter struct {
	Status   string    `json:"status,omitempty"`
	PromptID uuid.UUID `json:"prompt_id,omitempty"`
	Limit    int       `json:"limit,omitempty"`
	Offset   int       `json:"offset,omitempty"`
}
