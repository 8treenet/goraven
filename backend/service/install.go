package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util"

	"github.com/8treenet/freedom"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *InstallService {
			return &InstallService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *InstallService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// InstallService 系统初始化服务，处理数据库连接检查和初始化流程
type InstallService struct {
	Worker freedom.Worker
}

// CheckDB 测试数据库连接是否可用，打开连接后立即关闭
func (service *InstallService) CheckDB(req *vo.InstallDBCheckReq) error {
	db, err := service.openDB(req.DBType, req.DBAddr, req.DBPort, req.DBUser, req.DBPass, req.DBName)
	if err != nil {
		return err
	}
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
	return nil
}

// CheckRedis 测试 Redis 连接是否可用
func (service *InstallService) CheckRedis(req *vo.InstallRedisCheckReq) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     service.buildRedisAddr(req.RedisAddr, req.RedisPort),
		Password: req.RedisPass,
		DB:       req.RedisDB,
	})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}
	return nil
}

// installDone 是进程级的一次性安装守卫。
// 即使 config.yaml 写入成功但自重启失败，本进程也不会第二次执行安装；
// 重启之后则由 config.System.Initialized 阻止安装路由重新注册。
var installDone atomic.Bool

// Init 执行系统初始化：验证连接、写入配置、建表、创建超管用户、标记完成并重启服务
func (service *InstallService) Init(req *vo.InstallInitReq) (*vo.InstallInitRsp, error) {
	// 幂等守卫必须放在最前面：避免匿名重复调用重设超管口令。
	if config.Get().System.Initialized || installDone.Load() {
		return nil, errs.NewFormatError("system is already initialized", "系统已完成初始化")
	}
	if !util.IsValidUsername(req.Username) {
		return nil, errs.ErrInvalidUsername
	}

	db, err := service.openDB(req.DBType, req.DBAddr, req.DBPort, req.DBUser, req.DBPass, req.DBName)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		defer sqlDB.Close()
	}

	// 如果选择 Redis 缓存，测试 Redis 连接
	var redisAddr string
	if req.CacheType == "redis" {
		redisAddr = service.buildRedisAddr(req.RedisAddr, req.RedisPort)
		if err := service.CheckRedis(&vo.InstallRedisCheckReq{
			RedisAddr: req.RedisAddr,
			RedisPort: req.RedisPort,
			RedisPass: req.RedisPass,
			RedisDB:   req.RedisDB,
		}); err != nil {
			return nil, err
		}
	}

	// 写入 config.yaml（批量一次性落盘；不含 initialized，建表与超管创建成功后才标记完成）
	dsn := service.buildDSN(req.DBType, req.DBAddr, req.DBPort, req.DBUser, req.DBPass, req.DBName)
	cfg := config.Get()
	entries := []config.ConfigEntry{
		{Section: "system", Key: "language", Value: req.Language},
		{Section: "system", Key: "db_type", Value: req.DBType},
		{Section: "system", Key: "cache_type", Value: req.CacheType},
		{Section: "db", Key: "addr", Value: dsn},
	}
	if req.CacheType == "redis" {
		entries = append(entries,
			config.ConfigEntry{Section: "redis", Key: "addr", Value: redisAddr},
			config.ConfigEntry{Section: "redis", Key: "password", Value: req.RedisPass},
			config.ConfigEntry{Section: "redis", Key: "db", Value: fmt.Sprintf("%d", req.RedisDB)},
		)
	}
	if err := cfg.ModifyConfigs(entries); err != nil {
		return nil, fmt.Errorf("persist install config: %w", err)
	}
	cfg.System.Language = req.Language
	cfg.System.DBType = req.DBType

	// 建表
	if err = repository.Merge(db); err != nil {
		return nil, fmt.Errorf("init database schema: %w", err)
	}

	// 初始化数据
	if err = repository.Seed(db); err != nil {
		freedom.Logger().Errorf("seed data: %v", err)
	}

	passwordHash, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash super admin password: %w", err)
	}

	var userId string
	var existingUser po.User
	err = db.Where("super_admin = ?", 1).First(&existingUser).Error
	if err == nil {
		// 超管已存在，更新除 userId 外的所有字段
		existingUser.Username = req.Username
		existingUser.Password = passwordHash
		existingUser.Email = req.Email
		existingUser.Role = 1
		existingUser.SuperAdmin = 1
		existingUser.Status = 1
		if err := db.Save(&existingUser).Error; err != nil {
			return nil, fmt.Errorf("update super admin failed: %w", err)
		}
		userId = existingUser.UserId
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		// 超管不存在，新建
		userId = util.UUID()
		user := &po.User{
			UserId:     userId,
			Username:   req.Username,
			Password:   passwordHash,
			Email:      req.Email,
			Role:       1,
			SuperAdmin: 1,
			Status:     1,
		}
		if err := db.Create(user).Error; err != nil {
			return nil, fmt.Errorf("create super admin failed: %w", err)
		}
	} else {
		return nil, fmt.Errorf("query super admin failed: %w", err)
	}

	// 标记初始化完成：单独一次写入，确保上面全部成功后才落 initialized=true。
	// 写失败则不置内存标记、不置进程守卫，允许修复环境后重试。
	if err := cfg.ModifyConfig("system", "initialized", "true"); err != nil {
		return nil, fmt.Errorf("mark system initialized: %w", err)
	}
	cfg.System.Initialized = true
	installDone.Store(true)

	// 写入系统域名
	if req.Domain != "" {
		db.Model(&po.SystemSetting{}).Where("config_key = ?", "general.domain").Update("config_value", req.Domain)
	}

	// 延时重启，确保 HTTP 响应已发送
	go func() {
		time.Sleep(1 * time.Second)
		service.restartSelf()
	}()

	return &vo.InstallInitRsp{
		UserId:   userId,
		Username: req.Username,
	}, nil
}

