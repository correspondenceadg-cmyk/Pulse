package events

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/httpx"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	svc *Service
}

func NewHandlers(svc *Service) *Handlers {
	return &Handlers{svc: svc}
}

type eventReq struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Category    string     `json:"category"`
	Lat         float64    `json:"lat"`
	Lng         float64    `json:"lng"`
	StartsAt    time.Time  `json:"startsAt"`
	EndsAt      *time.Time `json:"endsAt,omitempty"`
	Venue       string     `json:"venue"`
	Address     string     `json:"address"`
	Visibility  string     `json:"visibility"`
}

type listResp struct {
	Items      []*Event `json:"items"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var req eventReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	owner := httpx.UserIDFrom(r.Context())
	e, err := h.svc.Create(r.Context(), owner, toInput(req))
	if err != nil {
		if errors.Is(err, ErrInvalid) {
			httpx.Error(w, r, http.StatusBadRequest, "invalid event")
			return
		}
		httpx.InternalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusCreated, e)
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	f := ListFilter{
		Category: q.Get("category"),
		Upcoming: q.Get("when") != "all",
	}

	if bbox := q.Get("bbox"); bbox != "" {
		parts := strings.Split(bbox, ",")
		if len(parts) != 4 {
			httpx.Error(w, r, http.StatusBadRequest, "bbox must be minLng,minLat,maxLng,maxLat")
			return
		}
		minLng, err1 := strconv.ParseFloat(parts[0], 64)
		minLat, err2 := strconv.ParseFloat(parts[1], 64)
		maxLng, err3 := strconv.ParseFloat(parts[2], 64)
		maxLat, err4 := strconv.ParseFloat(parts[3], 64)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			httpx.Error(w, r, http.StatusBadRequest, "bbox values must be numbers")
			return
		}
		f.MinLng, f.MinLat = &minLng, &minLat
		f.MaxLng, f.MaxLat = &maxLng, &maxLat
	}

	if cur := q.Get("cursor"); cur != "" {
		at, id, err := decodeCursor(cur)
		if err != nil {
			httpx.Error(w, r, http.StatusBadRequest, "invalid cursor")
			return
		}
		f.CursorAt, f.CursorID = &at, &id
	}

	if lim := q.Get("limit"); lim != "" {
		n, err := strconv.Atoi(lim)
		if err != nil || n < 1 {
			httpx.Error(w, r, http.StatusBadRequest, "invalid limit")
			return
		}
		f.Limit = n
	}

	items, err := h.svc.List(r.Context(), f)
	if err != nil {
		httpx.InternalError(w, r, err)
		return
	}

	resp := listResp{Items: items}
	if len(items) > 0 && len(items) == effectiveLimit(f.Limit) {
		last := items[len(items)-1]
		resp.NextCursor = encodeCursor(last.StartsAt, last.ID)
	}
	httpx.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	requester := httpx.UserIDFrom(r.Context())
	e, err := h.svc.Get(r.Context(), id, requester)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, r, http.StatusNotFound, "event not found")
			return
		}
		httpx.InternalError(w, r, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, e)
}

func (h *Handlers) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	owner := httpx.UserIDFrom(r.Context())

	var req eventReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	e, err := h.svc.Update(r.Context(), id, owner, toInput(req))
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalid):
			httpx.Error(w, r, http.StatusBadRequest, "invalid event")
		case errors.Is(err, ErrNotFound):
			httpx.Error(w, r, http.StatusNotFound, "event not found")
		default:
			httpx.InternalError(w, r, err)
		}
		return
	}
	httpx.WriteJSON(w, http.StatusOK, e)
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	owner := httpx.UserIDFrom(r.Context())

	if err := h.svc.Delete(r.Context(), id, owner); err != nil {
		if errors.Is(err, ErrNotFound) {
			httpx.Error(w, r, http.StatusNotFound, "event not found")
			return
		}
		httpx.InternalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toInput(req eventReq) CreateInput {
	return CreateInput{
		Title:       req.Title,
		Description: req.Description,
		Category:    req.Category,
		Lat:         req.Lat,
		Lng:         req.Lng,
		StartsAt:    req.StartsAt,
		EndsAt:      req.EndsAt,
		Venue:       req.Venue,
		Address:     req.Address,
		Visibility:  req.Visibility,
	}
}

func effectiveLimit(n int) int {
	if n <= 0 || n > 200 {
		return 50
	}
	return n
}

type cursorPayload struct {
	At time.Time `json:"at"`
	ID string    `json:"id"`
}

func encodeCursor(at time.Time, id string) string {
	b, _ := json.Marshal(cursorPayload{At: at, ID: id})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(raw string) (time.Time, string, error) {
	b, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return time.Time{}, "", err
	}
	var p cursorPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return time.Time{}, "", err
	}
	return p.At, p.ID, nil
}