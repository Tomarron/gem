// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"github.com/go-gem/gem"
	"github.com/go-gem/sessions"
)

// Sessions sessions middleware
type Sessions struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
}

// NewSessions returns a sessions middleware instance.
func NewSessions() *Sessions {
	return &Sessions{
		Skipper: defaultSkipper,
	}
}

// Handle implements the Middleware's Handle function.
func (m *Sessions) Handle(next gem.Handler) gem.Handler {
	if m.Skipper == nil {
		m.Skipper = defaultSkipper
	}

	return gem.HandlerFunc(func(ctx *gem.Context) {
		if m.Skipper(ctx) {
			next.Handle(ctx)
			return
		}

		defer func() {
			sessions.Save(ctx.RequestCtx)
			sessions.Clear(ctx.RequestCtx)
		}()

		next.Handle(ctx)
	})
}
