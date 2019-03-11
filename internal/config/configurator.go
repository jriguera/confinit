package config

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"confinit/pkg/log"
)

// Configurator is an interface to define a configurator factory
type Configurator interface {
	InitConfig() *Config
	CheckConfig(cfg *Config) error
	GetConfigFile(def bool) string
	LoadConfig(key string) (*Config, error)
	SaveConfig(cfg *Config) error
	BindFlagCommand(global bool, key string, cmd *cobra.Command) error
	BindFlagSet(flags *pflag.FlagSet) error
	GetConfigMap(ptr interface{}) (map[string]interface{}, error)
	Logger() log.Logger
}

// Configurator is the base class to
type configurator struct {
	// FileConfig is the name of the configuration file
	FileConfig    string
	PathConfig    string
	EnvPrefix     string
	ConsoleFormat string
	Version       string
	viper         *viper.Viper
	Log           log.Logger
}

// NewConfigurator creates a
func NewConfigurator(version, configArg string, cmd *cobra.Command) Configurator {
	c := configurator{
		FileConfig:    ConfigFile,
		PathConfig:    ConfigPath,
		EnvPrefix:     ConfigEnv,
		Version:       version,
		ConsoleFormat: "%localtime%%fields% %msg%",
		viper:         viper.New(),
		Log:           log.Standard(),
	}
	m := make(map[string]interface{})
	inspectConfig(reflect.ValueOf(new(Config)), "flag", "", ".", m)
	for key, value := range m {
		cmd.PersistentFlags().String(key, "", value.(string))
		if err := c.BindFlagCommand(true, key, cmd); err != nil {
			log.Panicf("Could not bind global flag: %s", err)
		}
	}
	// Bind config flag
	cfgFullPath := filepath.Join(ConfigPath, ConfigFile)
	msg := fmt.Sprintf("%s config file (default %s)", ConfigType, cfgFullPath)
	cmd.PersistentFlags().String(configArg, "", msg)
	c.BindFlagCommand(true, configArg, cmd)
	return &c
}

func (c *configurator) InitConfig() *Config {
	cfg := new(Config)
	cfg.SetDefaultConfig()
	// Viper initialize defaults
	m := make(map[string]interface{})
	inspectConfig(reflect.ValueOf(cfg), "", "", ".", m)
	for key, value := range m {
		if value != nil {
			c.viper.SetDefault(key, value)
		}
	}
	c.viper.SetConfigName(ConfigFile)
	c.viper.SetConfigType(ConfigType)
	c.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	c.viper.AddConfigPath(ConfigPath)
	// Logger initialize
	logger, _ := log.New(&log.Config{
		Level:         cfg.LogLevel,
		Output:        cfg.LogOutput,
		ConsoleFormat: c.ConsoleFormat,
	})
	c.Log = logger
	return cfg
}

func (c *configurator) Logger() log.Logger {
	return c.Log
}

func (c *configurator) CheckConfig(cfg *Config) error {
	err := cfg.Validate()
	return err
}

func (c *configurator) GetConfigFile(def bool) string {
	if def {
		return filepath.Join(ConfigPath, ConfigFile)
	}
	return filepath.Join(c.PathConfig, c.FileConfig)
}

