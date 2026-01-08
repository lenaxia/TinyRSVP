package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/models"
)

type QuestionHandlers struct {
	service events.QuestionService
}

func NewQuestionHandlers(service events.QuestionService) *QuestionHandlers {
	return &QuestionHandlers{
		service: service,
	}
}

func (h *QuestionHandlers) RegisterRoutes(r chi.Router) {
	r.Route("/api/events/{id}/questions", func(r chi.Router) {
		r.Post("/", h.CreateQuestion)
		r.Get("/", h.GetQuestions)
		r.Post("/reorder", h.ReorderQuestions)

		r.Route("/{qid}", func(r chi.Router) {
			r.Put("/", h.UpdateQuestion)
			r.Delete("/", h.DeleteQuestion)
		})
	})
}

type CreateQuestionRequest struct {
	QuestionText string               `json:"question_text"`
	QuestionType models.QuestionType  `json:"question_type"`
	Required     bool                 `json:"required"`
	Options      []string             `json:"options,omitempty"`
}

type UpdateQuestionRequest struct {
	QuestionText string               `json:"question_text"`
	QuestionType models.QuestionType  `json:"question_type"`
	Required     bool                 `json:"required"`
	Options      []string             `json:"options,omitempty"`
}

type ReorderQuestionsRequest struct {
	QuestionIDs []int64 `json:"question_ids"`
}

func (h *QuestionHandlers) CreateQuestion(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	var req CreateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	question := &models.PreferenceQuestion{
		EventID:      eventID,
		QuestionText: req.QuestionText,
		QuestionType: req.QuestionType,
		Required:     req.Required,
	}

	if len(req.Options) > 0 {
		if err := question.SetOptions(req.Options); err != nil {
			respondError(w, http.StatusBadRequest, "invalid options")
			return
		}
	}

	if err := h.service.AddQuestion(r.Context(), question); err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusCreated, question)
}

func (h *QuestionHandlers) GetQuestions(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	questions, err := h.service.GetQuestions(r.Context(), eventID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, questions)
}

func (h *QuestionHandlers) UpdateQuestion(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	questionIDStr := chi.URLParam(r, "qid")
	questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid question ID")
		return
	}

	var req UpdateQuestionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	question := &models.PreferenceQuestion{
		ID:           questionID,
		EventID:      eventID,
		QuestionText: req.QuestionText,
		QuestionType: req.QuestionType,
		Required:     req.Required,
	}

	if len(req.Options) > 0 {
		if err := question.SetOptions(req.Options); err != nil {
			respondError(w, http.StatusBadRequest, "invalid options")
			return
		}
	}

	if err := h.service.UpdateQuestion(r.Context(), question); err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, question)
}

func (h *QuestionHandlers) DeleteQuestion(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	_, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	questionIDStr := chi.URLParam(r, "qid")
	questionID, err := strconv.ParseInt(questionIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid question ID")
		return
	}

	if err := h.service.DeleteQuestion(r.Context(), questionID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *QuestionHandlers) ReorderQuestions(w http.ResponseWriter, r *http.Request) {
	eventIDStr := chi.URLParam(r, "id")
	eventID, err := strconv.ParseInt(eventIDStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "invalid event ID")
		return
	}

	var req ReorderQuestionsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.ReorderQuestions(r.Context(), eventID, req.QuestionIDs); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

