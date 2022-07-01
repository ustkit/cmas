package types

type Counter int64

type Gauge float64

type Value struct {
	CValue Counter `json:"delta,omitempty"`
	GValue Gauge   `json:"value,omitempty"`
	TValue string  `json:"type"`
}

type Values map[string]*Value

type ValueJSON struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Delta *Counter `json:"delta,omitempty"`
	Value *Gauge   `json:"value,omitempty"`
	Hash  string   `json:"hash,omitempty"`
}
