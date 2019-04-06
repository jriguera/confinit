package config

const (
	ConfigType string = "yaml"
	ConfigFile string = "config.yml"
	ConfigPath string = "/etc/confinit"
	ConfigEnv  string = "CONFINIT"
)

type Permissions struct {
	Glob  string `mapstructure:"glob" valid:"glob,configuration" default:"*"`
	Mode  string `mapstructure:"mode" default:"0"`
	User  string `mapstructure:"user" valid:"user"`
	Group string `mapstructure:"group" valid:"group"`
}

type DefaultMode struct {
	Dir  string `mapstructure:"folder" valid:"mode" default:"0"`
	File string `mapstructure:"file" valid:"mode" default:"0"`
}

type Default struct {
	Mode  DefaultMode `mapstructure:"mode" valid:"required"`
	Force *bool       `mapstructure:"force" default:"true"`
}

type Operation struct {
	DestinationPath string                 `mapstructure:"destination" valid:"configuration"`
	Default         Default                `mapstructure:"default"`
	Perms           []*Permissions         `mapstructure:"permissions"`
	Regex           string                 `mapstructure:"regex" default:".*"`
	Data            map[string]interface{} `mapstructure:"data"`
	Template        *bool                  `mapstructure:"template" default:"true"`
	DelExtension    *bool                  `mapstructure:"delextension" default:"true"`
	RenderCondition string                 `mapstructure:"condition"`
	PreDelete       *bool                  `mapstructure:"predelete" default:"true"`
	Command         *Runner                `mapstructure:"command" valid:"-"`
}

type MatchItem struct {
	Add  string `mapstructure:"add" valid:"glob" default:"*"`
	Skip string `mapstructure:"skip" valid:"glob"`
}

type Match struct {
	Folder MatchItem `mapstructure:"folder"`
	File   MatchItem `mapstructure:"file"`
}

type Process struct {
	Source      string       `mapstructure:"source" valid:"required"`
	Match       Match        `mapstructure:"match" valid:"required"`
	ExcludeDone *bool        `mapstructure:"excludedone" default:"true"`
	Operations  []*Operation `mapstructure:"operations" valid:"required"`
}

type Runner struct {
	Cmd     []string          `mapstructure:"cmd"`
	Timeout int               `mapstructure:"timeout" default:"300"`
	Env     map[string]string `mapstructure:"env"`
	Dir     string            `mapstructure:"dir"`
}

// Config the application's configuration
type Config struct {
	LogOutput string            `mapstructure:"logoutput" valid:"required" default:"split"`
	LogLevel  string            `mapstructure:"loglevel" valid:"in(debug|info|warn|error|panic|fatal),required" default:"info" flag:"program log level"`
	Env       map[string]string `mapstructure:"env"`
	DataFile  string            `mapstructure:"datafile" flag:"file for global template data key/values"`
	Start     *Runner           `mapstructure:"start"`
	Finish    *Runner           `mapstructure:"finish"`
	Process   []Process         `mapstructure:"process"`
}
