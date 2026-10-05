package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/http/middlewares"
	"AB_system/internal/http/observability"
	"AB_system/pkg/errs"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func writeError(c *gin.Context, err error) {
	ctx := c.Request.Context()
	status, msg := mapError(err)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(ctx, "internal error",
			"err", err,
			"trace_id", observability.GetTraceID(ctx),
		)
	}
	c.JSON(status, gin.H{
		"error":    msg,
		"trace_id": observability.GetTraceID(ctx),
	})
}

func writeBadRequest(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, gin.H{
		"error":    err.Error(),
		"trace_id": observability.GetTraceID(c.Request.Context()),
	})
}

func mapError(err error) (int, string) {
	switch {
	case errors.Is(err, errs.ErrFeatureFlagNotFound),
		errors.Is(err, errs.ErrUserNotFound),
		errors.Is(err, errs.ErrRoleNotFound),
		errors.Is(err, errs.ErrExperimentNotFound),
		errors.Is(err, errs.ErrRecordNotFound):
		return http.StatusNotFound, err.Error()

	case errors.Is(err, errs.ErrDuplicateEntry),
		errors.Is(err, errs.ErrEmailUniquenessFailed),
		errors.Is(err, errs.ErrConflict),
		errors.Is(err, errs.ErrExperimentNotEditable),
		errors.Is(err, errs.ErrInvalidTransition):
		return http.StatusConflict, err.Error()

	case errors.Is(err, errs.ErrPermissionDenied):
		return http.StatusForbidden, err.Error()

	case errors.Is(err, errs.ErrKeyIsEmpty),
		errors.Is(err, errs.ErrValueTypeIsEmpty),
		errors.Is(err, errs.ErrDefaultValueIsEmpty),
		errors.Is(err, errs.ErrInvalidField),
		errors.Is(err, errs.ErrIdIsEmpty),
		errors.Is(err, errs.ErrIdIsInvalid),
		errors.Is(err, errs.ErrEmailIsEmpty),
		errors.Is(err, errs.ErrNameIsEmpty),
		errors.Is(err, errs.ErrExperimentNameIsEmpty),
		errors.Is(err, errs.ErrExperimentVariantsEmpty),
		errors.Is(err, errs.ErrWeightsSumMismatch),
		errors.Is(err, errs.ErrControlVariantCount):
		return http.StatusBadRequest, err.Error()

	default:
		return http.StatusInternalServerError, "внутренняя ошибка сервера"
	}
}

func parseUUID(c *gin.Context, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param(param))
	if err != nil {
		writeError(c, errs.ErrIdIsInvalid)
		return uuid.Nil, false
	}
	return id, true
}
func currentUserID(c *gin.Context) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.GetHeader("X-User-Id"))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":    "нужен заголовок X-User-Id с UUID пользователя",
			"trace_id": observability.GetTraceID(c.Request.Context()),
		})
		return uuid.Nil, false
	}
	return id, true
}
func currentActor(c *gin.Context) (models.Actor, bool) {
	a, ok := middlewares.ActorFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":    "требуется аутентификация",
			"trace_id": observability.GetTraceID(c.Request.Context()),
		})
		return models.Actor{}, false
	}
	return a, true
}
