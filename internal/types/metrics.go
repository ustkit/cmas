package types

// Counter тип целочисленной метрики.
type Counter int64

// Gauge тип метрики с действительным значением.
type Gauge float64

// Value структура значения метрики.
type Value struct {
	CValue Counter `json:"delta,omitempty"`
	GValue Gauge   `json:"value,omitempty"`
	TValue string  `json:"type"`
}

// Values карта значений метрик с именами метрик как ключи.
type Values map[string]*Value

// ValueJSON структура метрики для передачи на сервер.
type ValueJSON struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *Counter `json:"delta,omitempty"`
	Value *Gauge   `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}

// RequestValueJSON структура метрики для запроса её значения с сервера.
type RequestValueJSON struct {
	ID    string `json:"id"`
	MType string `json:"type"`
}
