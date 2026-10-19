package handler

import (
	"AB_system/internal/domain/service"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GuardrailTriggerHandler struct {
	guardRailSvc *service.GuardrailService
}

func NewGuardrailTriggerHandler(
	guardRailSvc *service.GuardrailService) *GuardrailTriggerHandler {
	return &GuardrailTriggerHandler{
		guardRailSvc: guardRailSvc,
	}
}
func (h *GuardrailTriggerHandler) Register(router *gin.RouterGroup) {
	r := router.Group("guardrails")
	r.GET("/:experiment_id", h.GetTrigger)
	r.PUT("", h.CheckExperiments)
}
func (h *GuardrailTriggerHandler) GetTrigger(c *gin.Context) {
	experimentID, ok := parseUUID(c, "experiment_id")
	if !ok {
		return
	}
	guards, err := h.guardRailSvc.Get(c.Request.Context(), experimentID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, guards)
}
func (h *GuardrailTriggerHandler) CheckExperiments(c *gin.Context) {
	err := h.guardRailSvc.Check(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, nil)
}
