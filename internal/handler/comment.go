// Package handler - comment.go 处理评论相关 HTTP 请求。
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/internal/service"
	"github.com/jiujuan/go-ants/pkg/log"
)

// CommentHandler 评论 HTTP 处理器
type CommentHandler struct {
	svc service.CommentService
	log *log.Logger
}

// NewCommentHandler 创建评论处理器
func NewCommentHandler(svc service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc, log: log.DefaultLogger()}
}

// ListComments 获取指定留言下的评论列表（公开）
// GET /api/v1/messages/:id/comments?page=1&page_size=20
func (h *CommentHandler) ListComments(c *gin.Context) {
	messageID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, 400, "无效的留言 ID")
		return
	}

	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", 20)

	resp, err := h.svc.ListComments(c.Request.Context(), messageID, page, pageSize)
	if err != nil {
		handleError(c, h.log, "list comments", err)
		return
	}
	success(c, resp)
}

// CreateComment 发表评论（需登录）
// POST /api/v1/messages/:id/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, http.StatusUnauthorized, 401, "请先登录")
		return
	}

	messageID, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, 400, "无效的留言 ID")
		return
	}

	var req service.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	req.MessageID = messageID
	req.UserID = userID

	comment, err := h.svc.CreateComment(c.Request.Context(), &req)
	if err != nil {
		handleError(c, h.log, "create comment", err)
		return
	}
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "success", Data: comment})
}

// DeleteComment 删除评论（需登录，仅本人）
// DELETE /api/v1/messages/:id/comments/:cid
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, http.StatusUnauthorized, 401, "请先登录")
		return
	}

	commentID, err := parseUintParam(c, "cid")
	if err != nil {
		fail(c, http.StatusBadRequest, 400, "无效的评论 ID")
		return
	}

	req := &service.DeleteCommentRequest{ID: commentID, UserID: userID}
	if err := h.svc.DeleteComment(c.Request.Context(), req); err != nil {
		handleError(c, h.log, "delete comment", err)
		return
	}
	success(c, gin.H{"message": "评论已删除"})
}

// ===== 工具函数 =====

// queryInt 从 Query 参数读取 int，失败返回 defaultVal
func queryInt(c *gin.Context, key string, defaultVal int) int {
	s := c.DefaultQuery(key, "")
	if s == "" {
		return defaultVal
	}
	var n int
	if _, err := parseIntVal(s, &n); err != nil {
		return defaultVal
	}
	return n
}

func parseIntVal(s string, out *int) (int, error) {
	n := 0
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, &parseErr{}
		}
		n = n*10 + int(ch-'0')
	}
	*out = n
	return n, nil
}

type parseErr struct{}

func (e *parseErr) Error() string { return "parse error" }
