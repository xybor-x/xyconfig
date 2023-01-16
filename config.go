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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/fsnotify/fsnotify"
	"github.com/go-ini/ini"
	"github.com/joho/godotenv"
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
	ENV
)

var loggerName = "xybor.xyplatform.xyconfig"
var logger = xylog.GetLogger(loggerName)

var extensions = map[string]Format{
	".json": JSON,
	".ini":  INI,
	".env":  ENV,
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

	// timerWatchers tracks the waching of non-inotify instances.
	timerWatchers map[string]*time.Timer

	// watchInterval is used to choose the time interval to watch changes when
	// using Read method.
	watchInterval time.Duration

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
		config:        make(map[string]Value),
		hook:          make(map[string]func(Event)),
		timerWatchers: make(map[string]*time.Timer),
		watchInterval: 5 * time.Minute,
		lock:          &xylock.RWLock{},
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

	for k, w := range c.timerWatchers {
		w.Stop()
		delete(c.timerWatchers, k)
	}

	return err
}

// SetWatchInterval sets the time interval to watch the change when using Read()
// method.
func (c *Config) SetWatchInterval(d time.Duration) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.watchInterval = d
}

// UnWatch removes a filename from the watcher. This method also works with s3
// url. Put "env" as parameter if you want to stop watching environment
// variables of LoadEnv().
func (c *Config) UnWatch(filename string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if w, ok := c.timerWatchers[filename]; ok {
		w.Stop()
		delete(c.timerWatchers, filename)
		return nil
	}

	if c.watcher != nil {
		if err := c.watcher.Remove(filename); err != nil {
			return ConfigError.New(err)
		}
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

// ReadENV reads the config values from a byte array under ENV format.
func (c *Config) ReadENV(b []byte) error {
	var envmap, err = godotenv.Unmarshal(string(b))
	if err != nil {
		return ConfigError.New(err)
	}

	for k, v := range envmap {
		c.Set(k, v, false)
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
	case ENV:
		return c.ReadENV(b)
	default:
		return FormatError.New("unsupported format")
	}
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
		return FormatError.Newf("unknown extension: %s", filename)
	}

	if watch {
		if err := c.watchFile(filename); err != nil {
			return err
		}
	}

	if data, err := ioutil.ReadFile(filename); err != nil {
		if !os.IsNotExist(err) || !watch {
			return ConfigError.New(err)
		}
	} else if err := c.ReadBytes(fileFormat, data); err != nil {
		return err
	}

	return nil
}

// ReadS3 reads a file from AWS S3 bucket and watch for their changes every
// duration. Set the duration as zero if no need to watch the change.
//
// You must provide the aws credentials in ~/.aws/credentials. The AWS_REGION
// is required.
func (c *Config) ReadS3(url string, d time.Duration) error {
	var fileFormat = UnknownFormat
	for ext, format := range extensions {
		if strings.HasSuffix(url, ext) {
			fileFormat = format
		}
	}

	if fileFormat == UnknownFormat {
		return FormatError.Newf("unknown extension: %s", url)
	}

	if !strings.HasPrefix(url, "s3://") {
		return FormatError.Newf("can not parse the s3 url %s", url)
	}

	var path = url[5:]
	var bucket, item, found = strings.Cut(path, "/")
	if !found {
		return FormatError.Newf("not found item in path %s", path)
	}

	var sess, err = session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		return ConfigError.New(err)
	}

	var downloader = s3manager.NewDownloader(sess)
	var buf = aws.NewWriteAtBuffer([]byte{})
	_, err = downloader.Download(
		buf,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(item),
		})

	if d != 0 {
		c.lock.Lock()
		c.timerWatchers[url] = time.AfterFunc(d, func() { c.ReadS3(url, d) })
		c.lock.Unlock()
	}

	if err != nil {
		if d == 0 {
			return ConfigError.New(err)
		}
		return nil
	}

	return c.ReadBytes(fileFormat, buf.Bytes())
}

// LoadEnv loads all environment variables and watch for their changes every
// duration. Set the duration as zero if no need to watch the change.
func (c *Config) LoadEnv(d time.Duration) error {
	var envs = os.Environ()
	for i := range envs {
		var key, value, found = strings.Cut(envs[i], "=")
		if !found {
			return FormatError.Newf("invalid environment variable %s", envs[i])
		}
		c.Set(key, value, false)
	}

	if d != 0 {
		c.lock.Lock()
		c.timerWatchers["env"] = time.AfterFunc(d, func() { c.LoadEnv(d) })
		c.lock.Unlock()
	}

	return nil
}

// Read reads the config with any instance. If the instance is s3 url or
// environment variable, the watchInterval is used to choose the time interval
// for watching changes. If the instance is file path, it will watch the change
// if watchInterval > 0.
func (c *Config) Read(path string) error {
	switch {
	case path == "env":
		return c.LoadEnv(c.watchInterval)
	case strings.HasPrefix(path, "s3://"):
		return c.ReadS3(path, c.watchInterval)
	default:
		if c.watchInterval > 0 {
			return c.ReadFile(path, true)
		}
		return c.ReadFile(path, false)
	}
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

// ToMap converts current config to map.
func (c *Config) ToMap() map[string]any {
	var result = make(map[string]any)
	for k, v := range c.config {
		if c, ok := v.AsConfig(); ok {
			result[k] = c.ToMap()
		} else {
			result[k] = v.value
		}
	}
	return result
}

// initWatcher assigns a new watcher to Config. It also run a goroutine for
// handling watcher events.
func (c *Config) initWatcher() error {
	var watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return ConfigError.New(err)
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

	// Create the file if it does not exist, so the watcher will not raise an
	// exception.
	var ferr error
	if _, ferr = os.Open(filename); os.IsNotExist(ferr) {
		var dir = filepath.Dir(filename)
		fmt.Println(dir)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return err
		}

		if _, err := os.OpenFile(filename, os.O_CREATE, 0666); err != nil {
			return err
		}
	}

	if err := c.watcher.Add(filename); err != nil {
		return ConfigError.New(err)
	}

	// Remove the file after adding to watcher.
	if os.IsNotExist(ferr) {
		if err := os.Remove(filename); err != nil {
			return err
		}
	}

	return nil
}
