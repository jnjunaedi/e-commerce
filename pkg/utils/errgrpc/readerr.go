package errgrpc

import (
	"e-commerce/pkg/utils/apperror"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func ReadInternalError(err error) error {
	var errorGRPC error

	switch {
	case errors.Is(err, apperror.ErrBadRequest):
		errorGRPC = status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, apperror.ErrConflict):
		errorGRPC = status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, apperror.ErrNotFound):
		errorGRPC = status.Error(codes.NotFound, err.Error())
	case errors.Is(err, apperror.ErrUnauthorized):
		errorGRPC = status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, apperror.ErrForbidden):
		errorGRPC = status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, apperror.ErrInternalServer):
		errorGRPC = status.Error(codes.Internal, err.Error())
	default:
		errorGRPC = status.Error(codes.Internal, "error internal server")
	}

	return errorGRPC
}
