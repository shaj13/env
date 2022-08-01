// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package env implements environment variables parsing.

# Usage

Define env using env.String(), Bool(), Int(), etc.

This declares an integer env, N, stored in the pointer nEnv, with type *int:

	import "github.com/shaj13/env"
	var nEnv = env.Int("n", 1234, "usage message for env n")

If you like, you can bind the env to a variable using the Var() functions.

	var envvar int
	func init() {
		env.IntVar(&envvar, "envname", 1234, "usage message for envname")
	}

Or you can create custom envs that satisfy the Value interface (with
pointer receivers) and couple them to env parsing by

	env.Var(&envVal, "name", "usage message for name")

For such envs, the default value is just the initial value of the variable.

After all envs are defined, call

	env.Parse()

to parse the environment variables into the defined envs.

Envs may then be used directly. If you're using the envs themselves,
they are all pointers; if you bind to variables, they're values.

	fmt.Println("ip has value ", *ip)
	fmt.Println("envvar has value ", envvar)

Integer envs accept 1234, 0664, 0x1234 and may be negative.
Boolean envs may be:

	1, 0, t, f, T, F, true, false, TRUE, FALSE, True, False

Duration envs accept any input valid for time.ParseDuration.

The default set of environment variables (environ) is controlled
by top-level functions. The EnvSet type allows one to define
independent sets of envs. The methods of EnvSet are
analogous to the top-level functions for environ set.
*/
package env

import (
	"encoding"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// errParse is returned by Set if a env's value fails to parse, such as with an invalid integer for Int.
// It then gets wrapped through failf to provide more information.
var errParse = errors.New("parse error")

// errRange is returned by Set if a env's value is out of range.
// It then gets wrapped through failf to provide more information.
var errRange = errors.New("value out of range")

func numError(err error) error {
	ne, ok := err.(*strconv.NumError)
	if !ok {
		return err
	}
	if ne.Err == strconv.ErrSyntax {
		return errParse
	}
	if ne.Err == strconv.ErrRange {
		return errRange
	}
	return err
}

// -- bool Value
type boolValue bool

func newBoolValue(val bool, p *bool) *boolValue {
	*p = val
	return (*boolValue)(p)
}

func (b *boolValue) Set(s string) error {
	v, err := strconv.ParseBool(s)
	if err != nil {
		err = errParse
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) Get() interface{} { return bool(*b) }

func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }

// -- int Value
type intValue int

func newIntValue(val int, p *int) *intValue {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() interface{} { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

// -- int64 Value
type int64Value int64

func newInt64Value(val int64, p *int64) *int64Value {
	*p = val
	return (*int64Value)(p)
}

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = int64Value(v)
	return err
}

func (i *int64Value) Get() interface{} { return int64(*i) }

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }

// -- uint Value
type uintValue uint

func newUintValue(val uint, p *uint) *uintValue {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		err = numError(err)
	}
	*i = uintValue(v)
	return err
}

func (i *uintValue) Get() interface{} { return uint(*i) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- uint64 Value
type uint64Value uint64

func newUint64Value(val uint64, p *uint64) *uint64Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		err = numError(err)
	}
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() interface{} { return uint64(*i) }

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

// -- string Value
type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Get() interface{} { return string(*s) }

func (s *stringValue) String() string { return string(*s) }

// -- float64 Value
type float64Value float64

