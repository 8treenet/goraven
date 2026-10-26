package service

import (
	"goraven/backend/infra"
	"goraven/backend/po"
	"goraven/backend/repository"
	"goraven/backend/vo"
	"goraven/backend/vo/errs"
	"goraven/core/sandbox"
	"goraven/util"

	"github.com/8treenet/freedom"
)

func init() {
	freedom.Prepare(func(initiator freedom.Initiator) {
		initiator.BindService(func() *UserService {
			return &UserService{}
		})
		initiator.InjectController(func(ctx freedom.Context) (service *UserService) {
			initiator.FetchService(ctx, &service)
			return
		})
	})
}

// UserService 用户服务
type UserService struct {
	Worker   freedom.Worker
	UserRepo *repository.UserRepository
	Request  *infra.Request
}

func (service *UserService) Login(req *vo.UserLoginReq) (*vo.UserLoginRsp, error) {
	// 防护开关：全局密码错误累计达阈值时强制校验图形验证码
	if service.UserRepo.GetLoginFailureCount() >= repository.CaptchaThreshold {
		answer, ok := service.UserRepo.PopCaptchaAnswer(req.Username)
		if !ok || req.CaptchaAnswer == 0 {
			return nil, errs.ErrCaptchaRequired
		}
		if req.CaptchaAnswer != answer {
			return nil, errs.ErrCaptchaIncorrect
		}
	}

	user, err := service.UserRepo.FindByUsername(req.Username)
	if err != nil {
		service.UserRepo.RecordLoginFailure()
		return nil, errs.ErrInvalidCredentials
	}

	if user.Password != util.MD5(req.Password) {
		service.UserRepo.RecordLoginFailure()
		return nil, errs.ErrInvalidCredentials
	}

	if user.Status != po.UserStatusEnabled {
		return nil, infra.WrapperError(infra.Access_Disable, "account is disabled")
	}

	worker := service.Worker
	token, err := service.UserRepo.GenerateToken(
		user.UserId,
		worker.IrisContext().RemoteAddr(),
		worker.IrisContext().Request().UserAgent(),
	)
	if err != nil {
		return nil, err
	}

	return &vo.UserLoginRsp{
		AccessToken: token,
	}, nil
}

// GetCaptcha 拉取登录算术验证码
func (service *UserService) GetCaptcha(req *vo.CaptchaReq) (*vo.CaptchaRsp, error) {
	failureCount := service.UserRepo.GetLoginFailureCount()
	if failureCount < repository.CaptchaThreshold {
		return &vo.CaptchaRsp{Required: false}, nil
	}

	if failureCount > 120 {
		return &vo.CaptchaRsp{Required: true, Image1: "goraven", Image2: "good"}, nil
	}

	a, b, err := util.RandCaptchaPair()
	if err != nil {
		return nil, err
	}
	img1, err := util.RenderDigitBase64(a)
	if err != nil {
		return nil, err
	}
	img2, err := util.RenderDigitBase64(b)
	if err != nil {
		return nil, err
	}

	service.UserRepo.SaveCaptchaAnswer(req.Username, a+b)

	return &vo.CaptchaRsp{
		Required: true,
		Image1:   img1,
		Image2:   img2,
	}, nil
}

func (service *UserService) GetUserInfo() (*vo.UserInfoRsp, error) {
	userId := service.Request.GetUserId()
	if userId == "" {
		return nil, errs.ErrInvalidCredentials
	}

	user, err := service.UserRepo.FindByUserId(userId)
	if err != nil {
		return nil, errs.ErrInvalidCredentials
	}

	return &vo.UserInfoRsp{
		UserId:   user.UserId,
		Username: user.Username,
		Email:    user.Email,
		Role:     user.Role,
		Status:   user.Status,
		Nickname: user.Nickname,
		Avatar:   user.Avatar,
		Created:  user.Created,
	}, nil
}

// AdminGetUserDetail 管理员获取用户详情
func (service *UserService) AdminGetUserDetail(userId string) (*vo.AdminUserDetailRsp, error) {
	user, err := service.UserRepo.FindByUserId(userId)
	if err != nil {
		return nil, errs.ErrUserNotFound
	}

	sessionCount, lastActiveTime := service.UserRepo.GetUserSessionStats(userId)

	return &vo.AdminUserDetailRsp{
		UserId:          user.UserId,
		Username:        user.Username,
		Nickname:        user.Nickname,
		Email:           user.Email,
		Avatar:          user.Avatar,
		Role:            user.Role,
		Status:          user.Status,
		DailyTokenLimit: user.DailyTokenLimit,
		SessionCount:    sessionCount,
		LastActiveTime:  lastActiveTime,
		Created:         user.Created,
		Updated:         user.Updated,
	}, nil
}

// AdminListUsers 管理员获取用户列表（分页+搜索+角色筛选）
func (service *UserService) AdminListUsers(req *vo.AdminUserListReq) (*infra.PageResponse, error) {
	items, pr, err := service.UserRepo.PaginateUsers(req)
	if err != nil {
		return nil, err
	}
	return &infra.PageResponse{
		List:       items,
		TotalPage:  pr.TotalPage,
		TotalCount: pr.TotalCount,
		Page:       pr.Page,
		PageSize:   pr.PageSize,
	}, nil
}

// AdminBatchGetUsers 管理员根据 userId 列表批量查询用户
func (service *UserService) AdminBatchGetUsers(req *vo.AdminBatchUserReq) ([]vo.AdminUserItem, error) {
	users, err := service.UserRepo.FindByUserIds(req.UserIds)
	if err != nil {
		return nil, err
	}

	items := make([]vo.AdminUserItem, 0, len(users))
	for _, u := range users {
		items = append(items, vo.AdminUserItem{
			UserId:   u.UserId,
			Username: u.Username,
			Nickname: u.Nickname,
			Email:    u.Email,
			Avatar:   u.Avatar,
			Role:     u.Role,
			Status:   u.Status,
			Created:  u.Created,
		})
	}
	return items, nil
}

