package errgrpc

import (
	"e-commerce/pkg/utils"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func WriteGRPCError(w http.ResponseWriter, err error) {
	st, ok := status.FromError(err)

	if !ok {
		utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "terjadi kesalahan internal sistem"})
		return
	}

	var httpStatus int
	switch st.Code() {
	case codes.InvalidArgument:
		httpStatus = http.StatusBadRequest
	case codes.AlreadyExists:
		httpStatus = http.StatusConflict
	case codes.NotFound:
		httpStatus = http.StatusNotFound
	case codes.Unauthenticated:
		httpStatus = http.StatusUnauthorized
	case codes.PermissionDenied:
		httpStatus = http.StatusForbidden
	case codes.Internal:
		httpStatus = http.StatusInternalServerError
	default:
		httpStatus = http.StatusInternalServerError
	}

	w.Header().Set("Content-Type", "application/json")
	utils.WriteJSON(w, httpStatus, map[string]string{"error": st.Message()})
}
