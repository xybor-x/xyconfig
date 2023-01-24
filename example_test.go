package xyconfig_test

import (
	"fmt"

	"github.com/xybor-x/xyconfig"
)

func ExampleConfig() {
	// Name your config object for global usage.
	var config = xyconfig.GetConfig("example")

	// Read config from a map.
	config.ReadMap(0, map[string]any{
		"general": map[string]any{
			"timeout": 3.14,
		},
	})

	// It's also ok if you want to read from string, byte array, or file with
	// supported formats.
	config.ReadJSON(0, []byte(`{"system": "linux"}`))

	// Read the config from file.
	// config.ReadFile("foo.json", false)
	// Or read and watch the file change.
	// config.ReadFile("foo.json", true)

	// You can get a key-value pair by many ways.
	var timeout = config.MustGet("general.timeout").MustFloat()
	fmt.Println("general.timeout =", timeout)

	var general = config.MustGet("general").MustConfig()
	fmt.Println("[general].timeout =", general.MustGet("timeout").MustFloat())

	general = xyconfig.GetConfig("example.general")
	fmt.Println("{general}.timeout =", general.MustGet("timeout").MustFloat())

	// You also can check if the key exist and the value is expected type.
	var system, ok = config.Get("system")
	if !ok {
		fmt.Println("Key system doesn't exist")
	}

	systemVal, ok := system.AsString()
	if !ok {
		fmt.Println("Key system is not a stringn")
	}

	fmt.Println("system =", systemVal)

	// Output:
	// general.timeout = 3.14
	// [general].timeout = 3.14
	// {general}.timeout = 3.14
	// system = linux
}
