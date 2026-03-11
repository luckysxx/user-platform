package errs

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ToGRPCError(err error) error {
	if err == nil {
		return nil
	}

	customErr := ConvertToCustomError(err)
	if customErr == nil {
		return status.Error(codes.Internal, "internal error")
	}

	return status.Error(mapCodeToGRPC(customErr.Code), customErr.Msg)
}

func mapCodeToGRPC(code int) codes.Code {
	switch code {
	case ParamErr:
		return codes.InvalidArgument
	case Unauthorized:
		return codes.Unauthenticated
	case Forbidden:
		return codes.PermissionDenied
	case NotFound:
		return codes.NotFound
	case ServerErr:
		return codes.Internal
	default:
		return codes.Internal
	}
}
