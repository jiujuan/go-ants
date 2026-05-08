// Package middleware 提供 HTTP 中间件：跨域、请求日志、panic 恢复、JWT 鉴权。
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/auth"
	"github.com/jiujuan/go-ants/pkg/log"
)

// CORS 跨域中间件
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin,Content-Type,Authorization")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// Logger 请求日志中间件
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Infow("http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"latency", time.Since(start).String(),
			"client_ip", c.ClientIP(),
		)
	}
}

// Recovery panic 恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// JWTAuth JWT 鉴权中间件
// 验证 Authorization: Bearer <token>，并将 claims 写入 gin.Context
// 同时检查 token 是否在黑名单中（需注入 TokenBlacklistRepository）
func JWTAuth(jwtAuth *auth.JWT, blacklist domain.TokenBlacklistRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 token
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "缺少 Authorization header",
			})
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "Authorization 格式错误，应为 Bearer <token>",
			})
			return
		}
		tokenStr := parts[1]

		// 检查黑名单（已登出的 token）
		if blacklist != nil {
			inBlacklist, err := blacklist.Exists(tokenStr)
			if err == nil && inBlacklist {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"code":    int(domain.ErrCodeTokenInvalid),
					"message": "Token 已失效，请重新登录",
				})
				return
			}
		}

		// 验证 token
		claims, err := jwtAuth.VerifyToken(c.Request.Context(), tokenStr)
		if err != nil {
			code := int(domain.ErrCodeTokenInvalid)
			msg := "Token 无效"
			if err == auth.ErrExpiredToken {
				code = int(domain.ErrCodeTokenExpired)
				msg = "Token 已过期，请重新登录"
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    code,
				"message": msg,
			})
			return
		}

		// 将用户信息写入 context，供后续 handler 使用
		c.Set("claims", claims)
		c.Set("user_id", claims.Subject)
		c.Next()
	}
}
