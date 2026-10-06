package polls

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

type createReq struct {
	Question string   `json:"question"`
	Options  []string `json:"options"`
}

type voteReq struct {
	OptionID string `json:"optionId"`
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	requester := httpx.UserIDFrom(r.Context())

	var req createReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	p, err := h.svc.Create(r.Context(), eventID, requester, CreateInput{
		Question: req.Question,
		Options:  req.Options,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalid):
			httpx.Error(w, r, http.StatusBadRequest, "invalid poll")
		case errors.Is(err, ErrForbidden):
			httpx.Error(w, r, http.StatusForbidden, "not event owner")
		case errors.Is(err, ErrNotFound):
			httpx.Error(w, r, http.StatusNotFound, "event not found")
		default:
			httpx.InternalError(w, r, err)
		}
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, p)
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	items, err := h.svc.ListForEvent(r.Context(), eventID)
	if err != nil {
		httpx.InternalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handlers) Open(w http.ResponseWriter, r *http.Request) {
	pollID := chi.URLParam(r, "id")
	requester := httpx.UserIDFrom(r.Context())
	if err := h.svc.Open(r.Context(), pollID, requester); err != nil {
		writePollErr(w, r, err)
		return
	}
	h.writeState(w, r, pollID)
}

func (h *Handlers) Close(w http.ResponseWriter, r *http.Request) {
	pollID := chi.URLParam(r, "id")
	requester := httpx.UserIDFrom(r.Context())
	if err := h.svc.Close(r.Context(), pollID, requester); err != nil {
		writePollErr(w, r, err)
		return
	}
	h.writeState(w, r, pollID)
}

func (h *Handlers) Vote(w http.ResponseWriter, r *http.Request) {
	pollID := chi.URLParam(r, "id")
	userID := httpx.UserIDFrom(r.Context())

	var req voteReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}
	res, err := h.svc.Vote(r.Context(), pollID, req.OptionID, userID)
	if err != nil {
		writePollErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handlers) Results(w http.ResponseWriter, r *http.Request) {
	pollID := chi.URLParam(r, "id")
	userID := httpx.UserIDFrom(r.Context())
	res, err := h.svc.Results(r.Context(), pollID, userID)
	if err != nil {
		writePollErr(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func (h *Handlers) writeState(w http.ResponseWriter, r *http.Request, pollID string) {
	res, err := h.svc.Results(r.Context(), pollID, "")
	if err != nil {
		httpx.InternalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, res)
}

func writePollErr(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpx.Error(w, r, http.StatusNotFound, "poll not found")
	case errors.Is(err, ErrForbidden):
		httpx.Error(w, r, http.StatusForbidden, "not poll owner")
	case errors.Is(err, ErrNotOpen):
		httpx.Error(w, r, http.StatusConflict, "poll not open")
	case errors.Is(err, ErrInvalidOpt):
		httpx.Error(w, r, http.StatusBadRequest, "invalid option")
	case errors.Is(err, ErrAlreadyVote):
		httpx.Error(w, r, http.StatusConflict, "already voted")
	default:
		httpx.InternalError(w, r, err)
	}
}
