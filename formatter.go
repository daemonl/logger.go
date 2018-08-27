package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"time"
)

type JSONFormatter struct{}

func (JSONFormatter) Format(writer io.Writer, level int, name string, data map[string]interface{}) error {

	dataOut := map[string]interface{}{}
	for k, v := range data {
		dataOut[k] = v
	}
	for _, meta := range []string{"level", "message", "time"} {
		if val, ok := dataOut[meta]; ok {
			data["fields."+meta] = val
		}
	}
	dataOut["level"] = LevelNames[level]
	dataOut["message"] = name
	dataOut["time"] = time.Now().Format(time.RFC3339)
	return json.NewEncoder(writer).Encode(dataOut)
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

func (f MultilineFormatter) Format(writer io.Writer, level int, name string, data map[string]interface{}) error {

	fields := []logField{}

	for key, val := range data {
		valBytes, _ := json.Marshal(val)
		fields = append(fields, logField{
			order: f.getOrder(key),
			key:   key,
			val:   string(valBytes),
		})
	}

	sort.Sort(sortedFields(fields))

	levelColor := levelColors[level]

	if _, err := fmt.Fprintf(writer, "---------------\n\x1b[%dm%s\x1b[0m: %-44s\n", levelColor, LevelNames[level], name); err != nil {
		return err
	}
	for _, field := range fields {
		if _, err := fmt.Fprintf(writer, "    %s: %s\n", field.key, field.val); err != nil {
			return err
		}

	}

	return nil
}

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

var levelColors = map[int]int{
	DebugLevel: green,
	ErrorLevel: red,
	InfoLevel:  blue,
	WarnLevel:  yellow,
}
