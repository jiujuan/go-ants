// Package service - messageboard.go 提供留言板业务逻辑服务。
package service

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== 留言板 DTO =====

// CreateMessageRequest 创建留言请求
type CreateMessageRequest struct {
	Title   string `json:"title"   binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// UpdateMessageRequest 更新留言请求
type UpdateMessageRequest struct {
	ID      uint   `json:"-"`         // 由路由参数注入
	UserID  uint   `json:"-"`         // 由 JWT 注入，用于鉴权
	Title   string `json:"title"   binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"required,min=1,max=5000"`
}

// DeleteMessageRequest 删除留言请求
type DeleteMessageRequest struct {
	ID     uint `json:"-"`
	UserID uint `json:"-"`
}

// MessageItem 留言详情（响应 DTO）
type MessageItem struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	CommentCount int         `json:"comment_count"`
	CreatedAt    string      `json:"created_at"`
	UpdatedAt    string      `json:"updated_at"`
	Author       *AuthorInfo `json:"author,omitempty"`
}

// AuthorInfo 作者信息（嵌入 MessageItem / CommentItem）
type AuthorInfo struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
}

// MessageListResponse 留言列表响应
type MessageListResponse struct {
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
	Items    []*MessageItem `json:"items"`
}

// ===== 留言板服务接口 =====

// MessageBoardService 留言板服务接口
type MessageBoardService interface {
	// CreateMessage 创建留言（需登录）
	CreateMessage(ctx context.Context, userID uint, req *CreateMessageRequest) (*MessageItem, error)
	// UpdateMessage 更新留言（仅作者本人）
	UpdateMessage(ctx context.Context, req *UpdateMessageRequest) (*MessageItem, error)
	// DeleteMessage 删除留言（仅作者本人）
	DeleteMessage(ctx context.Context, req *DeleteMessageRequest) error
	// GetMessage 获取留言详情（公开）
	GetMessage(ctx context.Context, id uint) (*MessageItem, error)
	// ListMessages 获取留言列表（公开，可按用户筛选）
	ListMessages(ctx context.Context, page, pageSize int, userID uint) (*MessageListResponse, error)
}

// ===== 留言板服务实现 =====

type messageBoardServiceImpl struct {
	msgRepo    domain.MessageRepository
	msgDomain  *domain.MessageDomain
	log        *log.Logger
}

// NewMessageBoardService 创建留言板服务
func NewMessageBoardService(msgRepo domain.MessageRepository) MessageBoardService {
	return &messageBoardServiceImpl{
		msgRepo:   msgRepo,
		msgDomain: domain.NewMessageDomain(),
		log:       log.DefaultLogger(),
	}
}

// CreateMessage 创建留言
func (s *messageBoardServiceImpl) CreateMessage(ctx context.Context, userID uint, req *CreateMessageRequest) (*MessageItem, error) {
	if err := s.msgDomain.ValidateContent(req.Title, req.Content); err != nil {
		return nil, err
	}

	msg := &domain.Message{
		UserID:  userID,
		Title:   req.Title,
		Content: req.Content,
		Status:  domain.MessageStatusNormal,
	}
	if err := s.msgRepo.Create(msg); err != nil {
		s.log.Errorw("create message failed", "userID", userID, "err", err)
		return nil, fmt.Errorf("发布留言失败，请稍后重试")
	}

	s.log.Infow("message created", "msgID", msg.ID, "userID", userID)
	return toMessageItem(msg), nil
}

// UpdateMessage 更新留言（仅本人）
func (s *messageBoardServiceImpl) UpdateMessage(ctx context.Context, req *UpdateMessageRequest) (*MessageItem, error) {
	if err := s.msgDomain.ValidateContent(req.Title, req.Content); err != nil {
		return nil, err
	}

	msg, err := s.msgRepo.GetByID(req.ID)
	if err != nil {
		return nil, err
	}
	if msg.UserID != req.UserID {
		return nil, domain.ErrMessageForbidden
	}

	msg.Title = req.Title
	msg.Content = req.Content
	if err := s.msgRepo.Update(msg); err != nil {
		s.log.Errorw("update message failed", "msgID", req.ID, "err", err)
		return nil, fmt.Errorf("更新留言失败，请稍后重试")
	}

	updated, _ := s.msgRepo.GetByID(req.ID)
	return toMessageItem(updated), nil
}

// DeleteMessage 删除留言（仅本人）
func (s *messageBoardServiceImpl) DeleteMessage(ctx context.Context, req *DeleteMessageRequest) error {
	msg, err := s.msgRepo.GetByID(req.ID)
	if err != nil {
		return err
	}
	if msg.UserID != req.UserID {
		return domain.ErrMessageForbidden
	}

	if err := s.msgRepo.Delete(req.ID); err != nil {
		s.log.Errorw("delete message failed", "msgID", req.ID, "err", err)
		return fmt.Errorf("删除留言失败，请稍后重试")
	}

	s.log.Infow("message deleted", "msgID", req.ID, "userID", req.UserID)
	return nil
}

// GetMessage 获取留言详情（公开）
func (s *messageBoardServiceImpl) GetMessage(ctx context.Context, id uint) (*MessageItem, error) {
	msg, err := s.msgRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	return toMessageItem(msg), nil
}

// ListMessages 获取留言列表（公开）
func (s *messageBoardServiceImpl) ListMessages(ctx context.Context, page, pageSize int, userID uint) (*MessageListResponse, error) {
	opts := domain.MessageListOptions{
		Page:     page,
		PageSize: pageSize,
		UserID:   userID,
	}
	msgs, total, err := s.msgRepo.List(opts)
	if err != nil {
		s.log.Errorw("list messages failed", "err", err)
		return nil, fmt.Errorf("查询留言列表失败")
	}

	items := make([]*MessageItem, 0, len(msgs))
	for _, m := range msgs {
		items = append(items, toMessageItem(m))
	}

	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	return &MessageListResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Items:    items,
	}, nil
}

// ===== 工具函数 =====

func toMessageItem(m *domain.Message) *MessageItem {
	if m == nil {
		return nil
	}
	item := &MessageItem{
		ID:           m.ID,
		Title:        m.Title,
		Content:      m.Content,
		CommentCount: m.CommentCount,
		CreatedAt:    m.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    m.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if m.User != nil {
		item.Author = &AuthorInfo{
			ID:       m.User.ID,
			Username: m.User.Username,
			Nickname: m.User.Nickname,
			Avatar:   m.User.Avatar,
		}
	}
	return item
}

// ParseUserID 从字符串解析 userID（供 handler 使用）
func ParseUserID(s string) (uint, error) {
	id, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("无效的用户 ID")
	}
	return uint(id), nil
}
