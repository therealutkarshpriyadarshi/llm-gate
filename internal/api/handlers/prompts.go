package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/prompts"
)

// PromptsHandler handles prompt management HTTP requests
type PromptsHandler struct {
	manager *prompts.Manager
}

// NewPromptsHandler creates a new prompts handler
func NewPromptsHandler(manager *prompts.Manager) *PromptsHandler {
	return &PromptsHandler{
		manager: manager,
	}
}

// CreatePrompt handles POST /v1/prompts
func (h *PromptsHandler) CreatePrompt(w http.ResponseWriter, r *http.Request) {
	var req prompts.CreatePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	prompt, err := h.manager.CreatePrompt(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(prompt)
}

// GetPrompt handles GET /v1/prompts/{id}
func (h *PromptsHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	prompt, err := h.manager.GetPrompt(r.Context(), identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

// ListPrompts handles GET /v1/prompts
func (h *PromptsHandler) ListPrompts(w http.ResponseWriter, r *http.Request) {
	filter := prompts.PromptFilter{
		Name: r.URL.Query().Get("name"),
	}

	// Parse tags
	if tagsParam := r.URL.Query().Get("tags"); tagsParam != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsParam), &tags); err == nil {
			filter.Tags = tags
		}
	}

	// Parse is_active
	if activeParam := r.URL.Query().Get("is_active"); activeParam != "" {
		if active, err := strconv.ParseBool(activeParam); err == nil {
			filter.IsActive = &active
		}
	}

	// Parse pagination
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil {
			filter.Limit = limit
		}
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil {
			filter.Offset = offset
		}
	}

	prompts, err := h.manager.ListPrompts(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

// UpdatePrompt handles PUT /v1/prompts/{id}
func (h *PromptsHandler) UpdatePrompt(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	var req prompts.UpdatePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	prompt, err := h.manager.UpdatePrompt(r.Context(), identifier, req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompt)
}

// DeletePrompt handles DELETE /v1/prompts/{id}
func (h *PromptsHandler) DeletePrompt(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	if err := h.manager.DeletePrompt(r.Context(), identifier); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RenderPrompt handles POST /v1/prompts/{id}/render
func (h *PromptsHandler) RenderPrompt(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	var req prompts.RenderPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rendered, err := h.manager.RenderPrompt(r.Context(), identifier, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"rendered": rendered,
	})
}

// GetPromptVersions handles GET /v1/prompts/{id}/versions
func (h *PromptsHandler) GetPromptVersions(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	versions, err := h.manager.GetPromptVersions(r.Context(), identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versions)
}

// RenderPromptVersion handles POST /v1/prompts/{id}/versions/{version}/render
func (h *PromptsHandler) RenderPromptVersion(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")
	versionStr := chi.URLParam(r, "version")

	version, err := strconv.Atoi(versionStr)
	if err != nil {
		http.Error(w, "Invalid version number", http.StatusBadRequest)
		return
	}

	var req prompts.RenderPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	rendered, err := h.manager.RenderPromptVersion(r.Context(), identifier, version, req.Variables)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"rendered": rendered,
	})
}

// DeployPrompt handles POST /v1/prompts/deploy
func (h *PromptsHandler) DeployPrompt(w http.ResponseWriter, r *http.Request) {
	var req prompts.DeployPromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	deployment, err := h.manager.DeployPrompt(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deployment)
}

// GetDeployedPrompt handles GET /v1/prompts/{name}/deployed/{environment}
func (h *PromptsHandler) GetDeployedPrompt(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	environment := chi.URLParam(r, "environment")

	version, err := h.manager.GetDeployedPrompt(r.Context(), name, environment)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(version)
}

// CreateExperiment handles POST /v1/experiments
func (h *PromptsHandler) CreateExperiment(w http.ResponseWriter, r *http.Request) {
	var req prompts.CreateExperimentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	experiment, err := h.manager.CreateExperiment(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(experiment)
}

// GetExperiment handles GET /v1/experiments/{id}
func (h *PromptsHandler) GetExperiment(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	experiment, err := h.manager.GetExperiment(r.Context(), identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(experiment)
}

// ListExperiments handles GET /v1/experiments
func (h *PromptsHandler) ListExperiments(w http.ResponseWriter, r *http.Request) {
	filter := prompts.ExperimentFilter{
		Status: r.URL.Query().Get("status"),
	}

	// Parse prompt_id
	if promptIDParam := r.URL.Query().Get("prompt_id"); promptIDParam != "" {
		if promptID, err := uuid.Parse(promptIDParam); err == nil {
			filter.PromptID = promptID
		}
	}

	// Parse pagination
	if limitParam := r.URL.Query().Get("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil {
			filter.Limit = limit
		}
	}

	if offsetParam := r.URL.Query().Get("offset"); offsetParam != "" {
		if offset, err := strconv.Atoi(offsetParam); err == nil {
			filter.Offset = offset
		}
	}

	experiments, err := h.manager.ListExperiments(r.Context(), filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(experiments)
}

// StartExperiment handles POST /v1/experiments/{id}/start
func (h *PromptsHandler) StartExperiment(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	if err := h.manager.StartExperiment(r.Context(), identifier); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "running",
	})
}

// PauseExperiment handles POST /v1/experiments/{id}/pause
func (h *PromptsHandler) PauseExperiment(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	if err := h.manager.PauseExperiment(r.Context(), identifier); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "paused",
	})
}

// CompleteExperiment handles POST /v1/experiments/{id}/complete
func (h *PromptsHandler) CompleteExperiment(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	var req struct {
		WinnerVariantID *uuid.UUID `json:"winner_variant_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.manager.CompleteExperiment(r.Context(), identifier, req.WinnerVariantID); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "completed",
	})
}

// GetExperimentMetrics handles GET /v1/experiments/{id}/metrics
func (h *PromptsHandler) GetExperimentMetrics(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")

	metrics, err := h.manager.GetExperimentMetrics(r.Context(), identifier)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(metrics)
}

// DiffVersions handles GET /v1/prompts/{id}/diff/{version1}/{version2}
func (h *PromptsHandler) DiffVersions(w http.ResponseWriter, r *http.Request) {
	identifier := chi.URLParam(r, "id")
	version1Str := chi.URLParam(r, "version1")
	version2Str := chi.URLParam(r, "version2")

	version1, err := strconv.Atoi(version1Str)
	if err != nil {
		http.Error(w, "Invalid version1 number", http.StatusBadRequest)
		return
	}

	version2, err := strconv.Atoi(version2Str)
	if err != nil {
		http.Error(w, "Invalid version2 number", http.StatusBadRequest)
		return
	}

	diff, err := h.manager.DiffVersions(r.Context(), identifier, version1, version2)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"diff": diff,
	})
}