func newFloat64Value(val float64, p *float64) *float64Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		err = numError(err)
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() interface{} { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// -- time.Duration Value
type durationValue time.Duration

func newDurationValue(val time.Duration, p *time.Duration) *durationValue {
	*p = val
	return (*durationValue)(p)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		err = errParse
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() interface{} { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

// -- encoding.TextUnmarshaler Value
type textValue struct{ p encoding.TextUnmarshaler }

func newTextValue(val encoding.TextMarshaler, p encoding.TextUnmarshaler) textValue {
	ptrVal := reflect.ValueOf(p)
	if ptrVal.Kind() != reflect.Ptr {
		panic("variable value type must be a pointer")
	}
	defVal := reflect.ValueOf(val)
	if defVal.Kind() == reflect.Ptr {
		defVal = defVal.Elem()
	}
	if defVal.Type() != ptrVal.Type().Elem() {
		panic(fmt.Sprintf("default type does not match variable type: %v != %v", defVal.Type(), ptrVal.Type().Elem()))
	}
	ptrVal.Elem().Set(defVal)
	return textValue{p}
}

func (v textValue) Set(s string) error {
	return v.p.UnmarshalText([]byte(s))
}

func (v textValue) Get() interface{} {
	return v.p
}

func (v textValue) String() string {
	if m, ok := v.p.(encoding.TextMarshaler); ok {
		if b, err := m.MarshalText(); err == nil {
			return string(b)
		}
	}
	return ""
}

// -- func Value
type funcValue func(string) error

func (f funcValue) Set(s string) error { return f(s) }

func (f funcValue) String() string { return "" }

// Value is the interface to the dynamic value stored in a env.
// (The default value is represented as a string.)
//
// Set is called once, in environ order, for each env present.
// The env package may call the String method with a zero-valued receiver,
// such as a nil pointer.
type Value interface {
	String() string
	Set(string) error
}

// Getter is an interface that allows the contents of a Value to be retrieved.
// It wraps the Value interface, rather than being part of it, because it
// appeared after Go 1 and its compatibility rules. All Value types provided
// by this package satisfy the Getter interface, except the type used by Func.
type Getter interface {
	Value
	Get() interface{}
}

// ErrorHandling defines how EnvSet.Parse behaves if the parse fails.
type ErrorHandling int

// These constants cause EnvSet.Parse to behave as described if the parse fails.
const (
	ContinueOnError ErrorHandling = iota // Return a descriptive error.
	ExitOnError                          // Call os.Exit(2).
	PanicOnError                         // Call panic with a descriptive error.
)

// A EnvSet represents a set of defined envs. The zero value of a EnvSet
// has no prefix and has ContinueOnError error handling.
//
// Env names must be unique within a EnvSet. An attempt to define a env whose
// name is already in use will cause a panic.
//
// Env names and prefix uppercased automatically i.e (foo => FOO).
type EnvSet struct {
	// Usage is the function called when an error occurs while parsing envs.
	// The field is a function (not a method) that may be changed to point to
	// a custom error handler. What happens after Usage is called depends
	// on the ErrorHandling setting; for the "Environ", this defaults
	// to ExitOnError, which exits the program after calling Usage.
	Usage func()

	prefix        string
	parsed        bool
	actual        map[string]*Env
	formal        map[string]*Env
	envs          []string
	errorHandling ErrorHandling
	output        io.Writer // nil means stderr; use Output() accessor
}

// A Env represents the state of a environment variable.
type Env struct {
	Name     string // name as it appears on environ
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message
}

// sortEnvs returns the envs as a slice in lexicographical sorted order.
func sortEnvs(envs map[string]*Env) []*Env {
	result := make([]*Env, len(envs))
	i := 0
	for _, f := range envs {
		result[i] = f
		i++
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Output returns the destination for usage and error messages. os.Stderr is returned if
// output was not set or was set to nil.
func (e *EnvSet) Output() io.Writer {
	if e.output == nil {
		return os.Stderr
	}
	return e.output
}

// Prefix returns the prefix of the env set.
func (e *EnvSet) Prefix() string {
	return e.prefix
}

// ErrorHandling returns the error handling behavior of the env set.
func (e *EnvSet) ErrorHandling() ErrorHandling {
	return e.errorHandling
}

// SetOutput sets the destination for usage and error messages.
// If output is nil, os.Stderr is used.
func (e *EnvSet) SetOutput(output io.Writer) {
	e.output = output
}

// VisitAll visits the envs in lexicographical order, calling fn for each.
// It visits all envs, even those not set.
func (e *EnvSet) VisitAll(fn func(*Env)) {
	for _, env := range sortEnvs(e.formal) {
		fn(env)
	}
}

// VisitAll visits the "Environ" envs in lexicographical order, calling
// fn for each. It visits all envs, even those not set.
func VisitAll(fn func(*Env)) {
	Environ.VisitAll(fn)
}

// Visit visits the envs in lexicographical order, calling fn for each.
// It visits only those envs that have been set.
func (e *EnvSet) Visit(fn func(*Env)) {
	for _, env := range sortEnvs(e.actual) {
		fn(env)
	}
}

// Visit visits the "Environ" envs in lexicographical order, calling fn
// for each. It visits only those envs that have been set.
func Visit(fn func(*Env)) {
	Environ.Visit(fn)
}

// Lookup returns the Env structure of the named env, returning nil if none exists.
func (e *EnvSet) Lookup(name string) *Env {
	return e.formal[name]
}

// Lookup returns the Env structure of the named "Environ" env,
// returning nil if none exists.
func Lookup(name string) *Env {
	return Environ.formal[name]
}

// Set sets the value of the named env.
func (e *EnvSet) Set(name, value string) error {
	name = strings.ToUpper(name)
	env, ok := e.formal[name]
	if !ok {
		return fmt.Errorf("no such env %v", name)
	}
	err := env.Value.Set(value)
	if err != nil {
		return err
	}
	if e.actual == nil {
		e.actual = make(map[string]*Env)
	}
	e.actual[name] = env
	return nil
}

// Set sets the value of the named "Environ" env.
func Set(name, value string) error {
	return Environ.Set(name, value)
}

// isZeroValue determines whether the string represents the zero
// value for a env.
func isZeroValue(env *Env, value string) (ok bool, err error) {
	// Build a zero value of the env's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	typ := reflect.TypeOf(env.Value)
	var z reflect.Value
	if typ.Kind() == reflect.Ptr {
		z = reflect.New(typ.Elem())
	} else {
		z = reflect.Zero(typ)
	}
	// Catch panics calling the String method, which shouldn't prevent the
	// usage message from being printed, but that we should report to the
	// user so that they know to fix their code.
	defer func() {
		if e := recover(); e != nil {
			if typ.Kind() == reflect.Ptr {
				typ = typ.Elem()
			}
			err = fmt.Errorf("panic calling String method on zero %v for env %s: %v", typ, env.Name, e)
		}
	}()
	return value == z.Interface().(Value).String(), nil
}

// UnquoteUsage extracts a back-quoted name from the usage
// string for a env and returns it and the un-quoted usage.
// Given "a `name` to show" it returns ("name", "a name to show").
// If there are no back quotes, the name is an educated guess of the
// type of the env's value.
func UnquoteUsage(env *Env) (name string, usage string) {
	// Look for a back-quoted name, but avoid the strings package.
	usage = env.Usage
	for i := 0; i < len(usage); i++ {
		if usage[i] == '`' {
			for j := i + 1; j < len(usage); j++ {
				if usage[j] == '`' {
					name = usage[i+1 : j]
					usage = usage[:i] + name + usage[j+1:]
					return name, usage
				}
			}
			break // Only one back quote; use type name.
		}
	}
	// No explicit name, so use type if we can find one.
	name = "value"
	switch env.Value.(type) {
	case *boolValue:
		name = "bool"
	case *durationValue:
		name = "duration"
	case *float64Value:
		name = "float"
	case *intValue, *int64Value:
		name = "int"
	case *stringValue:
		name = "string"
	case *uintValue, *uint64Value:
		name = "uint"
	}

	return name, usage
}

// PrintDefaults prints, to standard error unless configured otherwise, the
// default values of all defined envs in the set. See the
// documentation for the global function PrintDefaults for more information.
func (e *EnvSet) PrintDefaults() {
	var isZeroValueErrs []error
	lines := make([]string, 0, len(e.formal))
	maxlen := 0
	prefix := e.prefix
	if prefix != "" {
		prefix = strings.TrimPrefix(e.prefix, "_")
		prefix = strings.ToUpper(prefix) + "_"
	}

	e.VisitAll(func(env *Env) {
		var b strings.Builder
		fmt.Fprintf(&b, "      %s%s", prefix, env.Name)
		name, usage := UnquoteUsage(env)
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}

		// This special character will be replaced with spacing once the
		// correct alignment is calculated
		b.WriteString("\x00")
		if b.Len() > maxlen {
			maxlen = b.Len()
		}

		b.WriteString(usage)

		// Print the default value only if it differs to the zero value
		// for this env type.
		if isZero, err := isZeroValue(env, env.DefValue); err != nil {
			isZeroValueErrs = append(isZeroValueErrs, err)
		} else if !isZero {
			if _, ok := env.Value.(*stringValue); ok {
				// put quotes on the value
				fmt.Fprintf(&b, " (default %q)", env.DefValue)
			} else {
				fmt.Fprintf(&b, " (default %v)", env.DefValue)
			}
		}

		lines = append(lines, b.String())
	})

	for _, line := range lines {
		sidx := strings.Index(line, "\x00")
		spacing := strings.Repeat(" ", maxlen-sidx)
		// maxlen + 2 comes from + 1 for the \x00 and + 1 for the (deliberate) off-by-one in maxlen-sidx
		indented := strings.Replace(line[sidx+1:], "\n", "\n"+strings.Repeat(" ", maxlen+2), -1)
		fmt.Fprintln(e.Output(), line[:sidx], spacing, indented)
	}

	// If calling String on any zero env.Values triggered a panic, print
	// the messages after the full set of defaults so that the programmer
	// knows to fix the panic.
	if errs := isZeroValueErrs; len(errs) > 0 {
		fmt.Fprintln(e.Output())
		for _, err := range errs {
			fmt.Fprintln(e.Output(), err)
		}
	}
}

// PrintDefaults prints, to standard error unless configured otherwise,
// a usage message showing the default settings of all defined envs.
// For an integer valued env x, the default output has the form
//
//	X int   usage-message-for-x (default 7)
//
// The usage message will appear on a the same line for anything.
// The parenthetical default is omitted if the default is the zero
// value for the type. The listed type, here int,
// can be changed by placing a back-quoted name in the env's usage
// string; the first such item in the message is taken to be a parameter
// name to show in the message and the back quotes are stripped from
// the message when displayed. For instance, given
//
//	env.String("DIR", "", "search `directory` for include files")
//
// the output will be
//
//	DIR directory   search directory for include files.
//
// To change the destination for env messages, call Environ.SetOutput.
func PrintDefaults() {
	Environ.PrintDefaults()
}

// defaultUsage is the default function to print a usage message.
func (e *EnvSet) defaultUsage() {
	if e.prefix == "" {
		fmt.Fprintf(e.Output(), "Usage:\n")
	} else {
		fmt.Fprintf(e.Output(), "Usage of %s:\n", e.prefix)
	}
	e.PrintDefaults()
}

// NOTE: Usage is not just defaultUsage(Environ)
// because it serves (via godoc env Usage) as the example
// for how to write your own usage function.

// Usage prints a usage message documenting all defined "Environ" envs
// to Environ's output, which by default is os.Stderr.
// It is called when an error occurs while parsing envs.
// The function is a variable that may be changed to point to a custom function.
// By default it prints a simple header and calls PrintDefaults; for details about the
// format of the output and how to control it, see the documentation for PrintDefaults.
// Custom usage functions may choose to exit the program; by default exiting
// happens anyway as the Environ's error handling strategy is set to
// ExitOnError.
var Usage = func() {
	fmt.Fprintf(Environ.Output(), "Usage of %s:\n", os.Args[0])
	PrintDefaults()
}

// NEnv returns the number of envs that have been set.
func (e *EnvSet) NEnv() int { return len(e.actual) }

// NEnv returns the number of "Environ" env that have been set.
func NEnv() int { return len(Environ.actual) }

// BoolVar defines a bool env with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the env.
func (e *EnvSet) BoolVar(p *bool, name string, value bool, usage string) {
	e.Var(newBoolValue(value, p), name, usage)
}

// BoolVar defines a bool env with specified name, default value, and usage string.
// The argument p points to a bool variable in which to store the value of the env.
func BoolVar(p *bool, name string, value bool, usage string) {
	Environ.Var(newBoolValue(value, p), name, usage)
}

// Bool defines a bool env with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the env.
func (e *EnvSet) Bool(name string, value bool, usage string) *bool {
	p := new(bool)
	e.BoolVar(p, name, value, usage)
	return p
}

// Bool defines a bool env with specified name, default value, and usage string.
// The return value is the address of a bool variable that stores the value of the env.
func Bool(name string, value bool, usage string) *bool {
	return Environ.Bool(name, value, usage)
}

// IntVar defines an int env with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the env.
func (e *EnvSet) IntVar(p *int, name string, value int, usage string) {
	e.Var(newIntValue(value, p), name, usage)
}

// IntVar defines an int env with specified name, default value, and usage string.
// The argument p points to an int variable in which to store the value of the env.
func IntVar(p *int, name string, value int, usage string) {
	Environ.Var(newIntValue(value, p), name, usage)
}

// Int defines an int env with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the env.
func (e *EnvSet) Int(name string, value int, usage string) *int {
	p := new(int)
	e.IntVar(p, name, value, usage)
	return p
}

// Int defines an int env with specified name, default value, and usage string.
// The return value is the address of an int variable that stores the value of the env.
func Int(name string, value int, usage string) *int {
	return Environ.Int(name, value, usage)
}

// Int64Var defines an int64 env with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the env.
func (e *EnvSet) Int64Var(p *int64, name string, value int64, usage string) {
	e.Var(newInt64Value(value, p), name, usage)
}

// Int64Var defines an int64 env with specified name, default value, and usage string.
// The argument p points to an int64 variable in which to store the value of the env.
func Int64Var(p *int64, name string, value int64, usage string) {
	Environ.Var(newInt64Value(value, p), name, usage)
}

// Int64 defines an int64 env with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the env.
func (e *EnvSet) Int64(name string, value int64, usage string) *int64 {
	p := new(int64)
	e.Int64Var(p, name, value, usage)
	return p
}

// Int64 defines an int64 env with specified name, default value, and usage string.
// The return value is the address of an int64 variable that stores the value of the env.
func Int64(name string, value int64, usage string) *int64 {
	return Environ.Int64(name, value, usage)
}

// UintVar defines a uint env with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the env.
func (e *EnvSet) UintVar(p *uint, name string, value uint, usage string) {
	e.Var(newUintValue(value, p), name, usage)
}

// UintVar defines a uint env with specified name, default value, and usage string.
// The argument p points to a uint variable in which to store the value of the env.
func UintVar(p *uint, name string, value uint, usage string) {
	Environ.Var(newUintValue(value, p), name, usage)
}

// Uint defines a uint env with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the env.
func (e *EnvSet) Uint(name string, value uint, usage string) *uint {
	p := new(uint)
	e.UintVar(p, name, value, usage)
	return p
}

// Uint defines a uint env with specified name, default value, and usage string.
// The return value is the address of a uint variable that stores the value of the env.
func Uint(name string, value uint, usage string) *uint {
	return Environ.Uint(name, value, usage)
}

// Uint64Var defines a uint64 env with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the env.
func (e *EnvSet) Uint64Var(p *uint64, name string, value uint64, usage string) {
	e.Var(newUint64Value(value, p), name, usage)
}

// Uint64Var defines a uint64 env with specified name, default value, and usage string.
// The argument p points to a uint64 variable in which to store the value of the env.
func Uint64Var(p *uint64, name string, value uint64, usage string) {
	Environ.Var(newUint64Value(value, p), name, usage)
}

// Uint64 defines a uint64 env with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the env.
func (e *EnvSet) Uint64(name string, value uint64, usage string) *uint64 {
	p := new(uint64)
	e.Uint64Var(p, name, value, usage)
	return p
}

// Uint64 defines a uint64 env with specified name, default value, and usage string.
// The return value is the address of a uint64 variable that stores the value of the env.
func Uint64(name string, value uint64, usage string) *uint64 {
	return Environ.Uint64(name, value, usage)
}

// StringVar defines a string env with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the env.
func (e *EnvSet) StringVar(p *string, name string, value string, usage string) {
	e.Var(newStringValue(value, p), name, usage)
}

// StringVar defines a string env with specified name, default value, and usage string.
// The argument p points to a string variable in which to store the value of the env.
func StringVar(p *string, name string, value string, usage string) {
	Environ.Var(newStringValue(value, p), name, usage)
}

// String defines a string env with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the env.
func (e *EnvSet) String(name string, value string, usage string) *string {
	p := new(string)
	e.StringVar(p, name, value, usage)
	return p
}

// String defines a string env with specified name, default value, and usage string.
// The return value is the address of a string variable that stores the value of the env.
func String(name string, value string, usage string) *string {
	return Environ.String(name, value, usage)
}

// Float64Var defines a float64 env with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the env.
func (e *EnvSet) Float64Var(p *float64, name string, value float64, usage string) {
	e.Var(newFloat64Value(value, p), name, usage)
}

// Float64Var defines a float64 env with specified name, default value, and usage string.
// The argument p points to a float64 variable in which to store the value of the env.
func Float64Var(p *float64, name string, value float64, usage string) {
	Environ.Var(newFloat64Value(value, p), name, usage)
}

// Float64 defines a float64 env with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the env.
func (e *EnvSet) Float64(name string, value float64, usage string) *float64 {
	p := new(float64)
	e.Float64Var(p, name, value, usage)
	return p
}

// Float64 defines a float64 env with specified name, default value, and usage string.
// The return value is the address of a float64 variable that stores the value of the env.
func Float64(name string, value float64, usage string) *float64 {
	return Environ.Float64(name, value, usage)
}

// DurationVar defines a time.Duration env with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the env.
// The env accepts a value acceptable to time.ParseDuration.
func (e *EnvSet) DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	e.Var(newDurationValue(value, p), name, usage)
}

// DurationVar defines a time.Duration env with specified name, default value, and usage string.
// The argument p points to a time.Duration variable in which to store the value of the env.
// The env accepts a value acceptable to time.ParseDuration.
func DurationVar(p *time.Duration, name string, value time.Duration, usage string) {
	Environ.Var(newDurationValue(value, p), name, usage)
}

// Duration defines a time.Duration env with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the env.
// The env accepts a value acceptable to time.ParseDuration.
func (e *EnvSet) Duration(name string, value time.Duration, usage string) *time.Duration {
	p := new(time.Duration)
	e.DurationVar(p, name, value, usage)
	return p
}

// Duration defines a time.Duration env with specified name, default value, and usage string.
// The return value is the address of a time.Duration variable that stores the value of the env.
// The env accepts a value acceptable to time.ParseDuration.
func Duration(name string, value time.Duration, usage string) *time.Duration {
	return Environ.Duration(name, value, usage)
}

// TextVar defines a env with a specified name, default value, and usage string.
// The argument p must be a pointer to a variable that will hold the value
// of the env, and p must implement encoding.TextUnmarshaler.
// If the env is used, the env value will be passed to p's UnmarshalText method.
// The type of the default value must be the same as the type of p.
func (e *EnvSet) TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	e.Var(newTextValue(value, p), name, usage)
}

// TextVar defines a env with a specified name, default value, and usage string.
// The argument p must be a pointer to a variable that will hold the value
// of the env, and p must implement encoding.TextUnmarshaler.
// If the env is used, the env value will be passed to p's UnmarshalText method.
// The type of the default value must be the same as the type of p.
func TextVar(p encoding.TextUnmarshaler, name string, value encoding.TextMarshaler, usage string) {
	Environ.Var(newTextValue(value, p), name, usage)
}

// Func defines a env with the specified name and usage string.
// Each time the env is seen, fn is called with the value of the env.
// If fn returns a non-nil error, it will be treated as a env value parsing error.
func (e *EnvSet) Func(name, usage string, fn func(string) error) {
	e.Var(funcValue(fn), name, usage)
}

// Func defines a env with the specified name and usage string.
// Each time the env is seen, fn is called with the value of the env.
// If fn returns a non-nil error, it will be treated as a env value parsing error.
func Func(name, usage string, fn func(string) error) {
	Environ.Func(name, usage, fn)
}

// Var defines a env with the specified name and usage string. The type and
// value of the env are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a env that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func (e *EnvSet) Var(value Value, name string, usage string) {
	// env must not contain "=".
	if strings.Contains(name, "=") {
		panic(e.sprintf("env %q contains =", name))
	}

	name = strings.ToUpper(name)

	// Remember the default value as a string; it won't change.
	env := &Env{name, usage, value, value.String()}
	_, alreadythere := e.formal[name]
	if alreadythere {
		var msg string
		if e.prefix == "" {
			msg = e.sprintf("env redefined: %s", name)
		} else {
			msg = e.sprintf("%s env redefined: %s", e.prefix, name)
		}
		panic(msg) // Happens only if envs are declared with identical names
	}
	if e.formal == nil {
		e.formal = make(map[string]*Env)
	}
	e.formal[name] = env
}

// Var defines a env with the specified name and usage string. The type and
// value of the env are represented by the first argument, of type Value, which
// typically holds a user-defined implementation of Value. For instance, the
// caller could create a env that turns a comma-separated string into a slice
// of strings by giving the slice the methods of Value; in particular, Set would
// decompose the comma-separated string into the slice.
func Var(value Value, name string, usage string) {
	Environ.Var(value, name, usage)
}

// sprintf formats the message, prints it to output, and returns it.
func (e *EnvSet) sprintf(format string, a ...interface{}) string {
	msg := fmt.Sprintf(format, a...)
	fmt.Fprintln(e.Output(), msg)
	return msg
}

// failf prints to standard error a formatted error and usage message and
// returns the error.
func (e *EnvSet) failf(format string, a ...interface{}) error {
	msg := e.sprintf(format, a...)
	e.usage()
	return errors.New(msg)
}

// usage calls the Usage method for the env set if one is specified,
// or the appropriate default usage function otherwise.
func (e *EnvSet) usage() {
	if e.Usage == nil {
		e.defaultUsage()
	} else {
		e.Usage()
	}
}

// parseOne parses one env. It reports whether a env was seen.
func (e *EnvSet) parseOne() (bool, error) {
	if len(e.envs) == 0 {
		return false, nil
	}

	s := e.envs[0]

	parts := strings.SplitN(s, "=", 2)
	if len(parts) != 2 {
		return false, e.failf("bad env syntax: %s", s)
	}

	e.envs = e.envs[1:]
	m := e.formal
	value := parts[1]
	name := parts[0]
	prefix := e.prefix

	if prefix != "" {
		prefix = strings.TrimPrefix(e.prefix, "_")
		prefix = strings.ToUpper(prefix) + "_"
		name = strings.TrimPrefix(parts[0], prefix)
	}

	if !strings.HasPrefix(parts[0], prefix) {
		return true, nil
	}

	env, alreadythere := m[name]
	if !alreadythere {
		//  e.failf("env provided but not defined: %s", name)
		// ignore not defined env.
		return true, nil
	}

	if err := env.Value.Set(value); err != nil {
		return false, e.failf("invalid value %q for env %s: %v", value, name, err)
	}

	if e.actual == nil {
		e.actual = make(map[string]*Env)
	}

	e.actual[name] = env
	return true, nil
}

// Parse parses env definitions from the envs list.
// Parse Must be called after all envs in the EnvSet
// are defined and before envs are accessed by the program.
func (e *EnvSet) Parse(envs []string) error {
	e.parsed = true
	e.envs = envs
	for {
		seen, err := e.parseOne()
		if seen {
			continue
		}
		if err == nil {
			break
		}
		switch e.errorHandling {
		case ContinueOnError:
			return err
		case ExitOnError:
			os.Exit(2)
		case PanicOnError:
			panic(err)
		}
	}
	return nil
}

// Parsed reports whether e.Parse has been called.
func (e *EnvSet) Parsed() bool {
	return e.parsed
}

// Parse parses the "Environ" envs from os.Environ(). Must be called
// after all envs are defined and before envs are accessed by the program.
func Parse() {
	// Ignore errors; Environ is set for ExitOnError.
	_ = Environ.Parse(os.Environ())
}

// Parsed reports whether the "Environ" envs have been parsed.
func Parsed() bool {
	return Environ.Parsed()
}

// Environ is the default env set, parsed from os.Environ().
// The top-level functions such as BoolVar, Parse, and so on are wrappers for the
// methods of Environ.
var Environ = NewEnvSet("", ExitOnError)

func init() {
	// Override generic EnvSet default Usage with call to global Usage.
	// Note: This is not Environ.Usage = Usage,
	// because we want any eventual call to use any updated value of Usage,
	// not the value it has when this line is run.
	Environ.Usage = environUsage
}

func environUsage() {
	Usage()
}

// NewEnvSet returns a new, empty env set with the specified prefix and
// error handling property. If the prefix is not empty, only env variables
// with the given prefix will be parsed, the prefix will be printed in the
// default usage message and in error messages.
func NewEnvSet(prefix string, errorHandling ErrorHandling) *EnvSet {
	e := new(EnvSet)
	e.Init(prefix, errorHandling)
	e.Usage = e.defaultUsage
	return e
}

// Init sets the prefix and error handling property for a env set.
// By default, the zero  EnvSet uses an empty prefix and the
// ContinueOnError error handling policy.
func (e *EnvSet) Init(prefix string, errorHandling ErrorHandling) {
	e.prefix = prefix
	e.errorHandling = errorHandling
}
