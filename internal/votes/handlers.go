package votes

import (
	"errors"
	"net/http"

	"github.com/correspondenceadg-cmyk/pulse/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

type voteReq struct {
	Value int `json:"value"`
}

func (h *Handlers) Vote(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	userID := httpx.UserIDFrom(r.Context())

	var req voteReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	stats, err := h.svc.Vote(r.Context(), eventID, userID, req.Value)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidValue):
			httpx.Error(w, r, http.StatusBadRequest, "value must be 1, -1, or 0")
		case errors.Is(err, ErrEventMissing):
			httpx.Error(w, r, http.StatusNotFound, "event not found")
		default:
			httpx.InternalError(w, r, err)
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, stats)
}

func (h *Handlers) Stats(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	userID := httpx.UserIDFrom(r.Context())

	stats, err := h.svc.Stats(r.Context(), eventID, userID)
	if err != nil {
		if errors.Is(err, ErrEventMissing) {
			httpx.Error(w, r, http.StatusNotFound, "event not found")
			return
		}
		httpx.InternalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, stats)
}
