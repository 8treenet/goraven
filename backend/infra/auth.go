package infra

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"goraven/backend/po"
	"goraven/config"
	"goraven/util"

	"github.com/8treenet/freedom"
	"gorm.io/gorm"
)

var authSingleton *Auth = new(Auth)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindInfra(true, authSingleton)
		initiator.InjectController(func(ctx freedom.Context) (auth *Auth) {
			initiator.FetchInfra(ctx, &auth)
			return
		})
	})
}

const (
	UserIdStoreKey     = "auth-user_id"
	UserRoleStoreKey   = "auth-userRole"
	UserNameStoreKey   = "auth-user_name"
	authCacheKeyPrefix = "auth:token:"
)

type cachedAuth struct {
	UserId   string `json:"userId"`
	UserName string `json:"userName"`
	Role     uint8  `json:"role"`
}

type Auth struct {
	freedom.Infra
}

func NewAuth(allMethod bool, skipPaths ...string) func(freedom.Context) {
	return func(ctx freedom.Context) {
		skip := !allMethod
		for _, v := range skipPaths {
			if matchSkipPath(ctx.Request().URL.Path, v) {
				skip = allMethod
				break
			}
		}
		if skip {
			ctx.Next()
			return
		}
		if ok := authSingleton.auth(ctx); !ok {
			ctx.JSON(ResBodyObject{Code: Access_Expire, Msg: "Authentication failed or token expired"})
			return
		}
		ctx.Next()
	}
}

// matchSkipPath 判断 URL path 是否在某个路径段边界上命中 skip 前缀 v。
// 仅在 v 作为完整路径段（前后为 "/" 或字符串首尾）时才匹配，避免子串误命中。
// 例如 v="/public" 命中 "/api/hfs/public/abc"，但不命中 "/api/hfs/private/x/public_y.md"。
func matchSkipPath(urlPath, v string) bool {
	if v == "" {
		return false
	}
	// v 可带前导 '/'，去掉后按段匹配，避免前导斜杠干扰边界判断
	needle := strings.TrimPrefix(v, "/")
	if needle == "" {
		return false
	}
	for from := 0; from <= len(urlPath)-len(needle); {
		idx := strings.Index(urlPath[from:], needle)
		if idx < 0 {
			return false
		}
		start := from + idx
		// 命中位置前必须是路径段边界（字符串开头或前一个字符为 '/'）
		boundaryBefore := start == 0 || urlPath[start-1] == '/'
		end := start + len(needle)
		// 命中位置后必须是路径段边界（字符串结尾或下一个字符为 '/'）
		boundaryAfter := end == len(urlPath) || urlPath[end] == '/'
		if boundaryBefore && boundaryAfter {
			return true
		}
		from = start + 1
	}
	return false
}

// InvalidateUserCache 清除用户所有 token 的缓存（不断开会话）
// 用于角色等属性变更后，使缓存立即失效，下次请求从 DB 重新加载
func (a *Auth) InvalidateUserCache(userId string) error {
	var tokens []string
	if err := a.db().Model(&po.UserAuth{}).Where("user_id = ?", userId).Pluck("access_token", &tokens).Error; err != nil {
		return err
	}
	for _, token := range tokens {
		a.deleteFromCache(token)
	}
	return nil
}

// DeleteUserTokens 删除用户所有 token（踢出所有设备）
// skipToken 为需要跳过的 token（如修改密码时保留当前会话），传空字符串则全部删除
func (a *Auth) DeleteUserTokens(userId string, skipToken string) error {
	var tokens []string
	if err := a.db().Model(&po.UserAuth{}).Where("user_id = ?", userId).Pluck("access_token", &tokens).Error; err != nil {
		return err
	}
	for _, token := range tokens {
		if skipToken != "" && token == skipToken {
			continue
		}
		a.deleteFromCache(token)
	}
	if skipToken != "" {
		return a.db().Where("user_id = ? AND access_token != ?", userId, skipToken).Delete(&po.UserAuth{}).Error
	}
	return a.db().Where("user_id = ?", userId).Delete(&po.UserAuth{}).Error
}

// DeleteToken 删除单个 AccessToken（退出登录时失效当前 token）
func (a *Auth) DeleteToken(token string) error {
	a.deleteFromCache(token)
	return a.db().Where("access_token = ?", token).Delete(&po.UserAuth{}).Error
}

