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

import (
	"fmt"
	"os"
	"os/user"
	"strconv"
)

type Perm struct {
	User  int
	Group int
	Mode  os.FileMode
}

func NewPerm(uid, gid string, mode os.FileMode) (*Perm, error) {
	currentUser, err := user.Current()
	if err != nil {
		err = fmt.Errorf("Could not get current user, %s", err)
		return nil, err
	}
	userID, _ := strconv.Atoi(currentUser.Uid)
	groupID, _ := strconv.Atoi(currentUser.Gid)
	if uid != "" {
		u, erru := user.LookupId(uid)
		if erru != nil {
			u, erru = user.Lookup(uid)
			if erru != nil {
				erru = fmt.Errorf("Invalid %s", erru)
				return nil, erru
			}
		}
		userID, _ = strconv.Atoi(u.Uid)
	}
	if gid != "" {
		g, errg := user.LookupGroupId(gid)
		if errg != nil {
			g, errg = user.LookupGroup(gid)
			if errg != nil {
				errg = fmt.Errorf("Invalid %s", errg)
				return nil, errg
			}
		}
		groupID, _ = strconv.Atoi(g.Gid)
	}
	p := Perm{
		User:  userID,
		Group: groupID,
		Mode:  mode,
	}
	return &p, nil
}

func (p *Perm) String() string {
	return fmt.Sprintf("%d:%d %s", p.User, p.Group, p.Mode.String())
}

func (p *Perm) Set(fullp string) error {
	var errd, errn error
	if p.Mode != 0 {
		errd = os.Chmod(fullp, p.Mode)
		if errd != nil {
			errd = fmt.Errorf("Cannot set mode (%s) to '%s': %s", p.Mode.String(), fullp, errd)
			return errd
		}
	}
	errn = os.Chown(fullp, p.User, p.Group)
	if errn != nil {
		errn = fmt.Errorf("Cannot set owner (%d) and/or group (%d) to '%s': %s", p.User, p.Group, fullp, errn)
		return errn
	}
	return nil
}
