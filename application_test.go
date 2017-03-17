// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"strconv"
	"testing"
	"time"
)

func TestNewApplication(t *testing.T) {
	filename := ""
	_, err := NewApplication(filename)
	if err == nil {
		t.Error("expected non-nil error, got nil error")
	}

	root := path.Join(os.TempDir(), "app-"+strconv.Itoa(time.Now().Nanosecond()))
	if err = os.MkdirAll(root, os.ModePerm); err != nil {
		t.Fatalf("failed to create application directory: %v", err)
	}

	filename = path.Join(root, "app.json")
	// write invalid json data
	if err = ioutil.WriteFile(filename, []byte(""), os.ModePerm); err != nil {
		t.Fatalf("failed to create configuration file: %v", err)
	}
	if _, err = NewApplication(filename); err == nil {
		t.Error("expected non-nil error, got nil error")
	}

	// write valid json data
	if err = ioutil.WriteFile(filename, []byte("{}"), os.ModePerm); err != nil {
		t.Fatalf("failed to create configuration file: %v", err)
	}
	if _, err = NewApplication(filename); err != nil {
		t.Errorf("expected nil error, got non-nil error: %v", err)
	}
}

func TestApplication_Init(t *testing.T) {
	var err error
	app := &Application{}

	nilCallback := func() error {
		return nil
	}
	errCallback := func() error {
		return errors.New("errCallback")
	}
	app.SetInitCallback(nilCallback)
	if len(app.initCallbacks) != 1 {
		t.Error("failed to set initialized callback")
	}
	if err = app.Init(); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	app.SetInitCallback(errCallback)
	if err = app.Init(); err == nil || err.Error() != errCallback().Error() {
		t.Errorf("expected error %v, got %v", errCallback(), err)
	}
}

func TestApplication_Close(t *testing.T) {
	var errs []error

	app := &Application{}
	if errs = app.Close(); len(errs) > 0 {
		t.Errorf("expected no error, got %v", errs)
	}

	nilCallback := func() error {
		return nil
	}
	errCallback := func() error {
		return errors.New("errCallback")
	}

	app.SetCloseCallback(nilCallback)
	app.SetCloseCallback(errCallback)

	if len(app.closeCallbacks) != 2 {
		t.Fatal("failed to set close callback")
	}

	errs = app.Close()
	if len(errs) != 1 || errs[0].Error() != errCallback().Error() {
		t.Errorf("expected close error %v, got %v", []error{errCallback()}, errs)
	}
}

func TestApplication_Router(t *testing.T) {
	app := &Application{}
	if app.router != app.Router() {
		t.Error("invalid instance of router")
	}
}

func TestApplication_Templates(t *testing.T) {
	app := &Application{}
	if app.templates != app.Templates() {
		t.Error("invalid instance of templates manager")
	}
}

func TestApplication_Component(t *testing.T) {
	var err error

	app := &Application{}

	db := "db"
	if err = app.SetComponent("db", db); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}

	// get existent component
	if v := app.Component("db"); !reflect.DeepEqual(db, v) {
		t.Errorf("expected component %v, got %v", db, v)
	}

	// get nonexistent component
	if v := app.Component("nonexistent"); v != nil {
		t.Errorf("expected nonexistent component, got %v", v)
	}

	// set another component using the same name.
	expectedErr := fmt.Errorf("the component named %q already exists", "db")
	if err = app.SetComponent("db", "another"); err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr, err)
	}
}

func TestApplication_initAssets(t *testing.T) {
	var err error
	app := &Application{
		router: NewRouter(),
	}
	if err = app.initAssets(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}

	app.AssetsOpt = AssetsOption{
		Dirs: map[string]string{
			"/assets": "/path/to/assets",
		},
	}
	// reset router
	app.router = NewRouter()
	if err = app.initAssets(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}

	app.AssetsOpt.HandlerOption = &HandlerOption{}
	// reset router
	app.router = NewRouter()
	if err = app.initAssets(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}
}

func TestApplication_initTemplates(t *testing.T) {
	opt := TemplatesOption{
		Root:      path.Join(os.TempDir(), strconv.Itoa(time.Now().Nanosecond())),
		Suffix:    ".tmpl",
		LayoutDir: "layout",
	}
	var err error
	app := &Application{
		TemplatesOpt: opt,
	}
	if err = app.initTemplates(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}
	if app.templates.Path != opt.Root || app.templates.LayoutDir != opt.LayoutDir || app.templates.Suffix != opt.Suffix {
		t.Error("failed to set templates option")
	}

	layoutDir := path.Join(opt.Root, opt.LayoutDir)
	layoutName := path.Join(layoutDir, "main"+opt.Suffix)
	app.TemplatesOpt.Layouts = []string{"main"}
	if err = app.initTemplates(); err == nil {
		t.Errorf("expected non-nil error, got %q", err)
	}

	// create layout file
	if err = os.MkdirAll(layoutDir, os.ModePerm); err != nil {
		t.Fatal("failed to create layout directory")
	}
	if err = ioutil.WriteFile(layoutName, []byte(""), os.ModePerm); err != nil {
		t.Fatal("failed to create layout file")
	}
	if err = app.initTemplates(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}

}

type errController struct {
	WebController
}

func (c *errController) Init(app *Application) error {
	return errors.New("error from testController")
}

func TestApplication_SetController(t *testing.T) {
	var err error
	app := &Application{
		router: NewRouter(),
	}
	app.SetController("/", &WebController{})
	if len(app.controllers) != 1 {
		t.Error("failed to set controller")
	}
	if err = app.InitControllers(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}

	c := &errController{}
	app.SetController("/test", c)
	if len(app.controllers) != 2 {
		t.Error("failed to set controller")
	}
	// reset router
	app.router = NewRouter()
	if err = app.InitControllers(); err == nil || err.Error() != c.Init(app).Error() {
		t.Errorf("expected error %q, got %q", c.Init(app), err)
	}
}

type testController struct {
	WebController
	methods        []string
	handlerOptions map[string]*HandlerOption
}

func (c *testController) Methods() []string {
	return c.methods
}

func (c *testController) HandlerOptions() map[string]*HandlerOption {
	return c.handlerOptions
}

func TestApplication_InitControllers(t *testing.T) {
	var err error
	app := &Application{
		router: NewRouter(),
	}

	tc := &testController{
		methods: []string{MethodGet, MethodPost, MethodDelete, MethodPut, MethodOptions, MethodHead, MethodPatch},
		handlerOptions: map[string]*HandlerOption{
			MethodGet: &HandlerOption{},
		},
	}
	app.SetController("/", tc)

	if err = app.InitControllers(); err != nil {
		t.Errorf("expected nil error, got %q", err)
	}

	tc.handlerOptions = nil
	if v := app.getHandlerOption(tc, MethodGet); !reflect.DeepEqual(v, emptyHandlerOption) {
		t.Errorf("expected handler option %v, got %v", emptyHandlerOption, v)
	}

	tc.methods = []string{"invalidMethod"}
	expectedErr := fmt.Errorf("unsupport method %q", "invalidMethod")
	if err = app.InitControllers(); err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr, err)
	}
}
