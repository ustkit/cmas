package agent

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ustkit/cmas/internal/types"
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		wantMetrics Metrics
	}{
		{
			name: "case 1",
			wantMetrics: Metrics{
				mu: &sync.Mutex{},
				Values: map[string]*types.Value{
					"Alloc":           {GValue: 0, TValue: "gauge"},
					"BuckHashSys":     {GValue: 0, TValue: "gauge"},
					"GCCPUFraction":   {GValue: 0, TValue: "gauge"},
					"GCSys":           {GValue: 0, TValue: "gauge"},
					"HeapAlloc":       {GValue: 0, TValue: "gauge"},
					"HeapIdle":        {GValue: 0, TValue: "gauge"},
					"HeapInuse":       {GValue: 0, TValue: "gauge"},
					"HeapObjects":     {GValue: 0, TValue: "gauge"},
					"HeapReleased":    {GValue: 0, TValue: "gauge"},
					"HeapSys":         {GValue: 0, TValue: "gauge"},
					"Lookups":         {GValue: 0, TValue: "gauge"},
					"MCacheInuse":     {GValue: 0, TValue: "gauge"},
					"MCacheSys":       {GValue: 0, TValue: "gauge"},
					"MSpanInuse":      {GValue: 0, TValue: "gauge"},
					"MSpanSys":        {GValue: 0, TValue: "gauge"},
					"Mallocs":         {GValue: 0, TValue: "gauge"},
					"NextGC":          {GValue: 0, TValue: "gauge"},
					"NumForcedGC":     {GValue: 0, TValue: "gauge"},
					"OtherSys":        {GValue: 0, TValue: "gauge"},
					"PauseTotalNs":    {GValue: 0, TValue: "gauge"},
					"StackInuse":      {GValue: 0, TValue: "gauge"},
					"StackSys":        {GValue: 0, TValue: "gauge"},
					"Sys":             {GValue: 0, TValue: "gauge"},
					"TotalAlloc":      {GValue: 0, TValue: "gauge"},
					"PollCount":       {CValue: 0, TValue: "counter"},
					"RandomValue":     {GValue: 0, TValue: "gauge"},
					"Frees":           {GValue: 0, TValue: "gauge"},
					"LastGC":          {GValue: 0, TValue: "gauge"},
					"NumGC":           {GValue: 0, TValue: "gauge"},
					"TotalMemory":     {GValue: 0, TValue: "gauge"},
					"FreeMemory":      {GValue: 0, TValue: "gauge"},
					"CPUutilization1": {GValue: 0, TValue: "gauge"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMetrics := NewMetrics()
			assert.Equal(t, tt.wantMetrics, gotMetrics)
		})
	}
}

func TestRuntimeUpdate(t *testing.T) {
	metrics := NewMetrics()
	metrics.RuntimeUpdate()
	counter := 0
	for _, m := range metrics.Values {
		if m.TValue == "gauge" && m.GValue != 0 {
			counter++
		}
		if m.TValue == "counter" && m.CValue != 0 {
			counter++
		}
	}
	assert.NotEqual(t, 0, counter)
}

func TestGopsutilUpdate(t *testing.T) {
	metrics := NewMetrics()
	metrics.Values["TotalMemory"].GValue = -1
	metrics.Values["FreeMemory"].GValue = -1
	metrics.Values["CPUutilization1"].GValue = -1
	err := metrics.GopsutilUpdate()
	require.NoError(t, err)
	assert.NotEqual(t, -1, metrics.Values["TotalMemory"].GValue)
	assert.NotEqual(t, -1, metrics.Values["FreeMemory"].GValue)
	assert.NotEqual(t, -1, metrics.Values["CPUutilization1"].GValue)
}

func TestCalcHash(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue types.Value
		hash        string
		key         string
	}{
		{
			name:        "case 1",
			metricName:  "Alloc",
			metricValue: types.Value{GValue: 7232459, TValue: "gauge"},
			key:         "SuperSecret007Key",
			hash:        "2cd7a0eac6aaf01da1ae757fa96ae4c6377c3cf1306bc4267378ad1eef11c66c",
		},
		{
			name:        "case 1",
			metricName:  "PollCount",
			metricValue: types.Value{CValue: 100500, TValue: "counter"},
			key:         "SuperSecret007Key",
			hash:        "d205b07d00b512d275dcdb2398118eae3444e9652a6d08549db9a5498c25b5bf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := calcHash(tt.metricName, &tt.metricValue, tt.key)
			assert.Equal(t, hash, tt.hash)
		})
	}

}

func BenchmarkGopsutilUpdate(b *testing.B) {
	triesN := 1000

	metrics := NewMetrics()

	b.ResetTimer()

	b.Run("chan", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			metrics.GopsutilUpdate()
		}
	})
}

func BenchmarkRuntimeUpdate(b *testing.B) {
	triesN := 1000

	metrics := NewMetrics()

	b.ResetTimer()

	b.Run("chan", func(b *testing.B) {
		for i := 0; i < triesN; i++ {
			metrics.RuntimeUpdate()
		}
	})
}
