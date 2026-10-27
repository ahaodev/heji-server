package domain

import (
	"context"
	"errors"
)

var (
	ErrRegisterUserExists  = errors.New("该手机号已被注册")
	ErrClientLoginDisabled = errors.New("客户端登录未启用")
)

// ClientRegisterRequest 是合記 Android 客户端的注册请求（tel + password，无邮箱）。
type ClientRegisterRequest struct {
	Name     string `json:"name,omitempty" binding:"omitempty,max=32"`
	Password string `json:"password" binding:"required,min=6,max=72"`
	Tel      string `json:"tel" binding:"required,max=20"`
}

// ClientLoginRequest 是合記 Android 客户端的登录请求。
type ClientLoginRequest struct {
	Tel      string `json:"tel" binding:"required,max=20"`
	Password string `json:"password" binding:"required"`
}

// MQTTBrokerInfo 是客户端连接通知通道所需的 broker 信息。
type MQTTBrokerInfo struct {
	Address string `json:"address"`
	TCPPort string `json:"tcp_port"`
	WSPort  string `json:"ws_port"`
}

// ClientAuthUseCase 是给合記 Android 客户端使用的账号用例。
// 登录返回裸 access token 字符串（客户端按 JWT payload 取 id/name），
// 与 Web 端的 /auth/login 契约（返回令牌对 + 验证码）刻意分开。
type ClientAuthUseCase interface {
	Register(ctx context.Context, req *ClientRegisterRequest) error
	Login(ctx context.Context, tel, password string, meta LoginMeta) (string, error)
}
