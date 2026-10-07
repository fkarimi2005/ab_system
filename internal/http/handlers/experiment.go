package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ExperimentHandler struct {
	svc *service.ExperimentService
}

func NewExperimentHandler(svc *service.ExperimentService) *ExperimentHandler {

	return &ExperimentHandler{svc: svc}
}

func (h *ExperimentHandler) Register(r *gin.RouterGroup) {
	write := middlewares.RequireRole(models.RoleAdmin, models.RoleExperimenter)
	g := r.Group("/experiments")

	g.POST("", write, h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", write, h.Update)

	g.POST("/:id/submit", write, h.transition(service.ActionSubmit))
	g.POST("/:id/rework", write, h.transition(service.ActionRework))
	g.POST("/:id/start", write, h.transition(service.ActionStart))
	g.POST("/:id/pause", write, h.transition(service.ActionPause))
	g.POST("/:id/resume", write, h.transition(service.ActionResume))
	g.POST("/:id/archive", write, h.transition(service.ActionArchive))
}

func (h *ExperimentHandler) transition(action service.Action) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := parseUUID(c, "id")
		if !ok {
			return
		}
		actor, ok := currentActor(c)
		if !ok {
			return
		}
		e, err := h.svc.Transition(c.Request.Context(), id, actor, action)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, dto.NewExperimentResponse(&e))
	}
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
