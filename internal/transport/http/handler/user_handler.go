package handler

import (
	"errors"

	"github.com/luckysxx/user-platform/internal/service"
	servicecontract "github.com/luckysxx/user-platform/internal/service/contract"
	httpdto "github.com/luckysxx/user-platform/internal/transport/http/dto"
	httperrs "github.com/luckysxx/user-platform/internal/transport/http/errs"
	"github.com/luckysxx/user-platform/internal/transport/http/response"
	"github.com/luckysxx/user-platform/internal/transport/http/validator"
	pkgerrs "github.com/luckysxx/user-platform/pkg/errs"

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
		h.logger.Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	user, err := h.svc.Register(c.Request.Context(), &servicecontract.RegisterCommand{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		h.logger.Error("用户注册失败", zap.Error(err))
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
		h.logger.Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	user, err := h.avc.Login(c.Request.Context(), &servicecontract.LoginCommand{
		Username: req.Username,
		Password: req.Password,
		AppCode:  req.AppCode,
	})
	if err != nil {
		h.logger.Error("用户登录失败", zap.Error(err))

		// 登录失败需要特殊处理，不暴露具体原因
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(c, pkgerrs.NewParamErr("用户名或密码错误", err))
			return
		}
		if errors.Is(err, service.ErrAppNotFound) {
			response.Error(c, pkgerrs.NewParamErr("应用不存在", err))
			return
		}

		// 其他错误统一转换
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
		h.logger.Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	token, err := h.avc.RefreshToken(c.Request.Context(), &servicecontract.RefreshTokenCommand{
		Token: req.Token,
	})
	if err != nil {
		h.logger.Error("刷新 Token 失败", zap.Error(err))
		response.Error(c, httperrs.ConvertToCustomError(err))
		return
	}
	response.Success(c, &httpdto.RefreshTokenResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	})
}
