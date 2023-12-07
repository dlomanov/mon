package handlers

import (
	"fmt"
	"github.com/dlomanov/mon/internal/metrics"
	"github.com/dlomanov/mon/internal/storage"
	"strconv"
)

func HandleGauge(metric metrics.Metric, db storage.Storage) error {
	_, err := strconv.ParseFloat(metric.Value, 64)
	if err != nil {
		return newValidationError("gauge type of value should be float")
	}

	db.Set(metric.Key(), metric.Value)
	return nil
}

func HandleCounter(metric metrics.Metric, db storage.Storage) error {
	valueInt, err := strconv.ParseInt(metric.Value, 10, 64)
	if err != nil {
		return newValidationError("counter type of value should be int")
	}

	old, ok := db.Get(metric.Key())
	if ok != true {
		db.Set(metric.Key(), metric.Value)
		return nil
	}

	oldInt, err := strconv.ParseInt(old, 10, 64)
	if err != nil {
		return err
	}

	newValueInt := valueInt + oldInt
	newValue := fmt.Sprintf("%d", newValueInt)
	db.Set(metric.Key(), newValue)
	return nil
}
