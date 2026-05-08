// Package domain 提供领域层：用户、留言板、评论的实体定义、仓储接口和业务规则。
package domain

import (
	"errors"
	"time"
)

// ===== 通用错误结构 =====

// DomainError 领域错误
type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string { return e.Message }

// ErrorCode 错误码
type ErrorCode int

const (
	// 通用
	ErrCodeUnknown ErrorCode = 10000
	// 用户
	ErrCodeUserNotFound     ErrorCode = 10001
	ErrCodeUserAlreadyExist ErrorCode = 10002
	ErrCodeInvalidPassword  ErrorCode = 10003
	ErrCodeInvalidEmail     ErrorCode = 10004
	ErrCodeInvalidUsername  ErrorCode = 10005
	ErrCodeUserDisabled     ErrorCode = 10006
	ErrCodeTokenInvalid     ErrorCode = 10007
	ErrCodeTokenExpired     ErrorCode = 10008
	// 留言板
	ErrCodeMessageNotFound   ErrorCode = 20001
	ErrCodeMessageForbidden  ErrorCode = 20002
	ErrCodeMessageContentErr ErrorCode = 20003
	// 评论
	ErrCodeCommentNotFound   ErrorCode = 30001
	ErrCodeCommentForbidden  ErrorCode = 30002
	ErrCodeCommentContentErr ErrorCode = 30003
)

// 预定义错误变量
var (
	ErrUserNotFound     = &DomainError{Code: ErrCodeUserNotFound, Message: "用户不存在"}
	ErrUserAlreadyExist = &DomainError{Code: ErrCodeUserAlreadyExist, Message: "用户已存在"}
	ErrInvalidPassword  = &DomainError{Code: ErrCodeInvalidPassword, Message: "密码不正确"}
	ErrInvalidEmail     = &DomainError{Code: ErrCodeInvalidEmail, Message: "邮箱格式不正确"}
	ErrInvalidUsername  = &DomainError{Code: ErrCodeInvalidUsername, Message: "用户名格式不正确"}
	ErrUserDisabled     = &DomainError{Code: ErrCodeUserDisabled, Message: "账户已被禁用"}
	ErrTokenInvalid     = &DomainError{Code: ErrCodeTokenInvalid, Message: "Token 无效"}
	ErrTokenExpired     = &DomainError{Code: ErrCodeTokenExpired, Message: "Token 已过期"}

	ErrMessageNotFound  = &DomainError{Code: ErrCodeMessageNotFound, Message: "留言不存在"}
	ErrMessageForbidden = &DomainError{Code: ErrCodeMessageForbidden, Message: "无权操作此留言"}

	ErrCommentNotFound  = &DomainError{Code: ErrCodeCommentNotFound, Message: "评论不存在"}
	ErrCommentForbidden = &DomainError{Code: ErrCodeCommentForbidden, Message: "无权操作此评论"}
)

// ===== 用户实体 =====

type UserStatus int

const (
	UserStatusActive   UserStatus = 1
	UserStatusInactive UserStatus = 2
)

