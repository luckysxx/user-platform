package handler

import (
	commonlogger "github.com/luckysxx/common/logger"
	"github.com/luckysxx/user-platform/internal/service"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	httpdto "github.com/luckysxx/user-platform/internal/transport/http/dto"
	httperrs "github.com/luckysxx/user-platform/internal/transport/http/errs"
	"github.com/luckysxx/user-platform/internal/transport/http/middleware"
	"github.com/luckysxx/user-platform/internal/transport/http/response"
	"github.com/luckysxx/user-platform/internal/transport/http/validator"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	avc    service.AuthService
	svc    service.UserService
	logger *zap.Logger
}

func NewUserHandler(svc service.UserService, avc service.AuthService, logger *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, avc: avc, logger: logger}
}

// @Summary      用户注册
// @Description  用户注册接口
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body dto.RegisterRequest true "注册信息"
// @Success      200  {object}  dto.RegisterResponse
// @Router       /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req httpdto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用 validator 翻译验证错误为友好提示
		errMsg := validator.TranslateValidationError(err)
		commonlogger.Ctx(c.Request.Context(), h.logger).Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	user, err := h.svc.Register(c.Request.Context(), &servicecontract.RegisterCommand{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		commonlogger.Ctx(c.Request.Context(), h.logger).Error("用户注册失败", zap.Error(err))
		response.Error(c, httperrs.ConvertToCustomError(err))
		return
	}
	response.Success(c, &httpdto.RegisterResponse{
		Email:    user.Email,
		UserID:   user.UserID,
		Username: user.Username,
	})
}

// @Summary      用户登录
// @Description  用户登录接口
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body dto.LoginRequest true "登录信息"
// @Success      200  {object}  dto.LoginResponse
// @Router       /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req httpdto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用 validator 翻译验证错误为友好提示
		errMsg := validator.TranslateValidationError(err)
		commonlogger.Ctx(c.Request.Context(), h.logger).Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	user, err := h.avc.Login(c.Request.Context(), &servicecontract.LoginCommand{
		Username: req.Username,
		Password: req.Password,
		AppCode:  req.AppCode,
	})
	if err != nil {
		commonlogger.Ctx(c.Request.Context(), h.logger).Error("用户登录失败", zap.Error(err))
		// 这里可以直接抛出，因为底层 Service 已经是 Domain Error 了
		response.Error(c, httperrs.ConvertToCustomError(err))
		return
	}
	response.Success(c, &httpdto.LoginResponse{
		AccessToken:  user.AccessToken,
		RefreshToken: user.RefreshToken,
		UserID:       user.UserID,
		Username:     user.Username,
	})
}

func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req httpdto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := validator.TranslateValidationError(err)
		commonlogger.Ctx(c.Request.Context(), h.logger).Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	token, err := h.avc.RefreshToken(c.Request.Context(), &servicecontract.RefreshTokenCommand{
		Token: req.Token,
	})
	if err != nil {
		commonlogger.Ctx(c.Request.Context(), h.logger).Error("刷新 Token 失败", zap.Error(err))
		response.Error(c, httperrs.ConvertToCustomError(err))
		return
	}
	response.Success(c, &httpdto.RefreshTokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}

// @Summary      用户登出
// @Description  登出特定设备
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body dto.LogoutRequest true "登出信息"
// @Success      200
// @Router       /users/logout [post]
func (h *UserHandler) Logout(c *gin.Context) {
	var req httpdto.LogoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errMsg := validator.TranslateValidationError(err)
		commonlogger.Ctx(c.Request.Context(), h.logger).Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	// 经过 Auth 中间件后，安全的获取身份
	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Unauthorized(c, "未授权的访问")
		return
	}

	err := h.avc.Logout(c.Request.Context(), &servicecontract.LogoutCommand{
		UserID:   userID,
		DeviceID: req.DeviceID,
	})
	if err != nil {
		commonlogger.Ctx(c.Request.Context(), h.logger).Error("登出失败", zap.Error(err))
		response.Error(c, err)
		return
	}
	response.Success(c, nil)
}
