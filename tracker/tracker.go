package tracker

import (
	"strings"
	"time"

	"gopkg.ecal.com/logger"
)

type Header struct {
	Version string         `json:"version"`
	Type    string         `json:"type"`
	Emitter EmitterDetails `json:"emitter"`
}

var DefaultHeader = Header{
	Version: "1.1",
	Type:    "system",
	Emitter: EmitterDetails{},
}

type EmitterDetails struct {
	System    string `json:"system"`
	Component string `json:"component"`
}

type postEvent struct {
	Header

	EmitTimestamp time.Time `json:"emit_timestamp"`

	Name string                 `json:"name"`
	Keys map[string]interface{} `json:"keys"`
	Data map[string]interface{} `json:"data"`
}

func RegisterLogHook(name string, version string) {
	hook := &LogWriter{
		Header: Header{
			Version: version,
			Type:    "system",
			Emitter: EmitterDetails{
				System:    name,
				Component: "",
			},
		},
	}
	logger.DefaultLogger.AddHook(hook)
}

type LogWriter struct {
	Header
	Publisher interface {
		Publish(interface{}) (string, error)
	}
}

func (lw LogWriter) Write(level int, name string, inputData map[string]interface{}) error {
	if level > logger.TrackLevel {
		return nil
	}

	keys := map[string]interface{}{}
	data := map[string]interface{}{}
	for k, v := range inputData {
		if strings.HasSuffix(k, "_id") {
			keys[k] = v
		} else {
			data[k] = v
		}
	}

	header := lw.Header

	if parts := strings.Split(name, "."); len(parts) == 2 {
		header.Emitter.Component = parts[0]
		name = parts[1]
	}

	_, err := lw.Publisher.Publish(postEvent{
		Header:        header,
		Name:          name,
		Data:          data,
		Keys:          keys,
		EmitTimestamp: time.Now(),
	})

	return err

}
