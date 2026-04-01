package log

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap/zapcore"
)

// TestInitZap_JSONFormat 测试 JSON 格式日志初始化
func TestInitZap_JSONFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	err := InitZap(
		WithOutput(buf),
		WithLevel(InfoLevel),
		WithFormat("json"),
	)
	if err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	Info("test message")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("日志输出不包含预期消息: %s", output)
	}
}

// TestInitZap_ConsoleFormat 测试 Console 格式日志初始化
func TestInitZap_ConsoleFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	err := InitZap(
		WithOutput(buf),
		WithLevel(InfoLevel),
		WithFormat("console"),
	)
	if err != nil {
		t.Fatalf("初始化日志失败: %v", err)
	}

	Info("console test")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "console test") {
		t.Errorf("日志输出不包含预期消息: %s", output)
	}
}

// TestInitZap_DebugLevel 测试 Debug 级别日志
func TestInitZap_DebugLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(DebugLevel),
	)

	Debug("debug message")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Errorf("Debug 级别日志未输出: %s", output)
	}
}

// TestInitZap_WarnLevel 测试 Warn 级别日志
func TestInitZap_WarnLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(WarnLevel),
	)

	Debug("debug message")
	Warn("warn message")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Errorf("Debug 级别日志不应该输出: %s", output)
	}
	if !strings.Contains(output, "warn message") {
		t.Errorf("Warn 级别日志应该输出: %s", output)
	}
}

// TestInitZap_ErrorLevel 测试 Error 级别日志
func TestInitZap_ErrorLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(ErrorLevel),
	)

	Info("info message")
	Error("error message")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if strings.Contains(output, "info message") {
		t.Errorf("Info 级别日志不应该输出: %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("Error 级别日志应该输出: %s", output)
	}
}

// TestWithOutput 测试自定义输出
func TestWithOutput(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	Info("custom output test")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "custom output test") {
		t.Errorf("自定义输出不工作: %s", output)
	}
}

// TestWithFormat 测试格式选项
func TestWithFormat(t *testing.T) {
	tests := []struct {
		name   string
		format string
	}{
		{"JSON", "json"},
		{"Console", "console"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			err := InitZap(
				WithOutput(buf),
				WithFormat(tt.format),
				WithLevel(InfoLevel),
			)
			if err != nil {
				t.Fatalf("初始化失败: %v", err)
			}
			Info("format test")
			time.Sleep(10 * time.Millisecond)
		})
	}
}

// TestInfo 测试 Info 日志方法
func TestInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	Info("info test", String("key", "value"))
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "info test") {
		t.Errorf("Info 日志未输出: %s", output)
	}
}

// TestInfof 测试 Infof 方法
func TestInfof(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	Infof("test %s", "message")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Infof 日志未输出: %s", output)
	}
}

// TestError 测试 Error 日志方法
func TestError(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(ErrorLevel),
	)

	Err := &testError{"test error"}
	Error("error occurred", Error(Err))
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "error occurred") {
		t.Errorf("Error 日志未输出: %s", output)
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// TestWarn 测试 Warn 日志方法
func TestWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(WarnLevel),
	)

	Warn("warning test", Int("code", 100))
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "warning test") {
		t.Errorf("Warn 日志未输出: %s", output)
	}
}

// TestDebug 测试 Debug 日志方法
func TestDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	_ = InitZap(
		WithOutput(buf),
		WithLevel(DebugLevel),
	)

	Debug("debug test", Int64("num", 123))
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "debug test") {
		t.Errorf("Debug 日志未输出: %s", output)
	}
}

// TestStringField 测试字符串字段
func TestStringField(t *testing.T) {
	field := String("key", "value")
	if field.Key != "key" || field.String != "value" {
		t.Errorf("String 字段创建失败: %v", field)
	}
}

// TestIntField 测试整数字段
func TestIntField(t *testing.T) {
	field := Int("count", 42)
	if field.Key != "count" || field.Integer != 42 {
		t.Errorf("Int 字段创建失败: %v", field)
	}
}

// TestInt64Field 测试 Int64 字段
func TestInt64Field(t *testing.T) {
	field := Int64("num", 12345)
	if field.Key != "num" || field.Integer != 12345 {
		t.Errorf("Int64 字段创建失败: %v", field)
	}
}

// TestFloat64Field 测试浮点数字段
func TestFloat64Field(t *testing.T) {
	field := Float64("rate", 3.14)
	if field.Key != "rate" || field.Float != 3.14 {
		t.Errorf("Float64 字段创建失败: %v", field)
	}
}

