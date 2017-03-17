// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestWebController(t *testing.T) {
	var err error
	var app *Application
	wc := &WebController{}

	if err = wc.Init(app); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}

	methods := wc.Methods()
	if !reflect.DeepEqual(methods, webControllerMethods) {
		t.Errorf("expected methods %v, got %v", webControllerMethods, methods)
	}

	handlerOptions := wc.HandlerOptions()
	if !reflect.DeepEqual(handlerOptions, webControllerHandlerOptions) {
		t.Errorf("expected handlerOptions %v, got %v", webControllerHandlerOptions, handlerOptions)
	}

	handlerFuncs := map[string]HandlerFunc{
		MethodGet:     wc.GET,
		MethodPost:    wc.POST,
		MethodDelete:  wc.DELETE,
		MethodPut:     wc.PUT,
		MethodHead:    wc.HEAD,
		MethodOptions: wc.OPTIONS,
		MethodPatch:   wc.PATCH,
	}

	for method, f := range handlerFuncs {
		resp := &httptest.ResponseRecorder{}
		ctx := &Context{
			Response: resp,
		}
		f(ctx)
		if resp.Body != nil {
			t.Error("expected %s response body is nil, got non nil", method)
		}
	}
}
