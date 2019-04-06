package program

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"confinit/internal/config"
	"confinit/pkg/fs"
	"confinit/pkg/fs/actions"
	"confinit/pkg/runner"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type Program struct {
	Build        string
	Config       *config.Config
	Data         map[string]interface{}
	ConfigArg    string
	Configurator config.Configurator
}

func NewProgram(build, version, configArg string, command *cobra.Command) *Program {
	p := Program{
		Build:        build,
		ConfigArg:    configArg,
		Configurator: config.NewConfigurator(version, configArg, command),
	}
	return &p
}

func (p *Program) Init() {
	p.Config = p.Configurator.InitConfig()
}

func (p *Program) LoadConfig() error {
	cfg, err := p.Configurator.LoadConfig(p.ConfigArg)
	if err == nil {
		log := p.Configurator.Logger()
		f := p.Configurator.GetConfigFile(false)
		log.Infof("Configuration loaded from file: %s", f)
		if errc := p.Configurator.CheckConfig(cfg); errc != nil {
			return errc
		}
		p.Config = cfg
		return p.LoadData()
	}
	return err
}

func (p *Program) LoadData() error {
	data := make(map[string]interface{})
	log := p.Configurator.Logger()
	if p.Config == nil {
		return nil
	}
	if p.Config.DataFile == "" {
		return nil
	}
	datafile := p.Config.DataFile
	if _, err := os.Stat(p.Config.DataFile); os.IsNotExist(err) {
		log.Errorf("Data file not loaded: %s", err)
		return err
	}
	log.Infof("Loading data from file: %s", datafile)
	dataloader := viper.New()
	basefile := filepath.Base(datafile)
	dataloader.SetConfigName(strings.TrimSuffix(basefile, filepath.Ext(basefile)))
	dataloader.AddConfigPath(filepath.Dir(datafile))
	dataloader.SetConfigType(filepath.Ext(basefile)[1:])
	if err := dataloader.ReadInConfig(); err != nil {
		log.Errorf("Error reading data file %s: %s", datafile, err)
		return err
	}
	if err := dataloader.Unmarshal(&data); err != nil {
		log.Fatalf("Format of data file not correct, %s", err.Error())
		return err
	}
	p.Data = data
	return nil
}

func (p *Program) GetJsonConfig() ([]byte, error) {
	cfg, err := p.Configurator.GetConfigMap(p.Config)
	if err == nil {
		return json.MarshalIndent(cfg, "", "  ")
	}
	return []byte{}, err
}

func (p *Program) RunAll() (err error) {
	// reset umask
	oldumask := syscall.Umask(0)
	defer syscall.Umask(oldumask)
	// set global env
	for key, value := range p.Config.Env {
		// viper bug: https://github.com/spf13/viper/issues/373
		os.Setenv(strings.ToUpper(key), value)
	}
	// program
	rcs := make(map[string]int)
	rcStart, errStart := p.RunStart()
	if rcStart >= 0 {
		rcs[fmt.Sprintf("%s_RC_START", config.ConfigEnv)] = rcStart
	}
	err = errStart
	if errStart == nil {
		rcP, errP := p.Process()
		rcs[fmt.Sprintf("%s_RC_PROCESS", config.ConfigEnv)] = rcP
		err = errP
	}
	_, errFinish := p.RunFinish(rcs)
	if errFinish != nil {
		if err == nil {
			err = errFinish
		}
	}
	return
}

