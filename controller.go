// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

type Controller interface {
	Init(app *Application) error
	Methods() []string
	HandlerOptions() map[string]*HandlerOption
	GET(ctx *Context)
	POST(ctx *Context)
	DELETE(ctx *Context)
	PUT(ctx *Context)
	HEAD(ctx *Context)
	OPTIONS(ctx *Context)
	PATCH(ctx *Context)
}

// WebController is an empty controller that implements
// Controller interface.
type WebController struct{}

// Init initialize the controller.
//
// It would be invoked when register a controller.
func (wc *WebController) Init(app *Application) error {
	return nil
}

var webControllerMethods = []string{MethodGet, MethodPost}

// Methods defines which handlers should be registered.
//
// The value should be the name of request method,
// case-sensitive, such as GET, POST.
// By default, GET and POST's handler will be registered.
func (wc *WebController) Methods() []string {
	return webControllerMethods
}

var webControllerHandlerOptions = map[string]*HandlerOption{}

// HandlerOptions defines handler's option.
//
// The key should be the name of request method,
// case-sensitive, such as GET, POST.
func (wc *WebController) HandlerOptions() map[string]*HandlerOption {
	return webControllerHandlerOptions
}

// GET implements Controller's GET method.
func (wc *WebController) GET(ctx *Context) {}

// POST implements Controller's POST method.
func (wc *WebController) POST(ctx *Context) {}

// DELETE implements Controller's DELETE method.
func (wc *WebController) DELETE(ctx *Context) {}

// PUT implements Controller's PUT method.
func (wc *WebController) PUT(ctx *Context) {}

// HEAD implements Controller's HEAD method.
func (wc *WebController) HEAD(ctx *Context) {}

// OPTIONS implements Controller's OPTIONS method.
func (wc *WebController) OPTIONS(ctx *Context) {}

// PATCH implements Controller's PATCH method.
func (wc *WebController) PATCH(ctx *Context) {}
