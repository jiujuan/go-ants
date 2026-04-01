package validator

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
)

// Validator 验证器实例
type Validator struct {
	validate *validator.Validate
	trans    ut.Translator
}

// New 创建新的验证器
func New(opts ...Option) (*Validator, error) {
	options := &Options{
		tagName: "validate",
		lang:    "en",
	}

	for _, opt := range opts {
		opt(options)
	}

	v := validator.New()
	v.SetTagName(options.tagName)

	// 设置翻译器
	uni := ut.New(en.New())
	trans, found := uni.GetTranslator(options.lang)
	if !found {
		return nil, fmt.Errorf("translator not found for locale: %s", options.lang)
	}

	// 注册翻译
	switch options.lang {
	case "zh":
		if err := zhTranslations.RegisterDefaultTranslations(v, trans); err != nil {
			return nil, err
		case "en":
		if err := enTranslations.RegisterDefaultTranslations(v, trans); err != nil {
			return nil, err
		}
	}

	return &Validator{
		validate: v,
		trans:    trans,
	}, nil
}

// Validate 验证结构体
func (v *Validator) Validate(s interface{}) error {
	err := v.validate.Struct(s)
	if err != nil {
		return v.translateErrors(err)
	}
	return nil
}

// ValidateVar 验证变量
func (v *Validator) ValidateVar(field interface{}, tag string) error {
	err := v.validate.Var(field, tag)
	if err != nil {
		return v.translateErrors(err)
	}
	return nil
}

// Struct 获取结构体验证元数据
func (v *Validator) Struct(s interface{}) error {
	return v.validate.Struct(s)
}

// Var 获取变量验证元数据
func (v *Validator) Var(field interface{}, tag string) error {
	return v.validate.Var(field, tag)
}

// translateErrors 翻译验证错误
func (v *Validator) translateErrors(err error) error {
	if errs, ok := err.(validator.ValidationErrors); ok {
		errMsgs := make([]string, 0, len(errs))
		for _, err := range errs {
			errMsgs = append(errMsgs, err.Translate(v.trans))
		}
		return &ValidationError{
			Errors: errMsgs,
		}
	}
	return err
}

// ValidationError 验证错误
type ValidationError struct {
	Errors []string
}

func (e *ValidationError) Error() string {
	return strings.Join(e.Errors, "; ")
}

// ===== 全局默认验证器 =====

var defaultValidator *Validator

func init() {
	var err error
	defaultValidator, err = New()
	if err != nil {
		panic(fmt.Sprintf("failed to init default validator: %v", err))
	}
}

// Default 获取默认验证器
func Default() *Validator {
	return defaultValidator
}

// Validate 使用默认验证器验证
func Validate(s interface{}) error {
	return defaultValidator.Validate(s)
}

// ValidateVar 使用默认验证器验证变量
func ValidateVar(field interface{}, tag string) error {
	return defaultValidator.ValidateVar(field, tag)
}

// ===== 自定义验证规则 =====

// RegisterValidation 注册自定义验证规则
func (v *Validator) RegisterValidation(tag string, fn validator.Func, callValidationEvenIfNull ...bool) error {
	return v.validate.RegisterValidation(tag, fn, callValidationEvenIfNull...)
}

// RegisterAlias 注册验证别名
func (v *Validator) RegisterAlias(alias, tag string) {
	v.validate.RegisterAlias(alias, tag)
}

// ===== 常用验证函数 =====

// IsEmail 验证邮箱格式
func IsEmail(email string) bool {
	err := defaultValidator.ValidateVar(email, "email")
	return err == nil
}

// IsURL 验证 URL 格式
func IsURL(url string) bool {
	err := defaultValidator.ValidateVar(url, "url")
	return err == nil
}

// IsPhone 验证手机号格式
func IsPhone(phone string) bool {
	err := defaultValidator.ValidateVar(phone, "e164")
	return err == nil
}

// IsUUID 验证 UUID 格式
func IsUUID(uuid string) bool {
	err := defaultValidator.ValidateVar(uuid, "uuid")
	return err == nil
}

// IsAlpha 验证是否为字母
func IsAlpha(str string) bool {
	err := defaultValidator.ValidateVar(str, "alpha")
	return err == nil
}

// IsNumeric 验证是否为数字
func IsNumeric(str string) bool {
	err := defaultValidator.ValidateVar(str, "numeric")
	return err == nil
}

// IsAlphanumeric 验证是否为字母数字
func IsAlphanumeric(str string) bool {
	err := defaultValidator.ValidateVar(str, "alphanum")
	return err == nil
}

// IsMinLength 验证最小长度
func IsMinLength(str string, length int) bool {
	err := defaultValidator.ValidateVar(str, fmt.Sprintf("min=%d", length))
	return err == nil
}

// IsMaxLength 验证最大长度
func IsMaxLength(str string, length int) bool {
	err := defaultValidator.ValidateVar(str, fmt.Sprintf("max=%d", length))
	return err == nil
}

// IsRange 验证范围
func IsRange(num interface{}, min, max float64) bool {
	err := defaultValidator.ValidateVar(num, fmt.Sprintf("min=%f,max=%f", min, max))
	return err == nil
}

// ===== 选项配置 =====

// Option 是验证器选项函数
type Option func(*Options)

type Options struct {
	tagName string
	lang    string
}

// WithTagName 设置验证标签名
func WithTagName(tagName string) Option {
	return func(o *Options) {
		o.tagName = tagName
	}
}

// WithLanguage 设置语言
func WithLanguage(lang string) Option {
	return func(o *Options) {
		o.lang = lang
	}
}

// ===== 验证结果辅助类型 =====

// ValidationResult 验证结果
type ValidationResult struct {
	Valid  bool
	Errors map[string]string
}

// NewValidationResult 创建验证结果
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}
}

// AddError 添加错误
func (r *ValidationResult) AddError(field, message string) {
	r.Valid = false
	r.Errors[field] = message
}

// GetError 获取错误消息
func (r *ValidationResult) GetError(field string) string {
	return r.Errors[field]
}

// HasError 检查是否有错误
func (r *ValidationResult) HasError() bool {
	return !r.Valid
}

// Error 返回错误字符串
func (r *ValidationResult) Error() string {
	if r.Valid {
		return ""
	}
	var msgs []string
	for _, msg := range r.Errors {
		msgs = append(msgs, msg)
	}
	return strings.Join(msgs, "; ")
}

// ValidateStruct 验证结构体并返回详细结果
func (v *Validator) ValidateStruct(s interface{}) *ValidationResult {
	result := NewValidationResult()
	err := v.Validate(s)
	if err != nil {
		result.Valid = false
		if verr, ok := err.(*ValidationError); ok {
			for _, e := range verr.Errors {
				parts := strings.SplitN(e, " ", 2)
				if len(parts) >= 2 {
					field := strings.Trim(parts[0], "\"")
					result.AddError(field, parts[1])
				}
			}
		}
	}
	return result
}

// GetStructFields 获取结构体字段信息
func GetStructFields(s interface{}) []FieldInfo {
	t := reflect.TypeOf(s)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	fields := make([]FieldInfo, 0)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fields = append(fields, FieldInfo{
			Name:    field.Name,
			Type:    field.Type.Name(),
			Tag:     string(field.Tag),
			JSONName: field.Tag.Get("json"),
		})
	}
	return fields
}

// FieldInfo 字段信息
type FieldInfo struct {
	Name    string
	Type    string
	Tag     string
	JSONName string
}
