package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/config"
	"goraven/util"
	"strings"
	"syscall"
	"time"

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

// Init 执行系统初始化：验证连接、写入配置、建表、创建超管用户、标记完成并重启服务
func (service *InstallService) Init(req *vo.InstallInitReq) (*vo.InstallInitRsp, error) {
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

	// 写入 config.yaml
	dsn := service.buildDSN(req.DBType, req.DBAddr, req.DBPort, req.DBUser, req.DBPass, req.DBName)
	cfg := config.Get()
	cfg.ModifyConfig("system", "language", req.Language)
	cfg.ModifyConfig("system", "db_type", req.DBType)
	cfg.ModifyConfig("system", "cache_type", req.CacheType)
	cfg.ModifyConfig("db", "addr", dsn)
	if req.CacheType == "redis" {
		cfg.ModifyConfig("redis", "addr", redisAddr)
		cfg.ModifyConfig("redis", "password", req.RedisPass)
		cfg.ModifyConfig("redis", "db", fmt.Sprintf("%d", req.RedisDB))
	}
	config.Get().System.Language = req.Language
	config.Get().System.DBType = req.DBType

	// 建表
	if err = repository.Merge(db); err != nil {
		freedom.Logger().Fatal(err.Error())
	}

	// 初始化数据
	if err = repository.Seed(db); err != nil {
		freedom.Logger().Errorf("seed data: %v", err)
	}

	var userId string
	var existingUser po.User
	err = db.Where("super_admin = ?", 1).First(&existingUser).Error
	if err == nil {
		// 超管已存在，更新除 userId 外的所有字段
		existingUser.Username = req.Username
		existingUser.Password = util.MD5(req.Password) // 前端传明文，后端做 MD5
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
			Password:   util.MD5(req.Password), // 前端传明文，后端做 MD5
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

	// 标记初始化完成
	cfg.ModifyConfig("system", "initialized", "true")

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

// restartSelf 通过 syscall.Exec 重启当前进程
func (service *InstallService) restartSelf() {
	self, _ := os.Executable()
	syscall.Exec(self, os.Args, os.Environ())
}

// buildRedisAddr 拼接 Redis 连接地址 host:port
func (service *InstallService) buildRedisAddr(addr string, port int) string {
	host, p := service.splitHostPort(addr, port)
	return fmt.Sprintf("%s:%s", host, p)
}
