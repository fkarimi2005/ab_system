package handler

import (
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DecideHandler struct {
	svc *service.DecideService
}

func NewDecideHandler(svc *service.DecideService) *DecideHandler {
	return &DecideHandler{svc: svc}
}

func (h *DecideHandler) Register(r *gin.RouterGroup) {
	r.POST("/decide", h.Decide)
}

func (h *DecideHandler) Decide(c *gin.Context) {
	var req dto.DecideRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	items, err := h.svc.Decide(c.Request.Context(), service.DecideInput{
		SubjectID:  req.SubjectID,
		Attributes: req.Attributes,
		FlagKeys:   req.FlagKeys,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewDecideResponse(items))
}
