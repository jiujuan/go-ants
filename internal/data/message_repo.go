// message_repo.go 实现 domain.MessageRepository。
//
// 负责留言（messages 表）的创建、查询、分页列表、更新、软删除
// 以及评论数的原子增减操作。
package data

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/jiujuan/go-ants/internal/domain"
)

// =============================================================================
// MessageRepo — 留言仓储
// =============================================================================

// MessageRepo 实现 domain.MessageRepository，封装 messages 表的读写操作。
// 删除采用软删除（更新 status 字段），不物理删除数据。
type MessageRepo struct {
	db *gorm.DB
}

// NewMessageRepo 创建 MessageRepo，从 Data 容器中获取 DB 连接。
func NewMessageRepo(data *Data) domain.MessageRepository {
	return &MessageRepo{db: data.db.GetDB()}
}

// Create 插入新留言记录。
func (r *MessageRepo) Create(msg *domain.Message) error {
	if err := r.db.Create(msg).Error; err != nil {
		return fmt.Errorf("创建留言失败: %w", err)
	}
	return nil
}

// GetByID 根据主键查询留言（仅状态正常的记录），同时预加载 User 作者信息。
// 记录不存在或已软删除时返回 domain.ErrMessageNotFound。
func (r *MessageRepo) GetByID(id uint) (*domain.Message, error) {
	var msg domain.Message
	err := r.db.Preload("User").
		Where("id = ? AND status = ?", id, domain.MessageStatusNormal).
		First(&msg).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrMessageNotFound
		}
		return nil, fmt.Errorf("查询留言失败: %w", err)
	}
	return &msg, nil
}

// List 分页查询留言列表（按 created_at 倒序），预加载 User 信息。
// opts.UserID > 0 时过滤指定用户的留言。
// 返回当前页数据、总条数和错误。
func (r *MessageRepo) List(opts domain.MessageListOptions) ([]*domain.Message, int64, error) {
	var msgs []*domain.Message
	var total int64

	query := r.db.Model(&domain.Message{}).
		Where("status = ?", domain.MessageStatusNormal)
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}

	// 先统计总数（不含 LIMIT/OFFSET）
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计留言数量失败: %w", err)
	}

	// 规范化分页参数
	page := opts.Page
	if page <= 0 {
		page = 1
	}
	pageSize := opts.PageSize
	if pageSize <= 0 || pageSize > 50 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&msgs).Error
	if err != nil {
		return nil, 0, fmt.Errorf("查询留言列表失败: %w", err)
	}
	return msgs, total, nil
}

// Update 更新留言的标题和内容。
// 调用方须在调用前完成权限校验（仅作者本人可更新）。
func (r *MessageRepo) Update(msg *domain.Message) error {
	if err := r.db.Model(msg).Updates(map[string]interface{}{
		"title":   msg.Title,
		"content": msg.Content,
	}).Error; err != nil {
		return fmt.Errorf("更新留言失败: %w", err)
	}
	return nil
}

// Delete 软删除留言（将 status 置为 MessageStatusDeleted）。
// 调用方须在调用前完成权限校验（仅作者本人可删除）。
func (r *MessageRepo) Delete(id uint) error {
	if err := r.db.Model(&domain.Message{}).
		Where("id = ?", id).
		Update("status", domain.MessageStatusDeleted).Error; err != nil {
		return fmt.Errorf("删除留言失败: %w", err)
	}
	return nil
}

// IncrCommentCount 原子地将指定留言的评论计数加 1。
// 在新评论创建成功后调用，维护 comment_count 冗余字段。
func (r *MessageRepo) IncrCommentCount(id uint) error {
	if err := r.db.Model(&domain.Message{}).
		Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error; err != nil {
		return fmt.Errorf("增加评论数失败: %w", err)
	}
	return nil
}

// DecrCommentCount 原子地将指定留言的评论计数减 1（下限为 0）。
// 在评论删除成功后调用，维护 comment_count 冗余字段。
func (r *MessageRepo) DecrCommentCount(id uint) error {
	if err := r.db.Model(&domain.Message{}).
		Where("id = ? AND comment_count > 0", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error; err != nil {
		return fmt.Errorf("减少评论数失败: %w", err)
	}
	return nil
}
