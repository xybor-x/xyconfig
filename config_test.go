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
	"io/ioutil"
	"testing"
	"time"

	"github.com/xybor-x/xycond"
	"github.com/xybor-x/xyconfig"
	"github.com/xybor-x/xyerror"
)

func TestConfigGetConfigSameName(t *testing.T) {
	xycond.ExpectEqual(
		xyconfig.GetConfig(t.Name()), xyconfig.GetConfig(t.Name())).Test(t)
}

func TestConfigGetConfigEmptyName(t *testing.T) {
	xycond.ExpectNotEqual(
		xyconfig.GetConfig(""), xyconfig.GetConfig("")).Test(t)
}

func TestConfigSet(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", "bar", true)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
}

func TestConfigSetSubConfig(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo.buzz", "bar", true)
	xycond.ExpectEqual(cfg.MustGet("foo.buzz").MustString(), "bar").Test(t)
	cfg.Set("foo.buzz.bar", "bar", true)
	xycond.ExpectEqual(cfg.MustGet("foo.buzz.bar").MustString(), "bar").Test(t)
}

func TestConfigSetWithHook(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var event xyconfig.Event
	cfg.AddHook("foo", func(e xyconfig.Event) {
		event = e
	})
	cfg.Set("foo", "bar", true)
	xycond.ExpectTrue(event.Old.IsNil()).Test(t)
	xycond.ExpectEqual(event.New.MustString(), "bar").Test(t)
	xycond.ExpectEqual(event.Key, t.Name()+".foo").Test(t)
}

func TestConfigReadMap(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadMap(map[string]any{
		"foo": "bar",
		"buzz": map[string]any{
			"bizz": "bemm",
		},
		"nil": nil,
	})

	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz.bizz").MustString(), "bemm").Test(t)
	xycond.ExpectTrue(cfg.MustGet("nil").IsNil()).Test(t)
}

func TestConfigReadJSON(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadJSON([]byte(`{"foo": "bar", "buzz": {"bizz": "bemm"}, "nil": null}`))

	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz.bizz").MustString(), "bemm").Test(t)
	xycond.ExpectTrue(cfg.MustGet("nil").IsNil()).Test(t)
}

func TestConfigReadJSONWithDotKey(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadJSON([]byte(`{"foo.buzz": "bar"}`))

	xycond.ExpectEqual(cfg.MustGet("foo.buzz").MustString(), "bar").Test(t)
}

func TestConfigReadJSONWithError(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.ReadJSON([]byte(`{""`))

	xycond.ExpectError(err, xyerror.ValueError).Test(t)
}

func TestConfigReadINI(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadINI([]byte("[foo]\nfizz=bar\n[buzz]\nbizz=bemm\nnil="))

	xycond.ExpectEqual(cfg.MustGet("foo.fizz").MustString(), "bar").Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz.bizz").MustString(), "bemm").Test(t)
	xycond.ExpectTrue(cfg.MustGet("buzz.nil").IsNil()).Test(t)
}

func TestConfigReadINIWithError(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.ReadINI([]byte("[foo]\nbar"))

	xycond.ExpectError(err, xyerror.ValueError).Test(t)
}

func TestConfigReadByteJSON(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadBytes(xyconfig.JSON, []byte(`{"foo": "bar", "buzz": {"bizz": "bemm"}, "nil": null}`))

	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz.bizz").MustString(), "bemm").Test(t)
	xycond.ExpectTrue(cfg.MustGet("nil").IsNil()).Test(t)
}

func TestConfigReadByteINI(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadBytes(xyconfig.INI, []byte("[foo]\nfizz=bar\n[buzz]\nbizz=bemm\nnil="))

	xycond.ExpectEqual(cfg.MustGet("foo.fizz").MustString(), "bar").Test(t)
	xycond.ExpectEqual(cfg.MustGet("buzz.bizz").MustString(), "bemm").Test(t)
	xycond.ExpectTrue(cfg.MustGet("buzz.nil").IsNil()).Test(t)
}

func TestConfigReadByteUnknown(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.ReadBytes(xyconfig.UnknownFormat, []byte(""))

	xycond.ExpectError(err, xyconfig.FormatError).Test(t)
}

func TestConfigRead(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.Read(xyconfig.UnknownFormat, "")

	xycond.ExpectError(err, xyconfig.FormatError).Test(t)
}

func TestConfigReadFileUnknownExt(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.ReadFile("foo.bar", false)

	xycond.ExpectError(err, xyconfig.ExtensionError).Test(t)
}

func TestConfigReadFileNotExist(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.ReadFile("foo.json", false)

	xycond.ExpectError(err, xyconfig.BaseError).Test(t)
}

func TestConfigReadFileWithChange(t *testing.T) {
	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "bar"}`), 0644)

	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadFile(t.Name()+".json", true)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)

	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "buzz"}`), 0644)
	time.Sleep(time.Millisecond)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "buzz").Test(t)
}

func TestConfigReadFileWithChangeUnWatch(t *testing.T) {
	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "bar"}`), 0644)

	var cfg = xyconfig.GetConfig(t.Name())
	cfg.NoWatch()
	cfg.ReadFile(t.Name()+".json", true)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)

	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "buzz"}`), 0644)
	time.Sleep(time.Millisecond)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
}

func TestConfigReadFileUnWatchAfterRead(t *testing.T) {
	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "bar"}`), 0644)

	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadFile(t.Name()+".json", true)
	cfg.NoWatch()
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)

	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "buzz"}`), 0644)
	time.Sleep(time.Millisecond)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
}

func TestConfigReadFileWithErrorFileAfterChange(t *testing.T) {
	ioutil.WriteFile(t.Name()+".json", []byte(`{"foo": "bar"}`), 0644)

	var cfg = xyconfig.GetConfig(t.Name())
	cfg.ReadFile(t.Name()+".json", true)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)

	ioutil.WriteFile(t.Name()+".json", []byte(`{"error`), 0644)
	time.Sleep(time.Millisecond)
	xycond.ExpectEqual(cfg.MustGet("foo").MustString(), "bar").Test(t)
}

func TestConfigReadFileErrorParse(t *testing.T) {
	ioutil.WriteFile(t.Name()+".json", []byte(`{"error`), 0644)

	var cfg = xyconfig.GetConfig(t.Name())
	var err = cfg.ReadFile(t.Name()+".json", false)

	xycond.ExpectError(err, xyerror.ValueError).Test(t)
}

func TestConfigMustGet(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", "bar", true)

	xycond.ExpectPanic(xyconfig.ConfigKeyError, func() { cfg.MustGet("bar") })
}

func TestConfigGetDefault(t *testing.T) {
	var cfg = xyconfig.GetConfig(t.Name())
	cfg.Set("foo", "bar", true)

	xycond.ExpectEqual(cfg.GetDefault("foo", "buzzz").MustString(), "bar").Test(t)
	xycond.ExpectEqual(cfg.GetDefault("bar", "buzzz").MustString(), "buzzz").Test(t)
}
