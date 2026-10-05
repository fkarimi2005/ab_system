package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"AB_system/pkg/errs"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type ExperimentHandler struct {
	svc       *service.ExperimentService
	lifecycle *service.LifecycleService
}

func NewExperimentHandler(svc *service.ExperimentService, lifecycle *service.LifecycleService) *ExperimentHandler {
	return &ExperimentHandler{svc: svc, lifecycle: lifecycle}
}
func (h *ExperimentHandler) Register(r *gin.RouterGroup) {
	g := r.Group("/experiments")
	write := middlewares.RequireRole(models.RoleAdmin, models.RoleExperimenter)
	g.POST("", write, h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", write, h.Update)

	g.POST("/:id/transition", write, h.Transition)
	reviewer := middlewares.RequireRole(models.RoleAdmin, models.RoleApprover)
	g.POST("/:id/approve", reviewer, h.decide(service.DecisionApprove))
	g.POST("/:id/reject", reviewer, h.decide(service.DecisionReject))
	g.POST("/:id/request-changes", reviewer, h.decide(service.DecisionRequestChanges))
	g.GET("/:id/approvals", h.Approvals)
	g.GET("/:id/versions", h.Versions)
	g.GET("/:id/versions/:version", h.Version)
}
func (h *ExperimentHandler) Create(c *gin.Context) {
	var req dto.CreateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		return
	}
	e, err := h.svc.CreateExperiment(c.Request.Context(), actor.ID, req.ToInput())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewExperimentResponse(e))
}
func (h *ExperimentHandler) Get(c *gin.Context) {
	ID, ok := parseUUID(c, "id")
	if !ok {
		return
	}

	experiment, err := h.svc.GetExperimentByID(c.Request.Context(), ID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewExperimentResponse(&experiment))

}
func (h *ExperimentHandler) List(c *gin.Context) {
	experiments, err := h.svc.GetExperiments(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	res := make([]dto.ExperimentResponse, 0, len(experiments))
	for i := range experiments {
		res = append(res, dto.NewExperimentResponse(&experiments[i]))
	}
	c.JSON(http.StatusOK, res)
}
func (h *ExperimentHandler) Update(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		return
	}
	var req dto.UpdateExperimentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}

	ctx := c.Request.Context()
	if err := h.svc.UpdateExperiment(ctx, id, actor, req.ToInput()); err != nil {
		writeError(c, err)
		return
	}
	e, err := h.svc.GetExperimentByID(ctx, id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewExperimentResponse(&e))
}

func (h *ExperimentHandler) Transition(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	actor, ok := currentActor(c)
	if !ok {
		return
	}
	var req dto.TransitionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	e, err := h.lifecycle.Transition(c.Request.Context(), id, actor, models.ExperimentStatus(req.Status))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewExperimentResponse(&e))
}

func (h *ExperimentHandler) decide(d service.Decision) gin.HandlerFunc {
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
		if c.Request.ContentLength != 0 {
			if err := c.ShouldBindJSON(&req); err != nil {
				writeBadRequest(c, err)
				return
			}
		}
		e, err := h.lifecycle.Decide(c.Request.Context(), id, actor, d, req.Comment)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.NewExperimentResponse(&e))
	}
}

func (h *ExperimentHandler) Approvals(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	list, err := h.lifecycle.Approvals(c.Request.Context(), id)
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

func (h *ExperimentHandler) Versions(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	list, err := h.svc.GetVersions(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	res := make([]dto.VersionResponse, 0, len(list))
	for _, v := range list {
		res = append(res, dto.NewVersionResponse(v))
	}
	c.JSON(http.StatusOK, res)
}

func (h *ExperimentHandler) Version(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	n, err := strconv.Atoi(c.Param("version"))
	if err != nil || n < 1 {
		writeError(c, errs.ErrInvalidField)
		return
	}
	v, err := h.svc.GetVersion(c.Request.Context(), id, n)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewVersionResponse(v))
}
