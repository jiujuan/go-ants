// Package handler - messageboard.go 处理留言板相关 HTTP 请求。
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jiujuan/go-ants/internal/service"
	"github.com/jiujuan/go-ants/pkg/log"
)

// MessageBoardHandler 留言板 HTTP 处理器
type MessageBoardHandler struct {
	svc service.MessageBoardService
	log *log.Logger
}

// NewMessageBoardHandler 创建留言板处理器
func NewMessageBoardHandler(svc service.MessageBoardService) *MessageBoardHandler {
	return &MessageBoardHandler{svc: svc, log: log.DefaultLogger()}
}

// ListMessages 获取留言列表（公开）
// GET /api/v1/messages?page=1&page_size=20&user_id=0
func (h *MessageBoardHandler) ListMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	filterUserID, _ := strconv.ParseUint(c.DefaultQuery("user_id", "0"), 10, 64)

	resp, err := h.svc.ListMessages(c.Request.Context(), page, pageSize, uint(filterUserID))
	if err != nil {
		handleError(c, h.log, "list messages", err)
		return
	}
	success(c, resp)
}

// GetMessage 获取留言详情（公开）
// GET /api/v1/messages/:id
func (h *MessageBoardHandler) GetMessage(c *gin.Context) {
	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, 400, "无效的留言 ID")
		return
	}

	msg, err := h.svc.GetMessage(c.Request.Context(), id)
	if err != nil {
		handleError(c, h.log, "get message", err)
		return
	}
	success(c, msg)
}

// CreateMessage 发布留言（需登录）
// POST /api/v1/messages
func (h *MessageBoardHandler) CreateMessage(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, http.StatusUnauthorized, 401, "请先登录")
		return
	}

	var req service.CreateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}

	msg, err := h.svc.CreateMessage(c.Request.Context(), userID, &req)
	if err != nil {
		handleError(c, h.log, "create message", err)
		return
	}
	c.JSON(http.StatusCreated, Response{Code: 0, Message: "success", Data: msg})
}

// UpdateMessage 更新留言（需登录，仅本人）
// PUT /api/v1/messages/:id
func (h *MessageBoardHandler) UpdateMessage(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, http.StatusUnauthorized, 401, "请先登录")
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, 400, "无效的留言 ID")
		return
	}

	var req service.UpdateMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, http.StatusBadRequest, 400, "请求参数错误: "+err.Error())
		return
	}
	req.ID = id
	req.UserID = userID

	msg, err := h.svc.UpdateMessage(c.Request.Context(), &req)
	if err != nil {
		handleError(c, h.log, "update message", err)
		return
	}
	success(c, msg)
}

// DeleteMessage 删除留言（需登录，仅本人）
// DELETE /api/v1/messages/:id
func (h *MessageBoardHandler) DeleteMessage(c *gin.Context) {
	userID, ok := currentUserID(c)
	if !ok {
		fail(c, http.StatusUnauthorized, 401, "请先登录")
		return
	}

	id, err := parseUintParam(c, "id")
	if err != nil {
		fail(c, http.StatusBadRequest, 400, "无效的留言 ID")
		return
	}

	req := &service.DeleteMessageRequest{ID: id, UserID: userID}
	if err := h.svc.DeleteMessage(c.Request.Context(), req); err != nil {
		handleError(c, h.log, "delete message", err)
		return
	}
	success(c, gin.H{"message": "留言已删除"})
}

// ===== 工具函数 =====

// currentUserID 从 gin.Context 中取出当前登录用户 ID（由 JWTAuth 中间件写入）
func currentUserID(c *gin.Context) (uint, bool) {
	uidStr, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	uid, err := service.ParseUserID(uidStr.(string))
	if err != nil {
		return 0, false
	}
	return uid, true
}

// parseUintParam 解析路由参数为 uint
func parseUintParam(c *gin.Context, key string) (uint, error) {
	val, err := strconv.ParseUint(c.Param(key), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(val), nil
}
