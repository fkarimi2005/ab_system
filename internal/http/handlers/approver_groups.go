package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ApproverGroupHandler struct {
	svc *service.ApproverGroupService
}

func NewApproverGroupHandler(svc *service.ApproverGroupService) *ApproverGroupHandler {
	return &ApproverGroupHandler{svc: svc}
}

func (h *ApproverGroupHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/approver-groups")
	g.PUT("/:experimenter_id", middlewares.RequireRole(models.RoleAdmin), h.Set)
	g.GET("/:experimenter_id", middlewares.RequireRole(models.RoleAdmin, models.RoleExperimenter), h.Get)
}

func (h *ApproverGroupHandler) Set(c *gin.Context) {
	id, ok := parseUUID(c, "experimenter_id")
	if !ok {
		return
	}
	var req dto.SetApproverGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	g, err := h.svc.Set(c.Request.Context(), id, req.ApproverIDs, req.MinApprovals)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewApproverGroupResponse(g))
}

func (h *ApproverGroupHandler) Get(c *gin.Context) {
	id, ok := parseUUID(c, "experimenter_id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		return
	}
	g, err := h.svc.Get(c.Request.Context(), actor, id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewApproverGroupResponse(g))
}
