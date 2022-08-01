// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// These examples demonstrate more intricate uses of the env package.
package env_test

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/shaj13/env"
)

var _ = species

// Example 1: A single string env called "species" with default value "gopher".
var species = env.String("species", "gopher", "the species we are studying")

// Example 2: A single string var env called "gopher_type".
// Must be set up with an init function.
var gopherType string

func init() {
	env.StringVar(&gopherType, "gopher_type", "pocket", "the variety of gopher")
}

// Example 3: A user-defined env type, a slice of durations.
type interval []time.Duration

// String is the method to format the env's value, part of the env.Value interface.
// The String method's output will be used in diagnostics.
func (i *interval) String() string {
	return fmt.Sprint(*i)
}

// Set is the method to set the env value, part of the env.Value interface.
// Set's argument is a string to be parsed to set the env.
// It's a comma-separated list, so we split it.
func (i *interval) Set(value string) error {
	if len(*i) > 0 {
		return errors.New("interval env already set")
	}
	for _, dt := range strings.Split(value, ",") {
		duration, err := time.ParseDuration(dt)
		if err != nil {
			return err
		}
		*i = append(*i, duration)
	}
	return nil
}

// Define a env to accumulate durations. Because it has a special type,
// we need to use the Var function and therefore create the env during
// init.

var intervalEnv interval

func init() {
	// Tie the environ to the intervalEnv variable and
	// set a usage message.
	env.Var(&intervalEnv, "delta_t", "comma-separated list of intervals to use between events")
}

func Example() {
	// All the interesting pieces are with the variables declared above, but
	// to enable the env package to see the env defined there, one must
	// execute, typically at the start of main (not init!):
	//	env.Parse()
	// We don't run it here because this is not a main function and
	// the testing suite has already parsed the envs.
}
