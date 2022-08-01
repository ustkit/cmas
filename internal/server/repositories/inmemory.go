package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/ustkit/cmas/internal/server/config"
	"github.com/ustkit/cmas/internal/types"
)

// RepoInMemory структура In-memory репозитория.
type RepoInMemory struct {
	mutex   *sync.RWMutex
	storage types.Values

	config *config.Config
}

// NewRepositoryInMemory возвращает структуру RepoInMemory.
func NewRepositoryInMemory(serverConfig *config.Config) RepoInMemory {
	return RepoInMemory{
		mutex:   &sync.RWMutex{},
		storage: make(types.Values),

		config: serverConfig,
	}
}

// Save сохраняет значение value метрики с именем name в репозитории.
func (mr RepoInMemory) Save(ctx context.Context, name string, value types.Value) error {
	mr.mutex.Lock()

	if _, ok := mr.storage[name]; !ok {
		mr.storage[name] = &value
		mr.mutex.Unlock()

		return nil
	}

	mr.storage[name].CValue += value.CValue
	mr.storage[name].GValue = value.GValue
	mr.storage[name].TValue = value.TValue
	mr.mutex.Unlock()

	if mr.config.StoreInterval == "0" {
		return mr.SaveToFile()
	}

	return nil
}

// SaveAll сохраняет значения метрик в репозитории.
func (mr RepoInMemory) SaveAll(ctx context.Context, values []types.ValueJSON) error {
	mr.mutex.Lock()

	for _, value := range values {
		var (
			delta types.Counter
			gauge types.Gauge
		)

		if value.Delta != nil {
			delta = *value.Delta
		}

		if value.Value != nil {
			gauge = *value.Value
		}

		if _, ok := mr.storage[value.ID]; !ok {
			mr.storage[value.ID] = &types.Value{TValue: value.MType, CValue: delta, GValue: gauge}

			continue
		}

		mr.storage[value.ID].CValue += *value.Delta
		mr.storage[value.ID].GValue = *value.Value
		mr.storage[value.ID].TValue = value.MType
	}

	mr.mutex.Unlock()

	if mr.config.StoreInterval == "0" {
		return mr.SaveToFile()
	}

	return nil
}

// FindByName находит метрику по имени name в репозитории.
func (mr RepoInMemory) FindByName(ctx context.Context, name string) (types.Value, error) {
	mr.mutex.RLock()
	defer mr.mutex.RUnlock()
	value, ok := mr.storage[name]

	if !ok {
		return types.Value{}, fmt.Errorf("metric %q not found", name)
	}

	return *value, nil
}

// FindAll возвращает все метрики из репозитория.
func (mr RepoInMemory) FindAll(ctx context.Context) (values types.Values, err error) {
	mr.mutex.RLock()
	values = mr.storage
	mr.mutex.RUnlock()

	return values, nil
}

// Restore восстанавливает метрики в репозитрии из файла заданого в Config.StoreFile.
func (mr RepoInMemory) Restore() (err error) {
	if !mr.config.Restore || mr.config.StoreFile == "" {
		return nil
	}

	file, err := os.Open(mr.config.StoreFile)
	if err != nil {
		return err
	}

	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&mr.storage)
	if err != nil {
		return err
	}

	return nil
}

// SaveToFile сохраняет метрики в файл заданный в Config.StoreFile.
func (mr RepoInMemory) SaveToFile() (err error) {
	file, err := os.Create(mr.config.StoreFile)
	if err != nil {
		return
	}

	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	encoder := json.NewEncoder(file)

	mr.mutex.RLock()
	err = encoder.Encode(mr.storage)
	mr.mutex.RUnlock()

	return
}

// Close закрывает репозиторий и высвобождает его ресурсы.
func (mr RepoInMemory) Close() error {
	return nil
}

// Ping возвращает непустую ошибку если репозитория работает нештатно.
func (mr RepoInMemory) Ping(ctx context.Context) error {
	return nil
}
