package middlewares

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/http/observability"
	"AB_system/pkg/errs"
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	UserHeader = "X-User-Id"
	actorKey   = "actor"
)

// UserFinder: репозиторий пользователей подходит под него без обёрток.
type UserFinder interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
}

func Authenticate(users UserFinder) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.GetHeader(UserHeader))
		if err != nil {
			deny(c, http.StatusUnauthorized, "нужен заголовок X-User-Id с UUID пользователя")
			return
		}

		u, err := users.GetUserByID(c.Request.Context(), id)
		if err != nil {
			if errors.Is(err, errs.ErrRecordNotFound) {
				deny(c, http.StatusUnauthorized, "пользователь не найден")
				return
			}
			slog.ErrorContext(c.Request.Context(), "auth: load user failed", "err", err)
			deny(c, http.StatusInternalServerError, "внутренняя ошибка сервера")
			return
		}
		if !u.IsActive {
			deny(c, http.StatusForbidden, "пользователь деактивирован")
			return
		}

		c.Set(actorKey, models.Actor{ID: u.ID, Role: u.Role.RoleName})
		c.Next()
	}
}

// RequireRole пропускает только перечисленные роли.
func RequireRole(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		a, ok := ActorFrom(c)
		if !ok {
			deny(c, http.StatusUnauthorized, "требуется аутентификация")
			return
		}
		if !allowed[a.Role] {
			deny(c, http.StatusForbidden, errs.ErrPermissionDenied.Error())
			return
		}
		c.Next()
	}
}

func ActorFrom(c *gin.Context) (models.Actor, bool) {
	v, ok := c.Get(actorKey)
	if !ok {
		return models.Actor{}, false
	}
	a, ok := v.(models.Actor)
	return a, ok
}

func deny(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error":    msg,
		"trace_id": observability.GetTraceID(c.Request.Context()),
	})
}
