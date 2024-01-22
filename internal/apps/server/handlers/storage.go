package handlers

import "github.com/dlomanov/mon/internal/entities"

type Storage interface {
	Set(metrics ...entities.Metric)
	Get(key entities.MetricsKey) (metric entities.Metric, ok bool)
	All() []entities.Metric
}
