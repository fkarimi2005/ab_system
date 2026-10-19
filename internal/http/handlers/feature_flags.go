package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

type FeatureFlagHandler struct {
	svc *service.FeatureFlagService
}

func NewFeatureFlagHandler(svc *service.FeatureFlagService) *FeatureFlagHandler {
	return &FeatureFlagHandler{svc: svc}
}

func (h *FeatureFlagHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/feature-flags")
	write := middlewares.RequireRole(models.RoleAdmin, models.RoleExperimenter)
	g.POST("", write, h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PATCH("/:id/default", write, h.UpdateDefault)
}

func (h *FeatureFlagHandler) Create(c *gin.Context) {
	var req dto.CreateFeatureFlagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	f, err := h.svc.CreateFeatureFlag(c.Request.Context(), &models.FeatureFlag{
		Key:          req.Key,
		ValueType:    req.ValueType,
		DefaultValue: req.DefaultValue,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewFeatureFlagResponse(f))
}

func (h *FeatureFlagHandler) List(c *gin.Context) {
	flags, err := h.svc.GetFeatureFlags(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	res := make([]dto.FeatureFlagResponse, 0, len(flags))
	for i := range flags {
		res = append(res, dto.NewFeatureFlagResponse(&flags[i]))
	}
	c.JSON(http.StatusOK, res)
}

func (h *FeatureFlagHandler) Get(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	f, err := h.svc.GetFeatureFlag(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewFeatureFlagResponse(&f))
}

func (h *FeatureFlagHandler) UpdateDefault(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateDefaultValueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	ctx := c.Request.Context()
	if err := h.svc.UpdateDefaultValue(ctx, id, req.Value); err != nil {
		writeError(c, err)
		return
	}
	f, err := h.svc.GetFeatureFlag(ctx, id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewFeatureFlagResponse(&f))
}
