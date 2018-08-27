package logger

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestMiddleware(t *testing.T) {

	logger, chEntries := testLogger()
	DefaultLogger = logger

	handler := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		FromContext(req.Context()).WithField("key", "value").Debug("Inside Handler")

	})

	ts := httptest.NewServer(Middleware(handler))

	u, _ := url.Parse(ts.URL)

	req, _ := http.NewRequest("GET", ts.URL, nil)
	req.Header.Set("User-Agent", "TESTAGENT")
	req.Header.Set("Referer", "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referer")
	req.Header.Set("X-Forwarded-For", "8.8.8.8")
	_, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Errorf(err.Error())
	}

	close(chEntries)

	debugEntry := <-chEntries
	if debugEntry.Name != "Inside Handler" {
		t.Errorf("Wrong entry: %s", debugEntry.Name)
	}
	mapCompare(t, debugEntry.Data, map[string]interface{}{
		"key": "value",
		"serving": map[string]interface{}{
			"path":   "/",
			"method": "GET",
			"query":  url.Values{},
			"host":   u.Host,
			"remote": "8.8.8.8",
		},
	})

	serveEntry := <-chEntries
	if serveEntry.Name != "HTTP Served Request" {
		t.Errorf("Wrong entry: %s", serveEntry.Name)
	}

	mapCompare(t, serveEntry.Data, map[string]interface{}{
		"statusCode": 200,
		"agent":      "TESTAGENT",
		"referer":    "https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referer",
		"serving": map[string]interface{}{
			"path":   "/",
			"method": "GET",
			"query":  url.Values{},
			"host":   u.Host,
			"remote": "8.8.8.8",
		},
	})

}

func mapCompare(t *testing.T, got map[string]interface{}, want map[string]interface{}) {
	t.Helper()
	for key, expectVal := range want {
		got, ok := got[key]
		if !ok {
			t.Errorf("Missing: %s", key)
		}

		if !reflect.DeepEqual(got, expectVal) {
			t.Errorf("%v != %v", got, expectVal)
		}
	}

}
