package prompts

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles database operations for prompts
type Repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new prompt repository
func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

// CreatePrompt creates a new prompt and its initial version
func (r *Repository) CreatePrompt(ctx context.Context, req CreatePromptRequest) (*Prompt, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create prompt
	prompt := &Prompt{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		Template:    req.Template,
		Variables:   req.Variables,
		Model:       req.Model,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Metadata:    req.Metadata,
		Tags:        req.Tags,
		CreatedBy:   req.CreatedBy,
		IsActive:    true,
	}

	variablesJSON, err := json.Marshal(prompt.Variables)
	if err != nil {
		return nil, fmt.Errorf("marshal variables: %w", err)
	}

	query := `
		INSERT INTO prompts (id, name, description, template, variables, model,
			temperature, max_tokens, top_p, metadata, tags, created_by, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		prompt.ID, prompt.Name, prompt.Description, prompt.Template, variablesJSON,
		prompt.Model, prompt.Temperature, prompt.MaxTokens, prompt.TopP,
		prompt.Metadata, prompt.Tags, prompt.CreatedBy, prompt.IsActive,
	).Scan(&prompt.CreatedAt, &prompt.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert prompt: %w", err)
	}

	// Create initial version
	versionQuery := `
		INSERT INTO prompt_versions (id, prompt_id, version, template, variables,
			model, temperature, max_tokens, top_p, metadata, change_log, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.Exec(ctx, versionQuery,
		uuid.New(), prompt.ID, 1, prompt.Template, variablesJSON,
		prompt.Model, prompt.Temperature, prompt.MaxTokens, prompt.TopP,
		prompt.Metadata, "Initial version", prompt.CreatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("insert prompt version: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return prompt, nil
}

// GetPrompt retrieves a prompt by ID
func (r *Repository) GetPrompt(ctx context.Context, id uuid.UUID) (*Prompt, error) {
	query := `
		SELECT id, name, description, template, variables, model, temperature,
			max_tokens, top_p, metadata, tags, created_at, updated_at, created_by, is_active
		FROM prompts
		WHERE id = $1
	`

	var prompt Prompt
	var variablesJSON []byte

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Template, &variablesJSON,
		&prompt.Model, &prompt.Temperature, &prompt.MaxTokens, &prompt.TopP,
		&prompt.Metadata, &prompt.Tags, &prompt.CreatedAt, &prompt.UpdatedAt,
		&prompt.CreatedBy, &prompt.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("prompt not found")
		}
		return nil, fmt.Errorf("query prompt: %w", err)
	}

	if err = json.Unmarshal(variablesJSON, &prompt.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &prompt, nil
}

// GetPromptByName retrieves a prompt by name
func (r *Repository) GetPromptByName(ctx context.Context, name string) (*Prompt, error) {
	query := `
		SELECT id, name, description, template, variables, model, temperature,
			max_tokens, top_p, metadata, tags, created_at, updated_at, created_by, is_active
		FROM prompts
		WHERE name = $1
	`

	var prompt Prompt
	var variablesJSON []byte

	err := r.pool.QueryRow(ctx, query, name).Scan(
		&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Template, &variablesJSON,
		&prompt.Model, &prompt.Temperature, &prompt.MaxTokens, &prompt.TopP,
		&prompt.Metadata, &prompt.Tags, &prompt.CreatedAt, &prompt.UpdatedAt,
		&prompt.CreatedBy, &prompt.IsActive,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("prompt not found")
		}
		return nil, fmt.Errorf("query prompt: %w", err)
	}

	if err = json.Unmarshal(variablesJSON, &prompt.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &prompt, nil
}

// ListPrompts retrieves prompts based on filters
func (r *Repository) ListPrompts(ctx context.Context, filter PromptFilter) ([]Prompt, error) {
	query := `
		SELECT id, name, description, template, variables, model, temperature,
			max_tokens, top_p, metadata, tags, created_at, updated_at, created_by, is_active
		FROM prompts
		WHERE 1=1
	`
	args := []any{}
	argPos := 1

	if filter.Name != "" {
		query += fmt.Sprintf(" AND name ILIKE $%d", argPos)
		args = append(args, "%"+filter.Name+"%")
		argPos++
	}

	if len(filter.Tags) > 0 {
		query += fmt.Sprintf(" AND tags && $%d", argPos)
		args = append(args, filter.Tags)
		argPos++
	}

	if filter.IsActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argPos)
		args = append(args, *filter.IsActive)
		argPos++
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filter.Limit)
		argPos++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filter.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query prompts: %w", err)
	}
	defer rows.Close()

	prompts := []Prompt{}
	for rows.Next() {
		var prompt Prompt
		var variablesJSON []byte

		err := rows.Scan(
			&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Template, &variablesJSON,
			&prompt.Model, &prompt.Temperature, &prompt.MaxTokens, &prompt.TopP,
			&prompt.Metadata, &prompt.Tags, &prompt.CreatedAt, &prompt.UpdatedAt,
			&prompt.CreatedBy, &prompt.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("scan prompt: %w", err)
		}

		if err = json.Unmarshal(variablesJSON, &prompt.Variables); err != nil {
			return nil, fmt.Errorf("unmarshal variables: %w", err)
		}

		prompts = append(prompts, prompt)
	}

	return prompts, nil
}

