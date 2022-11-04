// Copyright 2018 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env_test

import (
	"fmt"

	"github.com/shaj13/env"
)

type Config struct {
	Host string
	Port string
	// ....
}

func Example_struct() {
	cfg := new(Config)

	es := env.NewEnvSet("app", env.ExitOnError)
	es.StringVar(&cfg.Host, "host", "localhost", "App host name")
	es.StringVar(&cfg.Port, "port", "443", "App port")

	es.Parse([]string{"APP_HOST=env.localhost"})
	fmt.Printf(`%s:%s`, cfg.Host, cfg.Port)

	// Output:
	// env.localhost:443
}
