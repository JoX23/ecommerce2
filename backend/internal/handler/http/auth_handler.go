package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/middleware"
	"github.com/JoX23/go-without-magic/internal/service"
)

// AuthHandler handles registration, login and profile endpoints.
type AuthHandler struct {
	svc       *service.UserService
	jwtSecret string
	logger    *zap.Logger
}

func NewAuthHandler(svc *service.UserService, jwtSecret string, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, jwtSecret: jwtSecret, logger: logger}
}

// RegisterRoutes registers auth routes on the mux.
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.HandleFunc("POST /auth/register", h.Register)
	mux.HandleFunc("POST /auth/login", h.Login)
	mux.HandleFunc("POST /auth/logout", h.Logout)
	mux.Handle("GET /auth/me", authMiddleware(http.HandlerFunc(h.Me)))
}

// Register handles POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Name == "" || req.Password == "" {
		writeAuthError(w, http.StatusBadRequest, "email, name and password are required")
		return
	}

	user, err := h.svc.CreateUser(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrUserDuplicated):
			writeAuthError(w, http.StatusConflict, "email already registered")
		case errors.Is(err, domain.ErrInvalidUserEmail),
			errors.Is(err, domain.ErrInvalidUserName),
			errors.Is(err, domain.ErrInvalidUserPasswordHash):
			writeAuthError(w, http.StatusBadRequest, err.Error())
		default:
			h.logger.Error("register error", zap.Error(err))
			writeAuthError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	token, err := h.generateToken(user.ID.String(), user.Email)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		writeAuthError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true en producción con HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 horas
	})

	writeAuthJSON(w, http.StatusCreated, map[string]interface{}{
		"token": token,
		"user": authUserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeAuthError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		writeAuthError(w, http.StatusBadRequest, "email and password are required")
		return
	}

	user, err := h.svc.AuthenticateUser(r.Context(), req.Email, req.Password)
	if err != nil {
		// Don't leak whether email exists
		writeAuthError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}

	token, err := h.generateToken(user.ID.String(), user.Email)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		writeAuthError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true en producción con HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 horas
	})

	writeAuthJSON(w, http.StatusOK, map[string]interface{}{
		"token": token,
		"user": authUserResponse{
			ID:        user.ID.String(),
			Email:     user.Email,
			Name:      user.Name,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
	})
}

// Logout handles POST /auth/logout — clears the auth_token cookie.
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	w.WriteHeader(http.StatusNoContent)
}

// Me handles GET /auth/me (protected)
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		writeAuthError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			writeAuthError(w, http.StatusNotFound, "user not found")
			return
		}
		h.logger.Error("get user error", zap.Error(err))
		writeAuthError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeAuthJSON(w, http.StatusOK, authUserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
	})
}

// generateToken creates a signed HS256 JWT with 24h expiry.
func (h *AuthHandler) generateToken(userID, email string) (string, error) {
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
		"iat":   time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}

// ── Helpers ──────────────────────────────────────────────────────────────────

type authUserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt"`
}

func writeAuthJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func writeAuthError(w http.ResponseWriter, status int, msg string) {
	writeAuthJSON(w, status, map[string]string{"error": msg})
}
