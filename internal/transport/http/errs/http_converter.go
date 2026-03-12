package errs

import (
	"errors"

	"github.com/luckysxx/user-platform/internal/dberr"
	pkgerrs "github.com/luckysxx/user-platform/pkg/errs"
)

// ConvertToCustomError converts domain/storage errors to HTTP-facing business errors.
func ConvertToCustomError(err error) *pkgerrs.CustomError {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, dberr.ErrUsernameDuplicate):
		return pkgerrs.NewParamErr("用户名已存在", err)
	case errors.Is(err, dberr.ErrEmailDuplicate):
		return pkgerrs.NewParamErr("邮箱已存在", err)
	case errors.Is(err, dberr.ErrShortLinkDuplicate):
		return pkgerrs.NewParamErr("短链接已存在", err)
	case errors.Is(err, dberr.ErrDuplicateKey):
		return pkgerrs.NewParamErr("记录已存在", err)
	case errors.Is(err, dberr.ErrNoRows):
		return pkgerrs.New(pkgerrs.NotFound, "记录不存在", err)
	case errors.Is(err, dberr.ErrForeignKey):
		return pkgerrs.NewParamErr("关联记录不存在", err)
	case errors.Is(err, dberr.ErrCheckViolation):
		return pkgerrs.NewParamErr("数据校验失败", err)
	case errors.Is(err, dberr.ErrNotNullViolation):
		return pkgerrs.NewParamErr("必填字段缺失", err)
	case errors.Is(err, dberr.ErrInvalidData):
		return pkgerrs.NewParamErr("数据格式错误", err)
	case errors.Is(err, dberr.ErrConnection):
		return pkgerrs.New(pkgerrs.ServerErr, "数据库连接失败，请稍后重试", err)
	}

	return pkgerrs.NewServerErr(err)
}
