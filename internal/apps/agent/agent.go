package agent

import (
	"github.com/dlomanov/mon/internal/apps/agent/collector"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"
)

func Run(cfg Config) (err error) {
	m := collector.NewCollector(cfg.Addr, log.Default())
	reportTime := time.Now().Add(cfg.ReportInterval)

	for i := 0; i < math.MaxInt64; i++ {
		ms := runtime.MemStats{}
		runtime.ReadMemStats(&ms)

		m.UpdateGauge("Alloc", float64(ms.Alloc))
		m.UpdateGauge("BuckHashSys", float64(ms.BuckHashSys))
		m.UpdateGauge("Frees", float64(ms.Frees))
		m.UpdateGauge("GCCPUFraction", ms.GCCPUFraction)
		m.UpdateGauge("GCSys", float64(ms.GCSys))
		m.UpdateGauge("HeapAlloc", float64(ms.HeapAlloc))
		m.UpdateGauge("HeapIdle", float64(ms.HeapIdle))
		m.UpdateGauge("HeapInuse", float64(ms.HeapInuse))
		m.UpdateGauge("HeapObjects", float64(ms.HeapObjects))
		m.UpdateGauge("HeapReleased", float64(ms.HeapReleased))
		m.UpdateGauge("HeapSys", float64(ms.HeapSys))
		m.UpdateGauge("LastGC", float64(ms.LastGC))
		m.UpdateGauge("Lookups", float64(ms.Lookups))
		m.UpdateGauge("MCacheInuse", float64(ms.MCacheInuse))
		m.UpdateGauge("MCacheSys", float64(ms.MCacheSys))
		m.UpdateGauge("MSpanInuse", float64(ms.MSpanInuse))
		m.UpdateGauge("MSpanSys", float64(ms.MSpanSys))
		m.UpdateGauge("Mallocs", float64(ms.Mallocs))
		m.UpdateGauge("NextGC", float64(ms.NextGC))
		m.UpdateGauge("NumForcedGC", float64(ms.NumForcedGC))
		m.UpdateGauge("NumGC", float64(ms.NumGC))
		m.UpdateGauge("OtherSys", float64(ms.OtherSys))
		m.UpdateGauge("PauseTotalNs", float64(ms.PauseTotalNs))
		m.UpdateGauge("StackInuse", float64(ms.StackInuse))
		m.UpdateGauge("StackSys", float64(ms.StackSys))
		m.UpdateGauge("Sys", float64(ms.Sys))
		m.UpdateGauge("TotalAlloc", float64(ms.TotalAlloc))
		m.UpdateGauge("RandomValue", rand.Float64())
		m.UpdateCounter("PollCount", 1)
		m.LogUpdated()

		if reportTime.Compare(time.Now()) <= 0 {
			reportTime = time.Now().Add(cfg.ReportInterval)
			m.ReportMetrics()
		}

		time.Sleep(cfg.PollInterval)
	}
	return
}
