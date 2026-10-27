package controller

import (
	"net/http"

	"shadmin/domain"
	"shadmin/internal/conf"

	"github.com/gin-gonic/gin"
)

// ClientAuthController 服务合記 Android 客户端：注册、手机号登录、MQTT broker 下发。
// 这些端点是公开的，响应体沿用客户端期望的 domain.Response 结构，
// 但登录成功时 data 直接是 JWT 字符串（客户端从 payload 读取 id/name）。
type ClientAuthController struct {
	ClientAuthUseCase domain.ClientAuthUseCase
	Env               *conf.Env
}

// Register 客户端注册
//
//	@Summary		Register a client account
//	@Description	Register with phone number and password; binds the default client role
//	@Tags			Client
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.ClientRegisterRequest	true	"Registration data"
//	@Success		200		{object}	domain.Response					"Registration succeeded"
//	@Failure		400		{object}	domain.Response					"Invalid request"
//	@Failure		409		{object}	domain.Response					"Phone already registered"
//	@Router			/Register [post]
func (cc *ClientAuthController) Register(c *gin.Context) {
	var req domain.ClientRegisterRequest
	if !MustBindJSON(c, &req) {
		return
	}

	if err := cc.ClientAuthUseCase.Register(c.Request.Context(), &req); err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess("注册成功"))
}

// Login 客户端登录
//
//	@Summary		Client login
//	@Description	Log in with phone number and password; returns the access token string in data
//	@Tags			Client
//	@Accept			json
//	@Produce		json
//	@Param			request	body		domain.ClientLoginRequest	true	"Login credentials"
//	@Success		200		{object}	domain.Response{data=string}	"Access token"
//	@Failure		400		{object}	domain.Response				"Invalid request"
//	@Failure		401		{object}	domain.Response				"Invalid credentials"
//	@Failure		423		{object}	domain.Response				"Account temporarily locked"
//	@Router			/Login [post]
func (cc *ClientAuthController) Login(c *gin.Context) {
	var req domain.ClientLoginRequest
	if !MustBindJSON(c, &req) {
		return
	}

	meta := domain.LoginMeta{
		ClientIP:  c.ClientIP(),
		UserAgent: c.Request.Header.Get("User-Agent"),
	}
	token, err := cc.ClientAuthUseCase.Login(c.Request.Context(), req.Tel, req.Password, meta)
	if err != nil {
		RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, domain.RespSuccess(token))
}

// MQTTBroker 下发 MQTT broker 地址
//
//	@Summary		Get MQTT broker info
//	@Description	Return the broker address and ports the client uses for change notifications
//	@Tags			Client
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	domain.Response{data=domain.MQTTBrokerInfo}	"Broker info"
//	@Router			/mqtt/broker [get]
func (cc *ClientAuthController) MQTTBroker(c *gin.Context) {
	c.JSON(http.StatusOK, domain.RespSuccess(domain.MQTTBrokerInfo{
		Address: cc.Env.MQTTAddress,
		TCPPort: cc.Env.MQTTTCPPort,
		WSPort:  cc.Env.MQTTWSPort,
	}))
}
