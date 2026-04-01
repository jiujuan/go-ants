package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/jiujuan/go-ants/pkg/log"
)

var (
	// Version 版本号
	Version = "v0.1.0"
	// Commit Git commit
	Commit = ""
	// Date 构建日期
	Date = ""
)

// NewCommand 创建根命令
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ants",
		Short: "go-ants is a web application framework",
		Long: `go-ants is a modern Go web application framework
inspired by kratos, providing a complete solution 
for building microservices or monolithic applications.`,
		Version: Version,
	}

	cmd.AddCommand(
		NewVersionCommand(),
		NewNewCommand(),
		NewRunCommand(),
		NewGenCommand(),
	)

	return cmd
}

// NewVersionCommand 创建版本命令
func NewVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of ants",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("ants version %s", Version)
			if Commit != "" {
				fmt.Printf("\ncommit: %s", Commit)
			}
			if Date != "" {
				fmt.Printf("\ndate: %s", Date)
			}
			fmt.Println()
		},
	}
}

// NewNewCommand 创建新项目命令
func NewNewCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "new [project-name]",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		Run:   runNew,
	}

	cmd.Flags().StringVar(&projectName, "name", "", "project name")
	cmd.Flags().StringVar(&moduleName, "module", "", "module name (github.com/username/project)")
	cmd.Flags().StringVar(&directory, "dir", "", "project directory")
	cmd.Flags().BoolVar(&useWire, "wire", true, "use google wire for DI")
	cmd.Flags().BoolVar(&useSQL, "sql", true, "use gorm for database")

	return cmd
}

var (
	projectName string
	moduleName  string
	directory   string
	useWire     bool
	useSQL      bool
)

func runNew(cmd *cobra.Command, args []string) {
	name := args[0]
	if projectName == "" {
		projectName = name
	}

	// 确定模块名
	if moduleName == "" {
		moduleName = "github.com/jiujuan/" + name
	}

	// 确定目录
	dir := directory
	if dir == "" {
		dir = name
	}

	// 创建项目结构
	if err := createProject(dir, name, moduleName); err != nil {
		log.Error("create project failed", log.Error(err))
		os.Exit(1)
	}

	log.Info("project created successfully", log.String("dir", dir))
}

// createProject 创建项目结构
func createProject(dir, name, module string) error {
	// 创建目录结构
	dirs := []string{
		filepath.Join(dir, "cmd", name),
		filepath.Join(dir, "configs"),
		filepath.Join(dir, "internal", "domain"),
		filepath.Join(dir, "internal", "data"),
		filepath.Join(dir, "internal", "server"),
		filepath.Join(dir, "internal", "service"),
		filepath.Join(dir, "pkg"),
	}

	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}

	// 创建 go.mod 文件
	if err := createGoMod(dir, module); err != nil {
		return err
	}

	// 创建 main.go
	if err := createMainGo(dir, name); err != nil {
		return err
	}

	// 创建配置文件
	if err := createConfig(dir); err != nil {
		return err
	}

	// 创建业务代码
	if err := createInternal(dir, name); err != nil {
		return err
	}

	return nil
}

func createGoMod(dir, module string) error {
	content := fmt.Sprintf(`module %s

go 1.21

require (
	github.com/jiujuan/go-ants v0.1.0
)
`, module)

	return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(content), 0644)
}

func createMainGo(dir, name string) error {
	tmpl := `package main

import (
	"context"
	"os"

	"{{.Module}}/cmd/{{.Name}}"
	"github.com/jiujuan/go-ants/pkg/log"
)

func main() {
	ctx := context.Background()

	// 初始化日志
	log.InitZap()

	// 运行应用
	if err := cmd.Run(ctx, os.Args); err != nil {
		log.Error("application failed", log.Error(err))
		os.Exit(1)
	}
}
`

	content, err := executeTemplate(tmpl, map[string]string{
		"Name":   name,
		"Module": moduleName,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "cmd", name, "main.go"), []byte(content), 0644)
}

func createConfig(dir string) error {
	content := `server:
  addr: :8080
  mode: debug

database:
  driver: mysql
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  max_idle_conns: 10
  max_open_conns: 100

redis:
  addr: "localhost:6379"
  password: ""
  db: 0

log:
  level: debug
  format: json

auth:
  jwt_secret: "your-secret-key"
  jwt_expiration: 3600
`

	return os.WriteFile(filepath.Join(dir, "configs", "config.yaml"), []byte(content), 0644)
}

