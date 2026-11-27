package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"example.com/order-service/docs"
	"example.com/order-service/internal/order"
	"github.com/go-chi/chi/v5"
)

type OrderService interface {
	Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error)
	GetByID(ctx context.Context, id int64) (order.Order, error)
	List(ctx context.Context) ([]order.Order, error)
	UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error)
	History(ctx context.Context, id int64) ([]order.StatusChange, error)
}

type Handler struct {
	service OrderService
	logger  *slog.Logger
}

func NewRouter(service OrderService, logger *slog.Logger) http.Handler {
	h := &Handler{service: service, logger: logger}
	router := chi.NewRouter()

	router.Use(requestLoggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	router.Get("/health", h.health)
	router.Post("/api/v1/orders", h.create)
	router.Get("/api/v1/orders", h.list)
	router.Get("/api/v1/orders/{id}", h.get)
	router.Patch("/api/v1/orders/{id}/status", h.updateStatus)
	router.Get("/api/v1/orders/{id}/history", h.history)

	router.Get("/swagger/*", docs.ServeUI)
	router.Get("/swagger/doc.json", docs.ServeSpec)

	return router
}

// @Summary Проверка сервиса
// @Produce json
// @Success 200 {object} healthResponse
// @Failure 500 {object} errorResponse
// @Router /health [get]
func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

// @Summary Создать заказ
// @Accept json
// @Produce json
// @Param input body order.CreateOrderInput true "Заказ"
// @Success 201 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [post]
func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	var input order.CreateOrderInput
	if !readJSON(w, r, &input) {
		return
	}

	value, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, value)
}

// @Summary Список заказов
// @Produce json
// @Success 200 {array} order.Order
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders [get]
func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
	values, err := h.service.List(r.Context())
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, values)
}

// @Summary Получить заказ
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id} [get]
func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, ok := readID(w, r)
	if !ok {
		return
	}

	value, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// @Summary Изменить статус
// @Accept json
// @Produce json
// @Param id path int true "ID заказа"
// @Param input body order.UpdateStatusInput true "Статус"
// @Success 200 {object} order.Order
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 409 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/status [patch]
func (h *Handler) updateStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := readID(w, r)
	if !ok {
		return
	}

	var input order.UpdateStatusInput
	if !readJSON(w, r, &input) {
		return
	}

	value, err := h.service.UpdateStatus(r.Context(), id, input.Status)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, value)
}

// @Summary История статусов
// @Produce json
// @Param id path int true "ID заказа"
// @Success 200 {array} order.StatusChange
// @Failure 400 {object} errorResponse
// @Failure 404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /api/v1/orders/{id}/history [get]
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	id, ok := readID(w, r)
	if !ok {
		return
	}

	values, err := h.service.History(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, values)
}

type healthResponse struct {
	Status string `json:"status"`
}

type errorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(value)
}

func readJSON(w http.ResponseWriter, r *http.Request, value any) bool {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if decoder.Decode(value) != nil || decoder.Decode(new(any)) != io.EOF {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid JSON", "INVALID_JSON"})
		return false
	}

	return true
}

func readID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse{"invalid order id", "INVALID_ID"})
		return 0, false
	}

	return id, true
}

func (h *Handler) writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, order.ErrInvalidOrder):
		writeJSON(w, http.StatusBadRequest, errorResponse{order.ErrInvalidOrder.Error(), "INVALID_ORDER"})
	case errors.Is(err, order.ErrInvalidStatus):
		writeJSON(w, http.StatusBadRequest, errorResponse{order.ErrInvalidStatus.Error(), "INVALID_STATUS"})
	case errors.Is(err, order.ErrOrderNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{order.ErrOrderNotFound.Error(), "ORDER_NOT_FOUND"})
	case errors.Is(err, order.ErrInvalidStatusTransition):
		writeJSON(w, http.StatusConflict, errorResponse{order.ErrInvalidStatusTransition.Error(), "INVALID_STATUS_TRANSITION"})
	default:
		h.logger.Error("internal error", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{"internal server error", "INTERNAL_ERROR"})
	}
}
