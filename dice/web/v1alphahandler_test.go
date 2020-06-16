package web

import (
	"bytes"
	"context"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestPing(t *testing.T) {

	tests := []struct {
		name   string
		input  string
		output string
	}{
		{"Test return http_code", "", "200"},
		{"Test the content of response", "", "pong"},
	}

	r := Router(context.TODO())
	recorder := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	r.ServeHTTP(recorder, req)

	t.Run(tests[0].name, func(t *testing.T) {
		code, _ := strconv.Atoi(tests[0].output)
		assert.Equal(t, code, recorder.Code)
	})

	t.Run(tests[1].name, func(t *testing.T) {
		assert.Equal(t, tests[1].output, recorder.Body.String())
	})

}

func TestRouter(t *testing.T) {
	tests := []struct {
		name   string
		uri    string
		method string
		input  string
		code   int
		output string
	}{
		{"ping", "/ping", "GET", "", 200, "pong"},
		{"websocket", "/v1alpha1/ws", "GET", "", 200, ""},
		{"websocket+dry-run", "/v1alpha1/ws?dryRun=true", "GET", "", 200, ""},
		{"validate tile-spec", "/v1alpha1/tile", "POST", "nothing", 200, ""},
		{"validate tile-spec", "/v1alpha1/deployment", "POST", "nothing", 200, ""},
		{"generate tile with template", "/v1alpha1/tile", "GET", "", 200, ""},
		{"generate deployment with template", "/v1alpha1/deployment", "GET", "", 200, ""},
	}

	r := Router(context.TODO())
	recorder := httptest.NewRecorder()

	for i, test := range tests {
		req, _ := http.NewRequest(test.method, test.uri, bytes.NewReader([]byte(test.input)))
		r.ServeHTTP(recorder, req)

		t.Run(tests[i].name, func(t *testing.T) {
			assert.Equal(t, test.code, recorder.Code)
		})
	}

}
