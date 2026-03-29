package errs

import (
	"errors"

	pkgerrs "github.com/luckysxx/common/errs"
	"github.com/luckysxx/user-platform/internal/auth"
	"github.com/luckysxx/user-platform/internal/dberr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError maps domain/storage errors to gRPC status errors.
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	var customErr *pkgerrs.CustomError
	if errors.As(err, &customErr) {
		switch customErr.Code {
		case pkgerrs.ParamErr:
			return status.Error(codes.InvalidArgument, customErr.Msg)
		case pkgerrs.Unauthorized:
			return status.Error(codes.Unauthenticated, customErr.Msg)
		case pkgerrs.Forbidden:
			return status.Error(codes.PermissionDenied, customErr.Msg)
		case pkgerrs.NotFound:
			return status.Error(codes.NotFound, customErr.Msg)
		default:
			return status.Error(codes.Internal, customErr.Msg)
		}
	}

	switch {
	case errors.Is(err, auth.ErrInvalidOrExpiredToken):
		return status.Error(codes.Unauthenticated, "无效的刷新凭证或已过期")
	case errors.Is(err, dberr.ErrUsernameDuplicate),
		errors.Is(err, dberr.ErrEmailDuplicate),
		errors.Is(err, dberr.ErrDuplicateKey),
		errors.Is(err, dberr.ErrForeignKey),
		errors.Is(err, dberr.ErrCheckViolation),
		errors.Is(err, dberr.ErrNotNullViolation),
		errors.Is(err, dberr.ErrInvalidData):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, dberr.ErrNoRows):
		return status.Error(codes.NotFound, "记录不存在")
	case errors.Is(err, dberr.ErrConnection):
		return status.Error(codes.Unavailable, "数据库连接失败，请稍后重试")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
