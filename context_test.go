// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a MIT license
// that can be found in the LICENSE file.

package gem

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/go-gem/tests"
	"github.com/valyala/fasthttp"
)

type project struct {
	Name string `json:"name" xml:"name"`
}

var (
	proj = project{Name: "foo"}
	ch   = make(chan bool)
)

func TestContext_HTML(t *testing.T) {
	var err error

	router := NewRouter()
	router.GET("/html", func(ctx *Context) {
		ctx.HTML(fasthttp.StatusOK, "foo")
	})

	srv := New("", router.Handler())

	test1 := tests.New(srv, "/html")
	test1.Expect().Status(fasthttp.StatusOK).
		Header(HeaderContentType, HeaderContentTypeHTML).
		Body("foo")
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
}

func TestContext_IsAjax(t *testing.T) {
	var isAjax bool
	router := NewRouter()
	router.GET("/", func(ctx *Context) {
		isAjax = ctx.IsAjax()
	})

	srv := New("", router.Handler())

	test := tests.New(srv)
	test.Headers[HeaderXRequestedWith] = HeaderXMLHttpRequest
	test.Expect().Status(fasthttp.StatusOK)
	if err := test.Run(); err != nil {
		t.Error(err)
	}
	if !isAjax {
		t.Error("expected ctx.IsAjax is true, got false")
	}
}

func TestContext_JSON(t *testing.T) {
	respJson, err := json.Marshal(proj)
	if err != nil {
		t.Fatal(err)
	}

	router := NewRouter()
	router.GET("/json", func(ctx *Context) {
		ctx.JSON(fasthttp.StatusOK, proj)
	})
	router.GET("/json-error", func(ctx *Context) {
		ctx.JSON(fasthttp.StatusOK, ch)
	})

	srv := New("", router.Handler())

	test1 := tests.New(srv, "/json")
	test1.Expect().Status(fasthttp.StatusOK).
		Header(HeaderContentType, HeaderContentTypeJSON).
		Body(string(respJson))
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv, "/json-error")
	test2.Expect().Status(fasthttp.StatusInternalServerError)
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
}

func TestContext_JSONP(t *testing.T) {
	respJson, err := json.Marshal(proj)
	if err != nil {
		t.Fatal(err)
	}

	callback := []byte("callback")
	var respJsonp []byte
	respJsonp = append(respJsonp, callback...)
	respJsonp = append(respJsonp, "("...)
	respJsonp = append(respJsonp, respJson...)
	respJsonp = append(respJsonp, ")"...)

	router := NewRouter()
	router.GET("/jsonp", func(ctx *Context) {
		ctx.JSONP(fasthttp.StatusOK, proj, callback)
	})
	router.GET("/jsonp-error", func(ctx *Context) {
		ctx.JSONP(fasthttp.StatusOK, ch, callback)
	})

	srv := New("", router.Handler())

	test1 := tests.New(srv, "/jsonp")
	test1.Expect().Status(fasthttp.StatusOK).
		Header(HeaderContentType, HeaderContentTypeJSONP).
		Body(string(respJsonp))
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv, "/jsonp-error")
	test2.Expect().Status(fasthttp.StatusInternalServerError)
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
}

func TestContext_Logger(t *testing.T) {
	var equal bool
	router := NewRouter()

	srv := New("", router.Handler())

	router.GET("/", func(ctx *Context) {
		equal = ctx.Logger() == srv.logger
	})

	test := tests.New(srv)
	test.Headers[HeaderXRequestedWith] = HeaderXMLHttpRequest
	test.Expect().Status(fasthttp.StatusOK)
	if err := test.Run(); err != nil {
		t.Error(err)
	}
	if !equal {
		t.Error("expected ctx.Logger() == srv.logger: true, got false")
	}
}

func TestContext_Param(t *testing.T) {
	router := NewRouter()
	srv := New("", router.Handler())

	router.GET("/user/:name", func(ctx *Context) {
		ctx.HTML(fasthttp.StatusOK, ctx.Param("name"))
	})

	router.POST("/user/:name", func(ctx *Context) {
		ctx.SetUserValue("name", 1)
		ctx.HTML(fasthttp.StatusOK, ctx.Param("name"))
	})

	var err error

	test1 := tests.New(srv)
	test1.Url = "/user/foo"
	test1.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		body := string(resp.Body())
		if body != "foo" {
			return fmt.Errorf("expected body %q, got %q", "foo", body)
		}

		return nil
	})
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv)
	test2.Url = "/user/foo"
	test2.Method = MethodPost
	test2.Expect().Status(fasthttp.StatusOK).Custom(func(resp fasthttp.Response) error {
		body := string(resp.Body())
		if body != "" {
			return fmt.Errorf("expected empty body, got %q", body)
		}

		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
}

func TestContext_ParamInt(t *testing.T) {
	router := NewRouter()
	srv := New("", router.Handler())

	var page int

	router.GET("/list/:page", func(ctx *Context) {
		page = ctx.ParamInt("page")
	})

	var err error

	test1 := tests.New(srv)
	test1.Url = "/list/2"
	test1.Expect().Status(fasthttp.StatusOK)
	if err = test1.Run(); err != nil {
		t.Error(err)
	}
	if page != 2 {
		t.Errorf("expected page: %d, got %d", 2, page)
	}

	// empty page
	test2 := tests.New(srv)
	test2.Url = "/list/invalid_number"
	test2.Expect().Status(fasthttp.StatusOK)
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
	if page != 0 {
		t.Errorf("expected page: %d, got %d", 0, page)
	}
}

func TestContext_SessionsStore(t *testing.T) {
	var equal bool
	router := NewRouter()

	srv := New("", router.Handler())

	router.GET("/", func(ctx *Context) {
		equal = ctx.SessionsStore() == srv.sessionsStore
	})

	test := tests.New(srv)
	test.Headers[HeaderXRequestedWith] = HeaderXMLHttpRequest
	test.Expect().Status(fasthttp.StatusOK)
	if err := test.Run(); err != nil {
		t.Error(err)
	}
	if !equal {
		t.Error("expected ctx.SessionsStore() == srv.sessionsStore: true, got false")
	}
}

func TestContext_XML(t *testing.T) {
	xmlBytes, err := xml.Marshal(proj)
	if err != nil {
		t.Fatal(err)
	}

	respXml := append([]byte{}, xml.Header...)
	respXml = append(respXml, xmlBytes...)

	router := NewRouter()
	router.GET("/xml", func(ctx *Context) {
		ctx.XML(fasthttp.StatusOK, proj, xml.Header)
	})
	router.GET("/xml-error", func(ctx *Context) {
		ctx.XML(fasthttp.StatusOK, ch, xml.Header)
	})

	srv := New("", router.Handler())

	// XML
	test1 := tests.New(srv, "/xml")
	test1.Expect().Status(fasthttp.StatusOK).
		Header(HeaderContentType, HeaderContentTypeXML).
		Body(string(respXml))
	if err = test1.Run(); err != nil {
		t.Error(err)
	}

	test2 := tests.New(srv, "/xml-error")
	test2.Expect().Status(fasthttp.StatusInternalServerError).Custom(func(resp fasthttp.Response) error {
		return nil
	})
	if err = test2.Run(); err != nil {
		t.Error(err)
	}
}