// UpdatePrompt updates a prompt and creates a new version
func (r *Repository) UpdatePrompt(ctx context.Context, id uuid.UUID, req UpdatePromptRequest) (*Prompt, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Get current prompt
	prompt, err := r.GetPrompt(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update prompt fields
	if req.Description != "" {
		prompt.Description = req.Description
	}
	if req.Template != "" {
		prompt.Template = req.Template
	}
	if req.Variables != nil {
		prompt.Variables = req.Variables
	}
	if req.Model != "" {
		prompt.Model = req.Model
	}
	if req.Temperature != nil {
		prompt.Temperature = req.Temperature
	}
	if req.MaxTokens != nil {
		prompt.MaxTokens = req.MaxTokens
	}
	if req.TopP != nil {
		prompt.TopP = req.TopP
	}
	if req.Metadata != nil {
		prompt.Metadata = req.Metadata
	}
	if req.Tags != nil {
		prompt.Tags = req.Tags
	}

	variablesJSON, err := json.Marshal(prompt.Variables)
	if err != nil {
		return nil, fmt.Errorf("marshal variables: %w", err)
	}

	// Update prompt
	updateQuery := `
		UPDATE prompts
		SET description = $1, template = $2, variables = $3, model = $4,
			temperature = $5, max_tokens = $6, top_p = $7, metadata = $8, tags = $9
		WHERE id = $10
		RETURNING updated_at
	`

	err = tx.QueryRow(ctx, updateQuery,
		prompt.Description, prompt.Template, variablesJSON, prompt.Model,
		prompt.Temperature, prompt.MaxTokens, prompt.TopP, prompt.Metadata,
		prompt.Tags, id,
	).Scan(&prompt.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("update prompt: %w", err)
	}

	// Get next version number
	var nextVersion int
	err = tx.QueryRow(ctx,
		"SELECT COALESCE(MAX(version), 0) + 1 FROM prompt_versions WHERE prompt_id = $1",
		id,
	).Scan(&nextVersion)
	if err != nil {
		return nil, fmt.Errorf("get next version: %w", err)
	}

	// Create new version
	versionQuery := `
		INSERT INTO prompt_versions (id, prompt_id, version, template, variables,
			model, temperature, max_tokens, top_p, metadata, change_log, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = tx.Exec(ctx, versionQuery,
		uuid.New(), id, nextVersion, prompt.Template, variablesJSON,
		prompt.Model, prompt.Temperature, prompt.MaxTokens, prompt.TopP,
		prompt.Metadata, req.ChangeLog, req.UpdatedBy,
	)
	if err != nil {
		return nil, fmt.Errorf("insert prompt version: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return prompt, nil
}

// DeletePrompt soft deletes a prompt by setting is_active to false
func (r *Repository) DeletePrompt(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE prompts SET is_active = false WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete prompt: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("prompt not found")
	}

	return nil
}

// GetPromptVersion retrieves a specific version of a prompt
func (r *Repository) GetPromptVersion(ctx context.Context, promptID uuid.UUID, version int) (*PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, template, variables, model, temperature,
			max_tokens, top_p, metadata, change_log, created_at, created_by
		FROM prompt_versions
		WHERE prompt_id = $1 AND version = $2
	`

	var pv PromptVersion
	var variablesJSON []byte

	err := r.pool.QueryRow(ctx, query, promptID, version).Scan(
		&pv.ID, &pv.PromptID, &pv.Version, &pv.Template, &variablesJSON,
		&pv.Model, &pv.Temperature, &pv.MaxTokens, &pv.TopP,
		&pv.Metadata, &pv.ChangeLog, &pv.CreatedAt, &pv.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("prompt version not found")
		}
		return nil, fmt.Errorf("query prompt version: %w", err)
	}

	if err = json.Unmarshal(variablesJSON, &pv.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &pv, nil
}

// GetLatestPromptVersion retrieves the latest version of a prompt
func (r *Repository) GetLatestPromptVersion(ctx context.Context, promptID uuid.UUID) (*PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, template, variables, model, temperature,
			max_tokens, top_p, metadata, change_log, created_at, created_by
		FROM prompt_versions
		WHERE prompt_id = $1
		ORDER BY version DESC
		LIMIT 1
	`

	var pv PromptVersion
	var variablesJSON []byte

	err := r.pool.QueryRow(ctx, query, promptID).Scan(
		&pv.ID, &pv.PromptID, &pv.Version, &pv.Template, &variablesJSON,
		&pv.Model, &pv.Temperature, &pv.MaxTokens, &pv.TopP,
		&pv.Metadata, &pv.ChangeLog, &pv.CreatedAt, &pv.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("prompt version not found")
		}
		return nil, fmt.Errorf("query prompt version: %w", err)
	}

	if err = json.Unmarshal(variablesJSON, &pv.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &pv, nil
}

// ListPromptVersions retrieves all versions of a prompt
func (r *Repository) ListPromptVersions(ctx context.Context, promptID uuid.UUID) ([]PromptVersion, error) {
	query := `
		SELECT id, prompt_id, version, template, variables, model, temperature,
			max_tokens, top_p, metadata, change_log, created_at, created_by
		FROM prompt_versions
		WHERE prompt_id = $1
		ORDER BY version DESC
	`

	rows, err := r.pool.Query(ctx, query, promptID)
	if err != nil {
		return nil, fmt.Errorf("query prompt versions: %w", err)
	}
	defer rows.Close()

	versions := []PromptVersion{}
	for rows.Next() {
		var pv PromptVersion
		var variablesJSON []byte

		err := rows.Scan(
			&pv.ID, &pv.PromptID, &pv.Version, &pv.Template, &variablesJSON,
			&pv.Model, &pv.Temperature, &pv.MaxTokens, &pv.TopP,
			&pv.Metadata, &pv.ChangeLog, &pv.CreatedAt, &pv.CreatedBy,
		)
		if err != nil {
			return nil, fmt.Errorf("scan prompt version: %w", err)
		}

		if err = json.Unmarshal(variablesJSON, &pv.Variables); err != nil {
			return nil, fmt.Errorf("unmarshal variables: %w", err)
		}

		versions = append(versions, pv)
	}

	return versions, nil
}

// DeployPrompt deploys a prompt version to an environment
func (r *Repository) DeployPrompt(ctx context.Context, req DeployPromptRequest) (*PromptDeployment, error) {
	// Get prompt by name
	prompt, err := r.GetPromptByName(ctx, req.PromptName)
	if err != nil {
		return nil, err
	}

	// Get version (latest if not specified)
	var version *PromptVersion
	if req.Version > 0 {
		version, err = r.GetPromptVersion(ctx, prompt.ID, req.Version)
	} else {
		version, err = r.GetLatestPromptVersion(ctx, prompt.ID)
	}
	if err != nil {
		return nil, err
	}

	// Upsert deployment
	query := `
		INSERT INTO prompt_deployments (id, prompt_id, version_id, environment, deployed_by, status)
		VALUES ($1, $2, $3, $4, $5, 'active')
		ON CONFLICT (prompt_id, environment)
		DO UPDATE SET version_id = $3, deployed_at = CURRENT_TIMESTAMP,
			deployed_by = $5, status = 'active'
		RETURNING id, deployed_at
	`

	deployment := &PromptDeployment{
		ID:          uuid.New(),
		PromptID:    prompt.ID,
		VersionID:   version.ID,
		Environment: req.Environment,
		DeployedBy:  req.DeployedBy,
		Status:      "active",
	}

	err = r.pool.QueryRow(ctx, query,
		deployment.ID, deployment.PromptID, deployment.VersionID,
		deployment.Environment, deployment.DeployedBy,
	).Scan(&deployment.ID, &deployment.DeployedAt)
	if err != nil {
		return nil, fmt.Errorf("deploy prompt: %w", err)
	}

	return deployment, nil
}

// GetDeployedPrompt retrieves the deployed version of a prompt for an environment
func (r *Repository) GetDeployedPrompt(ctx context.Context, promptName, environment string) (*PromptVersion, error) {
	query := `
		SELECT pv.id, pv.prompt_id, pv.version, pv.template, pv.variables, pv.model,
			pv.temperature, pv.max_tokens, pv.top_p, pv.metadata, pv.change_log,
			pv.created_at, pv.created_by
		FROM prompt_versions pv
		JOIN prompt_deployments pd ON pv.id = pd.version_id
		JOIN prompts p ON pv.prompt_id = p.id
		WHERE p.name = $1 AND pd.environment = $2 AND pd.status = 'active'
	`

	var pv PromptVersion
	var variablesJSON []byte

	err := r.pool.QueryRow(ctx, query, promptName, environment).Scan(
		&pv.ID, &pv.PromptID, &pv.Version, &pv.Template, &variablesJSON,
		&pv.Model, &pv.Temperature, &pv.MaxTokens, &pv.TopP,
		&pv.Metadata, &pv.ChangeLog, &pv.CreatedAt, &pv.CreatedBy,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("no deployed version found")
		}
		return nil, fmt.Errorf("query deployed prompt: %w", err)
	}

	if err = json.Unmarshal(variablesJSON, &pv.Variables); err != nil {
		return nil, fmt.Errorf("unmarshal variables: %w", err)
	}

	return &pv, nil
}
