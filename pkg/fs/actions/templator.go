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
)

type Templator struct {
	*Replicator
	Data    map[string]string
	Env     map[string]string
	SkipExt bool
}

func NewTemplator(glob, dst string, force, skipext bool) (*Templator, error) {
	rpc, err := NewReplicator(glob, dst, fs.FsItemFile, force)
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
		Data:       make(map[string]string),
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

func (ft *Templator) AddData(data map[string]string) {
	for key, value := range data {
		ft.Data[key] = value
	}
}

type TemplateData struct {
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
	Data            map[string]string
	Env             map[string]string
}

func (ft *Templator) NewTemplateData(basedir, f string) *TemplateData {
	dstf := f
	if ft.SkipExt {
		dstf = strings.TrimSuffix(f, filepath.Ext(f))
	}
	fullpath := filepath.Join(basedir, f)
	dstpath := filepath.Join(ft.DstPath, dstf)
	data := TemplateData{
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
	tpl, err := template.New(name).Parse(value)
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
	tpl, err := template.ParseFiles(data.SourceFullPath)
	if err != nil {
		return err
	}
	if ft.FileMode != 0 {
		filemode = ft.FileMode
	}
	dst, err := os.OpenFile(data.Destination, os.O_RDWR|os.O_CREATE, filemode)
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

func (ft *Templator) Function(base string, path string, i os.FileInfo) (err error) {
	if i.IsDir() {
		// Using always default mode (is not replicate)
		err = ft.mkdir(filepath.Join(ft.DstPath, path), i.Mode())
	} else {
		tpldata := ft.NewTemplateData(base, path)
		err = ft.renderTemplate(tpldata, os.FileMode(0755), i.Mode())
		if err == nil {
			err = ft.applyPermissions(tpldata.Destination)
		}
	}
	return
}
