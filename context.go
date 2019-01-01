package logger

import (
	"context"

	"gopkg.daemonl.com/logger/trace"
)

var logContextKey = struct{}{}

func FromContext(ctx context.Context) Entry {
	entry, ok := ctx.Value(logContextKey).(Entry)
	if !ok {
		entry = New()
	}
	if traceKey, ok := trace.GetTrace(ctx); ok {
		entry = entry.WithField("trace", traceKey)
	}
	return entry
}

func WithEntry(ctx context.Context, entry Entry) context.Context {
	if entry == nil {
		entry = New()
	}
	return context.WithValue(ctx, logContextKey, entry)
}
