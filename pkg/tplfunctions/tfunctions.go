// Copyright (c) 2020 Jose Riguera <jriguera@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Most of these functions were copied and adapted from
// https://github.com/Masterminds/sprig
// in order to limit modules and extra dependencies
//
// Sprig
// Copyright (C) 2013 Masterminds
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
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
//
package tplfunctions

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"path"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"

	"github.com/google/uuid"
	"gopkg.in/yaml.v2"
)

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}
	return b, nil
}

// TemplateFuncMap returns a 'text/template'.FuncMap
func TemplateFuncMap() template.FuncMap {
	mapping := make(map[string]interface{}, len(templateFunctionsMapping))
	for k, v := range templateFunctionsMapping {
		mapping[k] = v
	}
	return template.FuncMap(mapping)
}

var templateFunctionsMapping = map[string]interface{}{
	// OS
	"env":       os.Getenv,
	"expandenv": os.ExpandEnv,
	// File Paths
	"base":  path.Base,
	"dir":   path.Dir,
	"clean": path.Clean,
	"ext":   path.Ext,
	"isAbs": path.IsAbs,
	// Date functions
	"date":        Date,
	"now":         Now,
	"dateConvert": DateConvert,
	"epoch":       Epoch,
	// Flow Control
	"loop":    Loop,
	"fail":    Fail,
	"ternary": Ternary,
	// Random, crypto
	"random":       GenerateRandom,
	"randomString": GenerateRandomString,
	"sha1sum":      Sha1sum,
	"sha256sum":    Sha256sum,
	"b64enc":       Base64Encode,
	"b64dec":       Base64Decode,
	"uuid":         UUID,
	// Strings
	"trim":       strings.TrimSpace,
	"upper":      strings.ToUpper,
	"lower":      strings.ToLower,
	"title":      strings.Title,
	"trimChars":  TrimChars,
	"trimSuffix": TrimSuffix,
	"trimPrefix": TrimPrefix,
	"contains":   Contains,
	"hasPrefix":  HasPrefix,
	"hasSuffix":  HasSuffix,
	"indent":     Indent,
	"quote":      Quote,
	"squote":     SQuote,
	"replace":    Replace,
	"toString":   ToString,
	"toBool":     ToBool,
	"ToFloat":    ToBool,
	"ToInt":      ToBool,
	// Math
	"add":   Add,
	"sub":   Sub,
	"div":   Div,
	"mod":   Mod,
	"mul":   Mul,
	"max":   Max,
	"min":   Min,
	"ceil":  Ceil,
	"floor": Floor,
	"round": Round,
	// Regex
	"regexMatch":      RegexMatch,
	"regexFindAll":    RegexFindAll,
	"regexReplaceAll": RegexReplaceAll,
	"regexSplit":      RegexSplit,
	// Strings and lists
	"split": Split,
	"join":  Join,
	"sort":  Sort,
	// JSON and YAML
	"toYAML": ToYAML,
	"toJSON": ToJSON,
	// Data Structures
	"list":    List,
	"last":    Last,
	"first":   First,
	"reverse": Reverse,
	"uniq":    Uniq,
	"has":     Has,
	"concat":  Concat,
	"append":  Append,
	"dict":    Dict,
	"get":     Get,
	"set":     Set,
	"unset":   Unset,
	"hasKey":  HasKey,
	"keys":    Keys,
	"values":  Values,
}

// Return current date/time. Use this in conjunction with other date functions.
func Now() time.Time {
	return time.Now()
}

// Given a format and a date, format the date string.
// Date can be a `time.Time` or an `int, int32, int64`.
// In the later case, it is treated as seconds since UNIX
// epoch.
func Date(fmt string, date interface{}) string {
	t := time.Now()
	switch date := date.(type) {
	case time.Time:
		t = date
	case *time.Time:
		t = *date
	case int64:
		t = time.Unix(date, 0)
	case int:
		t = time.Unix(int64(date), 0)
	case int32:
		t = time.Unix(int64(date), 0)
	}
	loc, err := time.LoadLocation("Local")
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}
	return t.In(loc).Format(fmt)
}

// Returns the seconds since the unix epoch for a time.Time.\
func Epoch(date time.Time) string {
	return strconv.FormatInt(date.Unix(), 10)
}

