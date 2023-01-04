// Copyright (c) 2022 xybor-x
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package xyconfig_test

import (
	"testing"

	"github.com/xybor-x/xycond"
	"github.com/xybor-x/xyconfig"
)

func TestValueMustConfig(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo.bar", "bar", true)

	xycond.ExpectEqual(
		cfg.MustGet("foo").MustConfig(), xyconfig.GetConfig(t.Name()+".foo")).Test(t)

	xycond.ExpectPanic(xyconfig.CastError,
		func() { cfg.MustGet("foo.bar").MustConfig() }).Test(t)
}

func TestValueAsConfig(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo.bar", "bar", true)
	var c, ok = cfg.MustGet("foo").AsConfig()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(c, xyconfig.GetConfig(t.Name()+".foo")).Test(t)

	_, ok = cfg.MustGet("foo.bar").AsConfig()
	xycond.ExpectFalse(ok).Test(t)
}

func TestValueMustInt(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", 1, true)
	cfg.Set("bar", "string", true)
	cfg.Set("buzz", "1", false)
	cfg.Set("bizz", "string", false)

	xycond.ExpectEqual(cfg.MustGet("foo").MustInt(), 1).Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("bar").MustInt() }).Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz").MustInt(), 1).Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("bizz").MustInt() }).Test(t)
}

func TestValueAsInt(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", 1, true)
	cfg.Set("bar", "string", true)
	cfg.Set("buzz", "1", false)
	cfg.Set("bizz", "string", false)
	cfg.Set("float", 1.0, false)

	var i int
	var ok bool

	i, ok = cfg.MustGet("foo").AsInt()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1)

	_, ok = cfg.MustGet("bar").AsInt()
	xycond.ExpectFalse(ok).Test(t)

	i, ok = cfg.MustGet("buzz").AsInt()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1)

	_, ok = cfg.MustGet("bizz").AsInt()
	xycond.ExpectFalse(ok).Test(t)

	i, ok = cfg.MustGet("float").AsInt()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1).Test(t)
}

func TestValueMustFloat(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", 1, true)
	cfg.Set("fizz", 1.2, true)
	cfg.Set("bar", "string", true)
	cfg.Set("buzz", "1.3", false)
	cfg.Set("bizz", "string", false)

	xycond.ExpectEqual(cfg.MustGet("foo").MustFloat(), 1.0).Test(t)
	xycond.ExpectEqual(cfg.MustGet("fizz").MustFloat(), 1.2).Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("bar").MustFloat() }).Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz").MustFloat(), 1.3).Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("bizz").MustFloat() }).Test(t)
}

func TestValueAsFloat(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", 1, true)
	cfg.Set("fizz", 1.2, true)
	cfg.Set("bar", "string", true)
	cfg.Set("buzz", "1.3", false)
	cfg.Set("bizz", "string", false)

	var i float64
	var ok bool

	i, ok = cfg.MustGet("foo").AsFloat()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1.0)

	i, ok = cfg.MustGet("fizz").AsFloat()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1.2)

	_, ok = cfg.MustGet("bar").AsFloat()
	xycond.ExpectFalse(ok).Test(t)

	i, ok = cfg.MustGet("buzz").AsFloat()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1.3)

	_, ok = cfg.MustGet("bizz").AsFloat()
	xycond.ExpectFalse(ok).Test(t)
}

func TestValueMustBool(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", true, true)
	cfg.Set("fizz", 1, true)
	cfg.Set("bar", "false", true)
	cfg.Set("buzz", "false", false)
	cfg.Set("bizz", "FALSE", false)

	xycond.ExpectEqual(cfg.MustGet("foo").MustBool(), true).Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("fizz").MustBool() }).Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("bar").MustBool() }).Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz").MustBool(), false).Test(t)
	xycond.ExpectEqual(cfg.MustGet("bizz").MustBool(), false).Test(t)
}

func TestValueAsBool(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", true, true)
	cfg.Set("fizz", 1, true)
	cfg.Set("bar", "false", true)
	cfg.Set("buzz", "false", false)
	cfg.Set("bizz", "FALSE", false)

	var b bool
	var ok bool

	b, ok = cfg.MustGet("foo").AsBool()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(b, true)

	_, ok = cfg.MustGet("fizz").AsBool()
	xycond.ExpectFalse(ok).Test(t)

	_, ok = cfg.MustGet("bar").AsBool()
	xycond.ExpectFalse(ok).Test(t)

	b, ok = cfg.MustGet("buzz").AsBool()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(b, false)

	b, ok = cfg.MustGet("bizz").AsBool()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(b, false)
}

func TestValueMustString(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", "string", true)
	cfg.Set("bar", 1, true)

	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "string").Test(t)
	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("bar").MustString() }).Test(t)
}

func TestValueAsString(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", "string", true)
	cfg.Set("bar", 1, true)

	var i string
	var ok bool

	i, ok = cfg.MustGet("foo").AsString()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(i, 1)

	_, ok = cfg.MustGet("bar").AsString()
	xycond.ExpectFalse(ok).Test(t)
}

func TestValueMustArray(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", []any{1, "foo"}, true)
	cfg.Set("bar", `[1, foo]`, false)
	cfg.Set("buzz", `[1, `, false)

	var array = cfg.MustGet("foo").MustArray()
	xycond.ExpectEqual(array[0].MustInt(), 1).Test(t)
	xycond.ExpectEqual(array[1].MustString(), "foo").Test(t)

	array = cfg.MustGet("bar").MustArray()
	xycond.ExpectEqual(array[0].MustInt(), 1).Test(t)
	xycond.ExpectEqual(array[1].MustString(), "foo").Test(t)

	xycond.ExpectPanic(xyconfig.CastError, func() { cfg.MustGet("buzz").MustArray() }).Test(t)
}

func TestValueAsArray(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", []any{1, "foo"}, true)
	cfg.Set("bar", `[1, foo]`, false)
	cfg.Set("buzz", `[1, `, true)
	cfg.Set("bizz", `[1, `, false)

	var array []xyconfig.Value
	var ok bool

	array, ok = cfg.MustGet("foo").AsArray()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(array[0].MustInt(), 1).Test(t)
	xycond.ExpectEqual(array[1].MustString(), "foo").Test(t)

	array, ok = cfg.MustGet("bar").AsArray()
	xycond.ExpectTrue(ok).Test(t)
	xycond.ExpectEqual(array[0].MustInt(), 1).Test(t)
	xycond.ExpectEqual(array[1].MustString(), "foo").Test(t)

	_, ok = cfg.MustGet("buzz").AsArray()
	xycond.ExpectFalse(ok).Test(t)

	_, ok = cfg.MustGet("bizz").AsArray()
	xycond.ExpectFalse(ok).Test(t)
}

func TestValueString(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", 1, true)
	cfg.Set("bar", "string", true)
	cfg.Set("buzz", 1.2, false)
	cfg.Set("bizz", []any{1, "foo"}, false)

	xycond.ExpectEqual(cfg.MustGet("foo").String(), "1").Test(t)
	xycond.ExpectEqual(cfg.MustGet("bar").String(), "string").Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz").String(), "1.2").Test(t)
	xycond.ExpectEqual(cfg.MustGet("bizz").String(), "[1 foo]").Test(t)
}
