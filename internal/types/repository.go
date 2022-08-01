package types

import "context"

// MetricRepo интерфейс репозитория метрик.
type MetricRepo interface {
	Save(context.Context, string, Value) error
	SaveAll(context.Context, []ValueJSON) error
	FindByName(context.Context, string) (Value, error)
	FindAll(context.Context) (Values, error)
	Restore() error
	SaveToFile() error
	Close() error
	Ping(context.Context) error
}
