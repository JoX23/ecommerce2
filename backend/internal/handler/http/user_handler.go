package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/service"
)

// UserHandler maneja las peticiones HTTP para el recurso User.
// Responsabilidad única: traducir HTTP ↔ servicio de dominio.
type UserHandler struct {
	svc    *service.UserService
	logger *zap.Logger
}

func NewUserHandler(svc *service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// RegisterRoutes registra todas las rutas de User en el mux dado.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /users", h.CreateUser)
	mux.HandleFunc("GET /users", h.ListUsers)
	mux.HandleFunc("GET /users/{id}", h.GetUser)
}

// CreateUser maneja POST /users
// Nota: este handler no está montado en main.go. Usar POST /auth/register.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeUserError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	e, err := h.svc.CreateUser(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		h.handleUserError(w, err)
		return
	}

	writeUserJSON(w, http.StatusCreated, toUserResponse(e))
}

// GetUser maneja GET /users/{id}
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	e, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleUserError(w, err)
		return
	}

	writeUserJSON(w, http.StatusOK, toUserResponse(e))
}

// ListUsers maneja GET /users
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	items, err := h.svc.ListUsers(r.Context())
	if err != nil {
		h.handleUserError(w, err)
		return
	}

	resp := make([]userResponse, 0, len(items))
	for _, e := range items {
		resp = append(resp, toUserResponse(e))
	}

	writeUserJSON(w, http.StatusOK, resp)
}

// ── Helpers ────────────────────────────────────────────────────────────────

type userResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

func toUserResponse(e *domain.User) userResponse {
	return userResponse{
		ID:        e.ID.String(),
		Email:     e.Email,
		Name:      e.Name,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *UserHandler) handleUserError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		writeUserError(w, http.StatusNotFound, "user not found")
	case errors.Is(err, domain.ErrUserDuplicated):
		writeUserError(w, http.StatusConflict, "user already exists")
	case errors.Is(err, domain.ErrInvalidUserEmail):
		writeUserError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidUserName):
		writeUserError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidUserPasswordHash):
		writeUserError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("unhandled error", zap.Error(err))
		writeUserError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeUserJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func writeUserError(w http.ResponseWriter, status int, msg string) {
	writeUserJSON(w, status, map[string]string{"error": msg})
}
