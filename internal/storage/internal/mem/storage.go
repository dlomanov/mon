package mem

import "github.com/dlomanov/mon/internal/entities"

func NewStorage() *Storage {
	s := make(Storage)
	return &s
}

type Storage map[string]entities.Metric

func (s *Storage) Set(metrics ...entities.Metric) {
	for _, v := range metrics {
		(*s)[v.String()] = v
	}
}

func (s *Storage) Get(keys ...entities.MetricsKey) []entities.Metric {
	result := make([]entities.Metric, 0, len(keys))
	for _, k := range keys {
		v, ok := (*s)[k.String()]
		if !ok {
			continue
		}
		result = append(result, v)
	}

	return result
}

func (s *Storage) All() []entities.Metric {
	result := make([]entities.Metric, 0, len(*s))
	for _, v := range *s {
		result = append(result, v)
	}

	return result
}
