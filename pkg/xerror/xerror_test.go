package xerror

import (
	"errors"
	"strings"
	"testing"
)

// ===== 构造函数测试 =====

func TestNewBiz(t *testing.T) {
	err := NewBiz(BizCodeNotFound, "用户不存在")

	if err.Level() != LevelBiz {
		t.Errorf("expect LevelBiz, got %v", err.Level())
	}
	if err.Code() != BizCodeNotFound {
		t.Errorf("expect BizCodeNotFound, got %v", err.Code())
	}
	if err.Message() != "用户不存在" {
		t.Errorf("unexpected message: %s", err.Message())
	}
	if len(err.Stack()) != 0 {
		t.Error("biz error should not capture stack by default")
	}
	if !err.IsBiz() {
		t.Error("IsBiz should return true")
	}
	if err.IsSys() {
		t.Error("IsSys should return false for biz error")
	}
}

func TestNewBizf(t *testing.T) {
	err := NewBizf(BizCodeBadRequest, "字段 %s 不合法", "email")
	if !strings.Contains(err.Message(), "email") {
		t.Errorf("message should contain 'email', got: %s", err.Message())
	}
}

func TestNewSys(t *testing.T) {
	cause := errors.New("connection refused")
	err := NewSys(cause)

	if err.Level() != LevelSys {
		t.Errorf("expect LevelSys, got %v", err.Level())
	}
	if err.Cause() != cause {
		t.Error("cause should be the original error")
	}
	if len(err.Stack()) == 0 {
		t.Error("sys error should capture stack")
	}
	if !err.IsSys() {
		t.Error("IsSys should return true")
	}
}

func TestNewFatal(t *testing.T) {
	cause := errors.New("out of memory")
	err := NewFatal(cause)

	if err.Level() != LevelFatal {
		t.Errorf("expect LevelFatal, got %v", err.Level())
	}
	if len(err.Stack()) == 0 {
		t.Error("fatal error should capture stack")
	}
}

// ===== Wrap 测试 =====

func TestWrap_NilReturnsNil(t *testing.T) {
	result := Wrap(nil, "msg")
	if result != nil {
		t.Error("Wrap(nil, ...) should return nil")
	}
}

func TestWrap_PlainError(t *testing.T) {
	orig := errors.New("io error")
	xe := Wrap(orig, "读取文件失败")

	if xe.Level() != LevelSys {
		t.Error("plain error should be wrapped as SysError")
	}
	if xe.Message() != "读取文件失败" {
		t.Errorf("unexpected message: %s", xe.Message())
	}
	if xe.Cause() != orig {
		t.Error("cause should be the original error")
	}
	if len(xe.Stack()) == 0 {
		t.Error("wrapped error should capture stack")
	}
}

func TestWrap_XErrorNotDoubleWrapped(t *testing.T) {
	orig := NewBiz(BizCodeNotFound, "资源不存在")
	wrapped := Wrap(orig, "再包一层")
	// 已经是 XError，应直接返回原值
	if wrapped != orig {
		t.Error("wrapping an XError should return the original XError")
	}
}

func TestWrapBiz(t *testing.T) {
	cause := errors.New("db error")
	xe := WrapBiz(cause, BizCodeServiceBusy, "服务繁忙")
	if xe.Level() != LevelBiz {
		t.Error("WrapBiz should produce a BizError")
	}
	if xe.Cause() != cause {
		t.Error("cause should be preserved")
	}
}

// ===== 链式 With 测试 =====

func TestWithChain(t *testing.T) {
	xe := NewBiz(BizCodeForbidden, "无权操作").
		WithDetail("user_id=1 resource_id=99").
		WithMsg("您没有权限执行此操作")

	if xe.Detail() != "user_id=1 resource_id=99" {
		t.Errorf("unexpected detail: %s", xe.Detail())
	}
	if xe.Message() != "您没有权限执行此操作" {
		t.Errorf("unexpected message: %s", xe.Message())
	}
}

func TestWithStack_ManualCapture(t *testing.T) {
	xe := NewBiz(BizCodeBadRequest, "参数错误").WithStack()
	if len(xe.Stack()) == 0 {
		t.Error("WithStack() should capture stack frames")
	}
}

// ===== As / 判断函数测试 =====

func TestAs_XError(t *testing.T) {
	xe := NewBiz(BizCodeNotFound, "not found")
	got, ok := As(xe)
	if !ok || got != xe {
		t.Error("As() should succeed for *XError")
	}
}

func TestAs_WrappedXError(t *testing.T) {
	inner := NewSys(errors.New("db"))
	outer := fmt.Errorf("outer: %w", inner)
	got, ok := As(outer)
	if !ok {
		t.Error("As() should unwrap to find *XError")
	}
	if got.Level() != LevelSys {
		t.Error("should recover the original XError")
	}
}

func TestIsNotFound(t *testing.T) {
	cases := []struct {
		err      error
		expected bool
	}{
		{NewBiz(BizCodeNotFound, "x"), true},
		{NewBiz(BizCodeUserNotFound, "x"), true},
		{NewBiz(BizCodeMessageNotFound, "x"), true},
		{NewBiz(BizCodeForbidden, "x"), false},
		{errors.New("plain"), false},
	}
	for _, c := range cases {
		if IsNotFound(c.err) != c.expected {
			t.Errorf("IsNotFound(%v) = %v, want %v", c.err, !c.expected, c.expected)
		}
	}
}

func TestIsForbidden(t *testing.T) {
	if !IsForbidden(NewBiz(BizCodeForbidden, "x")) {
		t.Error("BizCodeForbidden should be forbidden")
	}
	if !IsForbidden(NewBiz(BizCodeMessageForbidden, "x")) {
		t.Error("BizCodeMessageForbidden should be forbidden")
	}
	if IsForbidden(NewBiz(BizCodeNotFound, "x")) {
		t.Error("BizCodeNotFound should not be forbidden")
	}
}

