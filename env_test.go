// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package env_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/shaj13/env"
)

func boolString(s string) string {
	if s == "0" {
		return "false"
	}
	return "true"
}

func TestEverything(t *testing.T) {
	ResetForTesting(nil)
	Bool("test_bool", false, "bool value")
	Int("test_int", 0, "int value")
	Int64("test_int64", 0, "int64 value")
	Uint("test_uint", 0, "uint value")
	Uint64("test_uint64", 0, "uint64 value")
	String("test_string", "0", "string value")
	Float64("test_float64", 0, "float64 value")
	Duration("test_duration", 0, "time.Duration value")
	Func("test_func", "func value", func(string) error { return nil })

	m := make(map[string]*Env)
	desired := "0"
	visitor := func(f *Env) {
		if len(f.Name) > 5 && f.Name[0:5] == "TEST_" {
			m[f.Name] = f
			ok := false
			switch {
			case f.Value.String() == desired:
				ok = true
			case f.Name == "TEST_BOOL" && f.Value.String() == boolString(desired):
				ok = true
			case f.Name == "TEST_DURATION" && f.Value.String() == desired+"s":
				ok = true
			case f.Name == "TEST_FUNC" && f.Value.String() == "":
				ok = true
			}
			if !ok {
				t.Error("Visit: bad value", f.Value.String(), "for", f.Name)
			}
		}
	}
	VisitAll(visitor)
	if len(m) != 9 {
		t.Error("VisitAll misses some envs")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	m = make(map[string]*Env)
	Visit(visitor)
	if len(m) != 0 {
		t.Errorf("Visit sees unset envs")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now set all envs
	Set("test_bool", "true")
	Set("test_int", "1")
	Set("test_int64", "1")
	Set("test_uint", "1")
	Set("test_uint64", "1")
	Set("test_string", "1")
	Set("test_float64", "1")
	Set("test_duration", "1s")
	Set("test_func", "1")
	desired = "1"
	Visit(visitor)
	if len(m) != 9 {
		t.Error("Visit fails after set")
		for k, v := range m {
			t.Log(k, *v)
		}
	}
	// Now test they're visited in sort order.
	var envNames []string
	Visit(func(f *Env) { envNames = append(envNames, f.Name) })
	if !sort.StringsAreSorted(envNames) {
		t.Errorf("env names not sorted: %v", envNames)
	}
}

func TestGet(t *testing.T) {
	ResetForTesting(nil)
	Bool("test_bool", true, "bool value")
	Int("test_int", 1, "int value")
	Int64("test_int64", 2, "int64 value")
	Uint("test_uint", 3, "uint value")
	Uint64("test_uint64", 4, "uint64 value")
	String("test_string", "5", "string value")
	Float64("test_float64", 6, "float64 value")
	Duration("test_duration", 7, "time.Duration value")

	visitor := func(f *Env) {
		if len(f.Name) > 5 && f.Name[0:5] == "test_" {
			g, ok := f.Value.(Getter)
			if !ok {
				t.Errorf("Visit: value does not satisfy Getter: %T", f.Value)
				return
			}
			switch f.Name {
			case "test_bool":
				ok = g.Get() == true
			case "test_int":
				ok = g.Get() == int(1)
			case "test_int64":
				ok = g.Get() == int64(2)
			case "test_uint":
				ok = g.Get() == uint(3)
			case "test_uint64":
				ok = g.Get() == uint64(4)
			case "test_string":
				ok = g.Get() == "5"
			case "test_float64":
				ok = g.Get() == float64(6)
			case "test_duration":
				ok = g.Get() == time.Duration(7)
			}
			if !ok {
				t.Errorf("Visit: bad value %T(%v) for %s", g.Get(), g.Get(), f.Name)
			}
		}
	}
	VisitAll(visitor)
}

func TestUsage(t *testing.T) {
	called := false
	ResetForTesting(func() { called = true })
	if Environ.Parse([]string{"-x"}) == nil {
		t.Error("parse did not fail for unknown env")
	}
	if !called {
		t.Error("did not call Usage for unknown env")
	}
}

func testParse(f *EnvSet, t *testing.T) {
	if f.Parsed() {
		t.Error("f.Parse() = true before Parse")
	}
	boolEnv := f.Bool("bool", false, "bool value")
	intEnv := f.Int("int", 0, "int value")
	int64Env := f.Int64("int64", 0, "int64 value")
	uintEnv := f.Uint("uint", 0, "uint value")
	uint64Env := f.Uint64("uint64", 0, "uint64 value")
	stringEnv := f.String("string", "0", "string value")
	float64Env := f.Float64("float64", 0, "float64 value")
	durationEnv := f.Duration("duration", 5*time.Second, "time.Duration value")
	args := []string{
		"BOOL=true",
		"INT=22",
		"INT64=0x23",
		"UINT=24",
		"UINT64=25",
		"STRING=hello",
		"FLOAT64=2718e28",
		"DURATION=2m",
	}
	if err := f.Parse(args); err != nil {
		t.Fatal(err)
	}
	if !f.Parsed() {
		t.Error("f.Parse() = false after Parse")
	}
	if *boolEnv != true {
		t.Error("bool env should be true, is ", *boolEnv)
	}
	if *intEnv != 22 {
		t.Error("int env should be 22, is ", *intEnv)
	}
	if *int64Env != 0x23 {
		t.Error("int64 env should be 0x23, is ", *int64Env)
	}
	if *uintEnv != 24 {
		t.Error("uint env should be 24, is ", *uintEnv)
	}
	if *uint64Env != 25 {
		t.Error("uint64 env should be 25, is ", *uint64Env)
	}
	if *stringEnv != "hello" {
		t.Error("string env should be `hello`, is ", *stringEnv)
	}
	if *float64Env != 2718e28 {
		t.Error("float64 env should be 2718e28, is ", *float64Env)
	}
	if *durationEnv != 2*time.Minute {
		t.Error("duration env should be 2m, is ", *durationEnv)
	}
}

func TestParse(t *testing.T) {
	ResetForTesting(func() { t.Error("bad parse") })
	testParse(Environ, t)
}

func TestEnvSetParse(t *testing.T) {
	testParse(NewEnvSet("", ContinueOnError), t)
}

// Declare a user-defined env type.
type envVar []string

func (e *envVar) String() string {
	return fmt.Sprint([]string(*e))
}

func (e *envVar) Set(value string) error {
	*e = append(*e, value)
	return nil
}

func TestUserDefined(t *testing.T) {
	var envs EnvSet
	envs.Init("", ContinueOnError)
	envs.SetOutput(io.Discard)
	var v envVar
	envs.Var(&v, "v", "usage")
	if err := envs.Parse([]string{"V=1", "V=2", "V=3"}); err != nil {
		t.Error(err)
	}
	if len(v) != 3 {
		t.Fatal("expected 3 args; got ", len(v))
	}
	expect := "[1 2 3]"
	if v.String() != expect {
		t.Errorf("expected value %q got %q", expect, v.String())
	}
}

func TestUserDefinedFunc(t *testing.T) {
	envs := NewEnvSet("", ContinueOnError)
	envs.SetOutput(io.Discard)
	var ss []string
	envs.Func("v", "usage", func(s string) error {
		ss = append(ss, s)
		return nil
	})
	if err := envs.Parse([]string{"V=1", "V=2", "V=3"}); err != nil {
		t.Error(err)
	}
	if len(ss) != 3 {
		t.Fatal("expected 3 args; got ", len(ss))
	}
	expect := "[1 2 3]"
	if got := fmt.Sprint(ss); got != expect {
		t.Errorf("expected value %q got %q", expect, got)
	}

	// test Func error
	envs = NewEnvSet("", ContinueOnError)
	envs.SetOutput(io.Discard)
	envs.Func("v", "usage", func(s string) error {
		return fmt.Errorf("test error")
	})
	// env not set, so no error
	if err := envs.Parse(nil); err != nil {
		t.Error(err)
	}
	// env set, expect error
	if err := envs.Parse([]string{"V=1"}); err == nil {
		t.Error("expected error; got none")
	} else if errMsg := err.Error(); !strings.Contains(errMsg, "test error") {
		t.Errorf(`error should contain "test error"; got %q`, errMsg)
	}
}

func TestUserDefinedForEnviron(t *testing.T) {
	const help = "HELP"
	var result string
	ResetForTesting(func() { result = help })
	Usage()
	if result != help {
		t.Fatalf("got %q; expected %q", result, help)
	}
}

// Declare a user-defined boolean env type.
type boolEnvVar struct {
	count int
}

func (b *boolEnvVar) String() string {
	return fmt.Sprintf("%d", b.count)
}

func (b *boolEnvVar) Set(value string) error {
	if value == "true" {
		b.count++
	}
	return nil
}

func (b *boolEnvVar) IsBoolEnv() bool {
	return b.count < 4
}

func TestUserDefinedBool(t *testing.T) {
	var envs EnvSet
	envs.Init("", ContinueOnError)
	envs.SetOutput(io.Discard)
	var b boolEnvVar
	var err error
	envs.Var(&b, "b", "usage")
	if err = envs.Parse([]string{"B=true", "B=true", "B=true", "B=false", "B=true"}); err != nil {
		if b.count < 4 {
			t.Error(err)
		}
	}

	if b.count != 4 {
		t.Errorf("want: %d; got: %d", 4, b.count)
	}
}

func TestSetOutput(t *testing.T) {
	var envs EnvSet
	var buf bytes.Buffer
	envs.SetOutput(&buf)
	envs.Init("test", ContinueOnError)
	envs.Parse([]string{"-unknown"})
	if out := buf.String(); !strings.Contains(out, "-unknown") {
		t.Logf("expected output mentioning unknown; got %q", out)
	}
}

// zeroPanicker is a env.Value whose String method panics if its dontPanic
// field is false.
type zeroPanicker struct {
	dontPanic bool
	v         string
}

func (f *zeroPanicker) Set(s string) error {
	f.v = s
	return nil
}

func (f *zeroPanicker) String() string {
	if !f.dontPanic {
		panic("panic!")
	}
	return f.v
}

const defaultOutput = `      PRINTDEFAULTS_A bool              for bootstrapping, allow 'any' type
      PRINTDEFAULTS_ALONGENVNAME bool   disable bounds checking
      PRINTDEFAULTS_C bool              a boolean defaulting to true (default true)
      PRINTDEFAULTS_D path              set relative path for local imports
      PRINTDEFAULTS_E string            issue 23543 (default "0")
      PRINTDEFAULTS_F number            a non-zero number (default 2.7)
      PRINTDEFAULTS_G float             a float that defaults to zero
      PRINTDEFAULTS_M string            a multiline
                                        help
                                        string
      PRINTDEFAULTS_MAXT timeout        set timeout for dial
      PRINTDEFAULTS_N int               a non-zero int (default 27)
      PRINTDEFAULTS_O bool              a env
                                        multiline help string (default true)
      PRINTDEFAULTS_V list              a list of strings (default [a b])
      PRINTDEFAULTS_Z int               an int that defaults to zero
      PRINTDEFAULTS_ZP0 value           a env whose String method panics when it is zero
      PRINTDEFAULTS_ZP1 value           a env whose String method panics when it is zero

panic calling String method on zero env_test.zeroPanicker for env ZP0: panic!
panic calling String method on zero env_test.zeroPanicker for env ZP1: panic!
`

func TestPrintDefaults(t *testing.T) {
	fs := NewEnvSet("PrintDefaults", ContinueOnError)
	var buf bytes.Buffer
	fs.SetOutput(&buf)
	fs.Bool("A", false, "for bootstrapping, allow 'any' type")
	fs.Bool("Alongenvname", false, "disable bounds checking")
	fs.Bool("C", true, "a boolean defaulting to true")
	fs.String("D", "", "set relative `path` for local imports")
	fs.String("E", "0", "issue 23543")
	fs.Float64("F", 2.7, "a non-zero `number`")
	fs.Float64("G", 0, "a float that defaults to zero")
	fs.String("M", "", "a multiline\nhelp\nstring")
	fs.Int("N", 27, "a non-zero int")
	fs.Bool("O", true, "a env\nmultiline help string")
	fs.Var(&envVar{"a", "b"}, "V", "a `list` of strings")
	fs.Int("Z", 0, "an int that defaults to zero")
	fs.Var(&zeroPanicker{true, ""}, "ZP0", "a env whose String method panics when it is zero")
	fs.Var(&zeroPanicker{true, "something"}, "ZP1", "a env whose String method panics when it is zero")
	fs.Duration("maxT", 0, "set `timeout` for dial")
	fs.PrintDefaults()
	got := buf.String()
	if got != defaultOutput {
		t.Errorf("got:\n%q\nwant:\n%q", got, defaultOutput)
	}

	// out := "const defaultOutput = `" + got
	// os.WriteFile("./out", []byte(out), 0o600)
}

// Issue 19230: validate range of Int and Uint env values.
func TestIntEnvOverflow(t *testing.T) {
	if strconv.IntSize != 32 {
		return
	}
	ResetForTesting(nil)
	Int("i", 0, "")
	Uint("u", 0, "")
	if err := Set("i", "2147483648"); err == nil {
		t.Error("unexpected success setting Int")
	}
	if err := Set("u", "4294967296"); err == nil {
		t.Error("unexpected success setting Uint")
	}
}

// Issue 20998: Usage should respect Environ.output.
func TestUsageOutput(t *testing.T) {
	ResetForTesting(DefaultUsage)
	var buf bytes.Buffer
	Environ.SetOutput(&buf)
	Int("test_int", 0, "test")
	os.Setenv("TEST_INT", "string")
	defer func(old []string) {
		os.Unsetenv("TEST_INT")
		os.Args = old
	}(os.Args)
	os.Args = []string{"app"}
	Parse()
	const want = "invalid value \"string\" for env TEST_INT: parse error\nUsage of app:\n      TEST_INT int   test\n"
	if got := buf.String(); got != want {
		t.Errorf("output = %q; want %q", got, want)
	}
}

func TestGetters(t *testing.T) {
	expectedName := "env set"
	expectedErrorHandling := ContinueOnError
	expectedOutput := io.Writer(os.Stderr)
	es := NewEnvSet(expectedName, expectedErrorHandling)

	if es.Prefix() != expectedName {
		t.Errorf("unexpected name: got %s, expected %s", es.Prefix(), expectedName)
	}
	if es.ErrorHandling() != expectedErrorHandling {
		t.Errorf("unexpected ErrorHandling: got %d, expected %d", es.ErrorHandling(), expectedErrorHandling)
	}
	if es.Output() != expectedOutput {
		t.Errorf("unexpected output: got %#v, expected %#v", es.Output(), expectedOutput)
	}

	expectedName = "gopher"
	expectedErrorHandling = ExitOnError
	expectedOutput = os.Stdout
	es.Init(expectedName, expectedErrorHandling)
	es.SetOutput(expectedOutput)

	if es.Prefix() != expectedName {
		t.Errorf("unexpected name: got %s, expected %s", es.Prefix(), expectedName)
	}
	if es.ErrorHandling() != expectedErrorHandling {
		t.Errorf("unexpected ErrorHandling: got %d, expected %d", es.ErrorHandling(), expectedErrorHandling)
	}
	if es.Output() != expectedOutput {
		t.Errorf("unexpected output: got %v, expected %v", es.Output(), expectedOutput)
	}
}

func TestParseError(t *testing.T) {
	for _, typ := range []string{"BOOL", "INT", "INT64", "UINT", "UINT64", "FLOAT64", "DURATION"} {
		fs := NewEnvSet("", ContinueOnError)
		fs.SetOutput(io.Discard)
		_ = fs.Bool("bool", false, "")
		_ = fs.Int("int", 0, "")
		_ = fs.Int64("int64", 0, "")
		_ = fs.Uint("uint", 0, "")
		_ = fs.Uint64("uint64", 0, "")
		_ = fs.Float64("float64", 0, "")
		_ = fs.Duration("duration", 0, "")
		// Strings cannot give errors.
		args := []string{typ + "=x"}
		err := fs.Parse(args) // x is not a valid setting for any env.
		if err == nil {
			t.Errorf("Parse(%q)=%v; expected parse error", args, err)
			continue
		}
		if !strings.Contains(err.Error(), "invalid") || !strings.Contains(err.Error(), "parse error") {
			t.Errorf("Parse(%q)=%v; expected parse error", args, err)
		}
	}
}

func TestRangeError(t *testing.T) {
	bad := []string{
		"INT=123456789012345678901",
		"INT64=123456789012345678901",
		"UINT=123456789012345678901",
		"UINT64=123456789012345678901",
		"FLOAT64=1e1000",
	}
	for _, arg := range bad {
		fs := NewEnvSet("", ContinueOnError)
		fs.SetOutput(io.Discard)
		_ = fs.Int("int", 0, "")
		_ = fs.Int64("int64", 0, "")
		_ = fs.Uint("uint", 0, "")
		_ = fs.Uint64("uint64", 0, "")
		_ = fs.Float64("float64", 0, "")
		// Strings cannot give errors, and bools and durations do not return strconv.NumError.
		err := fs.Parse([]string{arg})
		if err == nil {
			t.Errorf("Parse(%q)=%v; expected range error", arg, err)
			continue
		}
		if !strings.Contains(err.Error(), "invalid") || !strings.Contains(err.Error(), "value out of range") {
			t.Errorf("Parse(%q)=%v; expected range error", arg, err)
		}
	}
}

func TestExitCode(t *testing.T) {
	magic := 123
	if os.Getenv("GO_CHILD_ENV") != "" {
		fs := NewEnvSet("", ExitOnError)
		if os.Getenv("GO_CHILD_ENV_HANDLE") != "" {
			var i int
			fs.IntVar(&i, os.Getenv("GO_CHILD_ENV_HANDLE"), 0, "")
		}
		fs.Parse([]string{os.Getenv("GO_CHILD_ENV_HANDLE") + "=" + os.Getenv("GO_CHILD_ENV")})
		os.Exit(magic)
	}

	tests := []struct {
		env        string
		envHandle  string
		expectExit int
	}{
		{
			env:        "1",
			envHandle:  "INT",
			expectExit: magic,
		},
		{
			env:        "not_int",
			envHandle:  "INT",
			expectExit: 2,
		},
		{
			env:        "not_int",
			expectExit: magic,
		},
	}

	for _, test := range tests {
		cmd := exec.Command(os.Args[0], "-test.run=TestExitCode")
		cmd.Env = []string{
			"GO_CHILD_ENV=" + test.env,
			"GO_CHILD_ENV_HANDLE=" + test.envHandle,
		}
		cmd.Run()
		got := cmd.ProcessState.ExitCode()
		// ExitCode is either 0 or 1 on Plan 9.
		if runtime.GOOS == "plan9" && test.expectExit != 0 {
			test.expectExit = 1
		}
		if got != test.expectExit {
			t.Errorf("unexpected exit code for test case %+v \n: got %d, expect %d",
				test, got, test.expectExit)
		}
	}
}

func mustPanic(t *testing.T, testName string, expected string, f func()) {
	t.Helper()
	defer func() {
		switch msg := recover().(type) {
		case nil:
			t.Errorf("%s\n: expected panic(%q), but did not panic", testName, expected)
		case string:
			if msg != expected {
				t.Errorf("%s\n: expected panic(%q), but got panic(%q)", testName, expected, msg)
			}
		default:
			t.Errorf("%s\n: expected panic(%q), but got panic(%T%v)", testName, expected, msg, msg)
		}
	}()
	f()
}

func TestInvalidEnv(t *testing.T) {
	tests := []struct {
		env      string
		errorMsg string
	}{
		{
			env:      "foo=bar",
			errorMsg: "env \"foo=bar\" contains =",
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("EnvSet.Var(&v, %q, \"\")", test.env)

		es := NewEnvSet("", ContinueOnError)
		buf := bytes.NewBuffer(nil)
		es.SetOutput(buf)

		mustPanic(t, testName, test.errorMsg, func() {
			var v envVar
			es.Var(&v, test.env, "")
		})
		if msg := test.errorMsg + "\n"; msg != buf.String() {
			t.Errorf("%s\n: unexpected output: expected %q, bug got %q", testName, msg, buf)
		}
	}
}

func TestRedefinedEnvs(t *testing.T) {
	tests := []struct {
		envSetName string
		errorMsg   string
	}{
		{
			envSetName: "",
			errorMsg:   "env redefined: FOO",
		},
		{
			envSetName: "es",
			errorMsg:   "es env redefined: FOO",
		},
	}

	for _, test := range tests {
		testName := fmt.Sprintf("env redefined in EnvSet(%q)", test.envSetName)

		fs := NewEnvSet(test.envSetName, ContinueOnError)
		buf := bytes.NewBuffer(nil)
		fs.SetOutput(buf)

		var v envVar
		fs.Var(&v, "foo", "")

		mustPanic(t, testName, test.errorMsg, func() {
			fs.Var(&v, "foo", "")
		})
		if msg := test.errorMsg + "\n"; msg != buf.String() {
			t.Errorf("%s\n: unexpected output: expected %q, bug got %q", testName, msg, buf)
		}
	}
}