// TestBoolField 测试布尔字段
func TestBoolField(t *testing.T) {
	field := Bool("flag", true)
	if field.Key != "flag" || field.Boolean != true {
		t.Errorf("Bool 字段创建失败: %v", field)
	}
}

// TestTimeField 测试时间字段
func TestTimeField(t *testing.T) {
	now := time.Now()
	field := Time("time", now)
	if field.Key != "time" {
		t.Errorf("Time 字段创建失败: %v", field)
	}
}

// TestDurationField 测试持续时间字段
func TestDurationField(t *testing.T) {
	field := Duration("duration", time.Second)
	if field.Key != "duration" {
		t.Errorf("Duration 字段创建失败: %v", field)
	}
}

// TestAnyField 测试任意类型字段
func TestAnyField(t *testing.T) {
	field := Any("data", map[string]string{"key": "value"})
	if field.Key != "data" {
		t.Errorf("Any 字段创建失败: %v", field)
	}
}

// TestWithLogger 测试带日志的上下文
func TestWithLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	ctx := context.Background()
	ctx = WithLogger(ctx, logger)

	retrievedLogger := FromContext(ctx)
	if retrievedLogger == nil {
		t.Error("从上下文获取日志失败")
	}
}

// TestFromContext_Default 测试默认上下文日志获取
func TestFromContext_Default(t *testing.T) {
	ctx := context.Background()
	logger := FromContext(ctx)
	if logger == nil {
		t.Error("默认日志获取失败")
	}
}

// TestNew_创建新日志实例
func TestNew(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, err := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
		WithFormat("console"),
	)
	if err != nil {
		t.Fatalf("创建新日志实例失败: %v", err)
	}

	logger.Info("new logger test")
	time.Sleep(10 * time.Millisecond)

	output := buf.String()
	if !strings.Contains(output, "new logger test") {
		t.Errorf("新日志实例不工作: %s", output)
	}
}

// TestSync 测试日志同步
func TestSync(t *testing.T) {
	err := Sync()
	if err != nil {
		t.Errorf("Sync 失败: %v", err)
	}
}

// TestGetGlobalLogger 测试获取全局日志
func TestGetGlobalLogger(t *testing.T) {
	logger := GetGlobalLogger()
	if logger == nil {
		t.Error("获取全局日志失败")
	}
}

// TestSetGlobalLogger 测试设置全局日志
func TestSetGlobalLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _ := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
	)

	SetGlobalLogger(logger)
	time.Sleep(10 * time.Millisecond)
}

// TestOptions 测试各种选项函数
func TestOptions(t *testing.T) {
	buf := &bytes.Buffer{}
	encoderConfig := defaultEncoderConfig()

	_, err := New(
		WithOutput(buf),
		WithLevel(InfoLevel),
		WithFormat("console"),
		WithName("test-logger"),
		WithEncoderConfig(encoderConfig),
	)
	if err != nil {
		t.Fatalf("选项函数测试失败: %v", err)
	}
}

// TestLoggerLevels 测试不同日志级别输出
func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		level     zapcore.Level
		debugMsg  string
		infoMsg   string
		warnMsg   string
		errorMsg  string
		debugWant bool
		infoWant  bool
		warnWant  bool
		errorWant bool
	}{
		{DebugLevel, "d", "i", "w", "e", true, true, true, true},
		{InfoLevel, "d", "i", "w", "e", false, true, true, true},
		{WarnLevel, "d", "i", "w", "e", false, false, true, true},
		{ErrorLevel, "d", "i", "w", "e", false, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			buf := &bytes.Buffer{}
			_ = InitZap(
				WithOutput(buf),
				WithLevel(tt.level),
			)

			Debug(tt.debugMsg)
			Info(tt.infoMsg)
			Warn(tt.warnMsg)
			Error(tt.errorMsg)
			time.Sleep(10 * time.Millisecond)

			output := buf.String()
			gotDebug := strings.Contains(output, tt.debugMsg)
			gotInfo := strings.Contains(output, tt.infoMsg)
			gotWarn := strings.Contains(output, tt.warnMsg)
			gotError := strings.Contains(output, tt.errorMsg)

			if gotDebug != tt.debugWant {
				t.Errorf("Debug 级别期望 %v, 实际 %v", tt.debugWant, gotDebug)
			}
			if gotInfo != tt.infoWant {
				t.Errorf("Info 级别期望 %v, 实际 %v", tt.infoWant, gotInfo)
			}
			if gotWarn != tt.warnWant {
				t.Errorf("Warn 级别期望 %v, 实际 %v", tt.warnWant, gotWarn)
			}
			if gotError != tt.errorWant {
				t.Errorf("Error 级别期望 %v, 实际 %v", tt.errorWant, gotError)
			}
		})
	}
}
