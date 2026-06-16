package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"raven/backend/po"
	"raven/backend/repository"
	"raven/backend/vo"
	"raven/backend/vo/errs"
	"raven/config"
	"raven/util"
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

type InstallService struct {
	Worker freedom.Worker
}

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

func (service *InstallService) CheckRedis(req *vo.InstallRedisCheckReq) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:		service.buildRedisAddr(req.RedisAddr, req.RedisPort),
		Password:	req.RedisPass,
		DB:		req.RedisDB,
	})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}
	return nil
}

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

	var redisAddr string
	if req.CacheType == "redis" {
		redisAddr = service.buildRedisAddr(req.RedisAddr, req.RedisPort)
		if err := service.CheckRedis(&vo.InstallRedisCheckReq{
			RedisAddr:	req.RedisAddr,
			RedisPort:	req.RedisPort,
			RedisPass:	req.RedisPass,
			RedisDB:	req.RedisDB,
		}); err != nil {
			return nil, err
		}
	}

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

	if err = repository.Merge(db); err != nil {
		freedom.Logger().Fatal(err.Error())
	}

	if err = repository.Seed(db); err != nil {
		freedom.Logger().Errorf("seed data: %v", err)
	}

	var userId string
	var existingUser po.User
	err = db.Where("super_admin = ?", 1).First(&existingUser).Error
	if err == nil {

		existingUser.Username = req.Username
		existingUser.Password = util.MD5(req.Password)
		existingUser.Email = req.Email
		existingUser.Role = 1
		existingUser.SuperAdmin = 1
		existingUser.Status = 1
		if err := db.Save(&existingUser).Error; err != nil {
			return nil, fmt.Errorf("update super admin failed: %w", err)
		}
		userId = existingUser.UserId
	} else if errors.Is(err, gorm.ErrRecordNotFound) {

		userId = util.UUID()
		user := &po.User{
			UserId:		userId,
			Username:	req.Username,
			Password:	util.MD5(req.Password),
			Email:		req.Email,
			Role:		1,
			SuperAdmin:	1,
			Status:		1,
		}
		if err := db.Create(user).Error; err != nil {
			return nil, fmt.Errorf("create super admin failed: %w", err)
		}
	} else {
		return nil, fmt.Errorf("query super admin failed: %w", err)
	}

	cfg.ModifyConfig("system", "initialized", "true")

	if req.Domain != "" {
		db.Model(&po.SystemSetting{}).Where("config_key = ?", "general.domain").Update("config_value", req.Domain)
	}

	go func() {
		time.Sleep(1 * time.Second)
		service.restartSelf()
	}()

	return &vo.InstallInitRsp{
		UserId:		userId,
		Username:	req.Username,
	}, nil
}

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

func (service *InstallService) splitHostPort(addr string, port int) (host, portStr string) {
	host = addr
	if host == "" {
		host = "localhost"
	}

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

func (service *InstallService) restartSelf() {
	self, _ := os.Executable()
	syscall.Exec(self, os.Args, os.Environ())
}

func (service *InstallService) buildRedisAddr(addr string, port int) string {
	host, p := service.splitHostPort(addr, port)
	return fmt.Sprintf("%s:%s", host, p)
}