// User 用户实体
type User struct {
	ID           uint       `gorm:"primaryKey;autoIncrement"`
	Username     string     `gorm:"uniqueIndex;size:50;not null"`
	Email        string     `gorm:"uniqueIndex;size:100;not null"`
	PasswordHash string     `gorm:"size:255;not null"`
	Nickname     string     `gorm:"size:100"`
	Avatar       string     `gorm:"size:255"`
	Status       UserStatus `gorm:"default:1"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (User) TableName() string { return "users" }

// ===== 留言板实体 =====

// MessageStatus 留言状态
type MessageStatus int

const (
	MessageStatusNormal  MessageStatus = 1 // 正常
	MessageStatusDeleted MessageStatus = 2 // 已删除
)

// Message 留言实体（对应 messages 表）
type Message struct {
	ID           uint          `gorm:"primaryKey;autoIncrement"`
	UserID       uint          `gorm:"not null;index"`           // 发布者 ID
	Title        string        `gorm:"size:200;not null"`         // 留言标题
	Content      string        `gorm:"type:text;not null"`        // 留言内容
	Status       MessageStatus `gorm:"default:1"`
	CommentCount int           `gorm:"default:0"`                 // 评论数量（冗余字段）
	CreatedAt    time.Time
	UpdatedAt    time.Time

	// 关联（不存库，仅供查询填充）
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Comments []Comment `gorm:"foreignKey:MessageID" json:"comments,omitempty"`
}

func (Message) TableName() string { return "messages" }

// ===== 评论实体 =====

// CommentStatus 评论状态
type CommentStatus int

const (
	CommentStatusNormal  CommentStatus = 1
	CommentStatusDeleted CommentStatus = 2
)

// Comment 评论实体（对应 comments 表）
type Comment struct {
	ID        uint          `gorm:"primaryKey;autoIncrement"`
	MessageID uint          `gorm:"not null;index"`    // 所属留言 ID
	UserID    uint          `gorm:"not null;index"`    // 评论者 ID
	Content   string        `gorm:"type:text;not null"` // 评论内容
	Status    CommentStatus `gorm:"default:1"`
	CreatedAt time.Time
	UpdatedAt time.Time

	// 关联（不存库，仅供查询填充）
	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Message *Message `gorm:"foreignKey:MessageID" json:"message,omitempty"`
}

func (Comment) TableName() string { return "comments" }

// ===== 用户仓储接口 =====

type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
}

// ===== Token 黑名单仓储接口 =====

type TokenBlacklistRepository interface {
	Add(token string, ttlSeconds int64) error
	Exists(token string) (bool, error)
}

// ===== 留言板仓储接口 =====

// MessageListOptions 留言列表查询选项
type MessageListOptions struct {
	Page     int
	PageSize int
	UserID   uint // 0 表示不过滤
}

// MessageRepository 留言仓储接口
type MessageRepository interface {
	Create(msg *Message) error
	GetByID(id uint) (*Message, error)
	List(opts MessageListOptions) ([]*Message, int64, error)
	Update(msg *Message) error
	Delete(id uint) error
	IncrCommentCount(id uint) error
	DecrCommentCount(id uint) error
}

// ===== 评论仓储接口 =====

// CommentListOptions 评论列表查询选项
type CommentListOptions struct {
	MessageID uint
	Page      int
	PageSize  int
}

// CommentRepository 评论仓储接口
type CommentRepository interface {
	Create(comment *Comment) error
	GetByID(id uint) (*Comment, error)
	ListByMessageID(opts CommentListOptions) ([]*Comment, int64, error)
	Delete(id uint) error
}

// ===== 用户领域服务 =====

type UserDomain struct{}

func NewUserDomain() *UserDomain { return &UserDomain{} }

// ValidateUsername 用户名规则：3-50 位，字母/数字/下划线
func (d *UserDomain) ValidateUsername(username string) error {
	if len(username) < 3 || len(username) > 50 {
		return errors.New("用户名长度必须在 3-50 个字符之间")
	}
	for _, c := range username {
		if !isAlphanumericOrUnderscore(c) {
			return errors.New("用户名只能包含字母、数字和下划线")
		}
	}
	return nil
}

// ValidatePassword 密码规则：8-64 位
func (d *UserDomain) ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 64 {
		return errors.New("密码长度必须在 8-64 个字符之间")
	}
	return nil
}

// ===== 留言板领域服务 =====

type MessageDomain struct{}

func NewMessageDomain() *MessageDomain { return &MessageDomain{} }

// ValidateContent 校验留言内容
func (d *MessageDomain) ValidateContent(title, content string) error {
	if len(title) == 0 || len(title) > 200 {
		return &DomainError{Code: ErrCodeMessageContentErr, Message: "留言标题长度必须在 1-200 个字符之间"}
	}
	if len(content) == 0 || len(content) > 5000 {
		return &DomainError{Code: ErrCodeMessageContentErr, Message: "留言内容长度必须在 1-5000 个字符之间"}
	}
	return nil
}

// ===== 评论领域服务 =====

type CommentDomain struct{}

func NewCommentDomain() *CommentDomain { return &CommentDomain{} }

// ValidateContent 校验评论内容
func (d *CommentDomain) ValidateContent(content string) error {
	if len(content) == 0 || len(content) > 2000 {
		return &DomainError{Code: ErrCodeCommentContentErr, Message: "评论内容长度必须在 1-2000 个字符之间"}
	}
	return nil
}

// ===== 工具函数 =====

func isAlphanumericOrUnderscore(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}
