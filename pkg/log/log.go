package log

import (
	"io/ioutil"
	"os"
	//"runtime"
	//"strings"

	"github.com/sirupsen/logrus"
)

// Config specifies all the parameters needed for logging
type Config struct {
	Level         string
	Output        string
	ConsoleFormat string
	Fields        Fields
}

// Fields aliases logrus.Fields
type Fields map[string]interface{}

type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	Panic(...interface{})
	Panicf(string, ...interface{})
	WithField(key string, value interface{}) Logger
	WithError(err error) Logger
	WithFields(fields Fields) Logger
}

type logger struct {
	entry  *logrus.Entry
	config *Config
}

var root *logrus.Logger
var logg *logger

func init() {
	Standard()
}

func Standard() Logger {
	root = logrus.StandardLogger()
	logg = &logger{
		entry: logrus.NewEntry(root),
		config: &Config{
			Level:         "info",
			Output:        "split",
			ConsoleFormat: "%localtime% %msg%%fields%",
		},
	}
	return logg
}

// New creates a logger based logging configuration and also adds
// a few default parameters. When stdout adds hooks to send logs to
// different destinations depending on level
func New(config *Config) (Logger, error) {
	root = logrus.New()
	if err := SetOutput(config.Output); err != nil {
		return nil, err
	}
	// Set level
	if err := SetLevel(config.Level); err != nil {
		return nil, err
	}
	console := false
	switch config.Output {
	case "stdout":
		console = true
	case "stderr":
		console = true
	case "split":
		console = true
	}
	if console {
		SetTextFormatter(config.ConsoleFormat)
	} else {
		SetJSONFormatter()
	}
	// Add global fields
	SetFields(config.Fields)
	logg = &logger{
		entry:  logrus.NewEntry(root),
		config: config,
	}
	return logg, nil
}

//
func SetLevel(level string) error {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		logrus.
			WithError(err).
			Error("Log level only allows: debug,info,warn,error,fatal,panic")
		return err
	}
	root.SetLevel(l)
	return nil
}

//
func SetOutput(output string) error {
	// Reset Hooks and Output
	root.SetOutput(ioutil.Discard)
	root.ReplaceHooks(make(logrus.LevelHooks))
	switch output {
	case "stdout":
		root.SetOutput(os.Stdout)
	case "stderr":
		root.SetOutput(os.Stderr)
	case "split":
		// Send logs with level higher than warning to stderr
		root.AddHook(StdErrorHook())
		// Send info and debug logs to stdout
		root.AddHook(StdOutHook())
	case "-", "", " ":
		root.SetOutput(ioutil.Discard)
	default:
		fd, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0664)
		// close fd!
		if err == nil {
			root.SetOutput(fd)
		} else {
			logrus.
				WithError(err).
				Errorf("Cannot open output log file: %s", output)
			return err
		}
	}
	return nil
}

//
func SetJSONFormatter() {
	root.SetFormatter(&logrus.JSONFormatter{})
}

//
func SetTextFormatter(f string) {
	root.SetFormatter(NewConsoleFormatter(true, ConsoleLogFormat(f)))
}

func SetFields(f Fields) {
	// Add global fields
	for key, value := range f {
		root.WithField(key, value)
	}
}

func Debug(args ...interface{}) {
	logg.Debug(args...)
}

func Debugf(format string, args ...interface{}) {
	logg.Debugf(format, args...)
}

func Info(args ...interface{}) {
	logg.Info(args...)
}

func Infof(format string, args ...interface{}) {
	logg.Infof(format, args...)
}

func Warn(args ...interface{}) {
	logg.Warn(args...)
}

func Warnf(format string, args ...interface{}) {
	logg.Warnf(format, args...)
}

func Error(args ...interface{}) {
	logg.Error(args...)
}

func Errorf(format string, args ...interface{}) {
	logg.Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func Fatal(args ...interface{}) {
	logg.Fatal(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func Fatalf(format string, args ...interface{}) {
	logg.Fatalf(format, args...)
}

// Panic logs a message at level Panic on the standard logger.
func Panic(args ...interface{}) {
	logg.Panic(args...)
}

// Panicf logs a message at level Panic on the standard logger.
func Panicf(format string, args ...interface{}) {
	logg.Panicf(format, args...)
}

func WithField(key string, value interface{}) Logger {
	return logg.WithField(key, value)
}

func WithError(err error) Logger {
	return logg.WithError(err)
}

func WithFields(fields Fields) Logger {
	return logg.WithFields(fields)
}

///////////////////////////////////////////////

func (l *logger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *logger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *logger) Warn(args ...interface{}) {
	l.entry.Warn(args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.entry.Warnf(format, args...)
}

func (l *logger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

// Fatal logs a message at level Fatal on the standard logger.
func (l *logger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

// Fatalf logs a message at level Fatal on the standard logger.
func (l *logger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

// Panic logs a message at level Panic on the standard logger.
func (l *logger) Panic(args ...interface{}) {
	l.entry.Panic(args...)
}

// Panicf logs a message at level Panic on the standard logger.
func (l *logger) Panicf(format string, args ...interface{}) {
	l.entry.Panicf(format, args...)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{
		entry:  l.entry.WithField(key, value),
		config: l.config,
	}
}

func (l *logger) WithError(err error) Logger {
	return &logger{
		entry:  l.entry.WithError(err),
		config: l.config,
	}
}

func (l *logger) WithFields(fields Fields) Logger {
	return &logger{
		entry:  l.entry.WithFields(logrus.Fields(fields)),
		config: l.config,
	}
}
