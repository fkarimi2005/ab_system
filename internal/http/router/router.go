package router

import (
	handler "AB_system/internal/http/handlers"
	"AB_system/internal/http/middlewares"

	"github.com/gin-gonic/gin"
)

type Registrar interface {
	Register(r *gin.RouterGroup)
}

type Deps struct {
	Health   *handler.HealthHandler
	Auth     gin.HandlerFunc
	Handlers []Registrar
}

func New(d Deps) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), middlewares.TraceID(), middlewares.RequestLogger())

	r.GET("/health", d.Health.Health)
	r.GET("/ready", d.Health.Ready)

	api := r.Group("/api/v1")
	api.Use(d.Auth)

	for _, h := range d.Handlers {
		h.Register(api)
	}

	return r
}
