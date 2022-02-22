package http

import (
	"net/http"

	"github.com/budhip/common/auth"
	"github.com/gorilla/handlers"
)

type Option func(http.Handler) http.Handler

func NewHandler(handler http.Handler, options ...Option) http.Handler {
	h := handler
	for _, option := range options {
		h = option(h)
	}

	return h
}

func Auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handler.ServeHTTP(w, auth.WithUserInfoRequestContext(r))
	})
}

func Recover(handler http.Handler) http.Handler {
	return handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(handler)
}

func WithDefault() Option {
	return func(h http.Handler) http.Handler {
		return handlers.CompressHandler(Auth(Recover(h)))
	}
}

func DefaultHandler(handler http.Handler) http.Handler {
	return NewHandler(handler, WithDefault())
}