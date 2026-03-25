package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	"github.com/JoX23/go-without-magic/internal/domain"
	"github.com/JoX23/go-without-magic/internal/service"
)

// ProductHandler handles HTTP requests for the Product resource.
type ProductHandler struct {
	svc    *service.ProductService
	logger *zap.Logger
}

func NewProductHandler(svc *service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{svc: svc, logger: logger}
}

// RegisterRoutes registers all Product routes on the mux.
// authMW protege las rutas de escritura (POST, PUT) requiriendo un JWT válido.
func (h *ProductHandler) RegisterRoutes(mux *http.ServeMux, authMW func(http.Handler) http.Handler) {
	mux.Handle("POST /products", authMW(http.HandlerFunc(h.CreateProduct)))
	mux.HandleFunc("GET /products", h.ListProducts)
	mux.HandleFunc("GET /products/{id}", h.GetProduct)
	mux.Handle("PUT /products/{id}", authMW(http.HandlerFunc(h.UpdateProduct)))
}

// CreateProduct handles POST /products
func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Sku         string   `json:"sku"`
		Name        string   `json:"name"`
		Price       float64  `json:"price"`
		Stock       int      `json:"stock"`
		Description *string  `json:"description"`
		ImageUrl    *string  `json:"image_url"`
		Status      string   `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProductError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	e, err := h.svc.CreateProduct(r.Context(), req.Sku, req.Name, req.Price, req.Stock, req.Description, req.ImageUrl, req.Status)
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	writeProductJSON(w, http.StatusCreated, toProductResponse(e))
}

// GetProduct handles GET /products/{id}
func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	e, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	writeProductJSON(w, http.StatusOK, toProductResponse(e))
}

// ListProducts handles GET /products — returns only published products with pagination.
func (h *ProductHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	params := parsePaginationParams(r, 20, 100)

	result, err := h.svc.ListPublished(r.Context(), params)
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	data := make([]productResponse, 0, len(result.Data))
	for _, e := range result.Data {
		data = append(data, toProductResponse(e))
	}

	writeProductJSON(w, http.StatusOK, map[string]any{
		"data":       data,
		"page":       result.Page,
		"limit":      result.Limit,
		"total":      result.Total,
		"totalPages": result.TotalPages,
	})
}

// UpdateProduct handles PUT /products/{id}
func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req struct {
		Name        *string  `json:"name"`
		Price       *float64 `json:"price"`
		Stock       *int     `json:"stock"`
		Description *string  `json:"description"`
		ImageUrl    *string  `json:"image_url"`
		Status      *string  `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeProductError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	e, err := h.svc.UpdateProduct(r.Context(), id, req.Name, req.Price, req.Stock, req.Description, req.ImageUrl, req.Status)
	if err != nil {
		h.handleProductError(w, err)
		return
	}

	writeProductJSON(w, http.StatusOK, toProductResponse(e))
}

// ── Helpers ────────────────────────────────────────────────────────────────

type productResponse struct {
	ID          string   `json:"id"`
	Sku         string   `json:"sku"`
	Name        string   `json:"name"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	Description *string  `json:"description,omitempty"`
	ImageUrl    *string  `json:"imageUrl,omitempty"`
	Status      string   `json:"status"`
	CreatedAt   string   `json:"createdAt"`
}

func toProductResponse(e *domain.Product) productResponse {
	return productResponse{
		ID:          e.ID.String(),
		Sku:         e.Sku,
		Name:        e.Name,
		Price:       e.Price,
		Stock:       e.Stock,
		Description: e.Description,
		ImageUrl:    e.ImageUrl,
		Status:      string(e.Status),
		CreatedAt:   e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (h *ProductHandler) handleProductError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		writeProductError(w, http.StatusNotFound, "product not found")
	case errors.Is(err, domain.ErrProductDuplicated):
		writeProductError(w, http.StatusConflict, "product already exists")
	case errors.Is(err, domain.ErrInvalidProductSku):
		writeProductError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidProductName):
		writeProductError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidProductPrice):
		writeProductError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidProductStock):
		writeProductError(w, http.StatusBadRequest, err.Error())
	default:
		h.logger.Error("unhandled error", zap.Error(err))
		writeProductError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeProductJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body) //nolint:errcheck
}

func writeProductError(w http.ResponseWriter, status int, msg string) {
	writeProductJSON(w, status, map[string]string{"error": msg})
}

// parsePaginationParams reads ?page= and ?limit= from the request query string.
// defaultLimit is the default limit; maxLimit is the maximum allowed limit.
func parsePaginationParams(r *http.Request, defaultLimit, maxLimit int) domain.PaginationParams {
	p := domain.PaginationParams{Page: 1, Limit: defaultLimit}

	if v := r.URL.Query().Get("page"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			p.Page = n
		}
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 1 {
			if n > maxLimit {
				n = maxLimit
			}
			p.Limit = n
		}
	}

	return p
}
