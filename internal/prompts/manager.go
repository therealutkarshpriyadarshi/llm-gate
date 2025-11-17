package prompts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Manager handles high-level prompt management operations
type Manager struct {
	repo       *Repository
	expRepo    *ExperimentRepository
	engine     *TemplateEngine
	abTesting  *ABTestingService
}

// NewManager creates a new prompt manager
func NewManager(pool *pgxpool.Pool) *Manager {
	repo := NewRepository(pool)
	expRepo := NewExperimentRepository(pool)
	engine := NewTemplateEngine()
	abTesting := NewABTestingService(expRepo)

	return &Manager{
		repo:      repo,
		expRepo:   expRepo,
		engine:    engine,
		abTesting: abTesting,
	}
}

// CreatePrompt creates a new prompt with validation
func (m *Manager) CreatePrompt(ctx context.Context, req CreatePromptRequest) (*Prompt, error) {
	// Validate template
	if err := m.engine.Validate(req.Template); err != nil {
		return nil, fmt.Errorf("invalid template: %w", err)
	}

	// Create prompt
	prompt, err := m.repo.CreatePrompt(ctx, req)
	if err != nil {
		return nil, err
	}

	return prompt, nil
}

// GetPrompt retrieves a prompt by ID or name
func (m *Manager) GetPrompt(ctx context.Context, identifier string) (*Prompt, error) {
	// Try parsing as UUID first
	if id, err := uuid.Parse(identifier); err == nil {
		return m.repo.GetPrompt(ctx, id)
	}

	// Otherwise treat as name
	return m.repo.GetPromptByName(ctx, identifier)
}

// UpdatePrompt updates a prompt and creates a new version
func (m *Manager) UpdatePrompt(ctx context.Context, identifier string, req UpdatePromptRequest) (*Prompt, error) {
	// Validate template if provided
	if req.Template != "" {
		if err := m.engine.Validate(req.Template); err != nil {
			return nil, fmt.Errorf("invalid template: %w", err)
		}
	}

	// Get prompt ID
	prompt, err := m.GetPrompt(ctx, identifier)
	if err != nil {
		return nil, err
	}

	// Update prompt
	return m.repo.UpdatePrompt(ctx, prompt.ID, req)
}

// ListPrompts lists prompts with optional filtering
func (m *Manager) ListPrompts(ctx context.Context, filter PromptFilter) ([]Prompt, error) {
	return m.repo.ListPrompts(ctx, filter)
}

// DeletePrompt soft deletes a prompt
func (m *Manager) DeletePrompt(ctx context.Context, identifier string) error {
	prompt, err := m.GetPrompt(ctx, identifier)
	if err != nil {
		return err
	}

	return m.repo.DeletePrompt(ctx, prompt.ID)
}

// RenderPrompt renders a prompt with variables
func (m *Manager) RenderPrompt(ctx context.Context, identifier string, variables map[string]any) (string, error) {
	prompt, err := m.GetPrompt(ctx, identifier)
	if err != nil {
		return "", err
	}

	// Render with validation
	return m.engine.RenderWithValidation(prompt.Template, variables, prompt.Variables)
}

// RenderPromptVersion renders a specific version of a prompt
func (m *Manager) RenderPromptVersion(ctx context.Context, identifier string, version int, variables map[string]any) (string, error) {
	prompt, err := m.GetPrompt(ctx, identifier)
	if err != nil {
		return "", err
	}

	// Get specific version
	pv, err := m.repo.GetPromptVersion(ctx, prompt.ID, version)
	if err != nil {
		return "", err
	}

	// Render with validation
	return m.engine.RenderWithValidation(pv.Template, variables, pv.Variables)
}

