package log

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// Terminal color constants
const (
	DebugLevelColor = 36
	InfoLevelColor  = 39
	WarnLevelColor  = 93
	ErrorLevelColor = 31
	FatalLevelColor = 31
	PanicLevelColor = 31
	CallerColor     = 37
	FieldsColor     = 96
	FieldKColor     = 93
	FieldVColor     = 95
	FieldSepColor   = 90
	TimestampColor  = 36
	RstColor        = 0
	DefaultColor    = 0
)

// logrus formatter for console
type ConsoleFormatter struct {
	// Timestamp format
	ConsoleTimestampFormat string
	// Level format
	ConsoleLevelFormat string
	// Available standard keys: time, msg, lvl
	// Also can include custom fields but limited to strings.
	// All of fields need to be wrapped inside %% i.e %time% %msg%
	ConsoleLogFormat    string
	ConsoleCallerFormat string
	// String value used to separate log fields
	ConsoleFieldSep string
	// String to separate KV fields
	ConsoleFieldKVSep string
	ConsoleFieldsWrap string
	// Enable color
	ConsoleLogColor bool
}

// OptionFormatter to pass to the constructor using Functional Options
type OptionFormatter func(*ConsoleFormatter)

// ConsoleTimestampFormat is a function used by users to set options.
func ConsoleTimestampFormat(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleTimestampFormat = s
		}
	}
}

// ConsoleLevelFormat is a function used by users to set options.
func ConsoleLevelFormat(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleLevelFormat = s
		}
	}
}

// ConsoleLogFormat is a function used by users to set options.
func ConsoleLogFormat(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleLogFormat = s
		}
	}
}

// ConsoleCallerFormat is a function used by users to set options.
func ConsoleCallerFormat(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleCallerFormat = s
		}
	}
}

// ConsoleFieldSep is a function used by users to set options.
func ConsoleFieldSep(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleFieldSep = s
		}
	}
}

// ConsoleFieldKVSep is a function used by users to set options.
func ConsoleFieldKVSep(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleFieldKVSep = s
		}
	}
}

// ConsoleFieldsWrap is a function used by users to set options.
func ConsoleFieldsWrap(s string) OptionFormatter {
	return func(f *ConsoleFormatter) {
		if s != "" {
			f.ConsoleFieldsWrap = s
		}
	}
}

// ConsoleFormatter is Constructor with default options
func NewConsoleFormatter(color bool, opts ...OptionFormatter) *ConsoleFormatter {
	f := &ConsoleFormatter{
		ConsoleTimestampFormat: "2018-01-02/15:04:05",
		ConsoleLevelFormat:     "%.1s",
		ConsoleLogFormat:       "%localtime% %LEVEL% %msg% %fields%",
		ConsoleCallerFormat:    "%file%:%line% %fun%()",
		ConsoleFieldSep:        ", ",
		ConsoleFieldKVSep:      ":",
		ConsoleFieldsWrap:      " [%s]", // "「%s」"
		ConsoleLogColor:        color,
	}
	// call option functions on instance to set options on it
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *ConsoleFormatter) setColor(entry *logrus.Entry) string {
	clr := DefaultColor
	switch entry.Level {
	case logrus.ErrorLevel:
		clr = ErrorLevelColor
	case logrus.PanicLevel:
		clr = PanicLevelColor
	case logrus.FatalLevel:
		clr = FatalLevelColor
	case logrus.WarnLevel:
		clr = WarnLevelColor
	case logrus.InfoLevel:
		clr = InfoLevelColor
	case logrus.DebugLevel:
		clr = DebugLevelColor
	}
	return fmt.Sprintf("\x1b[%dm", clr)
}

func (f *ConsoleFormatter) resetColor() string {
	return fmt.Sprintf("\x1b[%dm", RstColor)
}

func (f *ConsoleFormatter) colorize(s string, c int, e *logrus.Entry) string {
	if s != "" && f.ConsoleLogColor {
		if c != DefaultColor {
			return fmt.Sprintf("\x1b[%dm", c) + s + f.resetColor()
		}
		return f.setColor(e) + s + f.resetColor()
	}
	return s
}

