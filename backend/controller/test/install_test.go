package test

import (
	"raven/backend/vo"
	"testing"

	"github.com/8treenet/freedom/infra/requests"
)

var domain = "http://localhost:8000"

func TestInstallCheckDB(t *testing.T) {

	req := vo.InstallDBCheckReq{
		DBType:	"mysql",
		DBAddr:	"127.0.0.1",
		DBPort:	3306,
		DBUser:	"root",
		DBPass:	"",
		DBName:	"raven",
	}
	respBody, resp := requests.NewHTTPRequest(domain + "/api/install/check-db").Post().SetJSONBody(req).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("check-db response:", respBody)
}

func TestInstallCheckRedis(t *testing.T) {
	req := vo.InstallRedisCheckReq{
		RedisAddr:	"localhost",
		RedisPort:	6379,
		RedisPass:	"",
		RedisDB:	0,
	}
	respBody, resp := requests.NewHTTPRequest(domain + "/api/install/check-redis").Post().SetJSONBody(req).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("check-redis response:", respBody)
}

func TestInstallInit(t *testing.T) {
	req := vo.InstallInitReq{
		Language:	"zh",
		Username:	"admin",
		Password:	"e10adc3949ba59abbe56e057f20f883e",
		Email:		"",
		DBType:		"mysql",
		CacheType:	"local",
		DBAddr:		"127.0.0.1",
		DBPort:		3306,
		DBUser:		"root",
		DBPass:		"root7944",
		DBName:		"raven",
		RedisAddr:	"localhost",
		RedisPort:	6379,
		RedisPass:	"",
		RedisDB:	0,
	}
	respBody, resp := requests.NewHTTPRequest(domain + "/api/install/init").Post().SetJSONBody(req).ToString()
	if resp.Error != nil {
		t.Log("error:", resp.Error)
		return
	}
	t.Log("init response:", respBody)
}
