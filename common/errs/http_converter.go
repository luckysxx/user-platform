package errs

import (
	"errors"

	"github.com/luckysxx/user-platform/common/dberr"
)

// ConvertToCustomError 将标准 error 转换为 CustomError
// 这是一个辅助函数，用于 Handler 层统一处理错误转换
func ConvertToCustomError(err error) *CustomError {
	if err == nil {
		return nil
	}

	// 1. 检查数据库特定错误
	switch {
	case errors.Is(err, dberr.ErrUsernameDuplicate):
		return NewParamErr("用户名已存在", err)
	case errors.Is(err, dberr.ErrEmailDuplicate):
		return NewParamErr("邮箱已存在", err)
	case errors.Is(err, dberr.ErrShortLinkDuplicate):
		return NewParamErr("短链接已存在", err)
	case errors.Is(err, dberr.ErrDuplicateKey):
		return NewParamErr("记录已存在", err)
	case errors.Is(err, dberr.ErrNoRows):
		return New(NotFound, "记录不存在", err)
	case errors.Is(err, dberr.ErrForeignKey):
		return NewParamErr("关联记录不存在", err)
	case errors.Is(err, dberr.ErrCheckViolation):
		return NewParamErr("数据校验失败", err)
	case errors.Is(err, dberr.ErrNotNullViolation):
		return NewParamErr("必填字段缺失", err)
	case errors.Is(err, dberr.ErrInvalidData):
		return NewParamErr("数据格式错误", err)
	case errors.Is(err, dberr.ErrConnection):
		return New(ServerErr, "数据库连接失败，请稍后重试", err)
	}

	// 2. 其他未知错误统一作为系统错误
	return NewServerErr(err)
}
