package response

import (
	"encoding/json"
	"net/http"

	"github.com/nvhai248/ms-shared-package/common/apperror"
)

type APIResponse[T any] struct {
	Success   bool              `json:"success"`
	Message   string            `json:"message"`
	ErrorCode string            `json:"error_code,omitempty"`
	Data      T                 `json:"data,omitempty"`
	Detail    string            `json:"detail,omitempty"`
	Meta      map[string]any    `json:"meta,omitempty"`
}

func OK[T any](w http.ResponseWriter, data T, message string) {
	if message == "" {
		message = "success"
	}
	resp := APIResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
	}
	writeJSON(w, http.StatusOK, resp)
}

func Created[T any](w http.ResponseWriter, data T, message string) {
	if message == "" {
		message = "resource created"
	}
	resp := APIResponse[T]{
		Success: true,
		Message: message,
		Data:    data,
	}
	writeJSON(w, http.StatusCreated, resp)
}

func Error(w http.ResponseWriter, err error, locale apperror.Locale) {
	if locale == "" {
		locale = apperror.DefaultLocale
	}

	var appErr *apperror.AppError
	switch e := err.(type) {
	case *apperror.AppError:
		appErr = e
	default:
		appErr = apperror.FromGRPCError(err)
	}

	statusCode := appErr.Code
	if statusCode < 100 || statusCode > 599 {
		statusCode = http.StatusInternalServerError
	}

	resp := APIResponse[any]{
		Success:   false,
		ErrorCode: appErr.ErrCode,
		Message:   appErr.GetErrorMessage(locale),
		Detail:    appErr.Detail,
	}

	writeJSON(w, statusCode, resp)
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(payload)
}
