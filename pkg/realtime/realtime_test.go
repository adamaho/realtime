package realtime

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type Data struct {
	Count int `json:"count"`
}

func TestResponseWithoutStreamingHeader(t *testing.T) {
	initial, _ := json.Marshal(Data{Count: 0})
	rt := New()

	r := httptest.NewRequest(http.MethodGet, "/count", nil)
	w := httptest.NewRecorder()

	rt.Response(w, r, initial, "count", ResponseOptions())

	result := w.Result()
	defer result.Body.Close()

	actualContentType := result.Header.Get("Content-Type")
	expectedContentType := "application/json"
	if actualContentType != expectedContentType {
		t.Errorf("expected content-type to be %s, received %s", expectedContentType, actualContentType)
	}

	actualStatusCode := result.StatusCode
	expectedStatusCode := 200
	if actualStatusCode != expectedStatusCode {
		t.Errorf("expected status-code to be %d, received %d", expectedStatusCode, actualStatusCode)
	}

	actual, err := io.ReadAll(result.Body)
	if err != nil {
		t.Errorf("expected error to be nil, received %v", err)
	}

	if string(actual) != string(initial) {
		t.Errorf("expected response body to be %s, received %s", initial, actual)
	}
}
