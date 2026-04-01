package conf

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const (
	// EnvPrefix 环境变量前缀
	EnvPrefix = "ANTS"
)

// Config 是配置实例的包装器
type Config struct {
	v *viper.Viper
}

// New 创建新的配置实例
func New() *Config {
	v := viper.New()
	v.SetEnvPrefix(EnvPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	return &Config{v: v}
}

// NewWithConfig 使用现有的 viper 实例
func NewWithConfig(v *viper.Viper) *Config {
	return &Config{v: v}
}

// Load 加载配置文件
// 支持的格式: json, yaml, yml, toml, hcl, env, properties
// 搜索路径: 当前目录, ./configs, $HOME, etc.
func (c *Config) Load(configFile string, opts ...Option) error {
	options := &Options{
		configType: "yaml",
		dirs:       []string{".", "./configs", "./config"},
	}

	for _, opt := range opts {
		opt(options)
	}

	// 设置配置类型
	if options.configType != "" {
		c.v.SetConfigType(options.configType)
	}

	// 如果提供了配置文件路径
	if configFile != "" {
		c.v.SetConfigFile(configFile)
	} else {
		// 自动搜索配置文件
		c.v.AddConfigPath(options.dirs...)
		c.v.SetConfigName(options.configName)
	}

	// 设置默认值
	for key, value := range options.defaults {
		c.v.SetDefault(key, value)
	}

	// 读取配置
	if err := c.v.ReadInConfig(); err != nil {
		// 如果是配置文件未找到错误，且没有设置配置文件，则返回 nil
		if _, ok := err.(viper.ConfigFileNotFoundError); ok && configFile == "" {
			return nil
		}
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// 监听配置变化
	if options.watch {
		c.v.WatchConfig()
		if options.onConfigChange != nil {
			c.v.OnConfigChange(options.onConfigChange)
		}
	}

	return nil
}

// LoadFromBytes 从字节数组加载配置
func (c *Config) LoadFromBytes(data []byte, configType string) error {
	c.v.SetConfigType(configType)
	return c.v.ReadConfig(bytesToReader(data))
}

// LoadFromRemote 从远程源加载配置
func (c *Config) LoadFromRemote(opts ...RemoteOption) error {
	options := &RemoteOptions{
		provider: "etcd",
	}

	for _, opt := range opts {
		opt(options)
	}

	// 这里可以实现从 etcd, consul 等远程源加载配置
	// 目前为占位实现
	return nil
}

// Get 获取配置值
func (c *Config) Get(key string) interface{} {
	return c.v.Get(key)
}

// GetString 获取字符串配置
func (c *Config) GetString(key string) string {
	return c.v.GetString(key)
}

// GetInt 获取整数配置
func (c *Config) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetInt64 获取64位整数配置
func (c *Config) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

// GetBool 获取布尔配置
func (c *Config) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// GetStringSlice 获取字符串切片配置
func (c *Config) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// GetStringMap 获取字符串Map配置
func (c *Config) GetStringMap(key string) map[string]interface{} {
	return c.v.GetStringMap(key)
}

// GetDuration 获取时间持续配置
func (c *Config) GetDuration(key string) time.Duration {
	return c.v.GetDuration(key)
}

// Sub 获取子配置
func (c *Config) Sub(key string) *Config {
	return &Config{v: c.v.Sub(key)}
}

// Unmarshal 将配置解 marshal 到结构体
func (c *Config) Unmarshal(rawVal interface{}, opts ...Option) error {
	options := &structOptions{}

	for _, opt := range opts {
		opt(options)
	}

	return c.v.Unmarshal(rawVal, func(config *viper.Viper) error {
		configDecoder := viper.New()
		configDecoder.SetConfigType("yaml")
		// 使用 viper 的 DecoderConfig 选项
		return config.Unmarshal(rawVal)
	})
}

// UnmarshalKey 解码特定键的配置到结构体
func (c *Config) UnmarshalKey(key string, rawVal interface{}) error {
	return c.v.UnmarshalKey(key, rawVal)
}

// IsSet 检查键是否已设置
func (c *Config) IsSet(key string) bool {
	return c.v.IsSet(key)
}

// AllSettings 获取所有设置
func (c *Config) AllSettings() map[string]interface{} {
	return c.v.AllSettings()
}

// Set 设置配置值
func (c *Config) Set(key string, value interface{}) {
	c.v.Set(key, value)
}

// SetDefault 设置默认值
func (c *Config) SetDefault(key string, value interface{}) {
	c.v.SetDefault(key, value)
}

// GetViper 获取原始的 viper 实例
func (c *Config) GetViper() *viper.Viper {
	return c.v
}

// 辅助函数
func bytesToReader(data []byte) *strings.Reader {
	return strings.NewReader(string(data))
}

// Options 配置选项
type Options struct {
	configType     string
	configName     string
	dirs           []string
	defaults       map[string]interface{}
	watch          bool
	onConfigChange func(e fsnotify.Event)
}

// Option 是配置选项函数
type Option func(*Options)

// WithConfigType 设置配置类型
func WithConfigType(configType string) Option {
	return func(o *Options) {
		o.configType = configType
	}
}

// WithConfigName 设置配置名称
func WithConfigName(configName string) Option {
	return func(o *Options) {
		o.configName = configName
	}
}

// WithSearchDirs 设置搜索目录
func WithSearchDirs(dirs []string) Option {
	return func(o *Options) {
		o.dirs = dirs
	}
}

// WithDefault 设置默认值
func WithDefault(key string, value interface{}) Option {
	return func(o *Options) {
		if o.defaults == nil {
			o.defaults = make(map[string]interface{})
		}
		o.defaults[key] = value
	}
}

// WithDefaults 设置多个默认值
func WithDefaults(defaults map[string]interface{}) Option {
	return func(o *Options) {
		if o.defaults == nil {
			o.defaults = make(map[string]interface{})
		}
		for k, v := range defaults {
			o.defaults[k] = v
		}
	}
}

// WithWatch 监听配置变化
func WithWatch(watch bool) Option {
	return func(o *Options) {
		o.watch = watch
	}
}

// WithOnConfigChange 设置配置变化回调
func WithOnConfigChange(fn func(e interface{})) Option {
	return func(o *Options) {
		o.onConfigChange = func(e interface{}) {
			fn(e)
		}
	}
}

// RemoteOptions 远程配置选项
type RemoteOptions struct {
	provider  string
	endpoint  string
	secretKey string
	path      string
	timeout   time.Duration
}

// RemoteOption 是远程配置选项函数
type RemoteOption func(*RemoteOptions)

// WithRemoteProvider 设置远程配置提供者
func WithRemoteProvider(provider string) RemoteOption {
	return func(o *RemoteOptions) {
		o.provider = provider
	}
}

// WithRemoteEndpoint 设置远程端点
func WithRemoteEndpoint(endpoint string) RemoteOption {
	return func(o *RemoteOptions) {
		o.endpoint = endpoint
	}
}

// WithRemoteSecretKey 设置密钥
func WithRemoteSecretKey(secretKey string) RemoteOption {
	return func(o *RemoteOptions) {
		o.secretKey = secretKey
	}
}

// WithRemotePath 设置配置路径
func WithRemotePath(path string) RemoteOption {
	return func(o *RemoteOptions) {
		o.path = path
	}
}

// WithRemoteTimeout 设置超时时间
func WithRemoteTimeout(timeout time.Duration) RemoteOption {
	return func(o *RemoteOptions) {
		o.timeout = timeout
	}
}

// structOptions 结构体选项
type structOptions struct {
	decoderConfig *mapstructure.DecoderConfig
}

// StructOption 是结构体选项函数
type StructOption func(*structOptions)

// WithDecoderConfig 设置解码器配置
func WithDecoderConfig(config *mapstructure.DecoderConfig) StructOption {
	return func(o *structOptions) {
		o.decoderConfig = config
	}
}

// MustUnmarshal 必须解marshal，成功返回，否则panic
func (c *Config) MustUnmarshal(rawVal interface{}, opts ...Option) {
	if err := c.Unmarshal(rawVal, opts...); err != nil {
		panic(fmt.Sprintf("config unmarshal failed: %v", err))
	}
}

// GetSizeInBytes 获取字节大小
func (c *Config) GetSizeInBytes(key string) uintBytes {
	return uintBytes(c.v.GetString(key))
}

type uintBytes uint64

func (i uintBytes) String() string {
	return fmt.Sprintf("%d", i)
}

func (i uintBytes) Bytes() []byte {
	return []byte(i.String())
}
