package handler

import (
	"AB_system/internal/domain/models"
	"AB_system/internal/domain/service"
	service2 "AB_system/internal/domain/service/input"
	"AB_system/internal/http/dto"
	"AB_system/internal/http/middlewares"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UsersHandler struct {
	svc *service.UserService
}

func NewUsersHandler(svc *service.UserService) *UsersHandler {
	return &UsersHandler{svc: svc}
}
func (h *UsersHandler) Register(r *gin.RouterGroup) {
	//В Register стоят PUT и DELETE. Мы решили PATCH и без удаления (деактивация через is_active).
	g := r.Group("/users")
	admin := middlewares.RequireRole(models.RoleAdmin)
	g.POST("", admin, h.Create)
	g.GET("", admin, h.List)
	g.GET("/:id", admin, h.Get)
	g.PUT("/:id", admin, h.Update)
	g.DELETE("/:id", admin, h.Delete)
}
func (h *UsersHandler) Create(c *gin.Context) {
	var req dto.UserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	user, err := h.svc.CreateUser(c.Request.Context(), models.User{
		Email:    req.Email,
		Name:     req.Name,
		RoleID:   req.RoleID,
		IsActive: isActive,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, dto.NewUserResponse(user))
}
func (h *UsersHandler) List(c *gin.Context) {
	users, err := h.svc.GetAllUsers(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	responses := make([]dto.UserResponse, 0, len(users))
	for _, u := range users {
		responses = append(responses, dto.NewUserResponse(&u))
	}
	c.JSON(http.StatusOK, responses)
}
func (h *UsersHandler) Get(c *gin.Context) {

	ID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	user, err := h.svc.GetUserByID(c.Request.Context(), ID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewUserResponse(&user))
}
func (h *UsersHandler) Delete(c *gin.Context) {
	id, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	err := h.svc.DeleteUser(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
func (h *UsersHandler) Update(c *gin.Context) {
	ID, ok := parseUUID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeBadRequest(c, err)
		return
	}
	user, err := h.svc.UpdateUser(c.Request.Context(), ID, service2.UpdateUserInput{
		Name:     req.Name,
		RoleID:   req.RoleID,
		IsActive: req.IsActive,
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, dto.NewUserResponse(user))
}