// DateConvert converts a string to a date.
// The first argument is the date layout and the second the date string.
// It will return an error in case the string cannot be converted.
func DateConvert(fmt, str string) (time.Time, error) {
	return time.ParseInLocation(fmt, str, time.Local)
}

// GenerateRandomString returns a securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
// letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz-"
func GenerateRandom(letters string, n int) (string, error) {
	bytes, err := generateRandomBytes(n)
	if err != nil {
		return "", err
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes), nil
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomString(n int) (string, error) {
	b, err := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// receives a string, and computes it’s SHA256 digest.
func Sha256sum(input string) string {
	hash := sha256.Sum256([]byte(input))
	return hex.EncodeToString(hash[:])
}

// function receives a string, and computes it’s SHA1 digest.
func Sha1sum(input string) string {
	hash := sha1.Sum([]byte(input))
	return hex.EncodeToString(hash[:])
}

// uuid provides a safe and secure UUID v4 implementation
func UUID() string {
	return uuid.New().String()
}

// Encode with Base64
func Base64Encode(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

// Encode with Base64
func Base64Decode(v string) string {
	data, err := base64.StdEncoding.DecodeString(v)
	if err != nil {
		return err.Error()
	}
	return string(data)
}

// Remove given characters from the front or back of a string
func TrimChars(a, b string) string {
	return strings.Trim(b, a)
}

// Trim just the suffix from a string
func TrimSuffix(a, b string) string {
	return strings.TrimSuffix(b, a)
}

// Trim just the prefix from a string
func TrimPrefix(a, b string) string {
	return strings.TrimPrefix(b, a)
}

// Test to see if one string is contained inside of another
func Contains(substr string, str string) bool {
	return strings.Contains(str, substr)
}

// Test whether a string has a given prefix
func HasPrefix(substr string, str string) bool {
	return strings.HasPrefix(str, substr)
}

// Test whether a string has a given suffix
func HasSuffix(substr string, str string) bool {
	return strings.HasSuffix(str, substr)
}

// indents every line in a given string to the specified indent width.
func Indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return pad + strings.Replace(v, "\n", "\n"+pad, -1)
}

// Wraps the input with double quotes
func Quote(str ...interface{}) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("%q", ToString(s)))
		}
	}
	return strings.Join(out, " ")
}

// Wraps the input wiht sinble quote
func SQuote(str ...interface{}) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("'%v'", s))
		}
	}
	return strings.Join(out, " ")
}

// Perform simple string replacement. It takes three arguments:
// string to replace
// string to replace with
// source string
func Replace(old, new, src string) string {
	return strings.Replace(src, old, new, -1)
}

// Convert to a string
func ToString(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

// Control

// Loop generates a list of counting integers, defining a start, stop, and step:
// range $i, $e := until 0 5 1 generates "[0, 1, 2, 3, 4]"
func Loop(start, stop, step int) []int {
	v := []int{}
	if stop < start {
		if step >= 0 {
			return v
		}
		for i := start; i > stop; i += step {
			v = append(v, i)
		}
		return v
	}
	if step <= 0 {
		return v
	}
	for i := start; i < stop; i += step {
		v = append(v, i)
	}
	return v
}

// Returns the first value if the last value is true, otherwise returns the second value.
func Ternary(vt interface{}, vf interface{}, v bool) interface{} {
	if v {
		return vt
	}
	return vf
}

// Returns an error message
func Fail(msg string) (string, error) {
	return "", errors.New(msg)
}

// Math

// From html/template/content.go
// Copyright 2011 The Go Authors. All rights reserved.
// indirect returns the value, after dereferencing as many times
// as necessary to reach the base type (or nil).
func indirect(a interface{}) interface{} {
	if a == nil {
		return nil
	}
	if t := reflect.TypeOf(a); t.Kind() != reflect.Ptr {
		// Avoid creating a reflect.Value if it's not a pointer.
		return a
	}
	v := reflect.ValueOf(a)
	for v.Kind() == reflect.Ptr && !v.IsNil() {
		v = v.Elem()
	}
	return v.Interface()
}

// ToBool parses a string into a boolean
func ToBool(i interface{}) (bool, error) {
	i = indirect(i)
	switch b := i.(type) {
	case bool:
		return b, nil
	case nil:
		return false, nil
	case int:
		if i.(int) != 0 {
			return true, nil
		}
		return false, nil
	case string:
		return strconv.ParseBool(i.(string))
	default:
		return false, fmt.Errorf("ToBool, unable to cast %#v of type %T to bool", i, i)
	}
}

// ToFloat parses a string into a base 10 float
func ToFloat(i interface{}) (float64, error) {
	i = indirect(i)
	switch s := i.(type) {
	case float64:
		return s, nil
	case float32:
		return float64(s), nil
	case int:
		return float64(s), nil
	case int64:
		return float64(s), nil
	case int32:
		return float64(s), nil
	case int16:
		return float64(s), nil
	case int8:
		return float64(s), nil
	case uint:
		return float64(s), nil
	case uint64:
		return float64(s), nil
	case uint32:
		return float64(s), nil
	case uint16:
		return float64(s), nil
	case uint8:
		return float64(s), nil
	case string:
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return v, nil
		}
		return 0, fmt.Errorf("ToFloat, unable to cast %#v of type %T to float64", i, i)
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("ToFloat, unable to cast %#v of type %T to float64", i, i)
	}
}

