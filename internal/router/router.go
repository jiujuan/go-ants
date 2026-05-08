// Package router 统一注册所有路由，按公开/鉴权分组挂载。
package router

import (
	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/internal/handler"
	"github.com/jiujuan/go-ants/internal/middleware"
	"github.com/jiujuan/go-ants/pkg/auth"
)

// Register 注册所有路由
//
// 路由规划：
//
//	公开路由（无需 Token）：
//	  POST   /api/v1/users/register
//	  POST   /api/v1/users/login
//	  GET    /api/v1/messages            — 留言列表
//	  GET    /api/v1/messages/:id        — 留言详情
//	  GET    /api/v1/messages/:id/comments — 评论列表
//
//	鉴权路由（需 Bearer Token）：
//	  POST   /api/v1/users/logout
//	  POST   /api/v1/messages            — 发布留言
//	  PUT    /api/v1/messages/:id        — 更新留言
//	  DELETE /api/v1/messages/:id        — 删除留言
//	  POST   /api/v1/messages/:id/comments    — 发表评论
//	  DELETE /api/v1/messages/:id/comments/:cid — 删除评论
func Register(
	engine *gin.Engine,
	userHandler *handler.UserHandler,
	msgHandler *handler.MessageBoardHandler,
	commentHandler *handler.CommentHandler,
	jwtAuth *auth.JWT,
	blacklist domain.TokenBlacklistRepository,
) {
	// ===== 全局中间件 =====
	engine.Use(middleware.Recovery())
	engine.Use(middleware.CORS())
	engine.Use(middleware.Logger())

	// 健康检查
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "go-ants-user-messageboard"})
	})

	v1 := engine.Group("/api/v1")

	// ===== 公开路由 =====

	// 用户
	v1.POST("/users/register", userHandler.Register)
	v1.POST("/users/login", userHandler.Login)

	// 留言板（只读）
	v1.GET("/messages", msgHandler.ListMessages)
	v1.GET("/messages/:id", msgHandler.GetMessage)

	// 评论（只读）
	v1.GET("/messages/:id/comments", commentHandler.ListComments)

	// ===== 鉴权路由 =====
	auth := v1.Group("")
	auth.Use(middleware.JWTAuth(jwtAuth, blacklist))
	{
		// 用户
		auth.POST("/users/logout", userHandler.Logout)

		// 留言板（写操作）
		auth.POST("/messages", msgHandler.CreateMessage)
		auth.PUT("/messages/:id", msgHandler.UpdateMessage)
		auth.DELETE("/messages/:id", msgHandler.DeleteMessage)

		// 评论（写操作）
		auth.POST("/messages/:id/comments", commentHandler.CreateComment)
		auth.DELETE("/messages/:id/comments/:cid", commentHandler.DeleteComment)
	}
}
