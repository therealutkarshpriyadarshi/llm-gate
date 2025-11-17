package prompts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ExperimentRepository handles database operations for experiments
type ExperimentRepository struct {
	pool *pgxpool.Pool
}

// NewExperimentRepository creates a new experiment repository
func NewExperimentRepository(pool *pgxpool.Pool) *ExperimentRepository {
	return &ExperimentRepository{pool: pool}
}

// CreateExperiment creates a new A/B testing experiment
func (r *ExperimentRepository) CreateExperiment(ctx context.Context, req CreateExperimentRequest) (*Experiment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Validate traffic splits sum to 100
	totalSplit := 0
	for _, v := range req.Variants {
		totalSplit += v.TrafficSplit
	}
	if totalSplit != 100 {
		return nil, fmt.Errorf("traffic splits must sum to 100, got %d", totalSplit)
	}

	// Create experiment
	exp := &Experiment{
		ID:                uuid.New(),
		Name:              req.Name,
		Description:       req.Description,
		PromptID:          req.PromptID,
		Status:            "draft",
		TrafficPercentage: req.TrafficPercentage,
		StartDate:         req.StartDate,
		EndDate:           req.EndDate,
		CreatedBy:         req.CreatedBy,
		Metadata:          req.Metadata,
	}

	if exp.TrafficPercentage == 0 {
		exp.TrafficPercentage = 100
	}

	query := `
		INSERT INTO experiments (id, name, description, prompt_id, status,
			start_date, end_date, traffic_percentage, created_by, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		exp.ID, exp.Name, exp.Description, exp.PromptID, exp.Status,
		exp.StartDate, exp.EndDate, exp.TrafficPercentage, exp.CreatedBy, exp.Metadata,
	).Scan(&exp.CreatedAt, &exp.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert experiment: %w", err)
	}

	// Create variants
	variantQuery := `
		INSERT INTO experiment_variants (id, experiment_id, name, version_id, traffic_split)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	exp.Variants = make([]Variant, 0, len(req.Variants))
	for _, vc := range req.Variants {
		// Get version ID
		var versionID uuid.UUID
		err = tx.QueryRow(ctx,
			"SELECT id FROM prompt_versions WHERE prompt_id = $1 AND version = $2",
			req.PromptID, vc.Version,
		).Scan(&versionID)
		if err != nil {
			if err == pgx.ErrNoRows {
				return nil, fmt.Errorf("version %d not found for prompt", vc.Version)
			}
			return nil, fmt.Errorf("get version ID: %w", err)
		}

		variant := Variant{
			ID:           uuid.New(),
			ExperimentID: exp.ID,
			Name:         vc.Name,
			VersionID:    versionID,
			TrafficSplit: vc.TrafficSplit,
		}

		err = tx.QueryRow(ctx, variantQuery,
			variant.ID, variant.ExperimentID, variant.Name, variant.VersionID, variant.TrafficSplit,
		).Scan(&variant.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("insert variant: %w", err)
		}

		exp.Variants = append(exp.Variants, variant)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}

	return exp, nil
}

// GetExperiment retrieves an experiment by ID
func (r *ExperimentRepository) GetExperiment(ctx context.Context, id uuid.UUID) (*Experiment, error) {
	query := `
		SELECT id, name, description, prompt_id, status, start_date, end_date,
			traffic_percentage, created_at, updated_at, created_by, winner_variant_id, metadata
		FROM experiments
		WHERE id = $1
	`

	var exp Experiment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&exp.ID, &exp.Name, &exp.Description, &exp.PromptID, &exp.Status,
		&exp.StartDate, &exp.EndDate, &exp.TrafficPercentage,
		&exp.CreatedAt, &exp.UpdatedAt, &exp.CreatedBy, &exp.WinnerVariantID, &exp.Metadata,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("experiment not found")
		}
		return nil, fmt.Errorf("query experiment: %w", err)
	}

	// Get variants
	variants, err := r.GetExperimentVariants(ctx, id)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants

	return &exp, nil
}