func (f *ConsoleFormatter) fieldsColor(k, v string) string {
	if k != "" {
		sep := f.ConsoleFieldKVSep
		if f.ConsoleLogColor {
			k = fmt.Sprintf("\x1b[%dm", FieldKColor) + k + f.resetColor()
			v = fmt.Sprintf("\x1b[%dm", FieldVColor) + v + f.resetColor()
			sep = fmt.Sprintf("\x1b[%dm", FieldSepColor) + sep + f.resetColor()
		}
		return k + sep + v
	}
	return k
}

func (f *ConsoleFormatter) formatCaller(format string, e *logrus.Entry) string {
	output := format
	if e.HasCaller() && strings.Contains(format, "%caller%") {
		output = strings.Replace(f.ConsoleCallerFormat, "%file%", e.Caller.File, 1)
		line := fmt.Sprintf("%d", e.Caller.Line)
		output = strings.Replace(output, "%line%", line, 1)
		output = strings.Replace(output, "%function%", e.Caller.Function, 1)
		output = f.colorize(output, CallerColor, e)
		output = strings.Replace(format, "%caller%", output, 1)
	}
	return output
}

func (f *ConsoleFormatter) formatTime(format string, e *logrus.Entry) string {
	timestamp := e.Time
	if strings.Contains(format, "%localtime%") {
		timestamp = e.Time.Local()
		output := f.colorize(timestamp.Format(
			f.ConsoleTimestampFormat), TimestampColor, e,
		)
		return strings.Replace(format, "%localtime%", output, 1)
	}
	output := f.colorize(timestamp.Format(
		f.ConsoleTimestampFormat), TimestampColor, e,
	)
	return strings.Replace(format, "%time%", output, 1)
}

func (f *ConsoleFormatter) formatFields(format string, e *logrus.Entry) string {
	// To store the keys in slice in sorted order
	keys := []string{}
	for k := range e.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	output := format
	fields := ""
	for _, k := range keys {
		if s, ok := e.Data[k].(string); ok {
			if strings.Contains(output, "%"+k+"%") {
				output = strings.Replace(output, "%"+k+"%", s, 1)
			} else {
				if fields == "" {
					fields = f.fieldsColor(k, s)
				} else {
					fields = fields + f.ConsoleFieldSep + f.fieldsColor(k, s)
				}
			}
		}
	}
	if fields != "" {
		fields = f.colorize(fields, FieldsColor, e)
		fields = fmt.Sprintf(f.ConsoleFieldsWrap, fields)
	}
	return strings.Replace(output, "%fields%", fields, 1)
}

func (f *ConsoleFormatter) formatLevel(format string, e *logrus.Entry) string {
	levelText := e.Level.String()
	if strings.Contains(format, "%LEVEL%") {
		levelText = strings.ToUpper(levelText)
		output := f.colorize(fmt.Sprintf(
			f.ConsoleLevelFormat, levelText), DefaultColor, e,
		)
		return strings.Replace(format, "%LEVEL%", output, 1)
	}
	output := f.colorize(fmt.Sprintf(
		f.ConsoleLevelFormat, levelText), DefaultColor, e,
	)
	return strings.Replace(format, "%level%", output, 1)
}

// Format Logrus Formatter main function
func (f *ConsoleFormatter) Format(e *logrus.Entry) ([]byte, error) {
	buf := new(bytes.Buffer)
	// Remove a single newline if it already exists in the message to keep
	// the behavior of logrus text_formatter the same as the stdlib log package
	strings.TrimSuffix(e.Message, "\n")
	msg := f.colorize(e.Message, DefaultColor, e)
	output := strings.Replace(f.ConsoleLogFormat, "%msg%", msg, 1)
	output = f.formatTime(output, e)
	output = f.formatLevel(output, e)
	output = f.formatCaller(output, e)
	output = f.formatFields(output, e)
	fmt.Fprint(buf, output, "\n")
	return buf.Bytes(), nil
}
