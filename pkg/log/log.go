package log

import (
	"context"
	"io"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	// Field 是 zapcore.Field 的别名
	Field = zapcore.Field
	// Logger 是 zap.SugaredLogger 的别名
	Logger = zap.SugaredLogger
	// Config 是 zap.Config 的别名
	Config = zap.Config
)

// 全局默认日志实例
var defaultLogger *Logger
var globalLogger *zap.Logger

func init() {
	// 初始化默认日志实例
	_ = InitZap(WithOutput(os.Stdout), WithLevel(InfoLevel))
}

// InitZap 初始化 zap 日志
// Options 支持以下配置:
// - WithOutput: 设置输出位置
// - WithLevel: 设置日志级别
// - WithFormat: 设置日志格式 (json/console)
// - WithName: 设置日志名称
// - WithEncoderConfig: 自定义编码器配置
func InitZap(opts ...Option) error {
	options := &Options{
		level:         InfoLevel,
		format:        "console",
		output:        os.Stdout,
		encoderConfig: defaultEncoderConfig(),
	}

	for _, opt := range opts {
		opt(options)
	}

	var config Config
	if options.format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}
	config.EncoderConfig = *options.encoderConfig

	logger, err := config.Build()
	if err != nil {
		return err
	}

	globalLogger = logger
	defaultLogger = logger.Sugar()

	return nil
}

// defaultEncoderConfig 返回默认的编码器配置
func defaultEncoderConfig() *zapcore.EncoderConfig {
	return &zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// DefaultLogger 返回默认日志实例
func DefaultLogger() *Logger {
	if defaultLogger == nil {
		// 确保初始化
		_ = InitZap()
	}
	return defaultLogger
}

// WithLogger 返回带logger的context
func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, logger)
}

// FromContext 从context中获取logger
func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(loggerKey{}).(*Logger); ok && logger != nil {
		return logger
	}
	return defaultLogger
}

// 日志级别相关
const (
	DebugLevel  = zapcore.DebugLevel
	InfoLevel   = zapcore.InfoLevel
	WarnLevel   = zapcore.WarnLevel
	ErrorLevel  = zapcore.ErrorLevel
	DPanicLevel = zapcore.DPanicLevel
	PanicLevel  = zapcore.PanicLevel
	FatalLevel  = zapcore.FatalLevel
)

// 快捷方法
func Debug(msg string, fields ...Field) {
	defaultLogger.Debugw(msg, fields...)
}

func Debugf(template string, args ...interface{}) {
	defaultLogger.Debugf(template, args...)
}

func Info(msg string, fields ...Field) {
	defaultLogger.Infow(msg, fields...)
}

func Infof(template string, args ...interface{}) {
	defaultLogger.Infof(template, args...)
}

func Warn(msg string, fields ...Field) {
	defaultLogger.Warnw(msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	defaultLogger.Warnf(template, args...)
}

func Error(msg string, fields ...Field) {
	defaultLogger.Errorw(msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	defaultLogger.Errorf(template, args...)
}

func DPanic(msg string, fields ...Field) {
	defaultLogger.DPanicw(msg, fields...)
}

func DPanicf(template string, args ...interface{}) {
	defaultLogger.DPanicf(template, args...)
}

func Panic(msg string, fields ...Field) {
	defaultLogger.Panicw(msg, fields...)
}

func Panicf(template string, args ...interface{}) {
	defaultLogger.Panicf(template, args...)
}

func Fatal(msg string, fields ...Field) {
	defaultLogger.Fatalw(msg, fields...)
}

func Fatalf(template string, args ...interface{}) {
	defaultLogger.Fatalf(template, args...)
}

// 带有 context 的日志方法
func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Infow(msg, fields...)
}

func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	FromContext(ctx).Errorw(msg, fields...)
}

// 字段构建快捷函数
func String(key string, value string) Field {
	return zap.String(key, value)
}

func Strings(key string, values []string) Field {
	return zap.Strings(key, values)
}

func Int(key string, value int) Field {
	return zap.Int(key, value)
}

func Int64(key string, value int64) Field {
	return zap.Int64(key, value)
}

func Float64(key string, value float64) Field {
	return zap.Float64(key, value)
}

func Bool(key string, value bool) Field {
	return zap.Bool(key, value)
}

func Time(key string, value time.Time) Field {
	return zap.Time(key, value)
}

func Duration(key string, value time.Duration) Field {
	return zap.Duration(key, value)
}

func Error(err error) Field {
	return zap.Error(err)
}

func Any(key string, value interface{}) Field {
	return zap.Any(key, value)
}

// Sync 同步日志缓冲
func Sync() error {
	if globalLogger != nil {
		return globalLogger.Sync()
	}
	return nil
}

// New 创建新的日志实例
func New(opts ...Option) (*Logger, error) {
	options := &Options{
		level:         InfoLevel,
		format:        "console",
		output:        os.Stdout,
		encoderConfig: defaultEncoderConfig(),
	}

	for _, opt := range opts {
		opt(options)
	}

	var config Config
	if options.format == "json" {
		config = zap.NewProductionConfig()
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig = *options.encoderConfig

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}

// GetGlobalLogger 获取全局logger实例（用于低级操作）
func GetGlobalLogger() *zap.Logger {
	return globalLogger
}

// SetGlobalLogger 设置全局logger实例
func SetGlobalLogger(logger *zap.Logger) {
	globalLogger = logger
	defaultLogger = logger.Sugar()
}

// 类型定义
type loggerKey struct{}

type Options struct {
	level         zapcore.Level
	format        string
	output        io.Writer
	name          string
	encoderConfig *zapcore.EncoderConfig
}

type Option func(*Options)

// WithOutput 设置输出
func WithOutput(output io.Writer) Option {
	return func(o *Options) {
		o.output = output
	}
}

// WithLevel 设置级别
func WithLevel(level zapcore.Level) Option {
	return func(o *Options) {
		o.level = level
	}
}

// WithFormat 设置格式
func WithFormat(format string) Option {
	return func(o *Options) {
		o.format = format
	}
}

// WithName 设置名称
func WithName(name string) Option {
	return func(o *Options) {
		o.name = name
	}
}

// WithEncoderConfig 设置编码器配置
func WithEncoderConfig(cfg *zapcore.EncoderConfig) Option {
	return func(o *Options) {
		o.encoderConfig = cfg
	}
}