// GetExperimentByName retrieves an experiment by name
func (r *ExperimentRepository) GetExperimentByName(ctx context.Context, name string) (*Experiment, error) {
	query := `
		SELECT id, name, description, prompt_id, status, start_date, end_date,
			traffic_percentage, created_at, updated_at, created_by, winner_variant_id, metadata
		FROM experiments
		WHERE name = $1
	`

	var exp Experiment
	err := r.pool.QueryRow(ctx, query, name).Scan(
		&exp.ID, &exp.Name, &exp.Description, &exp.PromptID, &exp.Status,
		&exp.StartDate, &exp.EndDate, &exp.TrafficPercentage,
		&exp.CreatedAt, &exp.UpdatedAt, &exp.CreatedBy, &exp.WinnerVariantID, &exp.Metadata,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("experiment not found")
		}
		return nil, fmt.Errorf("query experiment: %w", err)
	}

	// Get variants
	variants, err := r.GetExperimentVariants(ctx, exp.ID)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants

	return &exp, nil
}

// ListExperiments retrieves experiments based on filters
func (r *ExperimentRepository) ListExperiments(ctx context.Context, filter ExperimentFilter) ([]Experiment, error) {
	query := `
		SELECT id, name, description, prompt_id, status, start_date, end_date,
			traffic_percentage, created_at, updated_at, created_by, winner_variant_id, metadata
		FROM experiments
		WHERE 1=1
	`
	args := []any{}
	argPos := 1

	if filter.Status != "" {
		query += fmt.Sprintf(" AND status = $%d", argPos)
		args = append(args, filter.Status)
		argPos++
	}

	if filter.PromptID != uuid.Nil {
		query += fmt.Sprintf(" AND prompt_id = $%d", argPos)
		args = append(args, filter.PromptID)
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
		return nil, fmt.Errorf("query experiments: %w", err)
	}
	defer rows.Close()

	experiments := []Experiment{}
	for rows.Next() {
		var exp Experiment
		err := rows.Scan(
			&exp.ID, &exp.Name, &exp.Description, &exp.PromptID, &exp.Status,
			&exp.StartDate, &exp.EndDate, &exp.TrafficPercentage,
			&exp.CreatedAt, &exp.UpdatedAt, &exp.CreatedBy, &exp.WinnerVariantID, &exp.Metadata,
		)
		if err != nil {
			return nil, fmt.Errorf("scan experiment: %w", err)
		}

		// Get variants for each experiment
		variants, err := r.GetExperimentVariants(ctx, exp.ID)
		if err != nil {
			return nil, err
		}
		exp.Variants = variants

		experiments = append(experiments, exp)
	}

	return experiments, nil
}

