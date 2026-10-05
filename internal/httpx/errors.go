package httpx

import (
	"log/slog"
	"net/http"
)

type ErrorBody struct {
	Type   string `json:"type,omitempty"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
}

func Error(w http.ResponseWriter, r *http.Request, status int, detail string) {
	WriteJSON(w, status, ErrorBody{
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}

func InternalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error",
		"err", err,
		"path", r.URL.Path,
		"request_id", RequestIDFrom(r.Context()),
	)
	Error(w, r, http.StatusInternalServerError, "internal error")
}
