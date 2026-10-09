package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

func includeArchived(c *gin.Context) bool { return c.Query("include_archived") == "true" }

// ---------- типы событий ----------

type EventTypeHandler struct {
	svc *service.EventTypeService
}

func NewEventTypeHandler(svc *service.EventTypeService) *EventTypeHandler {
	return &EventTypeHandler{svc: svc}
}

func (h *EventTypeHandler) Register(r *gin.RouterGroup) {
	write := middlewares.RequireRole(models.RoleAdmin, models.RoleExperimenter)
	g := r.Group("/event-types")
	g.POST("", write, h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", write, h.Update)
	g.POST("/:id/archive", write, h.Archive)
}

func (h *EventTypeHandler) Create(c *gin.Context) {
	var req dto.CreateEventTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	e, err := h.svc.Create(c.Request.Context(), service.CreateEventTypeInput{
		Key: req.Key, Name: req.Name, Description: req.Description,
		RequiresExposure: req.RequiresExposure, IsExposure: req.IsExposure,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewEventTypeResponse(*e))
}

func (h *EventTypeHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), includeArchived(c))
	if err != nil {
		writeError(c, err)
		return
	}
	res := make([]dto.EventTypeResponse, 0, len(list))
	for _, e := range list {
		res = append(res, dto.NewEventTypeResponse(e))
	}
	c.JSON(http.StatusOK, res)
}

func (h *EventTypeHandler) Get(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	e, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewEventTypeResponse(e))
}

func (h *EventTypeHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateEventTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	e, err := h.svc.Update(c.Request.Context(), id, service.UpdateEventTypeInput{
		Name: req.Name, Description: req.Description, RequiresExposure: req.RequiresExposure,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewEventTypeResponse(e))
}

func (h *EventTypeHandler) Archive(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	e, err := h.svc.Archive(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewEventTypeResponse(e))
}

// ---------- метрики ----------

type MetricHandler struct {
	svc *service.MetricService
}

func NewMetricHandler(svc *service.MetricService) *MetricHandler {
	return &MetricHandler{svc: svc}
}

func (h *MetricHandler) Register(r *gin.RouterGroup) {
	write := middlewares.RequireRole(models.RoleAdmin, models.RoleExperimenter)
	g := r.Group("/metrics")
	g.POST("", write, h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", write, h.Rename)
	g.POST("/:id/archive", write, h.Archive)
}

func (h *MetricHandler) Create(c *gin.Context) {
	var req dto.CreateMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	m, err := h.svc.Create(c.Request.Context(), service.CreateMetricInput{
		Key: req.Key, Name: req.Name, Kind: models.MetricKind(req.Kind),
		EventTypeID: req.EventTypeID, DenominatorEventTypeID: req.DenominatorEventTypeID,
		Percentile: req.Percentile, ValueField: req.ValueField,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewMetricResponse(*m))
}

func (h *MetricHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context(), includeArchived(c))
	if err != nil {
		writeError(c, err)
		return
	}
	res := make([]dto.MetricResponse, 0, len(list))
	for _, m := range list {
		res = append(res, dto.NewMetricResponse(m))
	}
	c.JSON(http.StatusOK, res)
}

func (h *MetricHandler) Get(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewMetricResponse(m))
}

func (h *MetricHandler) Rename(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.RenameMetricRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	m, err := h.svc.Rename(c.Request.Context(), id, req.Name)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewMetricResponse(m))
}

func (h *MetricHandler) Archive(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	m, err := h.svc.Archive(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewMetricResponse(m))
}
