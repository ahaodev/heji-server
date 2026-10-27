package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	"shadmin/domain"
	"shadmin/internal/constants"
)

// clientAuthUsecase 服务合記 Android 客户端。它复用登录安全（锁定计数）与登录日志，
// 但跳过滑块验证码 —— 移动客户端无法完成滑块交互，登录凭据本身即 tel + 密码。
type clientAuthUsecase struct {
	userRepository  domain.UserRepository
	roleRepository  domain.RoleRepository
	loginLogUsecase domain.LoginLogUseCase
	securityManager domain.LoginSecurity
	tokenService    domain.TokenIssuer
	contextTimeout  time.Duration
}

func NewClientAuthUsecase(
	userRepository domain.UserRepository,
	roleRepository domain.RoleRepository,
	loginLogUsecase domain.LoginLogUseCase,
	securityManager domain.LoginSecurity,
	tokenService domain.TokenIssuer,
	timeout time.Duration,
) domain.ClientAuthUseCase {
	return &clientAuthUsecase{
		userRepository:  userRepository,
		roleRepository:  roleRepository,
		loginLogUsecase: loginLogUsecase,
		securityManager: securityManager,
		tokenService:    tokenService,
		contextTimeout:  timeout,
	}
}

// Register 以手机号注册客户端账号，并绑定客户端角色（默认 user）。
func (u *clientAuthUsecase) Register(ctx context.Context, req *domain.ClientRegisterRequest) error {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	// 手机号是客户端唯一登录标识：已存在即视为重复注册
	if _, err := u.userRepository.GetByIdentifier(ctx, req.Tel); err == nil {
		return domain.ErrRegisterUserExists
	} else if !errors.Is(err, domain.ErrUserNotFound) {
		return fmt.Errorf("look up existing phone: %w", err)
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	username := req.Name
	if username == "" {
		username = req.Tel
	}

	roleIDs := []string{}
	if role, err := u.roleRepository.GetByName(ctx, domain.RoleNameClient); err == nil {
		roleIDs = append(roleIDs, role.ID)
	} else if !errors.Is(err, domain.ErrRoleNotFound) {
		return fmt.Errorf("resolve client role: %w", err)
	}

	user := &domain.User{
		Username: username,
		Phone:    req.Tel,
		Password: string(hashed),
		Status:   constants.UserStatusActive,
		Source:   constants.UserSourceLocal,
	}
	if err := u.userRepository.CreateWithRoles(ctx, user, roleIDs); err != nil {
		if errors.Is(err, domain.ErrUserExists) || errors.Is(err, domain.ErrUsernameExists) {
			return domain.ErrRegisterUserExists
		}
		return fmt.Errorf("create client user: %w", err)
	}
	return nil
}

// Login 校验手机号 + 密码并返回 access token。
// 失败计数与锁定复用 domain.LoginSecurity，避免客户端成为撞库入口。
func (u *clientAuthUsecase) Login(ctx context.Context, tel, password string, meta domain.LoginMeta) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, u.contextTimeout)
	defer cancel()

	if u.securityManager.IsLocked(ctx, tel) {
		return "", &domain.AccountLockedError{}
	}

	user, err := u.userRepository.GetByIdentifier(ctx, tel)
	if errors.Is(err, domain.ErrUserNotFound) {
		u.securityManager.RecordFailedAttempt(ctx, tel)
		u.recordLoginLog(ctx, meta, constants.StatusFailed, "用户不存在", "")
		return "", domain.ErrInvalidCredentials
	}
	if err != nil {
		return "", fmt.Errorf("get user by phone: %w", err)
	}

	if user.Status != constants.UserStatusActive {
		u.recordLoginLog(ctx, meta, constants.StatusFailed, "账户未启用或已停用", user.Email)
		return "", domain.ErrAccountInactive
	}
	if user.Password == "" {
		u.recordLoginLog(ctx, meta, constants.StatusFailed, "账户无本地密码", user.Email)
		return "", domain.ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		attempts := u.securityManager.RecordFailedAttempt(ctx, tel)
		u.recordLoginLog(ctx, meta, constants.StatusFailed, "密码错误", user.Email)
		if attempts >= u.securityManager.MaxFailures() {
			return "", &domain.AccountLockedError{}
		}
		return "", fmt.Errorf("%w，还可尝试 %d 次", domain.ErrInvalidCredentials, u.securityManager.MaxFailures()-attempts)
	}

	u.securityManager.RecordSuccessfulLogin(ctx, tel)

	tokens, err := u.tokenService.IssuePair(user)
	if err != nil {
		return "", err
	}
	u.recordLoginLog(ctx, meta, constants.StatusSuccess, "", user.Email)

	return tokens.AccessToken, nil
}

func (u *clientAuthUsecase) recordLoginLog(ctx context.Context, meta domain.LoginMeta, status, failureReason, email string) {
	writeLoginLog(ctx, u.loginLogUsecase, meta, constants.UserSourceLocal, status, failureReason, email)
}
