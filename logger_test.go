package logger

import "testing"

type captureWriter struct {
	ch chan testEntry
}

type testEntry struct {
	Level int
	Name  string
	Data  map[string]interface{}
}

func (cf captureWriter) Write(level int, name string, data map[string]interface{}) error {

	cf.ch <- testEntry{
		Data:  data,
		Level: level,
		Name:  name,
	}
	return nil
}

func testLogger() (Logger, chan testEntry) {
	chEntries := make(chan testEntry, 20)
	return &logger{
		rootEntry: map[string]interface{}{
			"test": "test",
		},
		level: DebugLevel,
		writers: []LogWriter{captureWriter{
			ch: chEntries,
		}},
	}, chEntries

}

func TestEntry(t *testing.T) {

	logger, entries := testLogger()

	go func() {
		e1 := logger.WithField("key", "value")
		e2 := e1.WithField("other", "val2")
		e1.Debug("Message")
		e2.Debug("e2")
		close(entries)
	}()

	entry := <-entries

	if entry.Level != DebugLevel {
		t.Errorf("Wrong level: %d %s", entry.Level, LevelNames[entry.Level])
	}

	if entry.Name != "Message" {
		t.Errorf("Wrong message: %s", entry.Name)
	}

	if v := entry.Data["key"]; v != "value" {
		t.Errorf("Explicit field: %v", v)
	}
	if _, ok := entry.Data["other"]; ok {
		t.Errorf("initial entry was modified by chain")
	}

	entry = <-entries

	if v := entry.Data["key"]; v != "value" {
		t.Errorf("extended did not contain initial: %v", v)
	}
	if v := entry.Data["other"]; v != "val2" {
		t.Errorf("extended did not contain extra: %v", v)
	}

}
