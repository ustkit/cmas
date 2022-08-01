package repositories

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/types"
)

func getConfig() *config.Config {
	serverConfig := &config.Config{}
	serverConfig.Address = "localhost:8080"
	serverConfig.Restore = true
	serverConfig.StoreInterval = "0"
	serverConfig.StoreFile = "cmas-metrics-db.json"

	return serverConfig
}

func TestRepoInMemory_Save(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue types.Value
		wantErrSave error
		wantErrFind error
	}{
		{
			name:        "case 1",
			metricName:  "Alloc",
			metricValue: types.Value{GValue: 234.12, TValue: "gauge"},
			wantErrSave: nil,
			wantErrFind: nil,
		},
		{
			name:        "case 2",
			metricName:  "Alloc",
			metricValue: types.Value{GValue: 234.12, TValue: "gauge"},
			wantErrSave: nil,
			wantErrFind: nil,
		},
		{
			name:        "case 3",
			metricName:  "PollCount",
			metricValue: types.Value{CValue: 589, TValue: "counter"},
			wantErrSave: nil,
			wantErrFind: nil,
		},
	}

	mr := NewRepositoryInMemory(getConfig())
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mr.Save(ctx, tt.metricName, tt.metricValue)
			assert.Equal(t, tt.wantErrSave, err)
			value, err := mr.FindByName(ctx, tt.metricName)
			assert.Equal(t, tt.wantErrFind, err)
			if err == nil {
				assert.Equal(t, value.GValue, tt.metricValue.GValue)
			}
		})
	}
}

func TestRepoInMemory_SaveAll(t *testing.T) {
	testValue := types.Gauge(234.12)
	testDelta := types.Counter(100500)
	tests := []struct {
		name        string
		metrics     []types.ValueJSON
		metricNames []string
		wantErrSave error
		wantErrFind error
	}{
		{
			name: "case 1",
			metrics: []types.ValueJSON{
				{ID: "Alloc", MType: "gauge", Value: &testValue, Delta: &testDelta},
				{ID: "PollCount", MType: "counter", Value: &testValue, Delta: &testDelta},
			},
			metricNames: []string{"Alloc", "PollCount"},
			wantErrSave: nil,
			wantErrFind: nil,
		},
	}

	mr := NewRepositoryInMemory(getConfig())
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mr.SaveAll(ctx, tt.metrics)
			assert.Equal(t, tt.wantErrSave, err)
			err = mr.SaveAll(ctx, tt.metrics)
			assert.Equal(t, tt.wantErrSave, err)
			for _, mn := range tt.metricNames {
				value, err := mr.FindByName(ctx, mn)
				assert.Equal(t, nil, err)
				assert.Equal(t, value.GValue, testValue)
				assert.Equal(t, value.CValue, 2*testDelta)
			}
		})
	}
}

func TestRepoInMemory_FindByName(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		metricValue types.Value
	}{
		{
			name:        "case 1",
			metricName:  "Unknow",
			metricValue: types.Value{},
		},
	}

	mr := NewRepositoryInMemory(getConfig())
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := mr.FindByName(ctx, tt.metricName)
			assert.NotNil(t, err)
			assert.Equal(t, tt.metricValue, value)
		})
	}
}

func TestRepoInMemory_FindAll(t *testing.T) {
	tests := []struct {
		name        string
		saveMetrics types.Values
		findMetrics types.Values
		wantErr     error
	}{
		{
			name: "case 1",
			saveMetrics: types.Values{
				"Alloc":     &types.Value{GValue: 234.12, TValue: "gauge"},
				"PollCount": &types.Value{GValue: 589.0, TValue: "counter"},
			},
			findMetrics: types.Values{
				"Alloc":     &types.Value{GValue: 234.12, TValue: "gauge"},
				"PollCount": &types.Value{GValue: 589.0, TValue: "counter"},
			},
			wantErr: nil,
		},
	}

	mr := NewRepositoryInMemory(getConfig())
	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for metricName, metricValue := range tt.saveMetrics {
				err := mr.Save(ctx, metricName, *metricValue)
				require.Nil(t, err)
			}
			findMetrics, err := mr.FindAll(ctx)
			assert.NoError(t, tt.wantErr, err)
			assert.Equal(t, tt.findMetrics, findMetrics)
		})
	}
}

func TestRepoInMemory_Resory(t *testing.T) {
	config := getConfig()
	mr := NewRepositoryInMemory(config)
	err := mr.Restore()
	require.Nil(t, err)
	os.Remove(config.StoreFile)
}

func TestRepoInMemory_Ping(t *testing.T) {
	mr := NewRepositoryInMemory(getConfig())
	err := mr.Ping(context.Background())
	require.Nil(t, err)
}

func TestRepoInMemory_Close(t *testing.T) {
	mr := NewRepositoryInMemory(getConfig())
	err := mr.Close()
	require.Nil(t, err)
}
