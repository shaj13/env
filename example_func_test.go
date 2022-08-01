// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env_test

import (
	"errors"
	"fmt"
	"net"
	"os"

	"github.com/shaj13/env"
)

func ExampleFunc() {
	es := env.NewEnvSet("Example", env.ContinueOnError)
	es.SetOutput(os.Stdout)
	var ip net.IP
	es.Func("IP", "`net.IP` to parse", func(s string) error {
		ip = net.ParseIP(s)
		if ip == nil {
			return errors.New("could not parse IP")
		}
		return nil
	})
	es.Parse([]string{"EXAMPLE_IP=127.0.0.1"})
	fmt.Printf("{ip: %v, loopback: %t}\n\n", ip, ip.IsLoopback())

	// 256 is not a valid IPv4 component
	es.Parse([]string{"EXAMPLE_IP=256.0.0.1"})
	fmt.Printf("{ip: %v, loopback: %t}\n\n", ip, ip.IsLoopback())

	// Output:
	// {ip: 127.0.0.1, loopback: true}
	//
	// invalid value "256.0.0.1" for env IP: could not parse IP
	// Usage of Example:
	//       EXAMPLE_IP net.IP   net.IP to parse
	// {ip: <nil>, loopback: false}
}
