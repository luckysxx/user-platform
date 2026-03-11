package handler

import (
	"errors"
	"github.com/luckysxx/user-platform/common/errs"
	"github.com/luckysxx/user-platform/common/response"
	"github.com/luckysxx/user-platform/common/validator"
	"github.com/luckysxx/user-platform/model"
	"github.com/luckysxx/user-platform/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	svc    service.UserService
	logger *zap.Logger
}

func NewUserHandler(svc service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{svc: svc, logger: logger}
}

// @Summary      用户注册
// @Description  用户注册接口
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body model.RegisterRequest true "注册信息"
// @Success      200  {object}  model.RegisterResponse
// @Router       /users/register [post]
func (h *UserHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用 validator 翻译验证错误为友好提示
		errMsg := validator.TranslateValidationError(err)
		h.logger.Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	user, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("用户注册失败", zap.Error(err))
		response.Error(c, errs.ConvertToCustomError(err))
		return
	}
	response.Success(c, user)
}

// @Summary      用户登录
// @Description  用户登录接口
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request body model.LoginRequest true "登录信息"
// @Success      200  {object}  model.LoginResponse
// @Router       /users/login [post]
func (h *UserHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 使用 validator 翻译验证错误为友好提示
		errMsg := validator.TranslateValidationError(err)
		h.logger.Warn("参数验证失败", zap.Error(err), zap.String("message", errMsg))
		response.BadRequest(c, errMsg)
		return
	}

	user, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("用户登录失败", zap.Error(err))

		// 登录失败需要特殊处理，不暴露具体原因
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(c, errs.NewParamErr("用户名或密码错误", err))
			return
		}

		// 其他错误统一转换
		response.Error(c, errs.ConvertToCustomError(err))
		return
	}
	response.Success(c, user)
}
