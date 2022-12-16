package benchmark

import (
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"testing"

	"github.com/xybor-x/xyconfig"
	"github.com/xybor-x/xylog"
)

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type ConfigFile struct {
	name    string
	cfg     map[string]any
	keys    []string
	cfgJSON []byte
	i       int
}

func (c *ConfigFile) New(n int) {
	c.i = 0
	for i := 0; i < n; i++ {
		c.cfg = c.createMap(n)
	}

	var err error
	c.cfgJSON, err = json.Marshal(c.cfg)

	if err != nil {
		panic(err)
	}
}

func (c *ConfigFile) Reset() {
	c.i = 0
}

func (c *ConfigFile) NextKey() string {
	c.i = (c.i + 1) % len(c.keys)
	return c.keys[c.i]
}

func (c *ConfigFile) Write() {
	if err := ioutil.WriteFile(c.name, c.cfgJSON, 0644); err != nil {
		panic(err)
	}
}

func (c *ConfigFile) createMap(n int) map[string]any {
	var m = make(map[string]any)
	for i := 0; i < n; i++ {
		var k = createRandomString()
		switch rand.Intn(4) {
		case 0:
			m[k] = rand.Int()
		case 1:
			m[k] = rand.Float64()
		case 2:
			m[k] = createRandomString()
		case 3:
			m[k] = c.createMap(n / 10)
		}
		c.keys = append(c.keys, k)
	}
	return m
}

func createRandomString() string {
	var b = make([]rune, rand.Intn(10)+rand.Intn(10))
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func init() {
	xylog.GetLogger("xybor.xyplatform.xyconfig").SetLevel(xylog.NOTLOG)
}

func BenchmarkGet(b *testing.B) {
	var file = &ConfigFile{name: b.Name() + ".json"}
	file.New(100)
	file.Write()

	var config = xyconfig.GetConfig(b.Name() + ".json")
	config.ReadFile(file.name, false)

	b.Run("Get", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			config.Get(file.NextKey())
		}
	})
}

func BenchmarkWatchChange(b *testing.B) {
	var files = make([]*ConfigFile, 0)
	var N = b.N
	for i := 0; i < N; i++ {
		var f = &ConfigFile{name: b.Name() + ".json"}
		files = append(files, f)
		f.New(100)
		f.Write()
	}

	var config = xyconfig.GetConfig(b.Name() + ".json")
	config.ReadFile(files[0].name, true)
	for _, k := range files[0].keys {
		config.AddHook(k, func(e xyconfig.Event) {})
	}

	b.Run("ChangeConfig", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			files[i%N].Write()
		}
	})

	config.CloseWatcher()

	b.Run("WriteFile", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			files[i%N].Write()
		}
	})
}
