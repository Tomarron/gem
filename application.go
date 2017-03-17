// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

// ApplicationCallback is type of func that defines
type ApplicationCallback func() error

// NewApplication
func NewApplication(filename string) (*Application, error) {
	var data []byte
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	app := Application{
		ServerOpt: ServerOption{
			Addr: ":8080",
		},
		AssetsOpt: AssetsOption{
			Root:          path.Join(filepath.Dir(filename), "assets"),
			HandlerOption: emptyHandlerOption,
		},
		TemplatesOpt: TemplatesOption{
			Root:      path.Join(filepath.Dir(filename), "templates"),
			Suffix:    ".html",
			LayoutDir: "layouts",
		},
		router:      NewRouter(),
		components:  make(map[string]interface{}),
		controllers: make(map[string]Controller),
	}

	if err = json.Unmarshal(data, &app); err != nil {
		return nil, err
	}

	app.initCallbacks = []ApplicationCallback{
		app.initAssets,
		app.initTemplates,
	}

	return &app, nil
}

type Application struct {
	ServerOpt ServerOption `json:"server"`
	AssetsOpt AssetsOption `json:"assets"`

	templates    *Templates
	TemplatesOpt TemplatesOption `json:"templates"`

	router *Router

	components map[string]interface{}

	controllers map[string]Controller

	initCallbacks  []ApplicationCallback
	closeCallbacks []ApplicationCallback
}

// SetInitCallback set user-defined initialized callback.
func (app *Application) SetInitCallback(callback ApplicationCallback) {
	app.initCallbacks = append(app.initCallbacks, callback)
}

// Init initialize application, All the initialized callbacks.
func (app *Application) Init() (err error) {
	for _, callback := range app.initCallbacks {
		if err = callback(); err != nil {
			return err
		}
	}

	return
}

func (app *Application) initAssets() error {
	for route, dir := range app.AssetsOpt.Dirs {
		if app.AssetsOpt.HandlerOption != nil {
			app.router.ServeFiles(route+"/*filepath", http.Dir(path.Join(app.AssetsOpt.Root, dir)), app.AssetsOpt.HandlerOption)
			continue
		}

		app.router.ServeFiles(route+"/*filepath", http.Dir(path.Join(app.AssetsOpt.Root, dir)))
	}

	return nil
}

func (app *Application) initTemplates() (err error) {
	app.templates = NewTemplates(app.TemplatesOpt.Root)
	if app.TemplatesOpt.Suffix != "" {
		app.templates.Suffix = app.TemplatesOpt.Suffix
	}
	if app.TemplatesOpt.LayoutDir != "" {
		app.templates.LayoutDir = app.TemplatesOpt.LayoutDir
	}

	for _, layout := range app.TemplatesOpt.Layouts {
		filenames := strings.Split(layout, ",")
		for i, filename := range filenames {
			filenames[i] = strings.TrimSpace(filename)
		}
		if err = app.templates.SetLayout(filenames...); err != nil {
			return err
		}
	}

	return nil
}

// SetCloseCallback set user-defined close callback.
func (app *Application) SetCloseCallback(callback ApplicationCallback) {
	app.closeCallbacks = append(app.closeCallbacks, callback)
}

// Close close application, all the close callbacks will be invoked.
func (app *Application) Close() (errs []error) {
	for _, callback := range app.closeCallbacks {
		if err := callback(); err != nil {
			errs = append(errs, err)
		}
	}

	return
}

// Router returns an instance of router.
func (app *Application) Router() *Router {
	return app.router
}

// Templates returns an instance of templates manager.
func (app *Application) Templates() *Templates {
	return app.templates
}

// Component returns a component via the given name.
func (app *Application) Component(name string) interface{} {
	return app.components[name]
}

// SetComponent set component by the given name and component.
//
// If the component already exists, returns an non-nil error.
func (app *Application) SetComponent(name string, component interface{}) error {
	if app.components == nil {
		app.components = make(map[string]interface{})
	}

	if _, ok := app.components[name]; ok {
		return fmt.Errorf("the component named %q already exists", name)
	}

	app.components[name] = component
	return nil
}

// SetController set controller with the given route's path
// and controller instance.
func (app *Application) SetController(path string, controller Controller) {
	if app.controllers == nil {
		app.controllers = map[string]Controller{
			path: controller,
		}

		return
	}

	app.controllers[path] = controller
}

// InitControllers initialize controllers.
func (app *Application) InitControllers() (err error) {
	for path, c := range app.controllers {
		if err = app.register(path, c); err != nil {
			return err
		}
	}

	return nil
}

func (app *Application) register(path string, controller Controller) error {
	if err := controller.Init(app); err != nil {
		return err
	}

	methods := controller.Methods()

	for _, method := range methods {
		switch method {
		case MethodGet:
			app.router.GET(path, controller.GET, app.getHandlerOption(controller, MethodGet))
		case MethodPost:
			app.router.POST(path, controller.POST, app.getHandlerOption(controller, MethodPost))
		case MethodPut:
			app.router.PUT(path, controller.PUT, app.getHandlerOption(controller, MethodPut))
		case MethodDelete:
			app.router.DELETE(path, controller.DELETE, app.getHandlerOption(controller, MethodDelete))
		case MethodHead:
			app.router.HEAD(path, controller.HEAD, app.getHandlerOption(controller, MethodHead))
		case MethodOptions:
			app.router.OPTIONS(path, controller.OPTIONS, app.getHandlerOption(controller, MethodOptions))
		case MethodPatch:
			app.router.PATCH(path, controller.PATCH, app.getHandlerOption(controller, MethodPatch))
		default:
			return fmt.Errorf("unsupport method %q", method)
		}
	}

	return nil
}

var emptyHandlerOption = &HandlerOption{}

func (app *Application) getHandlerOption(controller Controller, method string) *HandlerOption {
	options := controller.HandlerOptions()
	if options == nil {
		return emptyHandlerOption
	}

	if option, ok := options[method]; ok {
		return option
	}

	return emptyHandlerOption
}
