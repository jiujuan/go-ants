// Package main go-ants 用户服务入口（v0.3.1）
// 包含：用户注册/登录/登出 + 留言板 + 评论
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

// AppConfig 应用配置，对应 configs/config.yaml
type AppConfig struct {
	Server struct {
		Addr string `mapstructure:"addr"`
		Mode string `mapstructure:"mode"` // debug | release
	} `mapstructure:"server"`
	Database struct {
		Driver       string `mapstructure:"driver"`        // mysql | postgres
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
		Format string `mapstructure:"format"` // json | console
	} `mapstructure:"log"`
	Auth struct {
		JWTSecret            string `mapstructure:"jwt_secret"`
		JWTIssuer            string `mapstructure:"jwt_issuer"`
		JWTExpiration        int    `mapstructure:"jwt_expiration"`         // 秒，默认 86400 (24h)
		JWTRefreshExpiration int    `mapstructure:"jwt_refresh_expiration"` // 秒，默认 2592000 (30d)
	} `mapstructure:"auth"`
}

func main() {
	// 优雅退出信号
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	// ===== Step 1: 加载配置 =====
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
	// 自动建表：User、Message、Comment
	if err := db.AutoMigrate(
		&domain.User{},
		&domain.Message{},
		&domain.Comment{},
	); err != nil {
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
		// Redis 不可用时优雅降级（Token 黑名单功能关闭）
		log.Warnf("connect redis failed: %v, token blacklist disabled", err)
		redisClient = nil
	}

	// ===== Step 5: 初始化 JWT =====
	jwtSecret := appCfg.Auth.JWTSecret
	if jwtSecret == "" {
		jwtSecret = "change-me-in-production-please"
		log.Warn("jwt_secret not set, using insecure default")
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

	// ===== Step 6: 依赖注入 =====

	// 数据层
	dataLayer, dataCleanup, err := data.New(db, redisClient)
	if err != nil {
		log.Errorf("init data layer failed: %v", err)
		os.Exit(1)
	}
	defer dataCleanup()

	// 仓储
	userRepo := data.NewUserRepo(dataLayer)
	msgRepo := data.NewMessageRepo(dataLayer)
	commentRepo := data.NewCommentRepo(dataLayer)

	var blacklistRepo domain.TokenBlacklistRepository
	if redisClient != nil {
		blacklistRepo = data.NewTokenBlacklistRepo(dataLayer)
	}

	// 服务层
	userSvc := service.NewUserService(userRepo, blacklistRepo, jwtAuth)
	msgSvc := service.NewMessageBoardService(msgRepo)
	commentSvc := service.NewCommentService(commentRepo, msgRepo)

	// Handler 层
	userHandler := handler.NewUserHandler(userSvc)
	msgHandler := handler.NewMessageBoardHandler(msgSvc)
	commentHandler := handler.NewCommentHandler(commentSvc)

	// ===== Step 7: HTTP 服务器 =====
	addr := appCfg.Server.Addr
	if addr == "" {
		addr = ":8080"
	}
	if appCfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	ginServer := transport.NewGinServer("go-ants-service",
		transport.WithAddr(addr),
	)

	// 注册全部路由
	router.Register(ginServer.Engine(), userHandler, msgHandler, commentHandler, jwtAuth, blacklistRepo)

	// ===== Step 8: 启动应用 =====
	app, appCleanup := ants.New(
		ants.WithName("go-ants-service"),
		ants.WithLogger(log.DefaultLogger()),
		ants.WithComponents(ginServer),
	)
	defer appCleanup()

	log.Infof("go-ants service v0.3.1 starting — addr=%s mode=%s", addr, appCfg.Server.Mode)

	if err := app.Run(); err != nil {
		log.Errorf("application exit with error: %v", err)
		os.Exit(1)
	}
}
