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
package actions

import (
	"os"
)

type ActionRouter struct {
	*Runner
}

func NewActionRouter(glob, dst string, force, skipext, render bool) (*ActionRouter, error) {
	r, err := NewRunner(glob, dst, force, skipext, render)
	if err != nil {
		return nil, err
	}
	a := ActionRouter{
		Runner: r,
	}
	return &a, nil
}

func (a *ActionRouter) Function(base string, path string, i os.FileInfo) (err error) {
	if a.Cmd != "" {
		err = a.Runner.Function(base, path, i)
	} else {
		if a.DstPath != "" {
			if a.Render {
				err = a.Templator.Function(base, path, i)
			} else {
				err = a.Replicator.Function(base, path, i)
			}
		}
	}
	return
}
