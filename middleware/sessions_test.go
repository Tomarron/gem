// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package middleware

import (
	"errors"
	"fmt"
	"testing"

	"github.com/go-gem/gem"
	"github.com/go-gem/sessions"
	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

func TestSessions(t *testing.T) {
	var err error

	sessionStore := sessions.NewFilesystemStore("", []byte("test"))

	m := NewSessions()
	m.Skipper = nil

	router := gem.NewRouter()
	router.Use(m)
	router.GET("/", func(ctx *gem.Context) {
		ctx.HTML(200, "OK")
	})
	router.GET("/get", func(ctx *gem.Context) {
		session, _ := ctx.SessionsStore().Get(ctx.RequestCtx, "_user")
		ctx.HTML(200, fmt.Sprintf("%s", session.Values["name"]))
	})
	router.GET("/set", func(ctx *gem.Context) {
		session, _ := ctx.SessionsStore().Get(ctx.RequestCtx, "_user")
		session.Values["name"] = "foo"
	})

	srv := gem.New("", router.Handler())
	srv.SetSessionsStore(sessionStore)

	if m.Skipper == nil {
		t.Error(errSkipperNil)
	}

	cookieStr := ""

	test1 := tests.New(srv, "/set")
	test1.Expect().Status(200).Custom(func(resp fasthttp.Response) error {
		cookie := &fasthttp.Cookie{}
		cookie.SetKey("_user")
		if !resp.Header.Cookie(cookie) {
			return errors.New("failed to save sessions")
		}

		cookieStr = cookie.String()

		return nil
	})
	if err = test1.Run(); err != nil {
		t.Fatal(err)
	}

	test2 := tests.New(srv, "/get")
	test2.Headers["Cookie"] = cookieStr
	test2.Expect().Status(200).Body("foo")
	if err = test2.Run(); err != nil {
		t.Error(err)
	}

	// skip
	m.Skipper = alwaysSkipper
	test3 := tests.New(srv)
	test3.Expect().Status(200).Body("OK")
	if err = test3.Run(); err != nil {
		t.Error(err)
	}
}