func (a *Auth) GenerateAccessToken(userId string, expiresAt time.Time, clientIP, clientUA string) (string, error) {
	const maxTokensPerUser = 5

	var count int64
	if err := a.db().Model(&po.UserAuth{}).Where("user_id = ?", userId).Count(&count).Error; err != nil {
		return "", err
	}

	if count >= maxTokensPerUser {
		var oldestToken po.UserAuth
		if err := a.db().Where("user_id = ?", userId).Order("updated ASC").First(&oldestToken).Error; err == nil {
			a.deleteFromCache(oldestToken.AccessToken)
			a.db().Delete(&oldestToken)
		}
	}

	token := "rvn_" + util.UUID()

	userAuth := &po.UserAuth{
		UserId:      userId,
		AccessToken: token,
		ExpiresAt:   expiresAt,
		ClientIP:    clientIP,
		ClientUA:    clientUA,
	}

	if err := a.db().Create(userAuth).Error; err != nil {
		return "", err
	}

	return token, nil
}

func (a *Auth) auth(ctx freedom.Context) bool {
	worker := freedom.ToWorker(ctx)
	token := a.extractToken(ctx)

	previewUser := config.Get().Behavior.PreviewUser
	if previewUser != "" {
		token = "preview"
	}

	if token == "" {
		return false
	}

	if ca := a.getFromCache(token); ca != nil {
		worker.Store().Set(UserIdStoreKey, ca.UserId)
		worker.Store().Set(UserRoleStoreKey, ca.Role)
		worker.Store().Set(UserNameStoreKey, ca.UserName)
		return true
	}

	if previewUser != "" {
		var user po.User
		if err := a.db().First(&user, "user_id = ?", previewUser).Error; err != nil {
			return false
		}

		ca := &cachedAuth{UserId: user.UserId, Role: 1, UserName: user.Username}
		a.setToCache(token, ca, 30*time.Minute)
		worker.Store().Set(UserIdStoreKey, user.UserId)
		worker.Store().Set(UserRoleStoreKey, uint8(1))
		worker.Store().Set(UserNameStoreKey, ca.UserName)
		return true
	}

	var userAuth po.UserAuth
	if err := a.db().First(&userAuth, "access_token = ?", token).Error; err != nil {
		return false
	}

	if !userAuth.ExpiresAt.IsZero() && userAuth.ExpiresAt.Before(time.Now()) {
		return false
	}

	var user po.User
	if err := a.db().First(&user, "user_id = ?", userAuth.UserId).Error; err != nil {
		return false
	}

	ca := &cachedAuth{UserId: user.UserId, Role: user.Role, UserName: user.Username}
	a.setToCache(token, ca, 30*time.Minute)
	worker.Store().Set(UserIdStoreKey, ca.UserId)
	worker.Store().Set(UserRoleStoreKey, ca.Role)
	worker.Store().Set(UserNameStoreKey, ca.UserName)
	return true
}

func (a *Auth) extractToken(ctx freedom.Context) string {
	token := ctx.GetHeader("X-Access-Token")
	if token != "" {
		return token
	}

	authHeader := ctx.GetHeader("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	token = ctx.URLParamDefault("access_token", "")
	if token != "" {
		return token
	}

	return ""
}

func (a *Auth) getFromCache(token string) *cachedAuth {
	key := authCacheKeyPrefix + token
	data, err := a.Redis().Get(context.Background(), key).Bytes()
	if err != nil {
		return nil
	}
	var ca cachedAuth
	if err := json.Unmarshal(data, &ca); err != nil {
		return nil
	}
	return &ca
}

func (a *Auth) setToCache(token string, ca *cachedAuth, expiration time.Duration) {
	key := authCacheKeyPrefix + token
	data, _ := json.Marshal(ca)
	a.Redis().Set(context.Background(), key, data, expiration)
}

func (a *Auth) deleteFromCache(token string) {
	key := authCacheKeyPrefix + token
	a.Redis().Del(context.Background(), key)
}

func (a *Auth) db() *gorm.DB {
	var db *gorm.DB
	if err := a.FetchOnlyDB(&db); err != nil {
		panic(err)
	}
	return db
}
