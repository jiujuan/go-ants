// comment_repo.go 实现 domain.CommentRepository。
//
// 负责评论（comments 表）的创建、单条查询、
// 按留言 ID 分页查询和软删除操作。
package data

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/jiujuan/go-ants/internal/domain"
)

// =============================================================================
// CommentRepo — 评论仓储
// =============================================================================

// CommentRepo 实现 domain.CommentRepository，封装 comments 表的读写操作。
// 删除采用软删除（更新 status 字段），不物理删除数据。
type CommentRepo struct {
	db *gorm.DB
}

// NewCommentRepo 创建 CommentRepo，从 Data 容器中获取 DB 连接。
func NewCommentRepo(data *Data) domain.CommentRepository {
	return &CommentRepo{db: data.db.GetDB()}
}

// Create 插入新评论记录。
func (r *CommentRepo) Create(comment *domain.Comment) error {
	if err := r.db.Create(comment).Error; err != nil {
		return fmt.Errorf("创建评论失败: %w", err)
	}
	return nil
}

// GetByID 根据主键查询评论（仅状态正常的记录），同时预加载 User 评论者信息。
// 记录不存在或已软删除时返回 domain.ErrCommentNotFound。
func (r *CommentRepo) GetByID(id uint) (*domain.Comment, error) {
	var comment domain.Comment
	err := r.db.Preload("User").
		Where("id = ? AND status = ?", id, domain.CommentStatusNormal).
		First(&comment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrCommentNotFound
		}
		return nil, fmt.Errorf("查询评论失败: %w", err)
	}
	return &comment, nil
}

// ListByMessageID 分页查询指定留言下的评论列表（按 created_at 升序）。
// 预加载 User 评论者信息。
// 返回当前页数据、总条数和错误。
func (r *CommentRepo) ListByMessageID(opts domain.CommentListOptions) ([]*domain.Comment, int64, error) {
	var comments []*domain.Comment
	var total int64

	query := r.db.Model(&domain.Comment{}).
		Where("message_id = ? AND status = ?", opts.MessageID, domain.CommentStatusNormal)

	// 先统计总数（不含 LIMIT/OFFSET）
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计评论数量失败: %w", err)
	}

	// 规范化分页参数
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	err := query.Preload("User").
		Order("created_at ASC").
		Offset(offset).Limit(pageSize).
		Find(&comments).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询评论列表失败: %w", err)
	}
	return comments, total, nil
}

// Delete 软删除评论（将 status 置为 CommentStatusDeleted）。
// 调用方须在调用前完成权限校验（仅评论作者本人可删除）。
func (r *CommentRepo) Delete(id uint) error {
	if err := r.db.Model(&domain.Comment{}).
		Where("id = ?", id).
		Update("status", domain.CommentStatusDeleted).Error; err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}
	return nil
}