func (p *Program) operation(f *fs.Fs, c *config.Operation, excludes []string) ([]string, error) {
	errs := false
	log := p.Configurator.Logger()
	a, err := actions.NewActionRouter(c.Regex, c.DestinationPath,
		*c.Default.Force, *c.DelExtension, *c.Template, *c.PreDelete, excludes)
	if err != nil {
		return nil, err
	}
	dirmode, _ := strconv.ParseUint(c.Default.Mode.Dir, 8, 32)
	filemode, _ := strconv.ParseUint(c.Default.Mode.File, 8, 32)
	a.SetDefaultModes(os.FileMode(dirmode), os.FileMode(filemode))
	a.AddData(p.Data) // Global
	a.AddData(c.Data) // The one on this operation
	a.SetCondition(c.RenderCondition)
	for i, pe := range c.Perms {
		mode, _ := strconv.ParseUint(pe.Mode, 8, 32)
		errp := a.SetPermissions(pe.Glob, pe.User, pe.Group, os.FileMode(mode))
		if errp != nil {
			log.Errorf("Skipping permissions #%d: %s", i, errp)
			errs = true
		}
	}
	if c.Command != nil && len(c.Command.Cmd) > 0 {
		proc := runner.NewRunner(p.Configurator.Logger())
		proc.Command(c.Command.Cmd)
		envOS := make(map[string]string)
		for _, e := range os.Environ() {
			pair := strings.SplitN(e, "=", 2)
			envOS[pair[0]] = pair[1]
		}
		a.SetRunner(proc, envOS, c.Command.Timeout, c.Command.Dir)
		envC := make(map[string]string)
		for key, value := range c.Command.Env {
			// viper bug: https://github.com/spf13/viper/issues/373
			envC[strings.ToUpper(key)] = value
		}
		a.AddEnv(envC)
	}
	err = f.Run(a)
	if errs && err == nil {
		err = fmt.Errorf("Not all permissions were applied!")
	}
	return a.ListProcessed(), err
}

func (p *Program) Process() (int, error) {
	log := p.Configurator.Logger()
	errs := []error{}
	processed := []string{}
	for i, proc := range p.Config.Process {
		f := fs.New(
			fs.SkipDirGlob(proc.Match.Folder.Skip),
			fs.SkipFileGlob(proc.Match.File.Skip),
			fs.FileGlob(proc.Match.File.Add),
			fs.DirGlob(proc.Match.Folder.Add),
		)
		log.Infof("Scanning #%d path: %s", i+1, proc.Source)
		if err := f.Scan(proc.Source); err != nil {
			errs = append(errs, fmt.Errorf("#%d %s: %s", i+1, proc.Source, err))
			log.Error(err)
		}
		for j, oper := range proc.Operations {
			log.Infof("Processing #%d operation in %s", j+1, proc.Source)
			done, err := p.operation(f, oper, processed)
			if err != nil {
				errs = append(errs, fmt.Errorf("#%d %s: %s", i+1, proc.Source, err))
			}
			if *proc.ExcludeDone {
				processed = append(processed, done...)
			}
		}
	}
	if len(errs) > 0 {
		msg := ""
		for _, e := range errs {
			msg += fmt.Sprintf("%s\n", e.Error())
		}
		return 1, fmt.Errorf("%s", msg)
	}
	return 0, nil
}

func (p *Program) RunStart() (int, error) {
	if p.Config.Start != nil && len(p.Config.Start.Cmd) > 0 {
		log := p.Configurator.Logger()
		log.Infof("Running startup program: %s", p.Config.Start.Cmd)
		return p.runner(p.Config.Start, os.Environ()).Run()
	}
	return -1, nil
}

func (p *Program) RunFinish(rc map[string]int) (int, error) {
	if p.Config.Finish != nil && len(p.Config.Finish.Cmd) > 0 {
		log := p.Configurator.Logger()
		env := os.Environ()
		log.Infof("Running finish program %s", p.Config.Finish.Cmd)
		for key, value := range rc {
			env = append(env, fmt.Sprintf("%s=%d", key, value))
		}
		return p.runner(p.Config.Finish, env).Run()
	}
	return -1, nil
}

func (p *Program) runner(r *config.Runner, osEnv []string) *runner.Runner {
	env := make(map[string]string)
	for _, e := range osEnv {
		pair := strings.SplitN(e, "=", 2)
		env[pair[0]] = pair[1]
	}
	for key, value := range r.Env {
		// viper bug: https://github.com/spf13/viper/issues/373
		env[strings.ToUpper(key)] = value
	}
	procRunner := runner.NewRunner(p.Configurator.Logger())
	procRunner.SetEnv(env)
	procRunner.SetTimeout(r.Timeout)
	procRunner.SetDir(r.Dir)
	procRunner.Command(r.Cmd)
	return procRunner
}