// UpdateExperimentStatus updates the status of an experiment
func (r *ExperimentRepository) UpdateExperimentStatus(ctx context.Context, id uuid.UUID, status string) error {
	query := `UPDATE experiments SET status = $1 WHERE id = $2`
	result, err := r.pool.Exec(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("update experiment status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// SetExperimentWinner sets the winner variant for an experiment
func (r *ExperimentRepository) SetExperimentWinner(ctx context.Context, experimentID, variantID uuid.UUID) error {
	query := `
		UPDATE experiments
		SET winner_variant_id = $1, status = 'completed'
		WHERE id = $2
	`
	result, err := r.pool.Exec(ctx, query, variantID, experimentID)
	if err != nil {
		return fmt.Errorf("set experiment winner: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("experiment not found")
	}

	return nil
}

// GetExperimentVariants retrieves all variants for an experiment
func (r *ExperimentRepository) GetExperimentVariants(ctx context.Context, experimentID uuid.UUID) ([]Variant, error) {
	query := `
		SELECT id, experiment_id, name, version_id, traffic_split, created_at
		FROM experiment_variants
		WHERE experiment_id = $1
		ORDER BY traffic_split DESC
	`

	rows, err := r.pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("query variants: %w", err)
	}
	defer rows.Close()

	variants := []Variant{}
	for rows.Next() {
		var v Variant
		err := rows.Scan(&v.ID, &v.ExperimentID, &v.Name, &v.VersionID, &v.TrafficSplit, &v.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan variant: %w", err)
		}
		variants = append(variants, v)
	}

	return variants, nil
}

// RecordMetrics records performance metrics for a variant
func (r *ExperimentRepository) RecordMetrics(ctx context.Context, experimentID, variantID uuid.UUID, metrics ExperimentMetrics) error {
	query := `
		INSERT INTO experiment_metrics (id, experiment_id, variant_id, request_count,
			success_count, error_count, total_tokens, total_cost, avg_latency_ms,
			user_feedback_positive, user_feedback_negative, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := r.pool.Exec(ctx, query,
		uuid.New(), experimentID, variantID, metrics.RequestCount,
		metrics.SuccessCount, metrics.ErrorCount, metrics.TotalTokens,
		metrics.TotalCost, metrics.AvgLatencyMs, metrics.UserFeedbackPositive,
		metrics.UserFeedbackNegative, metrics.Metadata,
	)
	if err != nil {
		return fmt.Errorf("record metrics: %w", err)
	}

	return nil
}

// GetVariantMetrics retrieves aggregated metrics for a variant
func (r *ExperimentRepository) GetVariantMetrics(ctx context.Context, variantID uuid.UUID) (*ExperimentMetrics, error) {
	query := `
		SELECT
			COALESCE(SUM(request_count), 0) as request_count,
			COALESCE(SUM(success_count), 0) as success_count,
			COALESCE(SUM(error_count), 0) as error_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(AVG(avg_latency_ms), 0) as avg_latency_ms,
			COALESCE(SUM(user_feedback_positive), 0) as user_feedback_positive,
			COALESCE(SUM(user_feedback_negative), 0) as user_feedback_negative
		FROM experiment_metrics
		WHERE variant_id = $1
	`

	var metrics ExperimentMetrics
	metrics.VariantID = variantID

	err := r.pool.QueryRow(ctx, query, variantID).Scan(
		&metrics.RequestCount, &metrics.SuccessCount, &metrics.ErrorCount,
		&metrics.TotalTokens, &metrics.TotalCost, &metrics.AvgLatencyMs,
		&metrics.UserFeedbackPositive, &metrics.UserFeedbackNegative,
	)
	if err != nil {
		return nil, fmt.Errorf("query variant metrics: %w", err)
	}

	return &metrics, nil
}

// GetExperimentMetrics retrieves aggregated metrics for all variants in an experiment
func (r *ExperimentRepository) GetExperimentMetrics(ctx context.Context, experimentID uuid.UUID) (map[uuid.UUID]*ExperimentMetrics, error) {
	query := `
		SELECT
			variant_id,
			COALESCE(SUM(request_count), 0) as request_count,
			COALESCE(SUM(success_count), 0) as success_count,
			COALESCE(SUM(error_count), 0) as error_count,
			COALESCE(SUM(total_tokens), 0) as total_tokens,
			COALESCE(SUM(total_cost), 0) as total_cost,
			COALESCE(AVG(avg_latency_ms), 0) as avg_latency_ms,
			COALESCE(SUM(user_feedback_positive), 0) as user_feedback_positive,
			COALESCE(SUM(user_feedback_negative), 0) as user_feedback_negative
		FROM experiment_metrics
		WHERE experiment_id = $1
		GROUP BY variant_id
	`

	rows, err := r.pool.Query(ctx, query, experimentID)
	if err != nil {
		return nil, fmt.Errorf("query experiment metrics: %w", err)
	}
	defer rows.Close()

	metricsMap := make(map[uuid.UUID]*ExperimentMetrics)
	for rows.Next() {
		var metrics ExperimentMetrics
		metrics.ExperimentID = experimentID

		err := rows.Scan(
			&metrics.VariantID, &metrics.RequestCount, &metrics.SuccessCount,
			&metrics.ErrorCount, &metrics.TotalTokens, &metrics.TotalCost,
			&metrics.AvgLatencyMs, &metrics.UserFeedbackPositive, &metrics.UserFeedbackNegative,
		)
		if err != nil {
			return nil, fmt.Errorf("scan metrics: %w", err)
		}

		metricsMap[metrics.VariantID] = &metrics
	}

	return metricsMap, nil
}

// GetActiveExperimentForPrompt retrieves the active experiment for a prompt
func (r *ExperimentRepository) GetActiveExperimentForPrompt(ctx context.Context, promptID uuid.UUID) (*Experiment, error) {
	query := `
		SELECT id, name, description, prompt_id, status, start_date, end_date,
			traffic_percentage, created_at, updated_at, created_by, winner_variant_id, metadata
		FROM experiments
		WHERE prompt_id = $1 AND status = 'running'
		LIMIT 1
	`

	var exp Experiment
	err := r.pool.QueryRow(ctx, query, promptID).Scan(
		&exp.ID, &exp.Name, &exp.Description, &exp.PromptID, &exp.Status,
		&exp.StartDate, &exp.EndDate, &exp.TrafficPercentage,
		&exp.CreatedAt, &exp.UpdatedAt, &exp.CreatedBy, &exp.WinnerVariantID, &exp.Metadata,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No active experiment
		}
		return nil, fmt.Errorf("query active experiment: %w", err)
	}

	// Get variants
	variants, err := r.GetExperimentVariants(ctx, exp.ID)
	if err != nil {
		return nil, err
	}
	exp.Variants = variants

	return &exp, nil
}

// IncrementVariantMetric atomically increments a specific metric for a variant
func (r *ExperimentRepository) IncrementVariantMetric(ctx context.Context, variantID uuid.UUID, metricType string, value float64) error {
	// Create or update metrics entry for today
	query := `
		INSERT INTO experiment_metrics (id, experiment_id, variant_id, request_count,
			success_count, error_count, total_tokens, total_cost, avg_latency_ms,
			user_feedback_positive, user_feedback_negative, recorded_at, metadata)
		SELECT
			$1,
			ev.experiment_id,
			$2,
			CASE WHEN $3 = 'request' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'success' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'error' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'tokens' THEN $4 ELSE 0 END,
			CASE WHEN $3 = 'cost' THEN $4 ELSE 0 END,
			CASE WHEN $3 = 'latency' THEN $4 ELSE 0 END,
			CASE WHEN $3 = 'feedback_positive' THEN 1 ELSE 0 END,
			CASE WHEN $3 = 'feedback_negative' THEN 1 ELSE 0 END,
			CURRENT_DATE,
			'{}'::jsonb
		FROM experiment_variants ev
		WHERE ev.id = $2
		ON CONFLICT (experiment_id, variant_id, recorded_at)
		DO UPDATE SET
			request_count = experiment_metrics.request_count + CASE WHEN $3 = 'request' THEN 1 ELSE 0 END,
			success_count = experiment_metrics.success_count + CASE WHEN $3 = 'success' THEN 1 ELSE 0 END,
			error_count = experiment_metrics.error_count + CASE WHEN $3 = 'error' THEN 1 ELSE 0 END,
			total_tokens = experiment_metrics.total_tokens + CASE WHEN $3 = 'tokens' THEN $4 ELSE 0 END,
			total_cost = experiment_metrics.total_cost + CASE WHEN $3 = 'cost' THEN $4 ELSE 0 END,
			avg_latency_ms = CASE
				WHEN $3 = 'latency' THEN
					(experiment_metrics.avg_latency_ms * experiment_metrics.request_count + $4) /
					(experiment_metrics.request_count + 1)
				ELSE experiment_metrics.avg_latency_ms
			END,
			user_feedback_positive = experiment_metrics.user_feedback_positive + CASE WHEN $3 = 'feedback_positive' THEN 1 ELSE 0 END,
			user_feedback_negative = experiment_metrics.user_feedback_negative + CASE WHEN $3 = 'feedback_negative' THEN 1 ELSE 0 END
	`

	// Add unique constraint on (experiment_id, variant_id, recorded_at) to the schema
	_, err := r.pool.Exec(ctx, query, uuid.New(), variantID, metricType, value)
	if err != nil {
		return fmt.Errorf("increment variant metric: %w", err)
	}

	return nil
}