// ToInt parses a string into a base 10 int
func ToInt(i interface{}) (int64, error) {
	i = indirect(i)
	switch s := i.(type) {
	case int:
		return int64(s), nil
	case int64:
		return s, nil
	case int32:
		return int64(s), nil
	case int16:
		return int64(s), nil
	case int8:
		return int64(s), nil
	case uint:
		return int64(s), nil
	case uint64:
		return int64(s), nil
	case uint32:
		return int64(s), nil
	case uint16:
		return int64(s), nil
	case uint8:
		return int64(s), nil
	case float64:
		return int64(s), nil
	case float32:
		return int64(s), nil
	case string:
		v, err := strconv.ParseInt(s, 0, 0)
		if err == nil {
			return v, nil
		}
		return 0, fmt.Errorf("ToInt, unable to cast %#v of type %T to int64", i, i)
	case bool:
		if s {
			return 1, nil
		}
		return 0, nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("ToInt, unable to cast %#v of type %T to int64", i, i)
	}
}

// Add returns the sum of a and b
func Add(a, b interface{}) (r int64, err error) {
	r = 0
	if aa, err := ToInt(a); err == nil {
		if bb, err := ToInt(b); err == nil {
			r = aa + bb
		}
	}
	return
}

// Sub returns the difference of b from a
func Sub(a, b interface{}) (r int64, err error) {
	r = 0
	if aa, err := ToInt(a); err == nil {
		if bb, err := ToInt(b); err == nil {
			r = aa - bb
		}
	}
	return
}

// Div returns the division of b from a
func Div(a, b interface{}) (r int64, err error) {
	r = 0
	if aa, err := ToInt(a); err == nil {
		if bb, err := ToInt(b); err == nil {
			if bb == 0 {
				return 0, fmt.Errorf("Div, division by zero (%v / %v)", aa, bb)
			}
			r = aa / bb
		}
	}
	return
}

// Mod is Modulo of a % b
func Mod(a, b interface{}) (r int64, err error) {
	r = 0
	if aa, err := ToInt(a); err == nil {
		if bb, err := ToInt(b); err == nil {
			r = aa % bb
		}
	}
	return
}

// Mul returns the product of a and b
func Mul(a, b interface{}) (r int64, err error) {
	r = 0
	if aa, err := ToInt(a); err == nil {
		if bb, err := ToInt(b); err == nil {
			r = aa * bb
		}
	}
	return
}

// Max returns the largest of a series of integers
func Max(a interface{}, i ...interface{}) (r int64, err error) {
	r, err = ToInt(a)
	if err == nil {
		for _, b := range i {
			bb, err := ToInt(b)
			if err != nil {
				return r, err
			}
			if bb > r {
				r = bb
			}
		}
	}
	return r, err
}

// Min returns the smallest of a series of integers
func Min(a interface{}, i ...interface{}) (r int64, err error) {
	r, err = ToInt(a)
	if err == nil {
		for _, b := range i {
			bb, err := ToInt(b)
			if err != nil {
				return r, err
			}
			if bb < r {
				r = bb
			}
		}
	}
	return r, err
}

