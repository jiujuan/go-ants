// Package main 用户服务入口
// 基于 go-ants 框架实现用户注册、登录、登出 HTTP API
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	ants "github.com/jiujuan/go-ants/pkg/app"
	"github.com/jiujuan/go-ants/pkg/auth"
	"github.com/jiujuan/go-ants/pkg/conf"
	"github.com/jiujuan/go-ants/pkg/database"
	"github.com/jiujuan/go-ants/pkg/log"
	pkgredis "github.com/jiujuan/go-ants/pkg/redis"
	"github.com/jiujuan/go-ants/pkg/transport"

	"github.com/jiujuan/go-ants/internal/data"
	"github.com/jiujuan/go-ants/internal/domain"
	"github.com/jiujuan/go-ants/internal/handler"
	"github.com/jiujuan/go-ants/internal/router"
	"github.com/jiujuan/go-ants/internal/service"
)

// ===== 配置结构体 =====

// AppConfig 应用配置（对应 configs/config.yaml）
type AppConfig struct {
	Server struct {
		Addr string `mapstructure:"addr"`
		Mode string `mapstructure:"mode"`
	} `mapstructure:"server"`
	Database struct {
		Driver       string `mapstructure:"driver"`
		DSN          string `mapstructure:"dsn"`
		MaxIdleConns int    `mapstructure:"max_idle_conns"`
		MaxOpenConns int    `mapstructure:"max_open_conns"`
	} `mapstructure:"database"`
	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	} `mapstructure:"redis"`
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
	} `mapstructure:"log"`
	Auth struct {
		JWTSecret            string `mapstructure:"jwt_secret"`
		JWTIssuer            string `mapstructure:"jwt_issuer"`
		JWTExpiration        int    `mapstructure:"jwt_expiration"`         // 秒
		JWTRefreshExpiration int    `mapstructure:"jwt_refresh_expiration"` // 秒
	} `mapstructure:"auth"`
}

func main() {
	// 注册信号，实现优雅退出
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// ===== Step 1: 加载配置文件 =====
	cfg := conf.New()
	if err := cfg.Load("configs/config.yaml"); err != nil {
		log.Errorf("load config failed: %v", err)
		os.Exit(1)
	}
	var appCfg AppConfig
	if err := cfg.GetViper().Unmarshal(&appCfg); err != nil {
		log.Errorf("unmarshal config failed: %v", err)
		os.Exit(1)
	}

	// ===== Step 2: 初始化日志 =====
	if err := log.InitZap(log.WithFormat(appCfg.Log.Format)); err != nil {
		log.Errorf("init log failed: %v", err)
		os.Exit(1)
	}
	defer log.Sync()

	// ===== Step 3: 初始化数据库 =====
	driver := database.DBTypeMySQL
	if appCfg.Database.Driver == "postgres" {
		driver = database.DBTypePostgres
	}
	db, err := database.New(ctx,
		database.WithDriver(driver),
		database.WithDSN(appCfg.Database.DSN),
		database.WithMaxIdleConns(appCfg.Database.MaxIdleConns),
		database.WithMaxOpenConns(appCfg.Database.MaxOpenConns),
		database.WithDebug(appCfg.Server.Mode == "debug"),
	)
	if err != nil {
		log.Errorf("connect database failed: %v", err)
		os.Exit(1)
	}
	// 自动迁移建表
	if err := db.AutoMigrate(&domain.User{}); err != nil {
		log.Errorf("auto migrate failed: %v", err)
		os.Exit(1)
	}

	// ===== Step 4: 初始化 Redis =====
	redisAddr := appCfg.Redis.Addr
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisClient, err := pkgredis.New(ctx,
		pkgredis.WithAddr(redisAddr),
		pkgredis.WithPassword(appCfg.Redis.Password),
		pkgredis.WithDB(appCfg.Redis.DB),
	)
	if err != nil {
		// Redis 连接失败时打印警告，不阻断启动（黑名单功能降级）
		log.Warnf("connect redis failed: %v, token blacklist disabled", err)
		redisClient = nil
	}

	// ===== Step 5: 初始化 JWT 认证器 =====
	jwtSecret := appCfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production"
	}
	expiration := time.Duration(appCfg.Auth.JWTExpiration) * time.Second
	if expiration == 0 {
		expiration = 24 * time.Hour
	}
	refreshExpiration := time.Duration(appCfg.Auth.JWTRefreshExpiration) * time.Second
	if refreshExpiration == 0 {
		refreshExpiration = 30 * 24 * time.Hour
	}
	jwtAuth := auth.New(
		auth.WithHMACSigningKey(jwtSecret),
		auth.WithIssuer(appCfg.Auth.JWTIssuer),
		auth.WithExpiration(expiration),
		auth.WithRefreshExpiration(refreshExpiration),
	)

	// ===== Step 6: 依赖注入（手动 wire）=====
	// data 层
	dataLayer, dataCleanup, err := data.New(db, redisClient)
	if err != nil {
		log.Errorf("init data layer failed: %v", err)
		os.Exit(1)
	}
	defer dataCleanup()

	userRepo := data.NewUserRepo(dataLayer)
	var blacklistRepo domain.TokenBlacklistRepository
	if redisClient != nil {
		blacklistRepo = data.NewTokenBlacklistRepo(dataLayer)
	}

	// service 层
	userSvc := service.NewUserService(userRepo, blacklistRepo, jwtAuth)

	// handler 层
	userHandler := handler.NewUserHandler(userSvc)

	// ===== Step 7: 初始化 HTTP 服务器 =====
	addr := appCfg.Server.Addr
	if addr == "" {
		addr = ":8080"
	}
	if appCfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginServer := transport.NewGinServer("user-service",
		transport.WithAddr(addr),
	)

	// 注册路由
	router.Register(ginServer.Engine(), userHandler, jwtAuth, blacklistRepo)

	// ===== Step 8: 启动应用，管理生命周期 =====
	app, appCleanup := ants.New(
		ants.WithName("user-service"),
		ants.WithLogger(log.DefaultLogger()),
		ants.WithComponents(ginServer),
	)
	defer appCleanup()

	log.Infof("user service starting, addr=%s mode=%s", addr, appCfg.Server.Mode)

	if err := app.Run(); err != nil {
		log.Errorf("application exit with error: %v", err)
		os.Exit(1)
	}
}
