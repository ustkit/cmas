package agent

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"

	"github.com/ustkit/cmas/internal/agent/config"
	"github.com/ustkit/cmas/internal/types"
)

const (
	GAUGE   = "gauge"
	COUNTER = "counter"
)

type Metrics struct {
	mu     *sync.Mutex
	Values types.Values
}

func NewMetrics() (metrics Metrics) {
	metrics = Metrics{Values: make(map[string]*types.Value), mu: &sync.Mutex{}}
	metrics.Values["Alloc"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["BuckHashSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["GCCPUFraction"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["GCSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["HeapAlloc"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["HeapIdle"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["HeapInuse"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["HeapObjects"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["HeapReleased"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["HeapSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["Lookups"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["MCacheInuse"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["MCacheSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["MSpanInuse"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["MSpanSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["Mallocs"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["NextGC"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["NumForcedGC"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["OtherSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["PauseTotalNs"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["StackInuse"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["StackSys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["Sys"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["TotalAlloc"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["PollCount"] = &types.Value{CValue: 0, TValue: "counter"}
	metrics.Values["RandomValue"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["Frees"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["LastGC"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["NumGC"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["TotalMemory"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["FreeMemory"] = &types.Value{GValue: 0, TValue: "gauge"}
	metrics.Values["CPUutilization1"] = &types.Value{GValue: 0, TValue: "gauge"}

	return
}

func (metrics *Metrics) GopsutilUpdate() error {
	virtMem, err := mem.VirtualMemory()
	if err != nil {
		return err
	}

	cpuUtilization, err := cpu.Percent(time.Second, true)
	if err != nil {
		return err
	}

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.Values["TotalMemory"].GValue = types.Gauge(virtMem.Total)
	metrics.Values["FreeMemory"].GValue = types.Gauge(virtMem.Free)
	metrics.Values["CPUutilization1"].GValue = types.Gauge(cpuUtilization[0])

	return nil
}

func (metrics *Metrics) RuntimeUpdate() error {
	memStat := runtime.MemStats{}
	runtime.ReadMemStats(&memStat)

	metrics.mu.Lock()
	defer metrics.mu.Unlock()

	metrics.Values["Alloc"].GValue = types.Gauge(memStat.Alloc)
	metrics.Values["BuckHashSys"].GValue = types.Gauge(memStat.BuckHashSys)
	metrics.Values["GCCPUFraction"].GValue = types.Gauge(memStat.GCCPUFraction)
	metrics.Values["GCSys"].GValue = types.Gauge(memStat.GCSys)
	metrics.Values["HeapAlloc"].GValue = types.Gauge(memStat.HeapAlloc)
	metrics.Values["HeapIdle"].GValue = types.Gauge(memStat.HeapIdle)
	metrics.Values["HeapInuse"].GValue = types.Gauge(memStat.HeapInuse)
	metrics.Values["HeapObjects"].GValue = types.Gauge(memStat.HeapObjects)
	metrics.Values["HeapReleased"].GValue = types.Gauge(memStat.HeapReleased)
	metrics.Values["HeapSys"].GValue = types.Gauge(memStat.HeapSys)
	metrics.Values["Lookups"].GValue = types.Gauge(memStat.Lookups)
	metrics.Values["MCacheInuse"].GValue = types.Gauge(memStat.MCacheInuse)
	metrics.Values["MCacheSys"].GValue = types.Gauge(memStat.MCacheSys)
	metrics.Values["Mallocs"].GValue = types.Gauge(memStat.Mallocs)
	metrics.Values["NextGC"].GValue = types.Gauge(memStat.NextGC)
	metrics.Values["NumForcedGC"].GValue = types.Gauge(memStat.NumForcedGC)
	metrics.Values["OtherSys"].GValue = types.Gauge(memStat.OtherSys)
	metrics.Values["PauseTotalNs"].GValue = types.Gauge(memStat.PauseTotalNs)
	metrics.Values["StackInuse"].GValue = types.Gauge(memStat.StackInuse)
	metrics.Values["StackSys"].GValue = types.Gauge(memStat.StackSys)
	metrics.Values["Sys"].GValue = types.Gauge(memStat.Sys)
	metrics.Values["TotalAlloc"].GValue = types.Gauge(memStat.TotalAlloc)
	metrics.Values["Frees"].GValue = types.Gauge(memStat.Frees)
	metrics.Values["LastGC"].GValue = types.Gauge(memStat.LastGC)
	metrics.Values["NumGC"].GValue = types.Gauge(memStat.NumGC)

	return nil
}

func (metrics *Metrics) Send(ctx context.Context, client *http.Client, agentConfig *config.Config) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	rand.Seed(time.Now().UnixNano())
	metrics.Values["PollCount"].CValue++
	//nolint
	metrics.Values["RandomValue"].GValue = types.Gauge(rand.Float64())

	url := "http://" + agentConfig.Sever + "/update/"

	for name, value := range metrics.Values {
		name := name
		value := value

		go func(mName string, mValue *types.Value, url string) {
			var (
				req *http.Request
				err error
			)

			switch agentConfig.DataType {
			case "plain":
				req, err = requestPlain(ctx, mName, mValue, url)
				if err != nil {
					return
				}
			case "json":
				req, err = requestJSON(ctx, mName, mValue, url, agentConfig.Key)
				if err != nil {
					return
				}
			}

			resp, err := client.Do(req)
			if err != nil {
				return
			}

			resp.Body.Close()
		}(name, value, url)
	}
}

func (metrics *Metrics) SendBatch(ctx context.Context, client *http.Client, agentConfig *config.Config) {
	metrics.mu.Lock()
	defer metrics.mu.Unlock()
	rand.Seed(time.Now().UnixNano())
	metrics.Values["PollCount"].CValue++
	//nolint
	metrics.Values["RandomValue"].GValue = types.Gauge(rand.Float64())

	url := "http://" + agentConfig.Sever + "/updates/"

	req, err := requestJSONBatch(ctx, metrics.Values, url, agentConfig.Key)
	if err != nil {
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		return
	}

	resp.Body.Close()
}

func requestPlain(ctx context.Context, mName string, mValue *types.Value, url string) (req *http.Request, err error) {
	value := ""

	switch mValue.TValue {
	case GAUGE:
		value = strconv.FormatFloat(float64(mValue.GValue), 'f', -1, 64)
	case COUNTER:
		value = strconv.Itoa(int(mValue.CValue))
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, url+mValue.TValue+"/"+mName+"/"+value, &bytes.Buffer{})
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "text/plain")

	return
}

func requestJSON(ctx context.Context, mName string, mValue *types.Value, url, key string) (req *http.Request, err error) {
	value := &types.ValueJSON{ID: mName, MType: mValue.TValue}

	switch mValue.TValue {
	case GAUGE:
		value.Value = &mValue.GValue
	case COUNTER:
		value.Delta = &mValue.CValue
	}

	if key != "" {
		value.Hash = calcHash(mName, mValue, key)
	}

	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)

	err = encoder.Encode(value)
	if err != nil {
		return
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	return
}

func requestJSONBatch(ctx context.Context, metrics types.Values, url, key string) (req *http.Request, err error) {
	values := make([]types.ValueJSON, 0, len(metrics))

	for name, value := range metrics {
		valueJSON := types.ValueJSON{ID: name, MType: value.TValue, Delta: &value.CValue, Value: &value.GValue}

		if key != "" {
			valueJSON.Hash = calcHash(name, value, key)
		}

		values = append(values, valueJSON)
	}

	body := &bytes.Buffer{}
	encoder := json.NewEncoder(body)

	err = encoder.Encode(values)
	if err != nil {
		return
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	return
}

func calcHash(mName string, mValue *types.Value, key string) string {
	h := hmac.New(sha256.New, []byte(key))

	switch mValue.TValue {
	case GAUGE:
		fmt.Fprintf(h, "%s:gauge:%f", mName, mValue.GValue)
	case COUNTER:
		fmt.Fprintf(h, "%s:counter:%d", mName, mValue.CValue)
	}

	return hex.EncodeToString(h.Sum(nil))
}
