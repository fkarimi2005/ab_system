package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
)

type EventHandler struct {
	eventSvc *service.EventService
}

func NewEventHandler(eventSvc *service.EventService) *EventHandler {
	return &EventHandler{eventSvc: eventSvc}
}
func (h *EventHandler) Register(router *gin.RouterGroup) {
	r := router.Group("/events")
	r.POST("", h.Send)
	r.GET("/:experimenter_id", h.Get)
	r.GET("", h.List)
}
func (h *EventHandler) Send(c *gin.Context) {
	id, err := uuid.Parse(c.GetHeader(middlewares.UserHeader))
	if err != nil {
		return
	}
	var req dto.EventRequest
	if err := c.ShouldBind(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	event, err := h.eventSvc.Create(c.Request.Context(), models.Event{
		EventID:      req.EventID,
		ExperimentID: req.ExperimentID,
		UserID:       id,
		EventType:    req.EventType,
		DecisionID:   req.DecisionID,
		Properties:   req.Properties,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewEventResponse(&event))
}
func (h *EventHandler) Get(c *gin.Context) {
	ID, ok := parseUUID(c, "experimentID")
	if !ok {
		return
	}
	event, err := h.eventSvc.Get(c.Request.Context(), ID)
	if err != nil {
		writeBadRequest(c, err)
		return
	}
	c.JSON(200, dto.NewEventResponse(&event))

}
func (h *EventHandler) List(c *gin.Context) {
	events, err := h.eventSvc.List(c.Request.Context())
	if err != nil {
		writeBadRequest(c, err)
		return
	}
	response := make([]dto.EventResponse, len(events))
	for i, event := range events {
		response[i] = dto.NewEventResponse(&event)
	}
	c.JSON(200, response)
}
