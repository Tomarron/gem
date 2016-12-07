// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"encoding/base64"

	"github.com/go-gem/gem"
	"github.com/valyala/fasthttp"
)

// BasicAuth default configuration
var (
	BasicAuthOnValid = func(ctx *gem.Context, _ string) {}
)

// BasicAuth Basic Auth middleware
type BasicAuth struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper

	// Validator is a function to validate BasicAuth credentials.
	// Required.
	Validator func(username, password string) bool

	// OnValid will be invoked on when the Validator return true.
	//
	// It is easy to share the username with the other middlewares
	// by using ctx.SetUserValue.
	//
	// Optional.
	OnValid func(ctx *gem.Context, username string)
}

// NewBasicAuth returns BasicAuth instance by the
// given validator function.
func NewBasicAuth(validator func(username, password string) bool) *BasicAuth {
	return &BasicAuth{
		Skipper:   defaultSkipper,
		Validator: validator,
		OnValid:   BasicAuthOnValid,
	}
}

// Handle implements Middleware's Handle function.
func (m *BasicAuth) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}
	basicLen := len(gem.HeaderBasic)

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		auth := gem.Bytes2String(ctx.RequestCtx.Request.Header.Peek(gem.HeaderAuthorization))

		if auth == "" {
			ctx.HTML(fasthttp.StatusBadRequest, fasthttp.StatusMessage(fasthttp.StatusBadRequest))
			return
		}

		if len(auth) > basicLen+1 && auth[:basicLen] == gem.HeaderBasic {
			b, err := base64.StdEncoding.DecodeString(auth[basicLen+1:])
			if err != nil {
				ctx.Logger().Debugf("basic auth middleware: %s", err)
				ctx.HTML(fasthttp.StatusUnauthorized, fasthttp.StatusMessage(fasthttp.StatusUnauthorized))
				return
			}
			cred := string(b)
			for i := 0; i < len(cred); i++ {
				if cred[i] == ':' {
					// Verify credentials
					username := cred[:i]
					psw := cred[i+1:]
					if m.Validator(username, psw) {
						m.OnValid(ctx, username)
						next.Handle(ctx)
						return
					}
				}
			}
		}

		ctx.HTML(fasthttp.StatusUnauthorized, fasthttp.StatusMessage(fasthttp.StatusUnauthorized))
	})
}
