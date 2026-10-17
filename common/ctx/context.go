package ctx

import (
	"context"
	"net/http"

	"github.com/google/uuid"
	"github.com/nvhai248/ms-shared-package/common/apperror"
	"google.golang.org/grpc/metadata"
)

type contextKey string

const (
	RequestIDKey contextKey = "x-request-id"
	UserIDKey    contextKey = "x-user-id"
	RoleKey      contextKey = "x-user-role"
	UsernameKey  contextKey = "x-user-name"
	LocaleKey    contextKey = "x-locale"
)

func (k contextKey) String() string {
	return string(k)
}

func ContextWithValues(ctx context.Context, reqID, userID, username, role, locale string) context.Context {
	if reqID == "" {
		reqID = uuid.New().String()
	}
	ctx = context.WithValue(ctx, RequestIDKey, reqID)
	if userID != "" {
		ctx = context.WithValue(ctx, UserIDKey, userID)
	}
	if username != "" {
		ctx = context.WithValue(ctx, UsernameKey, username)
	}
	if role != "" {
		ctx = context.WithValue(ctx, RoleKey, role)
	}
	if locale != "" {
		ctx = context.WithValue(ctx, LocaleKey, locale)
	}
	return ctx
}

func RequestID(ctx context.Context) string {
	if val, ok := ctx.Value(RequestIDKey).(string); ok {
		return val
	}
	return ""
}

func UserID(ctx context.Context) string {
	if val, ok := ctx.Value(UserIDKey).(string); ok {
		return val
	}
	return ""
}

func Username(ctx context.Context) string {
	if val, ok := ctx.Value(UsernameKey).(string); ok {
		return val
	}
	return ""
}

func Role(ctx context.Context) string {
	if val, ok := ctx.Value(RoleKey).(string); ok {
		return val
	}
	return ""
}

func GetLocale(ctx context.Context) apperror.Locale {
	if val, ok := ctx.Value(LocaleKey).(string); ok && val != "" {
		return apperror.Locale(val)
	}
	return apperror.DefaultLocale
}

// ToGRPCMetadata creates an outgoing gRPC context with metadata headers attached
func ToOutgoingGRPCContext(ctx context.Context) context.Context {
	md := metadata.Pairs()
	if reqID := RequestID(ctx); reqID != "" {
		md.Set(RequestIDKey.String(), reqID)
	}
	if uID := UserID(ctx); uID != "" {
		md.Set(UserIDKey.String(), uID)
	}
	if uName := Username(ctx); uName != "" {
		md.Set(UsernameKey.String(), uName)
	}
	if role := Role(ctx); role != "" {
		md.Set(RoleKey.String(), role)
	}
	if loc := GetLocale(ctx); loc != "" {
		md.Set(LocaleKey.String(), string(loc))
	}
	return metadata.NewOutgoingContext(ctx, md)
}

// FromIncomingGRPCContext extracts request context from gRPC metadata into standard context
func FromIncomingGRPCContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return context.WithValue(ctx, RequestIDKey, uuid.New().String())
	}

	getMD := func(k contextKey) string {
		vals := md.Get(k.String())
		if len(vals) > 0 {
			return vals[0]
		}
		return ""
	}

	reqID := getMD(RequestIDKey)
	if reqID == "" {
		reqID = uuid.New().String()
	}

	return ContextWithValues(ctx, reqID, getMD(UserIDKey), getMD(UsernameKey), getMD(RoleKey), getMD(LocaleKey))
}

// HTTPContextMiddleware populates standard context from incoming HTTP headers
func HTTPContextMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.New().String()
		}
		w.Header().Set("X-Request-ID", reqID)

		ctx := ContextWithValues(
			r.Context(),
			reqID,
			r.Header.Get("X-User-ID"),
			r.Header.Get("X-User-Name"),
			r.Header.Get("X-User-Role"),
			r.Header.Get("Accept-Language"),
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
