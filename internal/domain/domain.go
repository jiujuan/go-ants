// Package domain 提供领域层，包含用户相关的核心业务规则、实体和仓储接口。
package domain

import (
	"errors"
	"time"
)

// ===== 错误定义 =====

// DomainError 领域错误
type DomainError struct {
	Code    ErrorCode
	Message string
}

func (e *DomainError) Error() string { return e.Message }

// ErrorCode 错误码
type ErrorCode int

const (
	ErrCodeUnknown          ErrorCode = 10000
	ErrCodeUserNotFound     ErrorCode = 10001
	ErrCodeUserAlreadyExist ErrorCode = 10002
	ErrCodeInvalidPassword  ErrorCode = 10003
	ErrCodeInvalidEmail     ErrorCode = 10004
	ErrCodeInvalidUsername  ErrorCode = 10005
	ErrCodeUserDisabled     ErrorCode = 10006
	ErrCodeTokenInvalid     ErrorCode = 10007
	ErrCodeTokenExpired     ErrorCode = 10008
)

// 预定义领域错误
var (
	ErrUserNotFound     = &DomainError{Code: ErrCodeUserNotFound, Message: "用户不存在"}
	ErrUserAlreadyExist = &DomainError{Code: ErrCodeUserAlreadyExist, Message: "用户已存在"}
	ErrInvalidPassword  = &DomainError{Code: ErrCodeInvalidPassword, Message: "密码不正确"}
	ErrInvalidEmail     = &DomainError{Code: ErrCodeInvalidEmail, Message: "邮箱格式不正确"}
	ErrInvalidUsername  = &DomainError{Code: ErrCodeInvalidUsername, Message: "用户名格式不正确"}
	ErrUserDisabled     = &DomainError{Code: ErrCodeUserDisabled, Message: "账户已被禁用"}
	ErrTokenInvalid     = &DomainError{Code: ErrCodeTokenInvalid, Message: "Token 无效"}
	ErrTokenExpired     = &DomainError{Code: ErrCodeTokenExpired, Message: "Token 已过期"}
)

// ===== 用户实体 =====

// UserStatus 用户状态
type UserStatus int

const (
	UserStatusActive   UserStatus = 1 // 正常
	UserStatusInactive UserStatus = 2 // 禁用
)

// User 用户领域实体（对应数据库 users 表）
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

// TableName 指定表名
func (User) TableName() string { return "users" }

// ===== 用户仓储接口（domain 层定义，data 层实现）=====

// UserRepository 用户仓储接口
type UserRepository interface {
	Create(user *User) error
	GetByID(id uint) (*User, error)
	GetByEmail(email string) (*User, error)
	GetByUsername(username string) (*User, error)
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)
}

// ===== Token 黑名单仓储接口 =====

// TokenBlacklistRepository Token 黑名单仓储（登出时将 Token 加入黑名单）
type TokenBlacklistRepository interface {
	// Add 将 token 加入黑名单，ttl 为过期时长（秒）
	Add(token string, ttlSeconds int64) error
	// Exists 判断 token 是否在黑名单中
	Exists(token string) (bool, error)
}

// ===== 用户领域服务 =====

// UserDomain 用户领域服务（纯业务规则，不依赖外部基础设施）
type UserDomain struct{}

// NewUserDomain 创建用户领域服务
func NewUserDomain() *UserDomain { return &UserDomain{} }

// ValidateUsername 校验用户名：3-50 位，字母/数字/下划线
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

// ValidatePassword 校验密码：8-64 位
func (d *UserDomain) ValidatePassword(password string) error {
	if len(password) < 8 || len(password) > 64 {
		return errors.New("密码长度必须在 8-64 个字符之间")
	}
	return nil
}

func isAlphanumericOrUnderscore(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') || c == '_'
}