// Floor returns the greatest float value less than or equal to input value
func Floor(a interface{}) (float64, error) {
	aa, err := ToFloat(a)
	if err == nil {
		return math.Floor(aa), nil
	}
	return 0, err
}

// Ceil returns the greatest float value greater than or equal to input value
func Ceil(a interface{}) (float64, error) {
	aa, err := ToFloat(a)
	if err == nil {
		return math.Ceil(aa), nil
	}
	return 0, err
}

// Round returns a float value with the remainder rounded to the given number to digits after the decimal point
func Round(a interface{}, p int, rOpt ...float64) (float64, error) {
	roundOn := .5
	if len(rOpt) > 0 {
		roundOn = rOpt[0]
	}
	val, err := ToFloat(a)
	if err != nil {
		return 0, err
	}
	places, _ := ToFloat(p)
	round := float64(0)
	pow := math.Pow(10, places)
	digit := pow * val
	_, div := math.Modf(digit)
	if div >= roundOn {
		round = math.Ceil(digit)
	} else {
		round = math.Floor(digit)
	}
	return round / pow, nil
}

// ToYAML converts the given structure into a deeply nested YAML string.
func ToYAML(m map[string]interface{}) (string, error) {
	result, err := yaml.Marshal(m)
	if err != nil {
		return "", fmt.Errorf("ToYAML: %s", err)
	}
	return string(bytes.TrimSpace(result)), nil
}

// ToJSON converts the given structure into a pretty JSON string.
func ToJSON(v interface{}) (string, error) {
	output, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func strslice(v interface{}) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))
		for _, s := range v {
			if s != nil {
				b = append(b, ToString(s))
			}
		}
		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)
			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, ToString(value))
				}
			}
			return b
		default:
			if v == nil {
				return []string{}
			}
			return []string{ToString(v)}
		}
	}
}

// Join a list of strings into a single string, with the given separator.
func Join(sep string, v interface{}) string {
	return strings.Join(strslice(v), sep)
}

// Split a string into a list of strings
func Split(sep, orig string) []string {
	return strings.Split(orig, sep)
}

// Returns a copy of the list in alphanum order
func Sort(list interface{}) []string {
	k := reflect.Indirect(reflect.ValueOf(list)).Kind()
	switch k {
	case reflect.Slice, reflect.Array:
		a := strslice(list)
		s := sort.StringSlice(a)
		s.Sort()
		return s
	}
	return []string{ToString(list)}
}

// Returns true if the input string contains any match of the regular expression
func RegexMatch(regex string, s string) (bool, error) {
	return regexp.MatchString(regex, s)
}

// Returns a slice of all matches of the regular expression in the input string.
// The last parameter n determines the number of substrings to return,
// where -1 means return all matches
func RegexFindAll(regex string, s string, n int) ([]string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return []string{}, err
	}
	return r.FindAllString(s, n), nil
}

// Returns a copy of the input string, replacing matches of the Regexp with the
// replacement string replacement. Inside string replacement, $ signs are
// interpreted as in Expand, so for instance $1 represents the text of the first submatch
func RegexReplaceAll(regex string, s string, repl string) (string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return "", err
	}
	return r.ReplaceAllString(s, repl), nil
}

// Slices the input string into substrings separated by the expression and
// returns a slice of the substrings between those expression matches.
// The last parameter n determines the number of substrings to return,
// where -1 means return all matches
func RegexSplit(regex string, s string, n int) ([]string, error) {
	r, err := regexp.Compile(regex)
	if err != nil {
		return []string{}, err
	}
	return r.Split(s, n), nil
}

// List type similar to arrays or slices, but lists are designed to be used
// as immutable data types
func List(v ...interface{}) []interface{} {
	return v
}

// Gets the last item on a list
func Last(list interface{}) (interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)
		l := l2.Len()
		if l == 0 {
			return nil, nil
		}
		return l2.Index(l - 1).Interface(), nil
	default:
		return nil, fmt.Errorf("Cannot find last on type %s", tp)
	}
}

