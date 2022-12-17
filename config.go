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

package xyconfig

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-ini/ini"
	"github.com/xybor-x/xyerror"
	"github.com/xybor-x/xylock"
	"github.com/xybor-x/xylog"
)

// Format represents supported file formats.
type Format int

// Supported file formats.
const (
	UnknownFormat Format = iota
	JSON
	INI
)

var loggerName = "xybor.xyplatform.xyconfig"
var logger = xylog.GetLogger(loggerName)

var extensions = map[string]Format{
	".json": JSON,
	".ini":  INI,
}

// Event represents for a changes in the config.
type Event struct {
	// Key is the key of value (including all parent keys with dot-separated).
	Key string

	// Old is the value before the change.
	Old Value

	// New is the value after the change.
	New Value
}

// Config contains configured values. It supports to read configuration files,
// watch their changes, and executes a custom hook when the change is applied.
//
// A Config can contain many key-value pairs, or other Config instances.
type Config struct {
	// name is the identification of Config. name of sub-Config include its
	// parent name with dot-separated.
	name string

	// config contains all key-value pairs.
	config map[string]Value

	// hook contains functions for hooking when a change is applied.
	hook map[string]func(Event)

	// watcher tracks changes of files.
	watcher *fsnotify.Watcher

	// lock avoids race condition.
	lock *xylock.RWLock
}

// globalLock avoids race condition of configMap
var globalLock = &xylock.RWLock{}

// configMap stores all Config instances created in program.
var configMap = map[string]*Config{}

// GetConfig gets the existing Config instance by the name or creates a new one
// if it hasn't existed yet.
//
// A Config instance is automatically created when new a dot-separated keys is
// added to the current Config. Its name is the dot-separated combination of the
// current Config's name and the its key.
//
// For example, when the key "system.delimiter" is added to a Config named
// "app", a new Config is automatically created with the name "app.system". This
// Config instance contains key-value pair of "delimiter".
func GetConfig(name string) *Config {
	var c = globalLock.RLockFunc(func() any {
		var c, ok = configMap[name]
		if ok {
			return c
		}
		return nil
	})

	if c != nil {
		return c.(*Config)
	}

	var cfg = &Config{
		config: make(map[string]Value),
		hook:   make(map[string]func(Event)),
		lock:   &xylock.RWLock{},
	}

	if name == "" {
		name = fmt.Sprintf("%p", cfg)
	}
	cfg.name = name

	globalLock.WLockFunc(func() {
		configMap[name] = cfg
	})
	return cfg
}

// CloseWatcher closes the watcher.
func (c *Config) CloseWatcher() error {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if c.watcher != nil {
		err = c.watcher.Close()
		c.watcher = nil
	}
	return err
}

// UnWatch removes a filename from the watcher.
func (c *Config) UnWatch(filename string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.watcher != nil {
		return c.watcher.Remove(filename)
	}
	return nil
}

// Set assigns the value to key. If the key doesn't exist, this method will
// create a new one, otherwise, it overrides the current value.
//
// The return value says if a hook function is executed for this change.
func (c *Config) Set(key string, value any, strict bool) bool {
	var old, ok = c.Get(key)
	if ok && old.value == value {
		return false
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	var before, after, found = strings.Cut(key, ".")
	var watched = false
	if !found {
		c.config[key] = Value{value: value, strict: strict}
	} else {
		if _, ok := c.config[before]; !ok {
			c.config[before] = Value{value: GetConfig(c.name + "." + before), strict: strict}
		}

		if _, ok := c.config[before].AsConfig(); !ok {
			c.config[before] = Value{value: GetConfig(c.name + "." + before), strict: strict}
		}

		watched = c.config[before].MustConfig().Set(after, value, strict)
	}

	if !watched {
		// Find the matched hook with the most detailed key.
		var prefix string
		var hook func(Event)
		for k, v := range c.hook {
			if k == "" || key == k || strings.HasPrefix(key, k+".") {
				if k == "" || len(k) > len(prefix) {
					prefix = k
					hook = v
				}
			}
		}

		if hook != nil {
			hook(Event{Old: old, New: Value{value, strict}, Key: c.name + "." + key})
			return true
		}
	}

	return false
}

// AddHook adds a hook function. This function will be executed when there is
// any change for values of the key.
//
// The hook function is executed according to the following priority:
//
// 1. If a key is hooked by many functions in some Config instances, the
// hook function of the Config being closest with the key is executed.
//
// 2. If a key is hooked by many functions in a Config instance, the hook
// function with the most detailed key is executed.
//
// Only one hook function is executed in a change.
//
// For example, a change is applied for the key "general.system.timeout":
//
// 1. For the first case, the func2 is executed because c2 is the closer Config
// with "timeout".
//    var c1 = xyconfig.GetConfig("config")
//    var c2 = xyconfig.GetConfig("config.general")
//    c1.AddHook("general.system", func1)
//    c2.AddHook("system", func2)
//
// 2. For the second case, the func2 is executed because "general.system" is
// the more detailed key.
//    var c = xyconfig.GetConfig("config")
//    c.AddHook("general", func1)
//    c.AddHook("general.system", func2)
//    c.AddHook("general.os", func3)
func (c *Config) AddHook(key string, f func(e Event)) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.hook[key] = f
}