// Load attempts to populate the struct with configuration values.
// The value passed to load must be a struct reference or an error
// will be returned.
func (c *configurator) LoadConfig(key string) (*Config, error) {
	cfg := new(Config)
	c.viper.AutomaticEnv()
	cfgfile := c.viper.GetString(key)
	basefile := filepath.Base(cfgfile)
	c.viper.SetConfigName(strings.TrimSuffix(basefile, filepath.Ext(basefile)))
	c.viper.AddConfigPath(filepath.Dir(cfgfile))
	if err := c.viper.MergeInConfig(); err != nil {
		log.Fatalf("Cannot read config file, %s", err.Error())
		return nil, err
	}
	ctxlog := log.WithField("configfile", c.viper.ConfigFileUsed())
	ctxlog.Debug("Loading configuration")
	if err := c.viper.Unmarshal(cfg); err != nil {
		ctxlog.Fatalf("Error parsing configuration file, %s", err.Error())
		return nil, err
	}
	// Set default config
	cfg.SetDefaultConfig()
	logger, err := log.New(&log.Config{
		Level:         cfg.LogLevel,
		Output:        cfg.LogOutput,
		ConsoleFormat: c.ConsoleFormat,
	})
	if err != nil {
		ctxlog.Errorf("Cannot to setup logging from config, %s", err.Error())
		return nil, err
	}
	cfgfile = c.viper.ConfigFileUsed()
	c.FileConfig = filepath.Base(cfgfile)
	c.PathConfig = filepath.Dir(cfgfile)
	c.Log = logger
	return cfg, nil
}

// Save attempts to populate the struct with configuration values.
// The value passed to load must be a struct reference or an error
// will be returned.
func (c *configurator) SaveConfig(cfg *Config) error {
	cfgfile := c.GetConfigFile(false)
	ctxlog := log.WithField("configfile", cfgfile)
	// If a config file is found, read it in.
	if err := c.viper.WriteConfigAs(cfgfile); err != nil {
		ctxlog.Errorf(err.Error())
		return err
	}
	return nil
}

// BindFlagCommand bind a command line argument with a config parameter
func (c *configurator) BindFlagCommand(global bool, key string, cmd *cobra.Command) error {
	var flag *pflag.Flag
	if global {
		flag = cmd.PersistentFlags().Lookup(key)
	} else {
		flag = cmd.Flags().Lookup(key)
	}
	return c.viper.BindPFlag(key, flag)
}

// BindFlagSet binds an existing set of pflags (pflag.FlagSet):
func (c *configurator) BindFlagSet(flags *pflag.FlagSet) error {
	return c.viper.BindPFlags(flags)
}

// GetConfigMap converts the configuration to a map
func (c *configurator) GetConfigMap(ptr interface{}) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	if reflect.TypeOf(ptr).Kind() != reflect.Ptr {
		err := fmt.Errorf("Not a struct pointer")
		log.Fatal(err.Error())
		return nil, err
	}
	inspectConfig(reflect.ValueOf(ptr), "", "", "", m)
	return m, nil
}

// Builds a map 'dict' from a struct with the fields tagged by 'flag', the map
// will filled will all keys prefixed by 'root' if is not empty. If 'sep'
// is defined it will be a flat map (k: v) otherwise is recursive.
func inspectConfig(val reflect.Value, flag, root, sep string, dict map[string]interface{}) {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		elm := val.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			val = elm
		}
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	// fields
	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)
		tags := typeField.Tag
		if valueField.Kind() == reflect.Interface &&
			!valueField.IsNil() {
			elm := valueField.Elem()
			if elm.Kind() == reflect.Ptr &&
				!elm.IsNil() &&
				elm.Elem().Kind() == reflect.Ptr {
				valueField = elm
			}
		}
		if valueField.Kind() == reflect.Ptr {
			valueField = valueField.Elem()
		}
		name := typeField.Name
		if m, ok := tags.Lookup("mapstructure"); ok {
			name = m
		}
		if !(root == "" || sep == "") {
			name = root + sep + name
		}
		if valueField.Kind() == reflect.Struct {
			if sep == "" {
				ndict := make(map[string]interface{})
				dict[name] = ndict
				inspectConfig(valueField, flag, root, sep, ndict)
			} else {
				inspectConfig(valueField, flag, name, sep, dict)
			}
		} else {
			//dict[name] = reflect.New(valueField.Type()).Elem().Interface()
			if flag != "" {
				if v, ok := tags.Lookup(flag); ok {
					dict[name] = v
				}
			} else {
				dict[name] = nil
				if valueField.IsValid() {
					dict[name] = valueField.Interface()
				}
			}
		}
	}
}
