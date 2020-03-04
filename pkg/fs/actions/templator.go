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
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	fs "confinit/pkg/fs"
	log "confinit/pkg/log"
	tfunc "confinit/pkg/tplfunctions"
)

type Templator struct {
	*Replicator
	Data    interface{}
	Env     map[string]string
	SkipExt bool
}

func NewTemplator(glob, dst string, force, skipext bool, excludes []string) (*Templator, error) {
	rpc, err := NewReplicator(glob, dst, fs.FsItemFile, force, excludes)
	if err != nil {
		return nil, err
	}
	env := make(map[string]string)
	for _, setting := range os.Environ() {
		pair := strings.SplitN(setting, "=", 2)
		env[pair[0]] = pair[1]
	}
	r := Templator{
		Replicator: rpc,
		Data:       nil,
		Env:        env,
		SkipExt:    skipext,
	}
	return &r, nil
}

func (ft *Templator) AddEnv(env map[string]string) {
	for key, value := range env {
		ft.Env[key] = value
	}
}

func (ft *Templator) AddData(data interface{}) (err error) {
	// convert to map[string]interface{} if
	// input is map[interface{}]interface{}
	switch y := data.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range y {
			switch k2 := k.(type) {
			case string:
				m[k2] = v
			default:
				m[fmt.Sprint(k)] = v
			}
		}
		data = m
	}
	if ft.Data == nil {
		// non initialized
		ft.Data = data
		return
	}
	switch ft.Data.(type) {
	case []interface{}:
		// both are list, append
		ft.Data = append(ft.Data.([]interface{}), data)
	case map[string]interface{}:
		switch y := data.(type) {
		case map[string]interface{}:
			// both are maps
			for k, v := range y {
				ft.Data.(map[string]interface{})[k] = v
			}
		default:
			// current ft.Data is a map
			// and new data is a not a map
			err = fmt.Errorf("Cannot add/mix Data source type Map with other Data source(s)")
		}
	default:
		err = fmt.Errorf("Cannot add/mix Data sources with different types")
	}
	return
}

type TemplateData struct {
	IsDir           bool
	Mode            string
	SourceBaseDir   string
	Source          string
	Filename        string
	SourceFile      string
	Path            string
	SourceFullPath  string
	SourcePath      string
	Ext             string
	DstBaseDir      string
	Destination     string
	DestinationPath string
	Data            interface{}
	Env             map[string]string
}

func (ft *Templator) NewTemplateData(basedir, f string, i os.FileMode) *TemplateData {
	dstf := f
	if ft.SkipExt && !i.IsDir() {
		dstf = strings.TrimSuffix(f, filepath.Ext(f))
	}
	fullpath := filepath.Join(basedir, f)
	dstpath := filepath.Join(ft.DstPath, dstf)
	data := TemplateData{
		IsDir:           i.IsDir(),
		Mode:            i.String(),
		SourceFile:      f,
		SourceBaseDir:   basedir,
		Source:          filepath.Base(f),
		Filename:        filepath.Base(dstf),
		Ext:             filepath.Ext(filepath.Base(dstf)),
		SourceFullPath:  fullpath,
		SourcePath:      filepath.Dir(fullpath),
		DstBaseDir:      ft.DstPath,
		Destination:     dstpath,
		DestinationPath: filepath.Dir(dstpath),
		Env:             ft.Env,
		Data:            ft.Data,
	}
	return &data
}

func (ft *Templator) renderTemplateString(name, value string, data *TemplateData) (string, error) {
	// A Buffer needs no initialization.
	var render bytes.Buffer
	tpl, err := template.New(name).Funcs(tfunc.TemplateFuncMap()).Parse(value)
	if err != nil {
		return "", err
	}
	if err := tpl.Execute(&render, data); err != nil {
		return "", err
	}
	return render.String(), nil
}

func (ft *Templator) renderTemplate(data *TemplateData, dirmode, filemode os.FileMode) error {
	if err := ft.mkdir(filepath.Dir(data.Destination), dirmode); err != nil {
		return err
	}
	tpl, err := template.New(data.Source).Funcs(tfunc.TemplateFuncMap()).ParseFiles(data.SourceFullPath)
	if err != nil {
		return err
	}
	if ft.FileMode != 0 {
		filemode = ft.FileMode
	}
	dst, err := os.OpenFile(data.Destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, filemode)
	if err != nil {
		err = fmt.Errorf("Cannot create file %s, %s", data.Destination, err)
		return err
	}
	defer dst.Close()
	if err := tpl.Execute(dst, data); err != nil {
		return err
	}
	log.Debugf("Successfully rendered template '%s' to '%s'", data.SourceFullPath, data.Destination)
	return nil
}

func (ft *Templator) Function(base string, path string, i os.FileMode) (dst string, err error) {
	if i.IsDir() {
		// Using always default mode (is not replicate)
		err = ft.mkdir(filepath.Join(ft.DstPath, path), i)
	} else {
		tpldata := ft.NewTemplateData(base, path, i)
		dst = tpldata.Destination
		err = ft.renderTemplate(tpldata, os.FileMode(0755), i)
		if err == nil {
			err = ft.applyPermissions(dst)
		}
	}
	return
}