func TestIsUnauthorized(t *testing.T) {
	if !IsUnauthorized(NewBiz(BizCodeTokenExpired, "x")) {
		t.Error("token expired should be unauthorized")
	}
	if IsUnauthorized(NewBiz(BizCodeNotFound, "x")) {
		t.Error("not found should not be unauthorized")
	}
}

// ===== 调用栈测试 =====

func TestStackFrames(t *testing.T) {
	xe := NewSys(errors.New("io"))

	if len(xe.Stack()) == 0 {
		t.Fatal("stack should not be empty")
	}

	found := false
	for _, f := range xe.Stack() {
		if strings.Contains(f.Function, "TestStackFrames") {
			found = true
			break
		}
	}
	if !found {
		t.Error("stack should contain the test function name")
	}

	// 每一帧都应该有文件名和行号
	for _, f := range xe.Stack() {
		if f.File == "" {
			t.Error("frame file should not be empty")
		}
		if f.Line == 0 {
			t.Error("frame line should not be 0")
		}
	}
}

func TestStackTrace_String(t *testing.T) {
	xe := NewSys(errors.New("x"))
	trace := xe.StackTrace()
	if trace == "" || trace == "<no stack>" {
		t.Error("stack trace should not be empty for sys error")
	}
	// 格式包含 # 编号
	if !strings.Contains(trace, "#0") {
		t.Errorf("stack trace should have frame indices: %s", trace)
	}
}

// ===== Recover（panic 捕获）测试 =====

func TestRecover_PanicWithError(t *testing.T) {
	var captured *XError

	func() {
		defer Recover(func(e *XError) { captured = e })
		panic(errors.New("something went wrong"))
	}()

	if captured == nil {
		t.Fatal("Recover should capture the panic")
	}
	if captured.Level() != LevelFatal {
		t.Errorf("panic should become FatalError, got %v", captured.Level())
	}
	if len(captured.Stack()) == 0 {
		t.Error("recovered panic should have stack")
	}
}

func TestRecover_PanicWithXError(t *testing.T) {
	orig := NewFatal(errors.New("fatal"))
	var captured *XError

	func() {
		defer Recover(func(e *XError) { captured = e })
		panic(orig)
	}()

	if captured != orig {
		t.Error("panicking with *XError should return the same XError")
	}
}

func TestRecover_PanicWithString(t *testing.T) {
	var captured *XError

	func() {
		defer Recover(func(e *XError) { captured = e })
		panic("index out of range")
	}()

	if captured == nil {
		t.Fatal("string panic should also be captured")
	}
	if captured.Level() != LevelFatal {
		t.Error("string panic should be FatalError")
	}
}

func TestRecover_NoPanic(t *testing.T) {
	var captured *XError

	func() {
		defer Recover(func(e *XError) { captured = e })
		// 不触发 panic
		_ = 1 + 1
	}()

	if captured != nil {
		t.Error("no panic should not trigger Recover callback")
	}
}

// ===== error 接口 / errors.Is 测试 =====

func TestErrorInterface(t *testing.T) {
	cause := errors.New("original")
	xe := NewSys(cause)

	// errors.Is 应该能找到原始 cause
	if !errors.Is(xe, cause) {
		t.Error("errors.Is should find cause through Unwrap")
	}
}

func TestErrorString(t *testing.T) {
	xe := NewBiz(BizCodeNotFound, "not found")
	s := xe.Error()
	if !strings.Contains(s, "BIZ") {
		t.Errorf("error string should contain level, got: %s", s)
	}
	if !strings.Contains(s, "not found") {
		t.Errorf("error string should contain message, got: %s", s)
	}
}

// ===== Format 输出测试 =====

func TestFormat_SysError(t *testing.T) {
	xe := NewSys(errors.New("connect refused")).
		WithMsg("数据库连接失败").
		WithDetail("host=127.0.0.1 port=3306")

	report := Format(xe)
	for _, substr := range []string{"SYS", "数据库连接失败", "host=127.0.0.1", "Stack"} {
		if !strings.Contains(report, substr) {
			t.Errorf("Format() should contain %q\nGot:\n%s", substr, report)
		}
	}
}

func TestFormat_NonXError(t *testing.T) {
	plain := errors.New("plain error")
	report := Format(plain)
	if !strings.Contains(report, "UNKNOWN") {
		t.Errorf("Format() for non-XError should contain UNKNOWN, got: %s", report)
	}
}

// ===== HTTPStatus 测试 =====

func TestHTTPStatus(t *testing.T) {
	cases := []struct {
		code     BizCode
		expected int
	}{
		{BizCodeNotFound, 404},
		{BizCodeForbidden, 403},
		{BizCodeUnauthorized, 401},
		{BizCodeConflict, 409},
		{BizCodeUnknown, 500}, // 未映射的默认 500
	}
	for _, c := range cases {
		xe := NewBiz(c.code, "test")
		if xe.HTTPStatus() != c.expected {
			t.Errorf("BizCode(%d).HTTPStatus() = %d, want %d", c.code, xe.HTTPStatus(), c.expected)
		}
	}
}

// ===== Level.String 测试 =====

func TestLevelString(t *testing.T) {
	if LevelBiz.String() != "BIZ" {
		t.Error("LevelBiz.String() should be BIZ")
	}
	if LevelSys.String() != "SYS" {
		t.Error("LevelSys.String() should be SYS")
	}
	if LevelFatal.String() != "FATAL" {
		t.Error("LevelFatal.String() should be FATAL")
	}
}

// 辅助：让 fmt.Errorf("%w") 在测试文件中编译通过
var _ = fmt.Errorf
