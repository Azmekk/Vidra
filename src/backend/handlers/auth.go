package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Azmekk/Vidra/backend/middleware"
	"github.com/Azmekk/Vidra/backend/services"
	"github.com/Azmekk/Vidra/backend/utils"
	"github.com/go-chi/chi/v5"
)

type AuthHandler struct {
	Auth *middleware.Auth
}

func NewAuthHandler(auth *middleware.Auth) *AuthHandler {
	return &AuthHandler{Auth: auth}
}

type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	CreatedAt string `json:"createdAt"`
}

type AuthStatusResponse struct {
	SetupRequired bool          `json:"setupRequired"`
	User          *UserResponse `json:"user,omitempty"`
}

func toUser(u services.User) *UserResponse {
	return &UserResponse{ID: u.ID, Username: u.Username, CreatedAt: u.CreatedAt}
}

// GetStatus godoc
// @Summary Get authentication status
// @Description Public. Tells the client whether first-run setup is needed and who is signed in.
// @ID getAuthStatus
// @Tags auth
// @Produce json
// @Success 200 {object} AuthStatusResponse
// @Router /api/auth/status [get]
func (h *AuthHandler) GetStatus(w http.ResponseWriter, r *http.Request) {
	required, err := h.Auth.Service.SetupRequired(r.Context())
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	resp := AuthStatusResponse{SetupRequired: required}
	if c, err := r.Cookie(middleware.SessionCookie); err == nil {
		if user, refreshed, err := h.Auth.Service.Authenticate(r.Context(), c.Value); err == nil {
			resp.User = toUser(user)
			if refreshed != nil {
				h.Auth.SetCookie(w, *refreshed)
			}
		}
	}
	utils.RespondWithJSON(w, http.StatusOK, resp)
}

