package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/Sirupsen/logrus"
	"gopkg.ecal.com/logger/trace"
)

var logContextKey = struct{}{}

var DefaultLogger = logrus.New()
var DefaultFields = logrus.Fields{}

func FromContext(ctx context.Context) *logrus.Entry {
	entry, ok := ctx.Value(logContextKey).(*logrus.Entry)
	if !ok {
		entry = New()
	}
	if traceKey, ok := trace.GetTrace(ctx); ok {
		entry = entry.WithField("trace", traceKey)
	}
	return entry
}

func WithEntry(ctx context.Context, entry *logrus.Entry) context.Context {
	if entry == nil {
		entry = New()
	}
	return context.WithValue(ctx, logContextKey, entry)
}

func New() *logrus.Entry {
	return DefaultLogger.WithFields(DefaultFields)
}

func FromEnvironment() *logrus.Logger {
	l := logrus.New()
	if os.Getenv("VERBOSE") == "true" {
		l.SetLevel(logrus.DebugLevel)
	}

	switch os.Getenv("LOG_FORMAT") {
	case "json":
		l.Formatter = &logrus.JSONFormatter{}
	case "pretty":
		l.Formatter = &logrus.TextFormatter{
			ForceColors: true,
		}
	case "text", "":
		l.Formatter = &logrus.TextFormatter{}
	case "multiline":
		l.Formatter = &MultilineFormatter{
			SpecialFields: []string{
				"function",
				"error",
				"schedule",
			},
		}
	}
	return l
}

func init() {
	DefaultLogger = FromEnvironment()
}

type MultilineFormatter struct {
	SpecialFields []string
}

func (f MultilineFormatter) getOrder(k string) int {
	for idx, entry := range f.SpecialFields {
		if k == entry {
			return idx
		}
	}

	return 99999
}

type sortedFields []logField

func (a sortedFields) Len() int      { return len(a) }
func (a sortedFields) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sortedFields) Less(i, j int) bool {
	if a[i].order != a[j].order {
		return a[i].order < a[j].order
	}
	return a[i].key < a[j].key
}

type logField struct {
	order int
	key   string
	val   string
}

func (f MultilineFormatter) Format(entry *logrus.Entry) ([]byte, error) {

	fields := []logField{}

	for key, val := range entry.Data {
		valBytes, _ := json.Marshal(val)
		fields = append(fields, logField{
			order: f.getOrder(key),
			key:   key,
			val:   string(valBytes),
		})
	}

	sort.Sort(sortedFields(fields))

	b := bytes.NewBuffer([]byte{})
	levelColor := levelColors[entry.Level]

	fmt.Fprintf(b, "---------------\n\x1b[%dm%s\x1b[0m: %-44s\n", levelColor, entry.Level.String(), entry.Message)
	for _, field := range fields {
		fmt.Fprintf(b, "    %s: %s\n", field.key, field.val)
	}

	return b.Bytes(), nil

}

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

var levelColors = map[logrus.Level]int{
	logrus.DebugLevel: green,
	logrus.ErrorLevel: red,
	logrus.InfoLevel:  blue,
	logrus.PanicLevel: red,
	logrus.WarnLevel:  yellow,
}
