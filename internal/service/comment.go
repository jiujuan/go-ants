// Package service - comment.go 提供评论业务逻辑服务。
package service

import (
	"context"
	"fmt"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== 评论 DTO =====

// CreateCommentRequest 创建评论请求
type CreateCommentRequest struct {
	MessageID uint   `json:"-"`         // 由路由参数注入
	UserID    uint   `json:"-"`         // 由 JWT 注入
	Content   string `json:"content" binding:"required,min=1,max=2000"`
}

// DeleteCommentRequest 删除评论请求
type DeleteCommentRequest struct {
	ID     uint // 评论 ID
	UserID uint // 操作用户 ID，用于鉴权
}

// CommentItem 评论详情（响应 DTO）
type CommentItem struct {
	ID        uint        `json:"id"`
	MessageID uint        `json:"message_id"`
	Content   string      `json:"content"`
	CreatedAt string      `json:"created_at"`
	Author    *AuthorInfo `json:"author,omitempty"`
}

// CommentListResponse 评论列表响应
type CommentListResponse struct {
	Total     int64          `json:"total"`
	Page      int            `json:"page"`
	PageSize  int            `json:"page_size"`
	MessageID uint           `json:"message_id"`
	Items     []*CommentItem `json:"items"`
}

// ===== 评论服务接口 =====

// CommentService 评论服务接口
type CommentService interface {
	// CreateComment 发表评论（需登录）
	CreateComment(ctx context.Context, req *CreateCommentRequest) (*CommentItem, error)
	// DeleteComment 删除评论（仅作者本人）
	DeleteComment(ctx context.Context, req *DeleteCommentRequest) error
	// ListComments 获取某条留言下的评论列表（公开）
	ListComments(ctx context.Context, messageID uint, page, pageSize int) (*CommentListResponse, error)
}

// ===== 评论服务实现 =====

type commentServiceImpl struct {
	commentRepo   domain.CommentRepository
	msgRepo       domain.MessageRepository
	commentDomain *domain.CommentDomain
	log           *log.Logger
}

// NewCommentService 创建评论服务
func NewCommentService(
	commentRepo domain.CommentRepository,
	msgRepo domain.MessageRepository,
) CommentService {
	return &commentServiceImpl{
		commentRepo:   commentRepo,
		msgRepo:       msgRepo,
		commentDomain: domain.NewCommentDomain(),
		log:           log.DefaultLogger(),
	}
}

// CreateComment 发表评论
func (s *commentServiceImpl) CreateComment(ctx context.Context, req *CreateCommentRequest) (*CommentItem, error) {
	// 1. 校验内容
	if err := s.commentDomain.ValidateContent(req.Content); err != nil {
		return nil, err
	}

	// 2. 确认留言存在
	if _, err := s.msgRepo.GetByID(req.MessageID); err != nil {
		return nil, err // ErrMessageNotFound 或其他错误
	}

	// 3. 创建评论
	comment := &domain.Comment{
		MessageID: req.MessageID,
		UserID:    req.UserID,
		Content:   req.Content,
		Status:    domain.CommentStatusNormal,
	}
	if err := s.commentRepo.Create(comment); err != nil {
		s.log.Errorw("create comment failed", "messageID", req.MessageID, "userID", req.UserID, "err", err)
		return nil, fmt.Errorf("发表评论失败，请稍后重试")
	}

	// 4. 更新留言评论数（允许失败，仅记录日志）
	if err := s.msgRepo.IncrCommentCount(req.MessageID); err != nil {
		s.log.Warnw("incr comment count failed", "messageID", req.MessageID, "err", err)
	}

	s.log.Infow("comment created", "commentID", comment.ID, "messageID", req.MessageID, "userID", req.UserID)
	return toCommentItem(comment), nil
}

// DeleteComment 删除评论（仅本人）
func (s *commentServiceImpl) DeleteComment(ctx context.Context, req *DeleteCommentRequest) error {
	comment, err := s.commentRepo.GetByID(req.ID)
	if err != nil {
		return err
	}
	if comment.UserID != req.UserID {
		return domain.ErrCommentForbidden
	}

	if err := s.commentRepo.Delete(req.ID); err != nil {
		s.log.Errorw("delete comment failed", "commentID", req.ID, "err", err)
		return fmt.Errorf("删除评论失败，请稍后重试")
	}

	// 更新留言评论数
	if err := s.msgRepo.DecrCommentCount(comment.MessageID); err != nil {
		s.log.Warnw("decr comment count failed", "messageID", comment.MessageID, "err", err)
	}

	s.log.Infow("comment deleted", "commentID", req.ID, "userID", req.UserID)
	return nil
}

// ListComments 获取留言下的评论列表
func (s *commentServiceImpl) ListComments(ctx context.Context, messageID uint, page, pageSize int) (*CommentListResponse, error) {
	// 确认留言存在
	if _, err := s.msgRepo.GetByID(messageID); err != nil {
		return nil, err
	}

	opts := domain.CommentListOptions{
		MessageID: messageID,
		Page:      page,
		PageSize:  pageSize,
	}
	comments, total, err := s.commentRepo.ListByMessageID(opts)
	if err != nil {
		s.log.Errorw("list comments failed", "messageID", messageID, "err", err)
		return nil, fmt.Errorf("查询评论失败")
	}

	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}

	items := make([]*CommentItem, 0, len(comments))
	for _, c := range comments {
		items = append(items, toCommentItem(c))
	}

	return &CommentListResponse{
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
		MessageID: messageID,
		Items:     items,
	}, nil
}

// ===== 工具函数 =====

func toCommentItem(c *domain.Comment) *CommentItem {
	if c == nil {
		return nil
	}
	item := &CommentItem{
		ID:        c.ID,
		MessageID: c.MessageID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt.Format("2006-01-02 15:04:05"),
	}
	if c.User != nil {
		item.Author = &AuthorInfo{
			ID:       c.User.ID,
			Username: c.User.Username,
			Nickname: c.User.Nickname,
			Avatar:   c.User.Avatar,
		}
	}
	return item
}
