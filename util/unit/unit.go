package unit_test

import (
	"context"
	"encoding/json"
	"raven/backend/infra"
	"raven/backend/repository"
	"raven/config"
	"time"

	"github.com/8treenet/freedom"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// set RAVEN_CONF from config.go -> ParseConfigPath
func GetUnitTest() freedom.UnitTest {
	//Create unit testing tools
	unitTest := freedom.NewUnitTest()
	unitTest.InstallDB(func() interface{} {
		dbConf := config.Get().DB
		dbConfAddr := dbConf.Addr
		dbType := config.Get().System.DBType

		var dialector gorm.Dialector
		switch dbType {
		case "mysql":
			dialector = mysql.Open(dbConfAddr)
		case "pg":
			dialector = postgres.Open(dbConfAddr)
		default: // sqlite
			dbConfAddr = config.Get().Paths.SqliteDB
			dialector = sqlite.Open(dbConfAddr)
		}

		db, err := gorm.Open(dialector, &gorm.Config{})
		if err != nil {
			freedom.Logger().Fatal(err.Error())
		}

		sqlDB, err := db.DB()
		if err != nil {
			freedom.Logger().Fatal(err.Error())
		}

		sqlDB.SetMaxOpenConns(dbConf.MaxOpenConns)
		sqlDB.SetMaxIdleConns(dbConf.MaxIdleConns)
		if dbConf.ConnMaxLifeTime > 0 {
			sqlDB.SetConnMaxLifetime(time.Duration(dbConf.ConnMaxLifeTime) * time.Second)
		}
		if dbConf.ConnMaxIdleTime > 0 {
			sqlDB.SetConnMaxIdleTime(time.Duration(dbConf.ConnMaxIdleTime) * time.Second)
		}

		if err = repository.Merge(db); err != nil {
			freedom.Logger().Fatal(err.Error())
		}

		if err = repository.Seed(db); err != nil {
			freedom.Logger().Errorf("seed data: %v", err)
		}
		return db
	})

	unitTest.InstallRedis(func() (client redis.Cmdable) {
		if config.Get().System.CacheType == "redis" {
			cfg := config.Get().Redis
			opt := &redis.Options{
				Addr:            cfg.Addr,
				Password:        cfg.Password,
				DB:              cfg.DB,
				MaxRetries:      cfg.MaxRetries,
				PoolSize:        cfg.PoolSize,
				ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
				WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
				ConnMaxIdleTime: time.Duration(cfg.IdleTimeout) * time.Second,
				ConnMaxLifetime: time.Duration(cfg.MaxConnAge) * time.Second,
				PoolTimeout:     time.Duration(cfg.PoolTimeout) * time.Second,
			}
			redisClient := redis.NewClient(opt)
			if e := redisClient.Ping(context.Background()).Err(); e != nil {
				freedom.Logger().Fatal(e.Error())
			}
			return redisClient
		}
		return infra.NewCacheWrapper(5*time.Minute, 10*time.Minute)
	})
	return unitTest
}

func JsonLog(obj interface{}) string {
	jdata, _ := json.Marshal(obj)
	return string(jdata)
}
