[![PkgGoDev](https://pkg.go.dev/badge/github.com/shaj13/env)](https://pkg.go.dev/github.com/shaj13/env)
[![Go Report Card](https://goreportcard.com/badge/github.com/shaj13/env)](https://goreportcard.com/report/github.com/shaj13/env)
[![Coverage Status](https://coveralls.io/repos/github/shaj13/env/badge.svg?branch=main)](https://coveralls.io/github/shaj13/env?branch=main)
[![CircleCI](https://circleci.com/gh/shaj13/env/tree/main.svg?style=svg)](https://circleci.com/gh/shaj13/env/tree/main)

# Env
> Declare environment variable like declaring flag.

Idiomatic go environment variable declaration and parsing.
 
## Installation 
Using env is easy. First, use go get to install the latest version of the library.

```sh
go get github.com/shaj13/env
```
Next, include env in your application:
```go
import (
    "github.com/shaj13/env"
)
```

## Usage
```go
package main

import (
	"errors"
	"fmt"
	"os"
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

func main() {
	os.Setenv("DELTA_T", "1s,2m,3h")

	// All the interesting pieces are with the variables declared above, but
	// to enable the env package to see the env defined there, one must
	// execute, typically at the start of main (not init!):
	env.Parse()

	fmt.Println("Interval: ", intervalEnv)   // print user defined env value
	fmt.Println("Gopher Type: ", gopherType) // print default env value
	fmt.Println("Species: ", *species)       // print default env value
	
	env.Usage() // print the usage
}
```

# Contributing
1. Fork it
2. Download your fork to your PC (`git clone https://github.com/your_username/env && cd env`)
3. Create your feature branch (`git checkout -b my-new-feature`)
4. Make changes and add them (`git add .`)
5. Commit your changes (`git commit -m 'Add some feature'`)
6. Push to the branch (`git push origin my-new-feature`)
7. Create new pull request

# License
env is released under the BSD 3-Clause license. See [LICENSE](https://github.com/shaj13/env/blob/main/LICENSE)
