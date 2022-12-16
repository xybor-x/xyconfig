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
	"fmt"
	"strconv"
	"strings"
)

// Value instance represents for values in Config.
type Value struct {
	value  any
	strict bool
}

// IsNil return true if value is nil.
func (v Value) IsNil() bool {
	return v.value == nil || (v.value == "" && !v.strict)
}

// AsConfig returns the value as *Config. The latter return value is false if
// failed to cast.
func (v Value) AsConfig() (*Config, bool) {
	var c, ok = v.value.(*Config)
	return c, ok
}

// MustConfig returns the value as *Config. It panics if failed to cast.
func (v Value) MustConfig() *Config {
	var c, ok = v.AsConfig()
	if !ok {
		panic(CastError.Newf("got a %T, not *Config", v.value))
	}
	return c
}

// AsInt returns the value as int. The latter return value is false if failed to
// cast.
func (v Value) AsInt() (int, bool) {
	if s, ok := v.value.(string); ok && !v.strict {
		var i, err = strconv.Atoi(s)
		if err != nil {
			return 0, false
		}
		return i, true
	}

	var i, ok = v.value.(int)
	return i, ok
}

// MustInt returns the value as int. It panics if failed to cast.
func (v Value) MustInt() int {
	var i, ok = v.AsInt()
	if !ok {
		panic(CastError.Newf("got a %T, not int", v.value))
	}
	return i
}

// AsFloat returns the value as float64. The latter return value is false if
// failed to cast.
func (v Value) AsFloat() (float64, bool) {
	if s, ok := v.value.(string); ok && !v.strict {
		var f, err = strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		return f, true
	}

	if i, ok := v.value.(int); ok {
		return float64(i), ok
	}

	var f, ok = v.value.(float64)
	return f, ok
}

// MustFloat returns the value as float64. It panics if failed to cast.
func (v Value) MustFloat() float64 {
	var f, ok = v.AsFloat()
	if !ok {
		panic(CastError.Newf("got a %T, not float64", v.value))
	}
	return f
}

// AsString returns the value as string. The latter return value is false if
// failed to cast.
func (v Value) AsString() (string, bool) {
	var s, ok = v.value.(string)
	if !ok && !v.strict {
		return fmt.Sprint(v.value), true
	}
	return s, ok
}

// MustString returns the value as string. It panics if failed to cast.
func (v Value) MustString() string {
	var s, ok = v.AsString()
	if !ok {
		panic(CastError.Newf("got a %T, not string", v.value))
	}
	return s
}

// AsArray returns the value as array. The latter return value is false if
// failed to cast.
func (v Value) AsArray() ([]Value, bool) {
	if s, ok := v.value.(string); ok && !v.strict {
		s = strings.Trim(s, " ")
		if s[0] != '[' || s[len(s)-1] != ']' {
			return nil, false
		}
		s = s[1 : len(s)-1]
		var elements []Value
		for _, e := range strings.Split(s, ",") {
			e = strings.Trim(e, " ")
			elements = append(elements, Value{value: e, strict: false})
		}

		return elements, true
	}

	if a, ok := v.value.([]any); ok {
		var elements []Value
		for _, e := range a {
			elements = append(elements, Value{value: e, strict: true})
		}
		return elements, true
	}

	return nil, false
}

// MustArray returns the value as array. It panics if failed to cast.
func (v Value) MustArray() []Value {
	var a, ok = v.AsArray()
	if !ok {
		panic(CastError.Newf("got a %T, not array", v.value))
	}
	return a
}

// String returns the string representation of Value.
func (v Value) String() string {
	return fmt.Sprint(v.value)
}
