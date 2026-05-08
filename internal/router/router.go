// Package router 提供路由注册，将 handler 与中间件挂载到 Gin 引擎。
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/internal/handler"
	"github.com/jiujuan/go-ants/internal/middleware"
	"github.com/jiujuan/go-ants/pkg/auth"
)

// Register 注册所有路由
// jwtAuth 和 blacklist 用于需要鉴权的路由（当前登出接口可选鉴权）
func Register(
	engine *gin.Engine,
	userHandler *handler.UserHandler,
	jwtAuth *auth.JWT,
	blacklist domain.TokenBlacklistRepository,
) {
	// ===== 全局中间件 =====
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())

	// ===== 基础路由 =====
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "user-service"})
	})

	// ===== API v1 =====
	v1 := engine.Group("/api/v1")

	// 公开路由（无需鉴权）
	users := v1.Group("/users")
	{
		users.POST("/register", userHandler.Register) // 注册
		users.POST("/login", userHandler.Login)       // 登录
	}

	// 需要鉴权的路由
	authUsers := v1.Group("/users")
	authUsers.Use(middleware.JWTAuth(jwtAuth, blacklist))
	{
		authUsers.POST("/logout", userHandler.Logout) // 登出（需携带有效 token）
	}
}
