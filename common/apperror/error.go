package apperror

import (
	"errors"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Locale string

const (
	LocaleVi      Locale = "vi"
	LocaleEn      Locale = "en"
	LocaleKo      Locale = "ko"
	DefaultLocale Locale = LocaleEn
)

type ErrorMessage struct {
	En string `json:"en"`
	Vi string `json:"vi,omitempty"`
	Ko string `json:"ko,omitempty"`
}

type AppError struct {
	Code       int          `json:"code"`
	ErrCode    string       `json:"error_code"`
	Message    ErrorMessage `json:"message"`
	Detail     string       `json:"detail,omitempty"`
	GRPCCode   codes.Code   `json:"grpc_code"`
	Underlying error        `json:"-"`
}

func (e *AppError) Error() string {
	if e.Detail != "" {
		return fmt.Sprintf("[%s] %s: %s", e.ErrCode, e.Message.En, e.Detail)
	}
	return fmt.Sprintf("[%s] %s", e.ErrCode, e.Message.En)
}

func (e *AppError) Unwrap() error {
	return e.Underlying
}

func (e *AppError) GetErrorMessage(locale Locale) string {
	switch locale {
	case LocaleVi:
		if e.Message.Vi != "" {
			return e.Message.Vi
		}
	case LocaleKo:
		if e.Message.Ko != "" {
			return e.Message.Ko
		}
	case LocaleEn:
		if e.Message.En != "" {
			return e.Message.En
		}
	}
	if e.Message.En != "" {
		return e.Message.En
	}
	return "An error occurred"
}

func (e *AppError) WithDetail(detail string) *AppError {
	clone := *e
	clone.Detail = detail
	return &clone
}

func (e *AppError) Wrap(err error) *AppError {
	clone := *e
	clone.Underlying = err
	if err != nil && clone.Detail == "" {
		clone.Detail = err.Error()
	}
	return &clone
}

func (e *AppError) ToGRPCStatus() *status.Status {
	c := e.GRPCCode
	if c == codes.OK {
		c = HTTPStatusToGRPCCode(e.Code)
	}
	return status.New(c, e.Error())
}

func HTTPStatusToGRPCCode(httpStatus int) codes.Code {
	switch httpStatus {
	case http.StatusOK:
		return codes.OK
	case http.StatusBadRequest:
		return codes.InvalidArgument
	case http.StatusUnauthorized:
		return codes.Unauthenticated
	case http.StatusForbidden:
		return codes.PermissionDenied
	case http.StatusNotFound:
		return codes.NotFound
	case http.StatusConflict:
		return codes.AlreadyExists
	case http.StatusRequestTimeout:
		return codes.DeadlineExceeded
	case http.StatusTooManyRequests:
		return codes.ResourceExhausted
	case http.StatusNotImplemented:
		return codes.Unimplemented
	case http.StatusServiceUnavailable:
		return codes.Unavailable
	case http.StatusGatewayTimeout:
		return codes.DeadlineExceeded
	default:
		return codes.Internal
	}
}

func GRPCCodeToHTTPStatus(code codes.Code) int {
	switch code {
	case codes.OK:
		return http.StatusOK
	case codes.InvalidArgument:
		return http.StatusBadRequest
	case codes.Unauthenticated:
		return http.StatusUnauthorized
	case codes.PermissionDenied:
		return http.StatusForbidden
	case codes.NotFound:
		return http.StatusNotFound
	case codes.AlreadyExists:
		return http.StatusConflict
	case codes.ResourceExhausted:
		return http.StatusTooManyRequests
	case codes.FailedPrecondition:
		return http.StatusBadRequest
	case codes.Aborted:
		return http.StatusConflict
	case codes.OutOfRange:
		return http.StatusBadRequest
	case codes.Unimplemented:
		return http.StatusNotImplemented
	case codes.Internal:
		return http.StatusInternalServerError
	case codes.Unavailable:
		return http.StatusServiceUnavailable
	case codes.DeadlineExceeded:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}

func FromGRPCError(err error) *AppError {
	if err == nil {
		return nil
	}
	st, ok := status.FromError(err)
	if !ok {
		return ErrInternalServerError.Wrap(err)
	}
	httpCode := GRPCCodeToHTTPStatus(st.Code())
	return &AppError{
		Code:     httpCode,
		ErrCode:  st.Code().String(),
		GRPCCode: st.Code(),
		Message: ErrorMessage{
			En: st.Message(),
			Vi: st.Message(),
			Ko: st.Message(),
		},
		Detail:     st.Message(),
		Underlying: err,
	}
}

var (
	ErrDuplicateValue = &AppError{
		Code:     http.StatusConflict,
		ErrCode:  "DUPLICATE_VALUE",
		GRPCCode: codes.AlreadyExists,
		Message: ErrorMessage{
			Vi: "Giá trị đã tồn tại",
			En: "Duplicate value",
			Ko: "값이 이미 존재합니다",
		},
	}

	ErrInternalServerError = &AppError{
		Code:     http.StatusInternalServerError,
		ErrCode:  "INTERNAL_SERVER_ERROR",
		GRPCCode: codes.Internal,
		Message: ErrorMessage{
			Vi: "Lỗi máy chủ nội bộ",
			En: "Internal server error",
			Ko: "내부 서버 오류",
		},
	}

	ErrInvalidRequest = &AppError{
		Code:     http.StatusBadRequest,
		ErrCode:  "INVALID_REQUEST",
		GRPCCode: codes.InvalidArgument,
		Message: ErrorMessage{
			Vi: "Yêu cầu không hợp lệ",
			En: "Invalid request payload",
			Ko: "잘못된 요청입니다",
		},
	}

	ErrNotFound = &AppError{
		Code:     http.StatusNotFound,
		ErrCode:  "NOT_FOUND",
		GRPCCode: codes.NotFound,
		Message: ErrorMessage{
			Vi: "Không tìm thấy dữ liệu",
			En: "Resource not found",
			Ko: "리소스를 찾을 수 없습니다",
		},
	}

	ErrUnauthorized = &AppError{
		Code:     http.StatusUnauthorized,
		ErrCode:  "UNAUTHORIZED",
		GRPCCode: codes.Unauthenticated,
		Message: ErrorMessage{
			Vi: "Không có quyền truy cập",
			En: "Unauthorized access",
			Ko: "인증되지 않은 접근입니다",
		},
	}

	ErrForbidden = &AppError{
		Code:     http.StatusForbidden,
		ErrCode:  "FORBIDDEN",
		GRPCCode: codes.PermissionDenied,
		Message: ErrorMessage{
			Vi: "Bị từ chối truy cập",
			En: "Access forbidden",
			Ko: "접근이 금지되었습니다",
		},
	}

	ErrLoginInformationInvalid = &AppError{
		Code:     http.StatusUnauthorized,
		ErrCode:  "LOGIN_INVALID",
		GRPCCode: codes.Unauthenticated,
		Message: ErrorMessage{
			Vi: "Thông tin đăng nhập không hợp lệ",
			En: "Invalid username or password",
			Ko: "로그인 정보가 유효하지 않습니다",
		},
	}
)

func Is(err, target error) bool {
	return errors.Is(err, target)
}