func createInternal(dir, name string) error {
	// 创建 domain
	domainContent := `package domain

type GreeterUsecase struct {
}

func NewGreeterUsecase() *GreeterUsecase {
	return &GreeterUsecase{}
}

func (u *GreeterUsecase) SayHello(name string) string {
	return "Hello, " + name
}
`
	os.WriteFile(filepath.Join(dir, "internal", "domain", "greeter.go"), []byte(domainContent), 0644)

	// 创建 service
	svcContent := `package service

import (
	"context"

	pb "{{.Module}}/internal/service"
	"{{.Module}}/internal/domain"
)

type GreeterServiceImpl struct {
	uc *domain.GreeterUsecase
}

func NewGreeterServiceImpl(uc *domain.GreeterUsecase) pb.GreeterService {
	return &GreeterServiceImpl{uc: uc}
}

func (s *GreeterServiceImpl) SayHello(ctx context.Context, req *pb.HelloReq) (*pb.HelloResp, error) {
	return &pb.HelloResp{
		Message: s.uc.SayHello(req.Name),
	}, nil
}
`
	svcContent, _ := executeTemplate(svcContent, map[string]string{
		"Module": moduleName,
	})
	os.WriteFile(filepath.Join(dir, "internal", "service", "greeter.go"), []byte(svcContent), 0644)

	// 创建 data
	dataContent := `package data

import (
	"context"
)

type GreeterRepo struct {
	db interface{}
}

func NewGreeterRepo(db interface{}) *GreeterRepo {
	return &GreeterRepo{db: db}
}

func (r *GreeterRepo) Save(ctx context.Context, name string) error {
	return nil
}

func (r *GreeterRepo) FindByName(ctx context.Context, name string) (string, error) {
	return name, nil
}
`
	os.WriteFile(filepath.Join(dir, "internal", "data", "greeter.go"), []byte(dataContent), 0644)

	// 创建 server
	serverContent := `package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/jiujuan/go-ants/pkg/transport"
)

type GreeterHandler struct {
	transport *transport.GinServer
}

func NewGreeterHandler(transport *transport.GinServer) *GreeterHandler {
	return &GreeterHandler{transport: transport}
}

func (h *GreeterHandler) RegisterRouter(r *gin.Engine) {
	r.GET("/hello", h.Hello)
}

func (h *GreeterHandler) Hello(c *gin.Context) {
	name := c.Query("name")
	c.JSON(200, gin.H{
		"message": "Hello, " + name,
	})
}
`
	os.WriteFile(filepath.Join(dir, "internal", "server", "greeter.go"), []byte(serverContent), 0644)

	return nil
}

// ====== cmd/main.go ======

func createMainGoFile(dir, name string) error {
	mainContent := `package main

import (
	"context"
	"os"

	"github.com/jiujuan/go-ants/pkg/app"
	"github.com/jiujuan/go-ants/pkg/conf"
	"github.com/jiujuan/go-ants/pkg/log"
	"github.com/jiujuan/go-ants/pkg/transport"
)

func main() {
	// 加载配置
	c := conf.New()
	if err := c.Load("configs/config.yaml"); err != nil {
		log.Error("load config failed", log.Error(err))
		os.Exit(1)
	}

	// 创建 HTTP 服务器
	engine := transport.NewGinServer("server",
		transport.WithAddr(":8080"),
	)

	// 注册路由
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to go-ants!",
		})
	})

	// 创建应用
	application, cleanup := app.New(
		app.WithName("go-ants"),
		app.WithLogger(log.DefaultLogger()),
		app.WithComponents(engine),
	)
	defer cleanup()

	// 运行应用
	if err := application.Run(); err != nil {
		log.Error("application failed", log.Error(err))
		os.Exit(1)
	}
}
`

	content, err := executeTemplate(mainContent, map[string]string{
		"Name": name,
	})
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "cmd", name, "main.go"), []byte(content), 0644)
}

// ====== cmd/run.go ======

func createRunGoFile(dir, name string) error {
	runContent := `package cmd

import (
	"context"
	"os"

	"github.com/jiujuan/go-ants/pkg/app"
	"github.com/jiujuan/go-ants/pkg/conf"
	"github.com/jiujuan/go-ants/pkg/log"
	"github.com/jiujuan/go-ants/pkg/transport"
	"github.com/gin-gonic/gin"
)

func Run(ctx context.Context, args []string) error {
	// 加载配置
	c := conf.New()
	if err := c.Load("configs/config.yaml"); err != nil {
		log.Error("load config failed", log.Error(err))
		return err
	}

	// 创建 HTTP 服务器
	engine := transport.NewGinServer("server",
		transport.WithAddr(":8080"),
	)

	// 注册路由
	engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to go-ants!",
		})
	})

	// 创建应用
	application, cleanup := app.New(
		app.WithName("go-ants"),
		app.WithLogger(log.DefaultLogger()),
		app.WithComponents(engine),
	)
	defer cleanup()

	// 运行应用
	return application.Run()
}
`
	return os.WriteFile(filepath.Join(dir, "cmd", "run.go"), []byte(runContent), 0644)
}

func executeTemplate(tmpl string, data map[string]string) (string, error) {
	t, err := template.New("tmpl").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// NewRunCommand 创建运行命令
func NewRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run",
		Short: "Run the application",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runRun(); err != nil {
				log.Error("run failed", log.Error(err))
				os.Exit(1)
			}
		},
	}
}

func runRun() error {
	// 这里可以添加更多的运行逻辑
	return nil
}

// NewGenCommand 创建代码生成命令
func NewGenCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gen [command]",
		Short: "Generate code",
	}

	cmd.AddCommand(
		NewGenCRUDCommand(),
		NewGenProtoCommand(),
	)

	return cmd
}

// NewGenCRUDCommand 创建 CRUD 生成命令
func NewGenCRUDCommand() *cobra.Command {
	var modelName string

	cmd := &cobra.Command{
		Use:   "crud [model-name]",
		Short: "Generate CRUD code",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			modelName = args[0]
			if err := generateCRUD(modelName); err != nil {
				log.Error("generate CRUD failed", log.Error(err))
				os.Exit(1)
			}
		},
	}

	return cmd
}

func generateCRUD(modelName string) error {
	log.Info("generating CRUD for", log.String("model", modelName))
	// 这里可以实现 CRUD 代码生成逻辑
	return nil
}

// NewGenProtoCommand 创建 Proto 生成命令
func NewGenProtoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "proto",
		Short: "Generate proto files",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("generating proto files")
		},
	}
}
