// Package domain 提供了业务逻辑层，是应用程序的核心业务规则所在。
package domain

import (
	"context"

	"github.com/jiujuan/go-ants/pkg/log"
)

// ===== 基础接口 =====

// Repository 数据仓库接口
type Repository interface {
	// Create 创建记录
	Create(ctx context.Context, entity interface{}) error
	// Update 更新记录
	Update(ctx context.Context, entity interface{}) error
	// Delete 删除记录
	Delete(ctx context.Context, id interface{}) error
	// GetByID 根据 ID 获取记录
	GetByID(ctx context.Context, id interface{}) (interface{}, error)
	// List 列表查询
	List(ctx context.Context, opts ...ListOption) ([]interface{}, error)
}

// ListOption 列表查询选项
type ListOption func(*ListOptions)

// ListOptions 列表查询配置
type ListOptions struct {
	Page     int
	PageSize int
	OrderBy  string
	Filters  map[string]interface{}
}

// WithPage 设置页码
func WithPage(page int) ListOption {
	return func(o *ListOptions) {
		o.Page = page
	}
}

// WithPageSize 设置每页大小
func WithPageSize(pageSize int) ListOption {
	return func(o *ListOptions) {
		o.PageSize = pageSize
	}
}

// WithOrderBy 设置排序字段
func WithOrderBy(orderBy string) ListOption {
	return func(o *ListOptions) {
		o.OrderBy = orderBy
	}
}

// WithFilter 设置过滤条件
func WithFilter(key string, value interface{}) ListOption {
	return func(o *ListOptions) {
		if o.Filters == nil {
			o.Filters = make(map[string]interface{})
		}
		o.Filters[key] = value
	}
}

// ===== 基础服务 =====

// BaseService 基础服务
type BaseService struct {
	log *log.Logger
}

// NewBaseService 创建基础服务
func NewBaseService() *BaseService {
	return &BaseService{
		log: log.DefaultLogger(),
	}
}

// Logger 获取日志实例
func (s *BaseService) Logger() *log.Logger {
	return s.log
}

// ContextWithLogger 将日志器添加到上下文
func ContextWithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return log.WithLogger(ctx, logger)
}

// ===== 通用业务逻辑 =====

// PaginationResult 分页结果
type PaginationResult struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

// NewPaginationResult 创建分页结果
func NewPaginationResult(total int64, page, pageSize int, data interface{}) *PaginationResult {
	return &PaginationResult{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		Data:     data,
	}
}

// ErrorCode 错误码
type ErrorCode int

const (
	// ErrCodeInternal 服务器内部错误
	ErrCodeInternal ErrorCode = 500
	// ErrCodeNotFound 资源未找到
	ErrCodeNotFound ErrorCode = 404
	// ErrCodeBadRequest 请求参数错误
	ErrCodeBadRequest ErrorCode = 400
	// ErrCodeUnauthorized 未授权
	ErrCodeUnauthorized ErrorCode = 401
	// ErrCodeForbidden 禁止访问
	ErrCodeForbidden ErrorCode = 403
	// ErrCodeConflict 资源冲突
	ErrCodeConflict ErrorCode = 409
)

// DomainError 业务错误
type DomainError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

func (e *DomainError) Error() string {
	return e.Message
}

// NewDomainError 创建业务错误
func NewDomainError(code ErrorCode, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// ErrInternal 创建内部错误
func ErrInternal(message string) *DomainError {
	return NewDomainError(ErrCodeInternal, message)
}

// ErrNotFound 创建未找到错误
func ErrNotFound(message string) *DomainError {
	return NewDomainError(ErrCodeNotFound, message)
}

// ErrBadRequest 创建请求错误
func ErrBadRequest(message string) *DomainError {
	return NewDomainError(ErrCodeBadRequest, message)
}

// ErrUnauthorized 创建未授权错误
func ErrUnauthorized(message string) *DomainError {
	return NewDomainError(ErrCodeUnauthorized, message)
}

// ErrForbidden 创建禁止错误
func ErrForbidden(message string) *DomainError {
	return NewDomainError(ErrCodeForbidden, message)
}

// ErrConflict 创建冲突错误
func ErrConflict(message string) *DomainError {
	return NewDomainError(ErrCodeConflict, message)
}

// ===== 业务操作辅助函数 =====

// HandleRepoError 处理仓库错误
func HandleRepoError(err error) *DomainError {
	if err == nil {
		return nil
	}

	// 可以根据具体错误类型进行映射
	// 这里简化处理
	log.Error("repository error",
		log.Error(err))

	return ErrInternal("database operation failed")
}

// HandleCacheError 处理缓存错误（不返回错误，仅记录日志）
func HandleCacheError(err error) {
	if err != nil {
		log.Warn("cache error",
			log.Error(err))
	}
}
