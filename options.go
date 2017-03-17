// Copyright 2016 The Gem Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license
// that can be found in the LICENSE file.

package gem

type ServerOption struct {
	Addr     string `json:"addr"`
	CertFile string `json:"cert_file"`
	KeyFile  string `json:"key_file"`
}

type AssetsOption struct {
	Root          string            `json:"root"`
	Dirs          map[string]string `json:"dirs"`
	HandlerOption *HandlerOption
}

type TemplatesOption struct {
	Root      string   `json:"root"`
	Suffix    string   `json:"suffix"`
	LayoutDir string   `json:"layout_dir"`
	Layouts   []string `json:"layouts"`
}
