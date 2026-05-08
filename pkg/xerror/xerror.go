// Package xerror 提供分级通用错误处理。
//
// 错误分两大类：
//   - SysError  系统错误：程序运行时异常（panic、IO、网络、数据库等底层错误）
//               附带调用栈帧（文件路径 + 行号 + 函数名）
//   - BizError  业务错误：可预期的业务逻辑错误（参数不合法、资源不存在、权限不足等）
//               包含 HTTP 状态码、业务错误码、可展示给用户的消息，以及可选的内部明细
//
// 典型使用方式：
//
//	// 业务错误
//	return xerror.NewBiz(xerror.BizCodeNotFound, "用户不存在").WithDetail("uid=42")
//
//	// 系统错误（自动捕获当前调用栈）
//	return xerror.NewSys(err).WithMsg("数据库连接失败")
//
//	// 包裹下层错误
//	return xerror.Wrap(err, "查询用户信息失败")
//
//	// 统一判断
//	if xe, ok := xerror.As(err); ok {
//	    log.Errorf("code=%d msg=%s stack=\n%s", xe.Code(), xe.Message(), xe.StackTrace())
//	}
package xerror

import (
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// =============================================================================
// 错误级别
// =============================================================================

// Level 错误级别
type Level int

const (
	LevelBiz    Level = 1 // 业务错误（可预期，提示用户）
	LevelSys    Level = 2 // 系统错误（不可预期，需要告警）
	LevelFatal  Level = 3 // 致命错误（程序无法继续，需要立即处理）
)

// String 返回级别名称
func (l Level) String() string {
	switch l {
	case LevelBiz:
		return "BIZ"
	case LevelSys:
		return "SYS"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// =============================================================================
// 业务错误码（BizCode）
// =============================================================================

// BizCode 业务错误码，用于前后端协议
type BizCode int

const (
	// ===== 通用 =====
	BizCodeOK             BizCode = 0     // 成功
	BizCodeUnknown        BizCode = 10000 // 未知错误
	BizCodeBadRequest     BizCode = 10001 // 请求参数错误
	BizCodeUnauthorized   BizCode = 10002 // 未登录/Token 无效
	BizCodeForbidden      BizCode = 10003 // 无权限
	BizCodeNotFound       BizCode = 10004 // 资源不存在
	BizCodeConflict       BizCode = 10005 // 资源冲突（如重复注册）
	BizCodeTooManyRequest BizCode = 10006 // 请求过于频繁
	BizCodeServiceBusy    BizCode = 10007 // 服务繁忙，请稍后重试

	// ===== 用户模块 11xxx =====
	BizCodeUserNotFound     BizCode = 11001
	BizCodeUserAlreadyExist BizCode = 11002
	BizCodePasswordWrong    BizCode = 11003
	BizCodeUserDisabled     BizCode = 11004
	BizCodeTokenExpired     BizCode = 11005
	BizCodeTokenInvalid     BizCode = 11006

	// ===== 留言板模块 12xxx =====
	BizCodeMessageNotFound  BizCode = 12001
	BizCodeMessageForbidden BizCode = 12002
	BizCodeMessageInvalid   BizCode = 12003

	// ===== 评论模块 13xxx =====
	BizCodeCommentNotFound  BizCode = 13001
	BizCodeCommentForbidden BizCode = 13002
	BizCodeCommentInvalid   BizCode = 13003

	// ===== 系统模块 9xxxx =====
	BizCodeDBError      BizCode = 90001 // 数据库错误
	BizCodeCacheError   BizCode = 90002 // 缓存错误
	BizCodeNetworkError BizCode = 90003 // 网络错误
	BizCodeIOError      BizCode = 90004 // IO 错误
	BizCodeTimeout      BizCode = 90005 // 超时
)

// bizCodeHTTPStatus 业务码对应的 HTTP 状态码
var bizCodeHTTPStatus = map[BizCode]int{
	BizCodeOK:               200,
	BizCodeBadRequest:       400,
	BizCodeUnauthorized:     401,
	BizCodeForbidden:        403,
	BizCodeNotFound:         404,
	BizCodeConflict:         409,
	BizCodeTooManyRequest:   429,
	BizCodeUserNotFound:     404,
	BizCodeUserAlreadyExist: 409,
	BizCodePasswordWrong:    401,
	BizCodeUserDisabled:     403,
	BizCodeTokenExpired:     401,
	BizCodeTokenInvalid:     401,
	BizCodeMessageNotFound:  404,
	BizCodeMessageForbidden: 403,
	BizCodeCommentNotFound:  404,
	BizCodeCommentForbidden: 403,
}

// HTTPStatus 根据 BizCode 返回对应 HTTP 状态码，默认 500
func (c BizCode) HTTPStatus() int {
	if code, ok := bizCodeHTTPStatus[c]; ok {
		return code
	}
	return 500
}

// =============================================================================
// 调用栈帧
// =============================================================================

// Frame 单个栈帧
type Frame struct {
	File     string // 文件完整路径
	ShortFile string // 文件短路径（从模块根相对路径）
	Line     int    // 行号
	Function string // 函数全路径
	ShortFn  string // 函数短名称（去掉包路径）
}

// String 返回可读栈帧字符串
func (f Frame) String() string {
	return fmt.Sprintf("%s:%d  %s", f.ShortFile, f.Line, f.ShortFn)
}

// Stack 调用栈（多帧）
type Stack []Frame

// String 格式化输出整个调用栈
func (s Stack) String() string {
	if len(s) == 0 {
		return "<no stack>"
	}
	var sb strings.Builder
	for i, f := range s {
		sb.WriteString(fmt.Sprintf("  #%d  %s\n", i, f.String()))
	}
	return sb.String()
}

// captureStack 捕获当前 goroutine 调用栈
// skip：跳过的帧数（0=captureStack 自身，1=调用 captureStack 的函数，以此类推）
// depth：最多捕获多少帧
func captureStack(skip, depth int) Stack {
	pcs := make([]uintptr, depth)
	n := runtime.Callers(skip+2, pcs)
	pcs = pcs[:n]

	frames := runtime.CallersFrames(pcs)
	var stack Stack
	for {
		f, more := frames.Next()
		// 过滤 Go 运行时内部帧
		if strings.Contains(f.Function, "runtime.") {
			if !more {
				break
			}
			continue
		}
		frame := Frame{
			File:     f.File,
			ShortFile: shortFilePath(f.File),
			Line:     f.Line,
			Function: f.Function,
			ShortFn:  shortFuncName(f.Function),
		}
		stack = append(stack, frame)
		if !more {
			break
		}
	}
	return stack
}

// shortFilePath 返回文件的短路径（保留最后 3 段路径）
func shortFilePath(file string) string {
	parts := strings.Split(filepath.ToSlash(file), "/")
	if len(parts) <= 3 {
		return file
	}
	return strings.Join(parts[len(parts)-3:], "/")
}

// shortFuncName 从函数全路径提取短函数名
func shortFuncName(fn string) string {
	// e.g. github.com/jiujuan/go-ants/internal/service.(*UserServiceImpl).Register
	// → service.(*UserServiceImpl).Register
	idx := strings.LastIndex(fn, "/")
	if idx >= 0 {
		return fn[idx+1:]
	}
	return fn
}

// =============================================================================
// XError —— 统一错误类型
// =============================================================================

// XError 统一错误类型，同时满足 error 接口
type XError struct {
	level     Level     // 错误级别
	code      BizCode   // 业务错误码（SysError 也记录，用于统一响应）
	message   string    // 面向用户的消息
	detail    string    // 内部明细（不对外暴露）
	cause     error     // 原始错误（链式包裹）
	stack     Stack     // 调用栈（SysError/FatalError 必填，BizError 可选）
	occurAt   time.Time // 发生时间
}

// Error 实现 error 接口
func (e *XError) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("[%s][%d] %s: %v", e.level, e.code, e.message, e.cause)
	}
	return fmt.Sprintf("[%s][%d] %s", e.level, e.code, e.message)
}

// Unwrap 实现 errors.Unwrap，支持 errors.Is / errors.As
func (e *XError) Unwrap() error { return e.cause }

// ===== 字段读取方法 =====

func (e *XError) Level() Level     { return e.level }
func (e *XError) Code() BizCode    { return e.code }
func (e *XError) Message() string  { return e.message }
func (e *XError) Detail() string   { return e.detail }
func (e *XError) Cause() error     { return e.cause }
func (e *XError) Stack() Stack     { return e.stack }
func (e *XError) OccurAt() time.Time { return e.occurAt }

// IsBiz 判断是否业务错误
func (e *XError) IsBiz() bool { return e.level == LevelBiz }

// IsSys 判断是否系统错误
func (e *XError) IsSys() bool { return e.level == LevelSys || e.level == LevelFatal }

// HTTPStatus 返回对应 HTTP 状态码
func (e *XError) HTTPStatus() int { return e.code.HTTPStatus() }

// StackTrace 返回格式化的调用栈字符串
func (e *XError) StackTrace() string { return e.stack.String() }

// ===== 链式 With 方法 =====

// WithDetail 附加内部明细（不对外暴露）
func (e *XError) WithDetail(detail string) *XError {
	e.detail = detail
	return e
}

// WithDetailf 附加格式化内部明细
func (e *XError) WithDetailf(format string, args ...interface{}) *XError {
	e.detail = fmt.Sprintf(format, args...)
	return e
}

// WithMsg 覆盖对外消息
func (e *XError) WithMsg(msg string) *XError {
	e.message = msg
	return e
}

// WithMsgf 覆盖格式化对外消息
func (e *XError) WithMsgf(format string, args ...interface{}) *XError {
	e.message = fmt.Sprintf(format, args...)
	return e
}

// WithCause 附加原始错误（不重新捕获栈）
func (e *XError) WithCause(err error) *XError {
	e.cause = err
	return e
}

// WithStack 手动附加调用栈（覆盖已有栈）
func (e *XError) WithStack() *XError {
	e.stack = captureStack(1, 32)
	return e
}

// =============================================================================
// 构造函数
// =============================================================================

// NewBiz 创建业务错误（不捕获调用栈，性能更好）
//
//	err := xerror.NewBiz(xerror.BizCodeNotFound, "用户不存在")
func NewBiz(code BizCode, message string) *XError {
	return &XError{
		level:   LevelBiz,
		code:    code,
		message: message,
		occurAt: time.Now(),
	}
}

// NewBizf 创建格式化业务错误
func NewBizf(code BizCode, format string, args ...interface{}) *XError {
	return NewBiz(code, fmt.Sprintf(format, args...))
}

// NewSys 创建系统错误（自动捕获调用栈）
//
//	err := xerror.NewSys(origErr).WithMsg("数据库查询失败")
func NewSys(cause error) *XError {
	return &XError{
		level:   LevelSys,
		code:    BizCodeUnknown,
		message: "系统繁忙，请稍后重试",
		cause:   cause,
		stack:   captureStack(1, 32),
		occurAt: time.Now(),
	}
}

// NewSysWithCode 创建带业务码的系统错误
func NewSysWithCode(code BizCode, cause error) *XError {
	return &XError{
		level:   LevelSys,
		code:    code,
		message: "系统繁忙，请稍后重试",
		cause:   cause,
		stack:   captureStack(1, 32),
		occurAt: time.Now(),
	}
}

// NewFatal 创建致命错误（自动捕获调用栈，通常触发告警）
func NewFatal(cause error) *XError {
	return &XError{
		level:   LevelFatal,
		code:    BizCodeUnknown,
		message: "系统发生严重错误",
		cause:   cause,
		stack:   captureStack(1, 64), // fatal 捕获更深的栈
		occurAt: time.Now(),
	}
}

// Wrap 包裹任意 error 为 XError（自动捕获调用栈）
// 如果 err 已经是 *XError，则直接返回；否则包装为 SysError
//
//	return xerror.Wrap(err, "查询用户失败")
func Wrap(err error, msg string) *XError {
	if err == nil {
		return nil
	}
	// 已经是 XError，不重复包裹（避免栈帧丢失）
	var xe *XError
	if errors.As(err, &xe) {
		return xe
	}
	return &XError{
		level:   LevelSys,
		code:    BizCodeUnknown,
		message: msg,
		cause:   err,
		stack:   captureStack(1, 32),
		occurAt: time.Now(),
	}
}

// Wrapf 带格式化消息的包裹
func Wrapf(err error, format string, args ...interface{}) *XError {
	return Wrap(err, fmt.Sprintf(format, args...))
}

// WrapBiz 将已知错误包裹为业务错误（不捕获栈）
func WrapBiz(err error, code BizCode, msg string) *XError {
	return &XError{
		level:   LevelBiz,
		code:    code,
		message: msg,
		cause:   err,
		occurAt: time.Now(),
	}
}

// =============================================================================
// 判断与断言
// =============================================================================

// As 尝试将 error 断言为 *XError（兼容链式 Unwrap）
func As(err error) (*XError, bool) {
	var xe *XError
	if errors.As(err, &xe) {
		return xe, true
	}
	return nil, false
}

// IsNotFound 判断是否"资源不存在"类错误
func IsNotFound(err error) bool {
	xe, ok := As(err)
	if !ok {
		return false
	}
	return xe.code == BizCodeNotFound ||
		xe.code == BizCodeUserNotFound ||
		xe.code == BizCodeMessageNotFound ||
		xe.code == BizCodeCommentNotFound
}

// IsForbidden 判断是否"无权限"类错误
func IsForbidden(err error) bool {
	xe, ok := As(err)
	if !ok {
		return false
	}
	return xe.code == BizCodeForbidden ||
		xe.code == BizCodeMessageForbidden ||
		xe.code == BizCodeCommentForbidden
}

// IsUnauthorized 判断是否"未认证"类错误
func IsUnauthorized(err error) bool {
	xe, ok := As(err)
	if !ok {
		return false
	}
	return xe.code == BizCodeUnauthorized ||
		xe.code == BizCodeTokenInvalid ||
		xe.code == BizCodeTokenExpired
}

// IsBizError 判断是否业务错误
func IsBizError(err error) bool {
	xe, ok := As(err)
	return ok && xe.IsBiz()
}

// IsSysError 判断是否系统错误
func IsSysError(err error) bool {
	xe, ok := As(err)
	return ok && xe.IsSys()
}

// =============================================================================
// 格式化输出
// =============================================================================

// Format 返回结构化的可读错误报告（用于日志记录）
func Format(err error) string {
	xe, ok := As(err)
	if !ok {
		return fmt.Sprintf("[UNKNOWN] %v", err)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("=== XError Report [%s] ===\n", xe.occurAt.Format("2006-01-02 15:04:05.000")))
	sb.WriteString(fmt.Sprintf("  Level  : %s\n", xe.level))
	sb.WriteString(fmt.Sprintf("  Code   : %d\n", xe.code))
	sb.WriteString(fmt.Sprintf("  Message: %s\n", xe.message))
	if xe.detail != "" {
		sb.WriteString(fmt.Sprintf("  Detail : %s\n", xe.detail))
	}
	if xe.cause != nil {
		sb.WriteString(fmt.Sprintf("  Cause  : %v\n", xe.cause))
	}
	if len(xe.stack) > 0 {
		sb.WriteString("  Stack  :\n")
		sb.WriteString(xe.stack.String())
	}
	return sb.String()
}

// =============================================================================
// PanicRecover —— 用于 defer 中捕获 panic 并转为 *XError
// =============================================================================

// Recover 在 defer 中捕获 panic，将其转为 *XError 并通过回调传出。
// 如果 panic 值本身是 *XError，则直接使用；否则包装为 FatalError。
//
// 典型使用：
//
//	func doSomething() (xe *xerror.XError) {
//	    defer xerror.Recover(func(e *xerror.XError) { xe = e })
//	    // ... 可能 panic 的代码
//	}
func Recover(callback func(*XError)) {
	r := recover()
	if r == nil {
		return
	}

	// 捕获 panic 发生时的栈（此处调用 captureStack 时 panic 已恢复，
	// 但 runtime.Stack 可在 recover() 之后调用）
	stack := captureStack(1, 64)

	var xe *XError
	switch v := r.(type) {
	case *XError:
		xe = v
		if len(xe.stack) == 0 {
			xe.stack = stack
		}
	case error:
		xe = &XError{
			level:   LevelFatal,
			code:    BizCodeUnknown,
			message: "程序发生 panic",
			cause:   v,
			stack:   stack,
			occurAt: time.Now(),
		}
	default:
		xe = &XError{
			level:   LevelFatal,
			code:    BizCodeUnknown,
			message: "程序发生 panic",
			cause:   fmt.Errorf("%v", v),
			stack:   stack,
			occurAt: time.Now(),
		}
	}

	if callback != nil {
		callback(xe)
	}
}

// RecoverToError 同 Recover，但返回 error（适合不需要详细字段的场景）
func RecoverToError(callback func(error)) {
	Recover(func(xe *XError) {
		if callback != nil {
			callback(xe)
		}
	})
}