// GetPromptVersions retrieves all versions of a prompt
func (m *Manager) GetPromptVersions(ctx context.Context, identifier string) ([]PromptVersion, error) {
	prompt, err := m.GetPrompt(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return m.repo.ListPromptVersions(ctx, prompt.ID)
}

// DeployPrompt deploys a prompt version to an environment
func (m *Manager) DeployPrompt(ctx context.Context, req DeployPromptRequest) (*PromptDeployment, error) {
	return m.repo.DeployPrompt(ctx, req)
}

// GetDeployedPrompt retrieves the deployed version of a prompt
func (m *Manager) GetDeployedPrompt(ctx context.Context, promptName, environment string) (*PromptVersion, error) {
	return m.repo.GetDeployedPrompt(ctx, promptName, environment)
}

// RenderDeployedPrompt renders the deployed version of a prompt
func (m *Manager) RenderDeployedPrompt(ctx context.Context, promptName, environment string, variables map[string]any) (string, error) {
	pv, err := m.GetDeployedPrompt(ctx, promptName, environment)
	if err != nil {
		return "", err
	}

	return m.engine.RenderWithValidation(pv.Template, variables, pv.Variables)
}

// CreateExperiment creates a new A/B testing experiment
func (m *Manager) CreateExperiment(ctx context.Context, req CreateExperimentRequest) (*Experiment, error) {
	return m.expRepo.CreateExperiment(ctx, req)
}

// GetExperiment retrieves an experiment by ID or name
func (m *Manager) GetExperiment(ctx context.Context, identifier string) (*Experiment, error) {
	// Try parsing as UUID first
	if id, err := uuid.Parse(identifier); err == nil {
		return m.expRepo.GetExperiment(ctx, id)
	}

	// Otherwise treat as name
	return m.expRepo.GetExperimentByName(ctx, identifier)
}

// ListExperiments lists experiments with optional filtering
func (m *Manager) ListExperiments(ctx context.Context, filter ExperimentFilter) ([]Experiment, error) {
	return m.expRepo.ListExperiments(ctx, filter)
}

// StartExperiment starts an experiment
func (m *Manager) StartExperiment(ctx context.Context, identifier string) error {
	exp, err := m.GetExperiment(ctx, identifier)
	if err != nil {
		return err
	}

	if exp.Status != "draft" && exp.Status != "paused" {
		return fmt.Errorf("experiment must be in draft or paused state to start")
	}

	return m.expRepo.UpdateExperimentStatus(ctx, exp.ID, "running")
}

// PauseExperiment pauses a running experiment
func (m *Manager) PauseExperiment(ctx context.Context, identifier string) error {
	exp, err := m.GetExperiment(ctx, identifier)
	if err != nil {
		return err
	}

	if exp.Status != "running" {
		return fmt.Errorf("experiment must be running to pause")
	}

	return m.expRepo.UpdateExperimentStatus(ctx, exp.ID, "paused")
}

// CompleteExperiment marks an experiment as completed
func (m *Manager) CompleteExperiment(ctx context.Context, identifier string, winnerVariantID *uuid.UUID) error {
	exp, err := m.GetExperiment(ctx, identifier)
	if err != nil {
		return err
	}

	if winnerVariantID != nil {
		return m.expRepo.SetExperimentWinner(ctx, exp.ID, *winnerVariantID)
	}

	return m.expRepo.UpdateExperimentStatus(ctx, exp.ID, "completed")
}

// GetExperimentMetrics retrieves metrics for an experiment
func (m *Manager) GetExperimentMetrics(ctx context.Context, identifier string) (map[uuid.UUID]*ExperimentMetrics, error) {
	exp, err := m.GetExperiment(ctx, identifier)
	if err != nil {
		return nil, err
	}

	return m.expRepo.GetExperimentMetrics(ctx, exp.ID)
}

// SelectVariant selects a variant for a request based on A/B testing rules
func (m *Manager) SelectVariant(ctx context.Context, promptName, userID string) (*PromptVersion, *Variant, error) {
	// Get prompt
	prompt, err := m.repo.GetPromptByName(ctx, promptName)
	if err != nil {
		return nil, nil, err
	}

	// Check if there's an active experiment
	exp, err := m.expRepo.GetActiveExperimentForPrompt(ctx, prompt.ID)
	if err != nil {
		return nil, nil, err
	}

	// If no active experiment, return latest version
	if exp == nil {
		pv, err := m.repo.GetLatestPromptVersion(ctx, prompt.ID)
		return pv, nil, err
	}

	// Select variant using A/B testing service
	variant := m.abTesting.SelectVariant(exp, userID)

	// Get the prompt version for the selected variant
	var pv *PromptVersion
	err = m.repo.pool.QueryRow(ctx,
		"SELECT id, prompt_id, version, template, variables, model, temperature, max_tokens, top_p, metadata, change_log, created_at, created_by FROM prompt_versions WHERE id = $1",
		variant.VersionID,
	).Scan(&pv.ID, &pv.PromptID, &pv.Version, &pv.Template, &pv.Variables, &pv.Model, &pv.Temperature, &pv.MaxTokens, &pv.TopP, &pv.Metadata, &pv.ChangeLog, &pv.CreatedAt, &pv.CreatedBy)
	if err != nil {
		return nil, nil, fmt.Errorf("get variant version: %w", err)
	}

	return pv, &variant, nil
}

// RenderPromptWithABTesting renders a prompt with A/B testing support
func (m *Manager) RenderPromptWithABTesting(ctx context.Context, promptName, userID string, variables map[string]any) (string, *Variant, error) {
	// Select variant
	pv, variant, err := m.SelectVariant(ctx, promptName, userID)
	if err != nil {
		return "", nil, err
	}

	// Render prompt
	rendered, err := m.engine.RenderWithValidation(pv.Template, variables, pv.Variables)
	if err != nil {
		return "", variant, err
	}

	return rendered, variant, nil
}

// RecordExperimentMetric records a metric for an experiment variant
func (m *Manager) RecordExperimentMetric(ctx context.Context, variantID uuid.UUID, metricType string, value float64) error {
	return m.expRepo.IncrementVariantMetric(ctx, variantID, metricType, value)
}

// DiffVersions compares two versions of a prompt
func (m *Manager) DiffVersions(ctx context.Context, identifier string, version1, version2 int) (string, error) {
	prompt, err := m.GetPrompt(ctx, identifier)
	if err != nil {
		return "", err
	}

	pv1, err := m.repo.GetPromptVersion(ctx, prompt.ID, version1)
	if err != nil {
		return "", fmt.Errorf("get version %d: %w", version1, err)
	}

	pv2, err := m.repo.GetPromptVersion(ctx, prompt.ID, version2)
	if err != nil {
		return "", fmt.Errorf("get version %d: %w", version2, err)
	}

	// Simple diff (in production, use a proper diff library)
	diff := fmt.Sprintf("Version %d:\n%s\n\n---\n\nVersion %d:\n%s",
		version1, pv1.Template, version2, pv2.Template)

	return diff, nil
}
