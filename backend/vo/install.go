package vo

import "errors"

type InstallDBCheckReq struct {
	DBType string `json:"dbType" validate:"required,oneof=sqlite mysql pg"`
	DBAddr string `json:"dbAddr"`
	DBPort int    `json:"dbPort"`
	DBUser string `json:"dbUser"`
	DBPass string `json:"dbPass"`
	DBName string `json:"dbName"`
}

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

type InstallRedisCheckReq struct {
	RedisAddr string `json:"redisAddr" validate:"required"`
	RedisPort int    `json:"redisPort" validate:"required"`
	RedisPass string `json:"redisPass"`
	RedisDB   int    `json:"redisDB"`
}

type InstallInitReq struct {
	Language string `json:"language" validate:"required,oneof=zh en"`

	Domain string `json:"domain"`

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

type InstallInitRsp struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
}
