package vo

import "errors"

// InstallDBCheckReq 数据库连接检查请求
type InstallDBCheckReq struct {
	DBType string `json:"dbType" validate:"required,oneof=sqlite mysql pg"`
	DBAddr string `json:"dbAddr"`
	DBPort int    `json:"dbPort"`
	DBUser string `json:"dbUser"`
	DBPass string `json:"dbPass"`
	DBName string `json:"dbName"`
}

// Check 条件校验：MySQL/PG 时 DBAddr、DBPort、DBName 必填
func (req *InstallDBCheckReq) Check() error {
	if req.DBType == "mysql" || req.DBType == "pg" {
		if req.DBAddr == "" {
			return errors.New("dbAddr is required")
		}
		if req.DBPort == 0 {
			return errors.New("dbPort is required")
		}
		if req.DBName == "" {
			return errors.New("dbName is required")
		}
	}
	return nil
}

// InstallRedisCheckReq Redis 连接检查请求
type InstallRedisCheckReq struct {
	RedisAddr string `json:"redisAddr" validate:"required"`
	RedisPort int    `json:"redisPort" validate:"required"`
	RedisPass string `json:"redisPass"`
	RedisDB   int    `json:"redisDB"`
}

// InstallInitReq 系统初始化请求
type InstallInitReq struct {
	Language string `json:"language" validate:"required,oneof=zh en"`

	Domain string `json:"domain"` // 选填，系统域名（生成外链等场景使用）

	Username string `json:"username" validate:"required,min=8,max=16"`
	Password string `json:"password" validate:"required,min=6,max=256"`
	Email    string `json:"email"`

	DBType string `json:"dbType" validate:"required,oneof=sqlite mysql pg"`
	DBAddr string `json:"dbAddr"`
	DBPort int    `json:"dbPort"`
	DBUser string `json:"dbUser"`
	DBPass string `json:"dbPass"`
	DBName string `json:"dbName"`

	CacheType string `json:"cacheType" validate:"required,oneof=local redis"`
	RedisAddr string `json:"redisAddr"`
	RedisPort int    `json:"redisPort"`
	RedisPass string `json:"redisPass"`
	RedisDB   int    `json:"redisDB"`
}

// Check 条件校验：MySQL/PG 时 DB 字段必填，Redis 时 Redis 字段必填
func (req *InstallInitReq) Check() error {
	if req.DBType == "mysql" || req.DBType == "pg" {
		if req.DBAddr == "" {
			return errors.New("dbAddr is required")
		}
		if req.DBPort == 0 {
			return errors.New("dbPort is required")
		}
		if req.DBUser == "" {
			return errors.New("dbUser is required")
		}
		if req.DBPass == "" {
			return errors.New("dbPass is required")
		}
		if req.DBName == "" {
			return errors.New("dbName is required")
		}
	}
	if req.CacheType == "redis" {
		if req.RedisAddr == "" {
			return errors.New("redisAddr is required")
		}
		if req.RedisPort == 0 {
			return errors.New("redisPort is required")
		}
		if req.RedisPass == "" {
			return errors.New("redisPass is required")
		}
	}
	return nil
}

// InstallInitRsp 系统初始化响应
type InstallInitRsp struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
}
