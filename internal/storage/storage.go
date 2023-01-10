package storage

type Gauge float64
type Counter uint64

// Storagier rules for interaction with the repository
type Storagier interface {
	SetStats(string, any, string)
	TakeStats() (map[string]Gauge, map[string]Counter)
	TakeThisStat(string) any
}

// MetricStorage storage for runtime metric
type MetricStorage struct {
	Metrics map[string]Metric
}

type Metric struct {
	ID    string   `json:"id"`
	MType string   `json:"type"`
	Hash  string   `json:"hash,omitempty"`
	Delta *Counter `json:"delta,omitempty"`
	Value *Gauge   `json:"value,omitempty"`
}

// NewStorage storage type constructor
func NewStorage() *MetricStorage {
	var m MetricStorage

	m.Metrics = make(map[string]Metric)
	return &m
}

// SetStat storing metrics in storage
func (m *MetricStorage) SetStat(metric *Metric) {
	if metric.MType == "gauge" {
		m.Metrics[metric.ID] = *metric
	}
	if metric.MType == "counter" {
		tmp := m.Metrics[metric.ID].Delta
		if tmp == nil {
			m.Metrics[metric.ID] = *metric
			return
		}

		*metric.Delta = *tmp + *metric.Delta
		m.Metrics[metric.ID] = *metric
	}
}

// StatStatusM (name) return: value
func (m *MetricStorage) StatStatusM(name, mType string) (value any) {
	if mType == "gauge" {
		if _, ok := m.Metrics[name]; ok {
			return *m.Metrics[name].Value
		}
	}
	if mType == "counter" {
		if _, ok := m.Metrics[name]; ok {
			return *m.Metrics[name].Delta
		}
	}

	return nil
}
