// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

var (
	testPath = path.Join(os.TempDir(), "templates")
)

func init() {
	var err error

	if err = os.MkdirAll(testPath, os.ModePerm); err != nil {
		panic(err)
	}

	if err = os.MkdirAll(path.Join(testPath, "layouts"), os.ModePerm); err != nil {
		panic(err)
	}

	var files = []struct {
		name string
		data []byte
		mode os.FileMode
	}{
		{
			path.Join(testPath, "layouts", "main.html"),
			[]byte(`<html><head></head><body>{{block "body" .}}No Body{{end}}<body></html>`),
			os.ModePerm,
		},
		{
			path.Join(testPath, "index.html"),
			[]byte(`{{define "body"}}hello wolrd{{end}}`),
			os.ModePerm,
		},
	}

	for _, file := range files {
		if err = ioutil.WriteFile(file.name, file.data, file.mode); err != nil {
			panic(err)
		}
	}
}

func TestTemplates_SetLayout(t *testing.T) {
	var err error
	ts := NewTemplates(testPath)

	if err = ts.SetLayout(); err == nil || err != errNoTemplateSpecified {
		t.Error("wrong hint for setting empty layout")
	}

	layout := "main"
	if err = ts.SetLayout(layout); err != nil {
		t.Fatal(err)
	}

	if err = ts.SetLayout(layout); err != nil {
		if err.Error() != fmt.Sprintf("the layout named %q already exists", layout+ts.Suffix) {
			t.Error("wrong hint for setting same layout")
		}
	} else {
		t.Error(err)
	}

	if _, err = ts.Layout(layout); err != nil {
		t.Error(err)
	}

	if _, err = ts.Layout("nonexistent"); err == nil || err.Error() != fmt.Sprintf("no layout named %q", "nonexistent") {
		t.Error("wrong hint for getting nonexistent layout")
	}
}

func TestTemplates_Render(t *testing.T) {
	var err error
	ts := NewTemplates(testPath)
	layoutName := "main"
	if err = ts.SetLayout(layoutName); err != nil {
		t.Fatal(err)
	}

	if _, err = ts.Render(layoutName); err == nil || err != errNoTemplateSpecified {
		t.Error("wrong hint for rendering empty template")
	}

	if _, err = ts.Render("nonexistentLayout", "index"); err == nil || err.Error() != fmt.Sprintf("no layout named %q", "nonexistentLayout") {
		t.Error("wrong hint for getting nonexistent layout")
	}

	index, err := ts.Render(layoutName, "index")
	if err != nil {
		t.Fatal(err)
	}

	wr := &bytes.Buffer{}
	if err = index.Execute(wr, nil); err != nil {
		t.Fatal(err)
	}

	html := `<html><head></head><body>hello wolrd<body></html>`
	if html != wr.String() {
		t.Errorf("expected html %q, got %q", html, wr.String())
	}

	// executed the layout and check again
	layout, err := ts.Layout(layoutName)
	if err != nil {
		t.Fatal(err)
	}
	layout.Execute(wr, nil)
	index, err = ts.Render(layoutName, "index")
	if err == nil || err.Error() != fmt.Sprintf("html/template: cannot Clone %q after it has executed", layoutName+ts.Suffix) {
		t.Error("wrong hint for cloning am executed layout")
	}
}

func TestTemplates_New(t *testing.T) {
	var err error
	ts := NewTemplates(testPath)

	if _, err = ts.New(); err == nil || err != errNoTemplateSpecified {
		t.Error("wrong hint for rendering empty template")
	}
}
