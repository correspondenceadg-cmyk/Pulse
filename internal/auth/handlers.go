package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/correspondenceadg-cmyk/pulse/internal/httpx"
)

const refreshCookieName = "pulse_rt"

type Handlers struct {
	svc      *Service
	secure   bool
}

func NewHandlers(svc *Service, secure bool) *Handlers {
	return &Handlers{svc: svc, secure: secure}
}

type registerReq struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type userResp struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	Role        string `json:"role"`
}

type tokenResp struct {
	AccessToken string   `json:"accessToken"`
	ExpiresIn   int      `json:"expiresIn"`
	User        userResp `json:"user"`
}

func (h *Handlers) Register(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	u, err := h.svc.Register(r.Context(), req.Email, req.Password, req.DisplayName)
	if err != nil {
		if errors.Is(err, ErrEmailTaken) {
			httpx.Error(w, r, http.StatusConflict, "email already registered")
			return
		}
		httpx.Error(w, r, http.StatusBadRequest, err.Error())
		return
	}

	httpx.WriteJSON(w, http.StatusCreated, toUserResp(u))
}

func (h *Handlers) Login(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := httpx.DecodeJSON(w, r, &req); err != nil {
		httpx.Error(w, r, http.StatusBadRequest, "invalid json")
		return
	}

	u, pair, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, ErrBadCredentials) {
			httpx.Error(w, r, http.StatusUnauthorized, "invalid credentials")
			return
		}
		httpx.InternalError(w, r, err)
		return
	}

	h.setRefreshCookie(w, pair)
	httpx.WriteJSON(w, http.StatusOK, tokenResp{
		AccessToken: pair.Access,
		ExpiresIn:   600,
		User:        toUserResp(u),
	})
}

func (h *Handlers) Refresh(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(refreshCookieName)
	if err != nil {
		httpx.Error(w, r, http.StatusUnauthorized, "no refresh token")
		return
	}

	u, pair, err := h.svc.Refresh(r.Context(), c.Value)
	if err != nil {
		switch {
		case errors.Is(err, ErrTokenReuse):
			h.clearRefreshCookie(w)
			httpx.Error(w, r, http.StatusUnauthorized, "token reuse detected")
		case errors.Is(err, ErrBadCredentials):
			h.clearRefreshCookie(w)
			httpx.Error(w, r, http.StatusUnauthorized, "invalid refresh token")
		default:
			httpx.InternalError(w, r, err)
		}
		return
	}

	h.setRefreshCookie(w, pair)
	httpx.WriteJSON(w, http.StatusOK, tokenResp{
		AccessToken: pair.Access,
		ExpiresIn:   600,
		User:        toUserResp(u),
	})
}

func (h *Handlers) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(refreshCookieName); err == nil {
		_ = h.svc.Logout(r.Context(), c.Value)
	}
	h.clearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) setRefreshCookie(w http.ResponseWriter, pair *TokenPair) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    pair.Refresh,
		Path:     "/api/auth",
		Expires:  pair.RefreshExp,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *Handlers) clearRefreshCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshCookieName,
		Value:    "",
		Path:     "/api/auth",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func toUserResp(u *User) userResp {
	return userResp{ID: u.ID, Email: u.Email, DisplayName: u.DisplayName, Role: u.Role}
}