package router

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/server/repositories"
)

func getConfig() *config.Config {
	serverConfig := &config.Config{}
	serverConfig.Address = "localhost:8080"
	serverConfig.Restore = true
	serverConfig.StoreInterval = "300s"
	serverConfig.StoreFile = "/tmp/cmas-metrics-db.json"

	return serverConfig
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body *bytes.Buffer) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)

	defer resp.Body.Close()

	return resp, string(respBody)
}

func TestRouterWithValidRepository(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		url    string
		body   []byte
		method string
		want   want
	}{
		{
			name:   "case 1",
			url:    "/update/gauge/Alloc/3459",
			body:   []byte(``),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/update/",
			body:   []byte(`{"id":"Alloc","type":"gauge","value":3459}`),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "{}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 3",
			url:    "/update/counter/PollCount/10",
			body:   []byte(``),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 4",
			url:    "/update/",
			body:   []byte(`{"id":"PollCount","type":"counter","delta":10}`),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "{}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 5",
			url:    "/",
			body:   []byte(``),
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "is not empty",
				contentType: "text/html; charset=utf-8",
			},
		},
		{
			name:   "case 6",
			url:    "/value/counter/PollCount",
			body:   []byte(``),
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "20\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 7",
			url:    "/value/",
			body:   []byte(`{"id":"PollCount","type":"counter"}`),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":20}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	config := getConfig()
	r := NewRouter(config, repositories.NewRepositoryInMemory(config))
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, bytes.NewBuffer(tt.body))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			if tt.want.response != "is not empty" {
				assert.Equal(t, tt.want.response, body)
			} else {
				assert.NotEmpty(t, body)
			}
			resp.Body.Close()
		})
	}
}
