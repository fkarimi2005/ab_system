package handler

import (
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"net/http"

	"github.com/gin-gonic/gin"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) Register(r *gin.RouterGroup) {
	r.POST("/events", h.Ingest)
	r.GET("/decisions/:id/attribution", h.Attribution)
}

// Ingest принимает пакет. Ответ 200 даже если часть событий отклонена:
// подробности по каждому — в rejected_details.
func (h *EventHandler) Ingest(c *gin.Context) {
	var req dto.IngestEventsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	res, err := h.svc.Ingest(c.Request.Context(), req.ToInput())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewIngestEventsResponse(res))
}

func (h *EventHandler) Attribution(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	a, err := h.svc.Attribution(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewAttributionResponse(a))
}
