package log

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
)

// WriterHook is a hook that writes logs of specified LogLevels to
// specified Writer. Code from https://github.com/sirupsen/logrus/issues/678
type WriterHook struct {
	Writer    io.Writer
	LogLevels []logrus.Level
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *WriterHook) Fire(entry *logrus.Entry) error {
	line, err := entry.String()
	if err != nil {
		return err
	}
	_, err = hook.Writer.Write([]byte(line))
	return err
}

// Levels define on which log levels this hook would trigger
func (hook *WriterHook) Levels() []logrus.Level {
	return hook.LogLevels
}

// StdErrorHook defines a hook for stderr for common error logs
func StdErrorHook(levels ...logrus.Level) *WriterHook {
	if len(levels) == 0 {
		levels = []logrus.Level{
			logrus.PanicLevel,
			logrus.FatalLevel,
			logrus.ErrorLevel,
			logrus.WarnLevel,
		}
	}
	return &WriterHook{
		Writer:    os.Stderr,
		LogLevels: levels,
	}
}

// StdOutHook defines a hook for stderr for common info logs
func StdOutHook(levels ...logrus.Level) *WriterHook {
	if len(levels) == 0 {
		levels = []logrus.Level{
			logrus.InfoLevel,
			logrus.DebugLevel,
		}
	}
	return &WriterHook{
		Writer:    os.Stdout,
		LogLevels: levels,
	}
}
