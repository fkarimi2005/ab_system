package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ApprovalHandler struct {
	approvals   *service.ApprovalService
	experiments *service.ExperimentService
}

func NewApprovalHandler(
	approvals *service.ApprovalService,
	experiments *service.ExperimentService,
) *ApprovalHandler {
	return &ApprovalHandler{approvals: approvals, experiments: experiments}
}

func (h *ApprovalHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/experiments")
	reviewer := middlewares.RequireRole(models.RoleAdmin, models.RoleApprover)

	g.POST("/:id/approve", reviewer, h.decide(h.approvals.Approve))
	g.POST("/:id/reject", reviewer, h.decide(h.approvals.Reject))
	g.POST("/:id/request-changes", reviewer, h.decide(h.approvals.RequestChanges))
	g.GET("/:id/approvals", h.List)
}

type decisionFunc func(ctx context.Context, id uuid.UUID, actor models.Actor, comment string) error

// decide превращает метод сервиса в gin-хендлер; три эндпоинта отличаются только им.
func (h *ApprovalHandler) decide(fn decisionFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseUUID(c, "id")
		if !ok {
			return
		}
		actor, ok := currentActor(c)
		if !ok {
			return
		}

		var req dto.DecisionRequest
		// тело необязательно (у approve комментария может не быть)
		if err := c.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
			writeBadRequest(c, err)
			return
		}

		ctx := c.Request.Context()
		if err := fn(ctx, id, actor, req.Comment); err != nil {
			writeError(c, err)
			return
		}

		// отдаём актуальное состояние: статус мог смениться из-за гонки
		e, err := h.experiments.GetExperimentByID(ctx, id)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.NewExperimentResponse(&e))
	}
}

func (h *ApprovalHandler) List(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	list, err := h.approvals.List(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	res := make([]dto.ApprovalResponse, 0, len(list))
	for _, a := range list {
		res = append(res, dto.NewApprovalResponse(a))
	}
	c.JSON(http.StatusOK, res)
}
