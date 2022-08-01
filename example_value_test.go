// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env_test

import (
	"fmt"
	"net/url"

	"github.com/shaj13/env"
)

type URLValue struct {
	URL *url.URL
}

func (v URLValue) String() string {
	if v.URL != nil {
		return v.URL.String()
	}
	return ""
}

func (v URLValue) Set(s string) error {
	if u, err := url.Parse(s); err != nil {
		return err
	} else {
		*v.URL = *u
	}
	return nil
}

var u = &url.URL{}

func ExampleValue() {
	fs := env.NewEnvSet("Example", env.ExitOnError)
	fs.Var(&URLValue{u}, "URL", "URL to parse")

	fs.Parse([]string{"EXAMPLE_URL=https://pkg.go.dev/github.com/shaj13/env"})
	fmt.Printf(`{scheme: %q, host: %q, path: %q}`, u.Scheme, u.Host, u.Path)

	// Output:
	// {scheme: "https", host: "pkg.go.dev", path: "/github.com/shaj13/env"}
}
