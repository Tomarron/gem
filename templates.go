// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"errors"
	"fmt"
	"html/template"
	"path"
	"path/filepath"
)

// NewTemplates returns a Templates instance with the given path
// and default options.
func NewTemplates(path string) *Templates {
	return &Templates{
		Path:      path,
		Suffix:    ".html",
		Delims:    []string{"{{", "}}"},
		LayoutDir: "layouts",
		layouts:   make(map[string]*template.Template),
	}
}

// Templates is a templates manager.
type Templates struct {
	Path      string
	Suffix    string
	Delims    []string
	FuncMap   template.FuncMap
	LayoutDir string
	layouts   map[string]*template.Template
}

var errNoTemplateSpecified = errors.New("no template file specified")

// SetLayout set layout.
func (ts *Templates) SetLayout(filenames ...string) error {
	for i, filename := range filenames {
		filenames[i] = path.Join(ts.Path, ts.LayoutDir, filename+ts.Suffix)
	}

	layout, err := ts.New(filenames...)
	if err != nil {
		return err
	}

	name := filepath.Base(filenames[0])
	if _, ok := ts.layouts[name]; ok {
		return fmt.Errorf("the layout named %q already exists", name)
	}

	ts.layouts[name] = layout

	return nil
}

// Layout get layout by the given name.

// if the layout does not exists, returns
// non-nil error.
func (ts *Templates) Layout(name string) (*template.Template, error) {
	if layout, ok := ts.layouts[name+ts.Suffix]; ok {
		return layout, nil
	}

	return nil, fmt.Errorf("no layout named %q", name)
}

// Filenames converts relative paths to absolute paths.
func (ts *Templates) Filenames(filenames ...string) []string {
	for i, filename := range filenames {
		filenames[i] = path.Join(ts.Path, filename+ts.Suffix)
	}

	return filenames
}

// New returns a template.Template instance with
// the Templates's option and template.ParseFiles.
//
// Note: filenames should be absolute paths,
// either uses Templates.Filenames or specifies manually.
func (ts *Templates) New(filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		return nil, errNoTemplateSpecified
	}

	name := filepath.Base(filenames[0])
	return template.New(name).
		Delims(ts.Delims[0], ts.Delims[1]).
		Funcs(ts.FuncMap).
		ParseFiles(filenames...)
}

// Render uses layout to render template.
func (ts *Templates) Render(layoutName string, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		return nil, errNoTemplateSpecified
	}

	layout, err := ts.Layout(layoutName)
	if err != nil {
		return nil, err
	}

	if layout, err = layout.Clone(); err != nil {
		return nil, err
	}

	return layout.ParseFiles(ts.Filenames(filenames...)...)
}
