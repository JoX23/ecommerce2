package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/middleware"
	"github.com/JoX23/go-without-magic/internal/service"
)

// OrderHandler handles HTTP requests for the Order resource.
type OrderHandler struct {
	svc    *service.OrderService
	logger *zap.Logger
}

func NewOrderHandler(svc *service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{svc: svc, logger: logger}
}

// RegisterRoutes registers all Order routes (all protected).
func (h *OrderHandler) RegisterRoutes(mux *http.ServeMux, authMiddleware func(http.Handler) http.Handler) {
	mux.Handle("POST /orders", authMiddleware(http.HandlerFunc(h.CreateOrder)))
	mux.Handle("GET /orders", authMiddleware(http.HandlerFunc(h.ListOrders)))
	mux.Handle("GET /orders/{id}", authMiddleware(http.HandlerFunc(h.GetOrder)))
}

// CreateOrder handles POST /orders
// Body: {"items": [{"productId": "<uuid>", "qty": 2}]}
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userIDStr == "" {
		writeOrderError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeOrderError(w, http.StatusUnauthorized, "invalid user id in token")
		return
	}

	var req struct {
		Items []struct {
			ProductId string `json:"productId"`
			Qty       int    `json:"qty"`
		} `json:"items"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOrderError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if len(req.Items) == 0 {
		writeOrderError(w, http.StatusBadRequest, "items cannot be empty")
		return
	}

	var itemRequests []domain.OrderItemRequest
	for _, item := range req.Items {
		pid, err := uuid.Parse(item.ProductId)
		if err != nil {
			writeOrderError(w, http.StatusBadRequest, "invalid productId: "+item.ProductId)
			return
		}
		if item.Qty <= 0 {
			writeOrderError(w, http.StatusBadRequest, "qty must be positive")
			return
		}
		itemRequests = append(itemRequests, domain.OrderItemRequest{
			ProductId: pid,
			Qty:       item.Qty,
		})
	}

	order, err := h.svc.CreateOrder(r.Context(), userID, itemRequests)
	if err != nil {
		h.handleOrderError(w, err)
		return
	}

	writeOrderJSON(w, http.StatusCreated, toOrderResponse(order))
}

// GetOrder handles GET /orders/{id}
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	userIDStr, _ := r.Context().Value(middleware.UserIDKey).(string)
	id := r.PathValue("id")

	order, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleOrderError(w, err)
		return
	}

	// Only allow the owning user to see the order
	if order.UserId.String() != userIDStr {
		writeOrderError(w, http.StatusForbidden, "forbidden")
		return
	}

	writeOrderJSON(w, http.StatusOK, toOrderResponse(order))
}

// ListOrders handles GET /orders — returns orders for the authenticated user with pagination.
func (h *OrderHandler) ListOrders(w http.ResponseWriter, r *http.Request) {
	userIDStr, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userIDStr == "" {
		writeOrderError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		writeOrderError(w, http.StatusUnauthorized, "invalid user id in token")
		return
	}

	params := parsePaginationParams(r, 20, 100)

	result, err := h.svc.ListByUserPaginated(r.Context(), userID, params)
	if err != nil {
		h.handleOrderError(w, err)
		return
	}

	data := make([]orderResponse, 0, len(result.Data))
	for _, e := range result.Data {
		data = append(data, toOrderResponse(e))
	}

	writeOrderJSON(w, http.StatusOK, map[string]any{
		"data":       data,
		"page":       result.Page,
		"limit":      result.Limit,
		"total":      result.Total,
		"totalPages": result.TotalPages,
	})
}

// ── Helpers ────────────────────────────────────────────────────────────────

type orderItemResponse struct {
	ProductId   string  `json:"productId"`
	ProductName string  `json:"productName"`
	Qty         int     `json:"qty"`
	UnitPrice   float64 `json:"unitPrice"`
	Subtotal    float64 `json:"subtotal"`
}

type orderResponse struct {
	ID        string              `json:"id"`
	UserId    string              `json:"userId"`
	Status    string              `json:"status"`
	Total     float64             `json:"total"`
	Items     []orderItemResponse `json:"items"`
	CreatedAt string              `json:"createdAt"`
}

func toOrderResponse(e *domain.Order) orderResponse {
	items := make([]orderItemResponse, 0, len(e.Items))
	for _, it := range e.Items {
		items = append(items, orderItemResponse{
			ProductId:   it.ProductId.String(),
			ProductName: it.ProductName,
			Qty:         it.Qty,
			UnitPrice:   it.UnitPrice,
			Subtotal:    it.Subtotal,
		})
	}
	return orderResponse{
		ID:        e.ID.String(),
		UserId:    e.UserId.String(),
		Status:    string(e.Status),
		Total:     e.Total,
		Items:     items,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *OrderHandler) handleOrderError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		writeOrderError(w, http.StatusNotFound, "order not found")
	case errors.Is(err, domain.ErrOrderDuplicated):
		writeOrderError(w, http.StatusConflict, "order already exists")
	case errors.Is(err, domain.ErrInvalidOrderUserId):
		writeOrderError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidOrderTotal):
		writeOrderError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrProductNotFound):
		writeOrderError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("unhandled error", zap.Error(err))
		writeOrderError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeOrderJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func writeOrderError(w http.ResponseWriter, status int, msg string) {
	writeOrderJSON(w, status, map[string]string{"error": msg})
}