type SetupRequest struct {
	Code     string `json:"code"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// Setup godoc
// @Summary Create the first account
// @Description Requires the one-time setup code printed in the server log.
// @ID setup
// @Tags auth
// @Accept json
// @Produce json
// @Param request body SetupRequest true "Account"
// @Success 201 {object} UserResponse
// @Failure 400 {object} utils.ErrorResponse
// @Failure 403 {object} utils.ErrorResponse
// @Failure 409 {object} utils.ErrorResponse
// @Router /api/auth/setup [post]
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	var req SetupRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	user, session, err := h.Auth.Service.Setup(r.Context(), req.Code, strings.TrimSpace(req.Username), req.Password, r.UserAgent())
	switch {
	case errors.Is(err, services.ErrSetupDone):
		utils.RespondWithError(w, http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalidSetupCode):
		utils.RespondWithError(w, http.StatusForbidden, err.Error())
	case err != nil:
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
	default:
		slog.Info("account created", "user", user.Username, "ip", middleware.ClientIPFrom(r.Context()))
		h.Auth.SetCookie(w, session)
		utils.RespondWithJSON(w, http.StatusCreated, toUser(user))
	}
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

// Login godoc
// @Summary Sign in
// @ID login
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credentials"
// @Success 200 {object} UserResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	user, session, err := h.Auth.Service.Login(r.Context(), strings.TrimSpace(req.Username), req.Password, req.Remember, r.UserAgent())
	if err != nil {
		utils.Annotate(w, "user", strings.TrimSpace(req.Username))
		utils.RespondWithError(w, http.StatusUnauthorized, services.ErrInvalidCredentials.Error())
		return
	}
	slog.Info("signed in", "user", user.Username, "ip", middleware.ClientIPFrom(r.Context()), "remember", req.Remember)
	h.Auth.SetCookie(w, session)
	utils.RespondWithJSON(w, http.StatusOK, toUser(user))
}

// Logout godoc
// @Summary Sign out
// @ID logout
// @Tags auth
// @Success 204
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if c, err := r.Cookie(middleware.SessionCookie); err == nil {
		_ = h.Auth.Service.Logout(r.Context(), c.Value)
		slog.Info("signed out", "ip", middleware.ClientIPFrom(r.Context()))
	}
	h.Auth.ClearCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

// ChangePassword godoc
// @Summary Change password
// @Description Signs out all other sessions.
// @ID changePassword
// @Tags auth
// @Accept json
// @Param request body ChangePasswordRequest true "Passwords"
// @Success 204
// @Failure 400 {object} utils.ErrorResponse
// @Failure 401 {object} utils.ErrorResponse
// @Router /api/auth/password [put]
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	user, _ := middleware.UserFrom(r.Context())
	err := h.Auth.Service.ChangePassword(r.Context(), user, middleware.SessionTokenFrom(r.Context()), req.CurrentPassword, req.NewPassword)
	switch {
	case errors.Is(err, services.ErrInvalidCredentials):
		utils.RespondWithError(w, http.StatusUnauthorized, "current password is incorrect")
	case err != nil:
		utils.RespondWithError(w, http.StatusBadRequest, err.Error())
	default:
		slog.Info("password changed, other sessions signed out", "user", user.Username)
		w.WriteHeader(http.StatusNoContent)
	}
}

type APITokenResponse struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	LastUsedAt *string `json:"lastUsedAt,omitempty"`
	CreatedAt  string  `json:"createdAt"`
}

type CreateAPITokenRequest struct {
	Name string `json:"name"`
}

type CreateAPITokenResponse struct {
	APITokenResponse
	Token string `json:"token"`
}

// ListAPITokens godoc
// @Summary List API tokens
// @ID listApiTokens
// @Tags auth
// @Produce json
// @Success 200 {array} APITokenResponse
// @Router /api/auth/tokens [get]
func (h *AuthHandler) ListAPITokens(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFrom(r.Context())
	rows, err := h.Auth.Service.ListAPITokens(r.Context(), user.ID)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tokens := make([]APITokenResponse, len(rows))
	for i, t := range rows {
		tokens[i] = APITokenResponse{ID: t.ID, Name: t.Name, LastUsedAt: t.LastUsedAt, CreatedAt: t.CreatedAt}
	}
	utils.RespondWithJSON(w, http.StatusOK, tokens)
}

// CreateAPIToken godoc
// @Summary Create an API token
// @Description The token is only returned once. Use it as "Authorization: Bearer <token>".
// @ID createApiToken
// @Tags auth
// @Accept json
// @Produce json
// @Param request body CreateAPITokenRequest true "Token name"
// @Success 201 {object} CreateAPITokenResponse
// @Failure 400 {object} utils.ErrorResponse
// @Router /api/auth/tokens [post]
func (h *AuthHandler) CreateAPIToken(w http.ResponseWriter, r *http.Request) {
	var req CreateAPITokenRequest
	if !utils.DecodeJSON(w, r, &req) {
		return
	}
	name := strings.TrimSpace(req.Name)
	if name == "" || len(name) > 64 {
		utils.RespondWithError(w, http.StatusBadRequest, "name must be between 1 and 64 characters")
		return
	}
	user, _ := middleware.UserFrom(r.Context())
	row, token, err := h.Auth.Service.CreateAPIToken(r.Context(), user.ID, name)
	if err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("API token created", "name", row.Name, "id", row.ID)
	utils.RespondWithJSON(w, http.StatusCreated, CreateAPITokenResponse{
		APITokenResponse: APITokenResponse{ID: row.ID, Name: row.Name, CreatedAt: row.CreatedAt},
		Token:            token,
	})
}

// DeleteAPIToken godoc
// @Summary Revoke an API token
// @ID deleteApiToken
// @Tags auth
// @Param id path string true "Token ID"
// @Success 204
// @Router /api/auth/tokens/{id} [delete]
func (h *AuthHandler) DeleteAPIToken(w http.ResponseWriter, r *http.Request) {
	user, _ := middleware.UserFrom(r.Context())
	if err := h.Auth.Service.DeleteAPIToken(r.Context(), user.ID, chi.URLParam(r, "id")); err != nil {
		utils.RespondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	slog.Info("API token deleted", "id", chi.URLParam(r, "id"))
	w.WriteHeader(http.StatusNoContent)
}