// ReadMap reads the config values from a map. If strict is false and the values
// of map are strings, it allows casting them to other types.
func (c *Config) ReadMap(m map[string]any) error {
	for k, v := range m {
		switch t := v.(type) {
		case map[string]any:
			var cfg = GetConfig(c.name + "." + k)
			if err := cfg.ReadMap(t); err != nil {
				return err
			}
			c.Set(k, cfg, true)
		default:
			c.Set(k, t, true)
		}
	}

	return nil
}

// ReadJSON reads the config values from a byte array under JSON format.
func (c *Config) ReadJSON(b []byte) error {
	var m map[string]any
	var err = json.Unmarshal(b, &m)
	if err != nil {
		return xyerror.ValueError.Newf("cannot parse json data (%v)", err)
	}

	return c.ReadMap(m)
}

// ReadINI reads the config values from a byte array under INI format.
func (c *Config) ReadINI(b []byte) error {
	var cfg, err = ini.Load(b)
	if err != nil {
		return xyerror.ValueError.New(err)
	}

	for _, section := range cfg.Sections() {
		for _, key := range section.Keys() {
			c.Set(section.Name()+"."+key.Name(), key.Value(), false)
		}
	}

	return nil
}

// ReadBytes reads the config values from a bytes array under any format.
func (c *Config) ReadBytes(format Format, b []byte) error {
	switch format {
	case JSON:
		return c.ReadJSON(b)
	case INI:
		return c.ReadINI(b)
	default:
		return FormatError.New("unsupported format")
	}
}

// Read reads the config values from a string under any format.
func (c *Config) Read(format Format, s string) error {
	return c.ReadBytes(format, []byte(s))
}

// ReadFile reads the config values from a file. If watch is true, it will
// reload config when the file is changed.
func (c *Config) ReadFile(filename string, watch bool) error {
	var fileFormat = UnknownFormat
	for ext, format := range extensions {
		if strings.HasSuffix(filename, ext) {
			fileFormat = format
		}
	}

	if fileFormat == UnknownFormat {
		return ExtensionError.Newf("unknown extension: %s", filename)
	}

	var data, err = ioutil.ReadFile(filename)
	if err != nil {
		return BaseError.New(err)
	}

	if err = c.ReadBytes(fileFormat, data); err != nil {
		return err
	}

	if watch {
		return c.watchFile(filename)
	}

	return nil
}

// Get returns the value assigned with the key. The latter returned value is
// false if they key doesn't exist.
func (c *Config) Get(key string) (Value, bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	var before, after, found = strings.Cut(key, ".")
	var v, ok = c.config[before]
	if !found {
		return v, ok
	}

	cfg, ok := v.AsConfig()
	if ok {
		return cfg.Get(after)
	}

	return Value{}, false
}

// MustGet returns the value assigned with the key. It panics if the key doesn't
// exist.
func (c *Config) MustGet(key string) Value {
	var v, ok = c.Get(key)
	if !ok {
		panic(ConfigKeyError.Newf("unknown key %s", key))
	}
	return v
}

// GetDefault returns the value assigned with the key. It returns the default
// value if the key doesn't exist.
func (c *Config) GetDefault(key string, def any) Value {
	var v, ok = c.Get(key)
	if !ok {
		return Value{value: def, strict: true}
	}
	return v
}

// initWatcher assigns a new watcher to Config. It also run a goroutine for
// handling watcher events.
func (c *Config) initWatcher() error {
	var watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return BaseError.New(err)
	}

	c.lock.WLockFunc(func() {
		c.watcher = watcher
	})

	go func() {
		var watcherEvents = watcher.Events
		var watcherErrors = watcher.Errors

		for {
			select {
			case event, ok := <-watcherEvents:
				if !ok {
					logger.Event("watcher-stop").Info()
					return
				}

				if event.Has(fsnotify.Write) {
					var err = c.ReadFile(event.Name, false)
					if err != nil {
						logger.Event("reload-error").
							Field("filename", event.Name).Field("error", err).Warning()
					} else {
						logger.Event("reload-config").Field("filename", event.Name).Info()
					}
				}

			case err, ok := <-watcherErrors:
				if !ok {
					logger.Event("watcher-stop").Info()
					return
				}
				logger.Event("watcher-error").Field("error", err).Warning()
			}
		}
	}()

	return nil
}

// watchFile adds filename to watcher. If the watcher has not initialized yet,
// create a new one.
func (c *Config) watchFile(filename string) error {
	var watcher = c.lock.RLockFunc(func() any {
		return c.watcher
	}).(*fsnotify.Watcher)

	if watcher == nil {
		if err := c.initWatcher(); err != nil {
			return err
		}
	}

	c.lock.Lock()
	defer c.lock.Unlock()
	if err := c.watcher.Add(filename); err != nil {
		return BaseError.New(err)
	}

	return nil
}
