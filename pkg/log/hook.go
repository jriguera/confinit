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
