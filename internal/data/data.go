// Package data 提供数据访问层，实现 domain 层定义的所有仓储接口。
package data

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/database"
	"github.com/jiujuan/go-ants/pkg/log"
	pkgredis "github.com/jiujuan/go-ants/pkg/redis"
)

// Data 数据层容器
type Data struct {
	db    *database.DB
	redis *pkgredis.Client
	log   *log.Logger
}

// New 创建数据层实例
func New(db *database.DB, redisClient *pkgredis.Client) (*Data, func(), error) {
	d := &Data{
		db:    db,
		redis: redisClient,
		log:   log.DefaultLogger(),
	}
	cleanup := func() {
		if d.db != nil {
			if err := d.db.Close(); err != nil {
				d.log.Errorw("close db failed", "err", err)
			}
		}
		if d.redis != nil {
			if err := d.redis.Close(); err != nil {
				d.log.Errorw("close redis failed", "err", err)
			}
		}
	}
	return d, cleanup, nil
}

// ===== 用户仓储实现 =====

type UserRepo struct{ db *gorm.DB }

func NewUserRepo(data *Data) domain.UserRepository {
	return &UserRepo{db: data.db.GetDB()}
}

func (r *UserRepo) Create(user *domain.User) error {
	if err := r.db.Create(user).Error; err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}
	return nil
}

func (r *UserRepo) GetByID(id uint) (*domain.User, error) {
	var user domain.User
	if err := r.db.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) GetByUsername(username string) (*domain.User, error) {
	var user domain.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("查询用户失败: %w", err)
	}
	return &user, nil
}

func (r *UserRepo) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询邮箱失败: %w", err)
	}
	return count > 0, nil
}

func (r *UserRepo) ExistsByUsername(username string) (bool, error) {
	var count int64
	if err := r.db.Model(&domain.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return false, fmt.Errorf("查询用户名失败: %w", err)
	}
	return count > 0, nil
}

// ===== Token 黑名单仓储实现 =====

const tokenBlacklistPrefix = "token:blacklist:"

type TokenBlacklistRepo struct{ redis *pkgredis.Client }

func NewTokenBlacklistRepo(data *Data) domain.TokenBlacklistRepository {
	return &TokenBlacklistRepo{redis: data.redis}
}

func (r *TokenBlacklistRepo) Add(token string, ttlSeconds int64) error {
	ctx := context.Background()
	key := tokenBlacklistPrefix + token
	if err := r.redis.Set(ctx, key, "1", time.Duration(ttlSeconds)*time.Second); err != nil {
		return fmt.Errorf("加入黑名单失败: %w", err)
	}
	return nil
}

func (r *TokenBlacklistRepo) Exists(token string) (bool, error) {
	ctx := context.Background()
	key := tokenBlacklistPrefix + token
	n, err := r.redis.Exists(ctx, key)
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, fmt.Errorf("查询黑名单失败: %w", err)
	}
	return n > 0, nil
}

// ===== 留言板仓储实现 =====

type MessageRepo struct{ db *gorm.DB }

func NewMessageRepo(data *Data) domain.MessageRepository {
	return &MessageRepo{db: data.db.GetDB()}
}

func (r *MessageRepo) Create(msg *domain.Message) error {
	if err := r.db.Create(msg).Error; err != nil {
		return fmt.Errorf("创建留言失败: %w", err)
	}
	return nil
}

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

func (r *MessageRepo) List(opts domain.MessageListOptions) ([]*domain.Message, int64, error) {
	var msgs []*domain.Message
	var total int64

	query := r.db.Model(&domain.Message{}).
		Where("status = ?", domain.MessageStatusNormal)
	if opts.UserID > 0 {
		query = query.Where("user_id = ?", opts.UserID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计留言数量失败: %w", err)
	}

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

func (r *MessageRepo) Update(msg *domain.Message) error {
	if err := r.db.Model(msg).Updates(map[string]interface{}{
		"title":   msg.Title,
		"content": msg.Content,
	}).Error; err != nil {
		return fmt.Errorf("更新留言失败: %w", err)
	}
	return nil
}

func (r *MessageRepo) Delete(id uint) error {
	if err := r.db.Model(&domain.Message{}).
		Where("id = ?", id).
		Update("status", domain.MessageStatusDeleted).Error; err != nil {
		return fmt.Errorf("删除留言失败: %w", err)
	}
	return nil
}

func (r *MessageRepo) IncrCommentCount(id uint) error {
	return r.db.Model(&domain.Message{}).
		Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
}

func (r *MessageRepo) DecrCommentCount(id uint) error {
	return r.db.Model(&domain.Message{}).
		Where("id = ? AND comment_count > 0", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count - 1")).Error
}

// ===== 评论仓储实现 =====

type CommentRepo struct{ db *gorm.DB }

func NewCommentRepo(data *Data) domain.CommentRepository {
	return &CommentRepo{db: data.db.GetDB()}
}

func (r *CommentRepo) Create(comment *domain.Comment) error {
	if err := r.db.Create(comment).Error; err != nil {
		return fmt.Errorf("创建评论失败: %w", err)
	}
	return nil
}

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

func (r *CommentRepo) ListByMessageID(opts domain.CommentListOptions) ([]*domain.Comment, int64, error) {
	var comments []*domain.Comment
	var total int64

	query := r.db.Model(&domain.Comment{}).
		Where("message_id = ? AND status = ?", opts.MessageID, domain.CommentStatusNormal)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("统计评论数量失败: %w", err)
	}

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

func (r *CommentRepo) Delete(id uint) error {
	if err := r.db.Model(&domain.Comment{}).
		Where("id = ?", id).
		Update("status", domain.CommentStatusDeleted).Error; err != nil {
		return fmt.Errorf("删除评论失败: %w", err)
	}
	return nil
}
