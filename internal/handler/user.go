// Package handler 提供 HTTP 请求处理层，负责解析请求、调用 service 层并统一响应。
package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/internal/service"
	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== 统一响应 =====

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: 0, Message: "success", Data: data})
}

func fail(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, Response{Code: code, Message: message})
}

// handleError 统一错误处理：领域错误 → 400，其他 → 500
func handleError(c *gin.Context, logger *log.Logger, operation string, err error) {
	if de, ok := err.(*domain.DomainError); ok {
		fail(c, http.StatusBadRequest, int(de.Code), de.Message)
		return
	}
	logger.Errorw(operation+" failed", "err", err)
	fail(c, http.StatusInternalServerError, 500, err.Error())
}

// ===== UserHandler =====

// UserHandler 用户 HTTP 处理器
type UserHandler struct {
	userService service.UserService
	log         *log.Logger
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
		log:         log.DefaultLogger(),
	}
}

// Register 用户注册
// POST /api/v1/users/register
func (h *UserHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	resp, err := h.userService.Register(c.Request.Context(), &req)
	if err != nil {
		handleError(c, h.log, "register", err)
		return
	}
	success(c, resp)
}

// Login 用户登录
// POST /api/v1/users/login
func (h *UserHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	resp, err := h.userService.Login(c.Request.Context(), &req)
	if err != nil {
		handleError(c, h.log, "login", err)
		return
	}
	success(c, resp)
}

// Logout 用户登出
// POST /api/v1/users/logout
// 需要在 Authorization header 中携带 Bearer <token>
func (h *UserHandler) Logout(c *gin.Context) {
	token := extractBearerToken(c)
	req := &service.LogoutRequest{AccessToken: token}
	if err := h.userService.Logout(c.Request.Context(), req); err != nil {
		handleError(c, h.log, "logout", err)
		return
	}
	success(c, gin.H{"message": "登出成功"})
}

// extractBearerToken 从 Authorization header 提取 Bearer token
func extractBearerToken(c *gin.Context) string {
	header := c.GetHeader("Authorization")
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return parts[1]
}
