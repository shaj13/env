// Copyright 2022 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env_test

import (
	"fmt"
	"net"
	"os"

	"github.com/shaj13/env"
)

func ExampleTextVar() {
	fs := env.NewEnvSet("Example", env.ContinueOnError)
	fs.SetOutput(os.Stdout)
	var ip net.IP
	fs.TextVar(&ip, "IP", net.IPv4(192, 168, 0, 100), "`net.IP` to parse")
	// fs.Parse([]string{"EXAMPLE_IP=127.0.0.1"})
	// fmt.Printf("{ip: %v}\n\n", ip)

	// 256 is not a valid IPv4 component
	ip = nil
	fs.Parse([]string{"EXAMPLE_IP=256.0.0.1"})
	fmt.Printf("{ip: %v}\n\n", ip)

	// Output:
	// invalid value "256.0.0.1" for env IP: invalid IP address: 256.0.0.1
	// Usage of Example:
	//       EXAMPLE_IP net.IP   net.IP to parse (default 192.168.0.100)
	// {ip: <nil>}
}
