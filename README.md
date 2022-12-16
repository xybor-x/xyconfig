[![xybor founder](https://img.shields.io/badge/xybor-huykingsofm-red)](https://github.com/huykingsofm)
[![Go Reference](https://pkg.go.dev/badge/github.com/xybor-x/xyconfig.svg)](https://pkg.go.dev/github.com/xybor-x/xyconfig)
[![GitHub Repo stars](https://img.shields.io/github/stars/xybor-x/xyconfig?color=yellow)](https://github.com/xybor-x/xyconfig)
[![GitHub top language](https://img.shields.io/github/languages/top/xybor-x/xyconfig?color=lightblue)](https://go.dev/)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/xybor-x/xyconfig)](https://go.dev/blog/go1.18)
[![GitHub release (release name instead of tag name)](https://img.shields.io/github/v/release/xybor-x/xyconfig?include_prereleases)](https://github.com/xybor-x/xyconfig/releases/latest)
[![Codacy Badge](https://app.codacy.com/project/badge/Grade/b50c3a932d5c4b1484901234e411e4a5)](https://www.codacy.com/gh/xybor-x/xyconfig/dashboard?utm_source=github.com&utm_medium=referral&utm_content=xybor-x/xyconfig&utm_campaign=Badge_Grade)
[![Codacy Badge](https://app.codacy.com/project/badge/Coverage/b50c3a932d5c4b1484901234e411e4a5)](https://www.codacy.com/gh/xybor-x/xyconfig/dashboard?utm_source=github.com&utm_medium=referral&utm_content=xybor-x/xyconfig&utm_campaign=Badge_Grade)
[![Go Report](https://goreportcard.com/badge/github.com/xybor-x/xyconfig)](https://goreportcard.com/report/github.com/xybor-x/xyconfig)

# Introduction

Package xyconfig supports to thread-safe read, control, and monitor
configuration files.

# Get started

```golang
var config = xyconfig.GetConfig("app")

// Read config from a string.
config.Read(xyconfig.JSON, `{"general": {"timeout": 3.14}}`)

// Read config from default.ini but do not watch the file.
config.ReadFile("config/default.ini", false)

// Read config from override.ini and watch the file change.
config.ReadFile("config/override.yml", true)

fmt.Println(config.MustGet("general.timeout").MustFloat())

config.AddHook("general.timeout", func (e xyconfig.Event) {
    var timeout, ok = e.New.AsFloat()
    if !ok {
        return
    }
    SetTimeoutToSomeThing(timeout)
})

config.AddHook("general", func (e Event) {
    var general = e.New.MustConfig()
    var timeout = general.MustGet("timeout").MustFloat()
    SetTimeoutToSomething(timeout)
})
```

# Benchmark

| Operation           |         Time | Objects Allocated |
| :------------------ | -----------: | ----------------: |
| Get                 |     70 ns/op |       0 allocs/op |
| ChangeConfig        | 413763 ns/op |     724 allocs/op |
| WriteFile           | 335892 ns/op |       3 allocs/op |
