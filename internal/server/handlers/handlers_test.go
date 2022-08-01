package handlers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/server/repositories"
	"github.com/ustkit/cmas/internal/types"
)

func getConfig() *config.Config {
	serverConfig := &config.Config{}
	serverConfig.Address = "localhost:8080"
	serverConfig.Restore = true
	serverConfig.StoreInterval = "300s"
	serverConfig.StoreFile = "/tmp/cmas-metrics-db.json"

	return serverConfig
}

type BrokenRepoInMemory struct {
}

func (mr BrokenRepoInMemory) Save(ctx context.Context, name string, value types.Value) error {
	return errors.New("operation not allowed")
}

func (mr BrokenRepoInMemory) SaveAll(ctx context.Context, values []types.ValueJSON) error {
	return nil
}

func (mr BrokenRepoInMemory) FindByName(ctx context.Context, name string) (types.Value, error) {
	return types.Value{}, fmt.Errorf("metric with %s not found", name)
}

func (mr BrokenRepoInMemory) FindAll(ctx context.Context) (types.Values, error) {
	return nil, errors.New("metrics not found")
}

func (mr BrokenRepoInMemory) Restore() error {
	return nil
}

func (mr BrokenRepoInMemory) SaveToFile() error {
	return nil
}

func (mr BrokenRepoInMemory) Close() error {
	return nil
}

func (mr BrokenRepoInMemory) Ping(ctx context.Context) error {
	return errors.New("no access repo")
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

func TestIndex_WithValidRepository(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "case 1",
			url:    "/update/gauge/Alloc/3459",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/update/counter/PollCount/10",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 3",
			url:    "/",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "is not empty",
				contentType: "text/html; charset=utf-8",
			},
		},
	}

	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Get("/", h.Index)
	r.Post("/update/{type}/{name}/{value}", h.UpdatePlain)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, &bytes.Buffer{})
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

