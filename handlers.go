package translatable

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

type Handler struct {
	repo   Repository
	config *Config
}

func NewHandler(repo Repository, config *Config) *Handler {
	return &Handler{
		repo:   repo,
		config: config,
	}
}

func getUserIDFromContext(r *http.Request) *uuid.UUID {
	if userID := r.Context().Value("user_id"); userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			return &uid
		}
		if uidStr, ok := userID.(string); ok {
			if uid, err := uuid.Parse(uidStr); err == nil {
				return &uid
			}
		}
	}
	return nil
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTranslatableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(h.config); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserIDFromContext(r)

	translatable, err := req.ToTranslatable(userID)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := h.repo.Create(r.Context(), translatable); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create translatable")
		return
	}

	respondJSON(w, http.StatusCreated, translatable)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.PathValue("id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	translatable, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if err.Error() == "translatable not found" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to get translatable")
		return
	}

	respondJSON(w, http.StatusOK, translatable)
}

func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	params := QueryParams{
		Limit:  20,
		Offset: 0,
	}

	if translatableIDStr := r.URL.Query().Get("translatable_id"); translatableIDStr != "" {
		if id, err := uuid.Parse(translatableIDStr); err == nil {
			params.TranslatableID = &id
		} else {
			respondError(w, http.StatusBadRequest, "Invalid translatable_id")
			return
		}
	}

	if translatable := r.URL.Query().Get("translatable"); translatable != "" {
		params.Translatable = &translatable
	}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if id, err := uuid.Parse(userIDStr); err == nil {
			params.UserID = &id
		} else {
			respondError(w, http.StatusBadRequest, "Invalid user_id")
			return
		}
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			params.Limit = limit
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			params.Offset = offset
		}
	}

	if err := params.Validate(); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	results, err := h.repo.Query(r.Context(), params)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to query translatable")
		return
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"data":   results,
		"limit":  params.Limit,
		"offset": params.Offset,
	})
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.PathValue("id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req UpdateTranslatableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := req.Validate(h.config); err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	userID := getUserIDFromContext(r)

	if err := h.repo.Update(r.Context(), id, req.Content, userID); err != nil {
		if err.Error() == "translatable not found" ||
		   err.Error() == "translatable not found or you don't have permission to update it" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to update translatable")
		return
	}

	translatable, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to get updated translatable")
		return
	}

	respondJSON(w, http.StatusOK, translatable)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		idStr = r.PathValue("id")
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	userID := getUserIDFromContext(r)

	if err := h.repo.Delete(r.Context(), id, userID); err != nil {
		if err.Error() == "translatable not found" ||
		   err.Error() == "translatable not found or you don't have permission to delete it" {
			respondError(w, http.StatusNotFound, err.Error())
			return
		}
		respondError(w, http.StatusInternalServerError, "Failed to delete translatable")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Translatable deleted successfully"})
}