// openDB 根据数据库类型和连接信息打开一个 gorm 数据库连接
func (service *InstallService) openDB(dbType, addr string, port int, user, pass, dbName string) (*gorm.DB, error) {
	dsn := service.buildDSN(dbType, addr, port, user, pass, dbName)
	var dialector gorm.Dialector
	switch dbType {
	case "mysql":
		dialector = mysql.Open(dsn)
	case "pg":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("unsupported db type: %s", dbType)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database connect failed: %w", err)
	}
	return db, nil
}

// buildDSN 根据数据库类型和连接信息构建 DSN 连接字符串，写入 config.yaml 的 db.addr 字段
func (service *InstallService) buildDSN(dbType, addr string, port int, user, pass, dbName string) string {
	host, p := service.splitHostPort(addr, port)
	switch dbType {
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pass, host, p, dbName)
	case "pg":
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, p, user, pass, dbName)
	case "sqlite":
		if addr != "" {
			return addr
		}
		return config.Get().Paths.SqliteDB
	}
	return addr
}

// splitHostPort 从地址和端口参数中分离 host 和 port，addr 可为空或含端口，port 参数优先
func (service *InstallService) splitHostPort(addr string, port int) (host, portStr string) {
	host = addr
	if host == "" {
		host = "localhost"
	}
	// 如果 addr 中包含端口，优先使用 addr 中的端口
	if strings.Contains(addr, ":") {
		h, p, err := net.SplitHostPort(addr)
		if err == nil {
			host = h
			portStr = p
			return
		}
	}
	if port > 0 {
		portStr = fmt.Sprintf("%d", port)
	} else {
		portStr = "3306"
	}
	return
}

// restartSelf 通过 syscall.Exec 重启当前进程。
// Exec 失败时必须退出：否则进程会带着已注册的安装路由继续运行，
// 而此时 DB 里超管已被重置，等于把「重复初始化」的窗口永久留着。
func (service *InstallService) restartSelf() {
	self, err := os.Executable()
	if err != nil {
		freedom.Logger().Errorf("resolve executable for restart: %v", err)
		os.Exit(1)
	}
	if err := syscall.Exec(self, os.Args, os.Environ()); err != nil {
		freedom.Logger().Errorf("restart after install failed: %v", err)
		os.Exit(1)
	}
}

// buildRedisAddr 拼接 Redis 连接地址 host:port
func (service *InstallService) buildRedisAddr(addr string, port int) string {
	host, p := service.splitHostPort(addr, port)
	return fmt.Sprintf("%s:%s", host, p)
}