// Gets the first item on a list
func First(list interface{}) (interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)
		l := l2.Len()
		if l == 0 {
			return nil, nil
		}
		return l2.Index(0).Interface(), nil
	default:
		return nil, fmt.Errorf("Cannot find first on type %s", tp)
	}
}

// Produce a new list with the reversed elements of the given list
func Reverse(v interface{}) ([]interface{}, error) {
	tp := reflect.TypeOf(v).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(v)
		l := l2.Len()
		// We do not sort in place because the incoming array should not be altered.
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[l-i-1] = l2.Index(i).Interface()
		}
		return nl, nil
	default:
		return nil, fmt.Errorf("Cannot find reverse on type %s", tp)
	}
}

// Generate a list with all of the duplicates removed
func Uniq(list interface{}) ([]interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)
		l := l2.Len()
		dest := []interface{}{}
		var item interface{}
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if has, err := Has(dest, item); err != nil {
				if !has {
					dest = append(dest, item)
				}
			} else {
				return nil, err
			}
		}
		return dest, nil
	default:
		return nil, fmt.Errorf("Cannot find uniq on type %s", tp)
	}
}

// Test to check if a list has a particular element.
func Has(needle interface{}, haystack interface{}) (bool, error) {
	if haystack == nil {
		return false, nil
	}
	tp := reflect.TypeOf(haystack).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(haystack)
		var item interface{}
		l := l2.Len()
		for i := 0; i < l; i++ {
			item = l2.Index(i).Interface()
			if reflect.DeepEqual(needle, item) {
				return true, nil
			}
		}
		return false, nil
	default:
		return false, fmt.Errorf("Cannot find Has on type %s", tp)
	}
}

// Concatenate arbitrary number of lists into one
func Concat(lists ...interface{}) interface{} {
	var res []interface{}
	for _, list := range lists {
		tp := reflect.TypeOf(list).Kind()
		switch tp {
		case reflect.Slice, reflect.Array:
			l2 := reflect.ValueOf(list)
			for i := 0; i < l2.Len(); i++ {
				res = append(res, l2.Index(i).Interface())
			}
		default:
			panic(fmt.Sprintf("Cannot concat type %s as list", tp))
		}
	}
	return res
}

// Append a new item to an existing list, creating a new list
func Append(list interface{}, v interface{}) ([]interface{}, error) {
	tp := reflect.TypeOf(list).Kind()
	switch tp {
	case reflect.Slice, reflect.Array:
		l2 := reflect.ValueOf(list)

		l := l2.Len()
		nl := make([]interface{}, l)
		for i := 0; i < l; i++ {
			nl[i] = l2.Index(i).Interface()
		}

		return append(nl, v), nil

	default:
		return nil, fmt.Errorf("Cannot push on type %s", tp)
	}
}

// Dict creates a map by passing it a list of pairs.
// $myDict := dict "name1" "value1" "name2" "value2" "name3" "value 3"
func Dict(v ...interface{}) map[string]interface{} {
	dict := map[string]interface{}{}
	lenv := len(v)
	for i := 0; i < lenv; i += 2 {
		key := ToString(v[i])
		if i+1 >= lenv {
			dict[key] = ""
			continue
		}
		dict[key] = v[i+1]
	}
	return dict
}

// Given a map and a key, get the value from the map
func Get(d map[string]interface{}, key string) interface{} {
	if val, ok := d[key]; ok {
		return val
	}
	return ""
}

// Add/set a new key/value pair to a dictionary
func Set(d map[string]interface{}, key string, value interface{}) map[string]interface{} {
	d[key] = value
	return d
}

// Delete the key from the map
func Unset(d map[string]interface{}, key string) map[string]interface{} {
	delete(d, key)
	return d
}

// Returns true if the given dict contains the given key
func HasKey(d map[string]interface{}, key string) bool {
	_, ok := d[key]
	return ok
}

// Returns a list of all of the keys in one or more dict types
// Order is not deterministic.
func Keys(dicts ...map[string]interface{}) []string {
	k := []string{}
	for _, dict := range dicts {
		for key := range dict {
			k = append(k, key)
		}
	}
	return k
}

// Returns a list of all of the values or one dict type
// Order is not deterministic.
func Values(dict map[string]interface{}) []interface{} {
	values := []interface{}{}
	for _, value := range dict {
		values = append(values, value)
	}
	return values
}
