package errs

import (
	"errors"

	"github.com/luckysxx/user-platform/internal/dberr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToGRPCError maps domain/storage errors to gRPC status errors.
func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, dberr.ErrUsernameDuplicate),
		errors.Is(err, dberr.ErrEmailDuplicate),
		errors.Is(err, dberr.ErrShortLinkDuplicate),
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
