// Copyright © 2019 Jose Riguera <jriguera@gmail.com>
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
package fs

// Package from https://github.com/zyedidia/glob

import (
	"regexp"
)

// Glob is a wrapper of *regexp.Regexp.
// It should contain a glob expression compiled into a regular expression.
type Glob struct {
	*regexp.Regexp
	Pattern string
}

// NewGlob a takes a glob expression as a string and transforms it
// into a *Glob object (which is really just a regular expression)
// Compile also returns a possible error.
func NewGlob(pattern string) (*Glob, error) {
	r, err := globToRegex(pattern)
	return &Glob{
		Regexp:  r,
		Pattern: pattern,
	}, err
}

func (g *Glob) String() string {
	return g.Pattern
}

func (g *Glob) GetRegexp() string {
	return g.Regexp.String()
}

func globToRegex(glob string) (*regexp.Regexp, error) {
	regex := ""
	inGroup := 0
	inClass := 0
	firstIndexInClass := -1
	arr := []byte(glob)
	for i := 0; i < len(arr); i++ {
		ch := arr[i]
		switch ch {
		case '\\':
			i++
			if i >= len(arr) {
				regex += "\\"
			} else {
				next := arr[i]
				switch next {
				case ',':
					// Nothing
				case 'Q', 'E':
					regex += "\\\\"
				default:
					regex += "\\"
				}
				regex += string(next)
			}
		case '*':
			if inClass == 0 {
				regex += ".*"
			} else {
				regex += "*"
			}
		case '?':
			if inClass == 0 {
				regex += "."
			} else {
				regex += "?"
			}
		case '[':
			inClass++
			firstIndexInClass = i + 1
			regex += "["
		case ']':
			inClass--
			regex += "]"
		case '.', '(', ')', '+', '|', '^', '$', '@', '%':
			if inClass == 0 || (firstIndexInClass == i && ch == '^') {
				regex += "\\"
			}
			regex += string(ch)
		case '!':
			if firstIndexInClass == i {
				regex += "^"
			} else {
				regex += "!"
			}
		case '{':
			inGroup++
			regex += "("
		case '}':
			inGroup--
			regex += ")"
		case ',':
			if inGroup > 0 {
				regex += "|"
			} else {
				regex += ","
			}
		default:
			regex += string(ch)
		}
	}
	return regexp.Compile("^" + regex + "$")
}
