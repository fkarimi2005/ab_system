package handler

import (
	"AB_system/internal/domain/service"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ReportHandler struct {
	reportSvc *service.ReportService
}

func NewReportHandler(reportSvc *service.ReportService) *ReportHandler {
	return &ReportHandler{reportSvc: reportSvc}
}
func (h *ReportHandler) Register(router *gin.RouterGroup) {
	r := router.Group("reports")
	r.GET("", h.Lists)
	r.GET("/:experiment_id", h.GetReport)
	r.PUT("", h.Update)
}
func (h *ReportHandler) GetReport(c *gin.Context) {
	experimentID, ok := parseUUID(c, "experiment_id")
	fmt.Println("ID------------", experimentID)
	if !ok {
		return
	}
	reports, err := h.reportSvc.Report(c.Request.Context(), experimentID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, reports)
}
func (h *ReportHandler) Update(c *gin.Context) {
	if err := h.reportSvc.Update(c.Request.Context()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, nil)
}
func (h *ReportHandler) Lists(c *gin.Context) {
	reports, err := h.reportSvc.Reports(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, reports)
}
