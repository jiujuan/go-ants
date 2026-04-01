// Package service 提供了应用服务层，负责业务逻辑的编排和数据转换。
package service

import (
	"context"

	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/pkg/log"
)

// Service 服务层实例
type Service struct {
	log *log.Logger
}

// New 创建服务层实例
func New() *Service {
	return &Service{
		log: log.DefaultLogger(),
	}
}

// Logger 获取日志实例
func (s *Service) Logger() *log.Logger {
	return s.log
}

// ContextWithLogger 将日志器添加到上下文
func ContextWithLogger(ctx context.Context, logger *log.Logger) context.Context {
	return log.WithLogger(ctx, logger)
}

// ===== DTO 定义 =====

// Request 请求 DTO 基类
type Request interface{}

// Response 响应 DTO 基类
type Response interface{}

// PageRequest 分页请求
type PageRequest struct {
	Page     int `json:"page" form:"page"`
	PageSize int `json:"page_size" form:"page_size"`
}

// PageResponse 分页响应
type PageResponse struct {
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
	Data     interface{} `json:"data"`
}

// DefaultPage 默认分页参数
const (
	DefaultPage     = 1
	DefaultPageSize = 10
	MaxPageSize     = 100
)

// NormalizePage 规范化分页参数
func NormalizePage(req *PageRequest) {
	if req.Page < 1 {
		req.Page = DefaultPage
	}
	if req.PageSize < 1 {
		req.PageSize = DefaultPageSize
	}
	if req.PageSize > MaxPageSize {
		req.PageSize = MaxPageSize
	}
}

// ===== 错误处理 =====

// ErrorResponse 错误响应
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewErrorResponse 创建错误响应
func NewErrorResponse(code int, message string) *ErrorResponse {
	return &ErrorResponse{
		Code:    code,
		Message: message,
	}
}

// HandleError 处理错误
func HandleError(err error) *ErrorResponse {
	if err == nil {
		return nil
	}

	if domainErr, ok := err.(*domain.DomainError); ok {
		return NewErrorResponse(int(domainErr.Code), domainErr.Message)
	}

	return NewErrorResponse(500, err.Error())
}

// ===== 服务接口定义 =====

// GreeterService 问候服务接口
type GreeterService interface {
	// SayHello 打招呼
	SayHello(ctx context.Context, req *HelloReq) (*HelloResp, error)
}

// HelloReq 打招呼请求
type HelloReq struct {
	Name string `json:"name" form:"name" binding:"required"`
}

// HelloResp 打招呼响应
type HelloResp struct {
	Message string `json:"message"`
}

// GreeterServiceImpl 问候服务实现
type GreeterServiceImpl struct{}

// NewGreeterService 创建问候服务
func NewGreeterService() GreeterService {
	return &GreeterServiceImpl{}
}

// SayHello 打招呼
func (s *GreeterServiceImpl) SayHello(ctx context.Context, req *HelloReq) (*HelloResp, error) {
	return &HelloResp{
		Message: "Hello, " + req.Name + "!",
	}, nil
}
