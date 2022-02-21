package context

import (
	gctx "context"
)

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

var (
	// CtxUserID is context key for user id
	CtxUserID = contextKey("user_id")
	// CtxMobile is context key for mobile number
	CtxMobile = contextKey("phone_number")
	// CtxName is context key for name
	CtxName = contextKey("name")
	// CtxEmail is context key for email
	CtxEmail = contextKey("email")
	// CtxUserInfo is context key for user info
	CtxUserInfo = contextKey("user_info")
)

// GetContextAsString return context value as type string
func GetContextAsString(ctx gctx.Context, ctxKey contextKey) string {
	if val := ctx.Value(ctxKey); val != nil {
		str, ok := val.(string)
		if ok {
			return str
		}
	}

	return ""
}