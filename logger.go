package logger

import (
	"io"
	"os"
	"strings"
)

const (
	TrackLevel = iota
	ErrorLevel
	WarnLevel
	InfoLevel
	DebugLevel
)

var LevelNames = map[int]string{
	TrackLevel: "track",
	ErrorLevel: "error",
	WarnLevel:  "warn",
	InfoLevel:  "info",
	DebugLevel: "debug",
}

// Entry is the stub of a new log entry being built. The called methods should
// all return a copy of the object, not modify the original
type Entry interface {
	WithField(string, interface{}) Entry
	WithFields(map[string]interface{}) Entry

	Error(string)
	Debug(string)
	Info(string)
}

type Logger interface {
	WithFields(map[string]interface{}) Entry
	WithField(string, interface{}) Entry

	AddHook(LogWriter)

	log(level int, name string, entry map[string]interface{})
}

type entry struct {
	fields map[string]interface{}
	logger Logger
}

func (e entry) log(level int, name string) {
	e.logger.log(level, name, e.fields)
}

// WithField retuns a clone of Entry with the field set to value
func (e entry) WithField(key string, value interface{}) Entry {
	return e.WithFields(map[string]interface{}{
		key: value,
	})
}

// WithFields returns a clone of Entry with the fields set, overriding any
// existing fields
func (e entry) WithFields(fields map[string]interface{}) Entry {
	mergedFields := map[string]interface{}{}
	for k, v := range e.fields {
		mergedFields[k] = v
	}
	for k, v := range fields {
		mergedFields[k] = v
	}
	return entry{
		logger: e.logger,
		fields: mergedFields,
	}
}

func (e entry) Error(name string) {
	e.log(ErrorLevel, name)
}
func (e entry) Debug(name string) {
	e.log(DebugLevel, name)
}
func (e entry) Info(name string) {
	e.log(InfoLevel, name)
}
func (e entry) Warn(name string) {
	e.log(WarnLevel, name)
}
func (e entry) Track(name string) {
	e.log(TrackLevel, name)
}

type logger struct {
	rootEntry map[string]interface{}
	level     int
	writers   []LogWriter
}

type Formatter interface {
	Format(w io.Writer, level int, name string, data map[string]interface{}) error
}

type LogWriter interface {
	Write(level int, name string, data map[string]interface{}) error
}

type writerWriter struct {
	io.Writer
	Formatter
}

func (ww writerWriter) Write(level int, name string, data map[string]interface{}) error {
	return ww.Formatter.Format(ww.Writer, level, name, data)
}

func (l *logger) WithFields(fields map[string]interface{}) Entry {
	return entry{
		logger: l,
		fields: l.rootEntry,
	}.WithFields(fields)
}

func (l *logger) WithField(key string, val interface{}) Entry {
	return entry{
		logger: l,
		fields: l.rootEntry,
	}.WithField(key, val)
}

func (l *logger) AddHook(w LogWriter) {
	l.writers = append(l.writers, w)
}

func (l *logger) log(level int, name string, data map[string]interface{}) {
	for _, writer := range l.writers {
		writer.Write(level, name, data)
	}
}

var DefaultLogger Logger
var DefaultFields = map[string]interface{}{}

func Setup(appName string, version string) {
	DefaultFields["app-name"] = appName
	DefaultFields["app-version"] = version
}

func New() Entry {
	return DefaultLogger.WithFields(DefaultFields)
}

func FromEnvironment() Logger {
	stderrWriter := writerWriter{
		Writer:    os.Stderr,
		Formatter: JSONFormatter{},
	}
	l := &logger{}
	if os.Getenv("VERBOSE") == "true" {
		l.level = DebugLevel
	}

	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "error":
		l.level = ErrorLevel
	case "debug":
		l.level = DebugLevel
	case "info":
		l.level = InfoLevel
	default:
		l.level = TrackLevel
	}

	switch os.Getenv("LOG_FORMAT") {
	case "json":
		stderrWriter.Formatter = &JSONFormatter{}
	case "multiline":
		stderrWriter.Formatter = &MultilineFormatter{
			SpecialFields: []string{
				"function",
				"error",
			},
		}
	}
	l.writers = []LogWriter{stderrWriter}
	return l
}

func init() {
	DefaultLogger = FromEnvironment()
}
