-- Migration: Create Prompt Management Tables
-- Version: 001
-- Description: Creates tables for prompt storage, versioning, and A/B testing

-- Prompts table: stores prompt templates and metadata
CREATE TABLE IF NOT EXISTS prompts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    template TEXT NOT NULL,
    variables JSONB DEFAULT '[]'::jsonb,
    model VARCHAR(100),
    temperature DECIMAL(3,2),
    max_tokens INTEGER,
    top_p DECIMAL(3,2),
    metadata JSONB DEFAULT '{}'::jsonb,
    tags TEXT[],
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    is_active BOOLEAN DEFAULT true
);

-- Prompt versions table: maintains version history
CREATE TABLE IF NOT EXISTS prompt_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_id UUID NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    version INTEGER NOT NULL,
    template TEXT NOT NULL,
    variables JSONB DEFAULT '[]'::jsonb,
    model VARCHAR(100),
    temperature DECIMAL(3,2),
    max_tokens INTEGER,
    top_p DECIMAL(3,2),
    metadata JSONB DEFAULT '{}'::jsonb,
    change_log TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    UNIQUE(prompt_id, version)
);

-- Prompt deployments table: tracks deployments across environments
CREATE TABLE IF NOT EXISTS prompt_deployments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    prompt_id UUID NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    version_id UUID NOT NULL REFERENCES prompt_versions(id) ON DELETE CASCADE,
    environment VARCHAR(50) NOT NULL, -- dev, staging, prod
    deployed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deployed_by VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active', -- active, inactive, rolled_back
    rollback_version_id UUID REFERENCES prompt_versions(id),
    UNIQUE(prompt_id, environment)
);

-- Experiments table: A/B testing configuration
CREATE TABLE IF NOT EXISTS experiments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    prompt_id UUID NOT NULL REFERENCES prompts(id) ON DELETE CASCADE,
    status VARCHAR(50) DEFAULT 'draft', -- draft, running, paused, completed
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    traffic_percentage INTEGER DEFAULT 100 CHECK (traffic_percentage >= 0 AND traffic_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(255),
    winner_variant_id UUID,
    metadata JSONB DEFAULT '{}'::jsonb
);

-- Experiment variants table: different prompt versions being tested
CREATE TABLE IF NOT EXISTS experiment_variants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    version_id UUID NOT NULL REFERENCES prompt_versions(id) ON DELETE CASCADE,
    traffic_split INTEGER NOT NULL CHECK (traffic_split >= 0 AND traffic_split <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(experiment_id, name)
);

-- Experiment metrics table: stores performance metrics for variants
CREATE TABLE IF NOT EXISTS experiment_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    experiment_id UUID NOT NULL REFERENCES experiments(id) ON DELETE CASCADE,
    variant_id UUID NOT NULL REFERENCES experiment_variants(id) ON DELETE CASCADE,
    request_count INTEGER DEFAULT 0,
    success_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    total_cost DECIMAL(10,6) DEFAULT 0,
    avg_latency_ms DECIMAL(10,2) DEFAULT 0,
    user_feedback_positive INTEGER DEFAULT 0,
    user_feedback_negative INTEGER DEFAULT 0,
    recorded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    metadata JSONB DEFAULT '{}'::jsonb,
    UNIQUE(experiment_id, variant_id, recorded_at)
);

-- Indexes for performance
CREATE INDEX idx_prompts_name ON prompts(name);
CREATE INDEX idx_prompts_tags ON prompts USING GIN(tags);
CREATE INDEX idx_prompts_active ON prompts(is_active);
CREATE INDEX idx_prompt_versions_prompt_id ON prompt_versions(prompt_id);
CREATE INDEX idx_prompt_versions_version ON prompt_versions(prompt_id, version);
CREATE INDEX idx_prompt_deployments_prompt_env ON prompt_deployments(prompt_id, environment);
CREATE INDEX idx_experiments_prompt_id ON experiments(prompt_id);
CREATE INDEX idx_experiments_status ON experiments(status);
CREATE INDEX idx_experiment_variants_experiment_id ON experiment_variants(experiment_id);
CREATE INDEX idx_experiment_metrics_experiment_id ON experiment_metrics(experiment_id);
CREATE INDEX idx_experiment_metrics_variant_id ON experiment_metrics(variant_id);
CREATE INDEX idx_experiment_metrics_recorded_at ON experiment_metrics(recorded_at);

-- Function to auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Triggers for updated_at
CREATE TRIGGER update_prompts_updated_at BEFORE UPDATE ON prompts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_experiments_updated_at BEFORE UPDATE ON experiments
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE prompts IS 'Stores prompt templates with versioning support';
COMMENT ON TABLE prompt_versions IS 'Maintains complete version history of prompts';
COMMENT ON TABLE prompt_deployments IS 'Tracks prompt deployments across environments';
COMMENT ON TABLE experiments IS 'Manages A/B testing experiments for prompts';
COMMENT ON TABLE experiment_variants IS 'Defines variants (prompt versions) for each experiment';
COMMENT ON TABLE experiment_metrics IS 'Stores performance metrics for experiment variants';