func TestIndex_WithBrokenRepository(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "case 1",
			url:    "/update/gauge/alloc/3459",
			method: http.MethodPost,
			want: want{
				code:        500,
				response:    "operation not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/update/counter/pollcount/10",
			method: http.MethodPost,
			want: want{
				code:        500,
				response:    "operation not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 3",
			url:    "/",
			method: http.MethodGet,
			want: want{
				code:        204,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	config := getConfig()

	r := chi.NewRouter()
	h := NewHandler(config, BrokenRepoInMemory{})
	r.Get("/", h.Index)
	r.Post("/update/{type}/{name}/{value}", h.UpdatePlain)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, &bytes.Buffer{})
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestUpdatePlain_WithValidRepository(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "case 1",
			url:    "/update/gauge/Alloc/3459",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/update/counter/PollCount/10",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 3",
			url:    "/update/gauge/",
			method: http.MethodPost,
			want: want{
				code:        404,
				response:    "404 page not found\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 4",
			url:    "/update/gauge/Alloc/none",
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "incorrect value\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 5",
			url:    "/update/counter/Alloc/none",
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "incorrect value\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 6",
			url:    "/update/unknow/metric/10",
			method: http.MethodPost,
			want: want{
				code:        501,
				response:    "unknown data type\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Post("/update/{type}/{name}/{value}", h.UpdatePlain)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, &bytes.Buffer{})
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

func TestUpdatePlain_WithBrokenRepository(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "case 1",
			url:    "/update/gauge/alloc/3459",
			method: http.MethodPost,
			want: want{
				code:        500,
				response:    "operation not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/update/counter/pollcount/10",
			method: http.MethodPost,
			want: want{
				code:        500,
				response:    "operation not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	config := getConfig()

	r := chi.NewRouter()
	h := NewHandler(config, BrokenRepoInMemory{})
	r.Post("/update/{type}/{name}/{value}", h.UpdatePlain)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, &bytes.Buffer{})
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestValuePlain_WithValidRepository(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{
			name:   "case 1",
			url:    "/update/gauge/Alloc/3459",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/update/counter/PollCount/10",
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 3",
			url:    "/value/gauge/Alloc",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "3459\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 4",
			url:    "/value/unknow/Alloc",
			method: http.MethodGet,
			want: want{
				code:        404,
				response:    "\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 5",
			url:    "/value/gauge/metric",
			method: http.MethodGet,
			want: want{
				code:        404,
				response:    "\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 6",
			url:    "/value/counter/PollCount",
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "10\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 7",
			url:    "/value/unknow/PollCount",
			method: http.MethodGet,
			want: want{
				code:        404,
				response:    "\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			name:   "case 8",
			url:    "/value/counter/metric",
			method: http.MethodGet,
			want: want{
				code:        404,
				response:    "\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Post("/update/{type}/{name}/{value}", h.UpdatePlain)
	r.Get("/value/{type}/{name}", h.ValuePlain)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, &bytes.Buffer{})
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

func TestUpdateJSON_WithValidRepository(t *testing.T) {
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
			name:   "case 2",
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
			name:   "case 3",
			url:    "/update/",
			body:   []byte(`{"type":"gauge"}`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"metric name empity\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 4",
			url:    "/update/",
			body:   []byte(`{"type":"counter"}`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"metric name empity\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 5",
			url:    "/update/",
			body:   []byte(`{"id":"Alloc","type":"gauge","value":"none"}`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"json: cannot unmarshal string into Go struct field ValueJSON.value of type types.Gauge\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 6",
			url:    "/update/",
			body:   []byte(`{"id":"Alloc","type":"counter","delta":"none"}`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"json: cannot unmarshal string into Go struct field ValueJSON.delta of type types.Counter\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 7",
			url:    "/update/",
			body:   []byte(`{"id":"metric","type":"unknow","value":10}`),
			method: http.MethodPost,
			want: want{
				code:        501,
				response:    "{\"error\":\"unknown data type\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 8",
			url:    "/update/",
			body:   []byte(`{"type":"gauge","value":945}`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"metric name empity\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 9",
			url:    "/update/",
			body:   []byte(`{"type":"counter","delta":43}`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"metric name empity\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Post("/update/", h.UpdateJSON)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, bytes.NewBuffer(tt.body))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestUpdateJSONBacth_WithValidRepository(t *testing.T) {
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
			url:    "/updates/",
			body:   []byte(`[{"id":"Alloc","type":"gauge","value":3459},{"id":"PollCount","type":"counter","delta":10}]`),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "{}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 2",
			url:    "/updates/",
			body:   []byte(`[{"type":"gauge"},{"id":"PollCount","type":"counter","delta":10}]`),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"metric name empity\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 3",
			url:    "/updates/",
			body:   []byte(`[{"id":"metric","type":"unknow","value":10}]`),
			method: http.MethodPost,
			want: want{
				code:        501,
				response:    "{\"error\":\"unknown data type for metric\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
	}

	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Post("/updates/", h.UpdateJSONBatch)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, bytes.NewBuffer(tt.body))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestValueJSON_WithValidRepository(t *testing.T) {
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
			name:   "case 2",
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
			name:   "case 3",
			url:    "/value/",
			body:   []byte(`{"id":"Alloc","type":"gauge"}`),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "{\"id\":\"Alloc\",\"type\":\"gauge\",\"value\":3459}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 4",
			url:    "/value/",
			body:   []byte(`{"id":"PollCount","type":"counter"}`),
			method: http.MethodPost,
			want: want{
				code:        200,
				response:    "{\"id\":\"PollCount\",\"type\":\"counter\",\"delta\":10}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 5",
			url:    "/value/",
			body:   []byte(``),
			method: http.MethodPost,
			want: want{
				code:        400,
				response:    "{\"error\":\"EOF\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
		{
			name:   "case 6",
			url:    "/value/",
			body:   []byte(`{"id":"PollCount","type":"gauge"}`),
			method: http.MethodPost,
			want: want{
				code:        404,
				response:    "{\"error\":\"metric not found\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Post("/update/", h.UpdateJSON)
	r.Post("/value/", h.ValueJSON)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, bytes.NewBuffer(tt.body))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestPing_WithValidRepository(t *testing.T) {
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
			url:    "/ping/",
			body:   []byte(``),
			method: http.MethodGet,
			want: want{
				code:        200,
				response:    "{}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	config := getConfig()
	repo := repositories.NewRepositoryInMemory(config)

	r := chi.NewRouter()
	h := NewHandler(config, repo)
	r.Get("/ping/", h.Ping)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, bytes.NewBuffer(tt.body))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestPing_WithBrokenRepository(t *testing.T) {
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
			url:    "/ping/",
			body:   []byte(``),
			method: http.MethodGet,
			want: want{
				code:        500,
				response:    "{\"error\":\"no access repo\"}\n",
				contentType: "application/json; charset=utf-8",
			},
		},
	}
	config := getConfig()

	r := chi.NewRouter()
	h := NewHandler(config, BrokenRepoInMemory{})

	r.Get("/ping/", h.Ping)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, body := testRequest(t, ts, tt.method, tt.url, bytes.NewBuffer(tt.body))
			assert.Equal(t, tt.want.code, resp.StatusCode)
			assert.Equal(t, tt.want.contentType, resp.Header.Get("Content-type"))
			assert.Equal(t, tt.want.response, body)
			resp.Body.Close()
		})
	}
}

func TestHashFunctions(t *testing.T) {
	gaugeValue := float64(223.4)
	deltaValue := int64(865)

	tests := []struct {
		name      string
		valueJSON types.ValueJSON
		checkHash bool
	}{
		{
			name:      "case 1",
			valueJSON: types.ValueJSON{ID: "Alloc", MType: "gauge", Value: (*types.Gauge)(&gaugeValue), Hash: "370ef738b508dbb9e2217d0ce8d40e49a2c81ca45635445a8af51be5dfee1514"},
			checkHash: true,
		},
		{
			name:      "case 2",
			valueJSON: types.ValueJSON{ID: "PollCount", MType: "counter", Delta: (*types.Counter)(&deltaValue), Hash: "8580208d895ad7171645ff28624871dd50fd373a22a4c23b6edfe6905c146cc4"},
			checkHash: true,
		},
		{
			name:      "case 3",
			valueJSON: types.ValueJSON{ID: "UnknownMetric", MType: "unknown", Delta: (*types.Counter)(&deltaValue), Hash: ""},
			checkHash: false,
		},
	}
	key := "eiDagh8t"
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, calcHash(tt.valueJSON, key), tt.valueJSON.Hash)
			assert.Equal(t, checkHash(tt.valueJSON, key), tt.checkHash)
		})
	}
}