// AdminCreateUser 管理员创建用户
// 系统允许多位管理员，仅初始化时的超级管理员唯一
func (service *UserService) AdminCreateUser(req *vo.AdminCreateUserReq) error {
	if !util.IsValidUsername(req.Username) {
		return errs.ErrInvalidUsername
	}

	_, err := service.UserRepo.FindByUsername(req.Username)
	if err == nil {
		return errs.ErrUsernameAlreadyExists
	}

	user := &po.User{
		UserId:   util.UUID(),
		Username: req.Username,
		Password: util.MD5(req.Password),
		Nickname: req.Nickname,
		Role:     req.Role,
		Status:   po.UserStatusEnabled,
	}

	return service.UserRepo.CreateUser(user)
}

// AdminUpdateUser 管理员编辑用户（昵称/邮箱/角色/状态）
// 禁用用户时自动清除该用户所有 token 使其立即失效
func (service *UserService) AdminUpdateUser(userId string, req *vo.AdminUpdateUserReq) error {
	user, err := service.UserRepo.FindByUserId(userId)
	if err != nil {
		return errs.ErrUserNotFound
	}
	if user.SuperAdmin == po.UserIsSuperAdmin && req.DailyTokenLimit != nil {
		user.DailyTokenLimit = *req.DailyTokenLimit
		if err := service.UserRepo.UpdateUser(user); err != nil {
			return err
		}
	}
	if user.SuperAdmin == po.UserIsSuperAdmin && req.DailyTokenLimit == nil {
		return errs.ErrCannotEditSuperAdmin
	}

	user.Nickname = req.Nickname
	user.Email = req.Email
	if req.Role != nil {
		user.Role = *req.Role
	}
	if req.Status != nil {
		user.Status = *req.Status
	}
	if req.DailyTokenLimit != nil {
		user.DailyTokenLimit = *req.DailyTokenLimit
	}

	if err := service.UserRepo.UpdateUser(user); err != nil {
		return err
	}

	if req.Role != nil {
		service.UserRepo.Auth.InvalidateUserCache(userId)
	}
	if req.Status != nil && *req.Status == po.UserStatusDisabled {
		service.UserRepo.Auth.DeleteUserTokens(userId, "")
	}

	return nil
}

// AdminResetPassword 管理员重置用户密码
func (service *UserService) AdminResetPassword(userId string, req *vo.AdminResetPasswordReq) error {
	user, err := service.UserRepo.FindByUserId(userId)
	if err != nil {
		return errs.ErrUserNotFound
	}
	if user.SuperAdmin == po.UserIsSuperAdmin {
		return errs.ErrCannotResetSuperAdminPassword
	}

	if err := service.UserRepo.UpdatePassword(userId, util.MD5(req.Password)); err != nil {
		return err
	}
	service.UserRepo.Auth.DeleteUserTokens(userId, "")
	return nil
}

// AdminDeleteUser 管理员删除用户（软删除）
// 禁止删除超级管理员，删除后清除该用户所有 token
func (service *UserService) AdminDeleteUser(userId string) error {
	user, err := service.UserRepo.FindByUserId(userId)
	if err != nil {
		return errs.ErrUserNotFound
	}

	if user.SuperAdmin == po.UserIsSuperAdmin {
		return errs.ErrCannotDeleteSuperAdmin
	}

	if err := service.UserRepo.SoftDeleteUser(userId); err != nil {
		return err
	}

	service.UserRepo.Auth.DeleteUserTokens(userId, "")

	sb, err := sandbox.NewSandbox(user.Username)
	if err == nil {
		_ = sb.Delete()
	}

	return nil
}

// UpdateProfile 用户修改个人资料（昵称/邮箱/头像）
func (service *UserService) UpdateProfile(req *vo.UserProfileReq) error {
	userId := service.Request.GetUserId()
	if userId == "" {
		return errs.ErrInvalidCredentials
	}

	updates := make(map[string]interface{})
	if req.Nickname != nil {
		if *req.Nickname == "" {
			return errs.NewFormatError("nickname cannot be empty", "昵称不能为空")
		}
		updates["nickname"] = *req.Nickname
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Avatar != nil {
		updates["avatar"] = *req.Avatar
	}

	return service.UserRepo.UpdateProfile(userId, updates)
}

// ChangePassword 用户修改密码
// 校验当前密码是否正确，新密码不能与当前密码相同
func (service *UserService) ChangePassword(req *vo.UserPasswordReq) error {
	userId := service.Request.GetUserId()
	if userId == "" {
		return errs.ErrInvalidCredentials
	}

	user, err := service.UserRepo.FindByUserId(userId)
	if err != nil {
		return errs.ErrInvalidCredentials
	}

	if user.Password != util.MD5(req.CurrentPassword) {
		return errs.ErrPasswordIncorrect
	}

	if req.CurrentPassword == req.NewPassword {
		return errs.ErrPasswordSameAsCurrent
	}

	if err := service.UserRepo.UpdatePassword(userId, util.MD5(req.NewPassword)); err != nil {
		return err
	}

	service.UserRepo.Auth.DeleteUserTokens(userId, service.Request.GetToken())
	return nil
}

// Logout 退出登录，失效当前 token
func (service *UserService) Logout() error {
	token := service.Request.GetToken()
	if token == "" {
		return nil
	}
	return service.UserRepo.Auth.DeleteToken(token)
}
