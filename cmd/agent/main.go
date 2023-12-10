package main

import (
	"github.com/dlomanov/mon/internal/handlers/metrics/counter"
	"github.com/dlomanov/mon/internal/handlers/metrics/gauge"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"
)

const (
	pollInterval   = time.Second * 2
	reportInterval = time.Second * 10
)

func main() {
	m := NewMon(log.Default())
	reportTime := time.Now().Add(reportInterval)

	for i := 0; i < math.MaxInt64; i++ {
		ms := runtime.MemStats{}
		runtime.ReadMemStats(&ms)

		m.UpdateGauge(gauge.Metric{Name: "Alloc", Value: float64(ms.Alloc)})
		m.UpdateGauge(gauge.Metric{Name: "BuckHashSys", Value: float64(ms.BuckHashSys)})
		m.UpdateGauge(gauge.Metric{Name: "Frees", Value: float64(ms.Frees)})
		m.UpdateGauge(gauge.Metric{Name: "GCCPUFraction", Value: float64(ms.GCCPUFraction)})
		m.UpdateGauge(gauge.Metric{Name: "GCSys", Value: float64(ms.GCSys)})
		m.UpdateGauge(gauge.Metric{Name: "HeapAlloc", Value: float64(ms.HeapAlloc)})
		m.UpdateGauge(gauge.Metric{Name: "HeapIdle", Value: float64(ms.HeapIdle)})
		m.UpdateGauge(gauge.Metric{Name: "HeapInuse", Value: float64(ms.HeapInuse)})
		m.UpdateGauge(gauge.Metric{Name: "HeapObjects", Value: float64(ms.HeapObjects)})
		m.UpdateGauge(gauge.Metric{Name: "HeapReleased", Value: float64(ms.HeapReleased)})
		m.UpdateGauge(gauge.Metric{Name: "HeapSys", Value: float64(ms.HeapSys)})
		m.UpdateGauge(gauge.Metric{Name: "LastGC", Value: float64(ms.LastGC)})
		m.UpdateGauge(gauge.Metric{Name: "Lookups", Value: float64(ms.Lookups)})
		m.UpdateGauge(gauge.Metric{Name: "MCacheInuse", Value: float64(ms.MCacheInuse)})
		m.UpdateGauge(gauge.Metric{Name: "MCacheSys", Value: float64(ms.MCacheSys)})
		m.UpdateGauge(gauge.Metric{Name: "MSpanInuse", Value: float64(ms.MSpanInuse)})
		m.UpdateGauge(gauge.Metric{Name: "MSpanSys", Value: float64(ms.MSpanSys)})
		m.UpdateGauge(gauge.Metric{Name: "Mallocs", Value: float64(ms.Mallocs)})
		m.UpdateGauge(gauge.Metric{Name: "NextGC", Value: float64(ms.NextGC)})
		m.UpdateGauge(gauge.Metric{Name: "NumForcedGC", Value: float64(ms.NumForcedGC)})
		m.UpdateGauge(gauge.Metric{Name: "NumGC", Value: float64(ms.NumGC)})
		m.UpdateGauge(gauge.Metric{Name: "OtherSys", Value: float64(ms.OtherSys)})
		m.UpdateGauge(gauge.Metric{Name: "PauseTotalNs", Value: float64(ms.PauseTotalNs)})
		m.UpdateGauge(gauge.Metric{Name: "StackInuse", Value: float64(ms.StackInuse)})
		m.UpdateGauge(gauge.Metric{Name: "StackSys", Value: float64(ms.StackSys)})
		m.UpdateGauge(gauge.Metric{Name: "Sys", Value: float64(ms.Sys)})
		m.UpdateGauge(gauge.Metric{Name: "TotalAlloc", Value: float64(ms.TotalAlloc)})
		m.UpdateGauge(gauge.Metric{Name: "RandomValue", Value: rand.Float64()})
		m.UpdateCounter(counter.Metric{Name: "PollCount", Value: 1})
		m.Updated()

		if reportTime.Compare(time.Now()) <= 0 {
			reportTime = time.Now().Add(reportInterval)
			m.ReportMetrics()
		}

		time.Sleep(pollInterval)
	}
}
