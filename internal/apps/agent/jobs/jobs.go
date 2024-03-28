package jobs

import (
	"context"
	"log"
	"math"
	"math/rand"
	"runtime"
	"time"

	"github.com/dlomanov/mon/internal/apps/agent/collector"
	"github.com/dlomanov/mon/internal/entities"
	"github.com/shirou/gopsutil/v3/mem"
	"go.uber.org/zap"
)

type Report func(map[string]entities.Metric)

func CollectMetrics(
	ctx context.Context,
	cfg collector.Config,
	logger *zap.Logger,
	report Report,
) {
	c := collector.NewCollector(logger)
	reportTime := time.Now().Add(cfg.ReportInterval)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for i := 0; i < math.MaxInt64 && ctx.Err() == nil; i++ {
		ms := runtime.MemStats{}
		runtime.ReadMemStats(&ms)

		c.UpdateGauge("Alloc", float64(ms.Alloc))
		c.UpdateGauge("BuckHashSys", float64(ms.BuckHashSys))
		c.UpdateGauge("Frees", float64(ms.Frees))
		c.UpdateGauge("GCCPUFraction", ms.GCCPUFraction)
		c.UpdateGauge("GCSys", float64(ms.GCSys))
		c.UpdateGauge("HeapAlloc", float64(ms.HeapAlloc))
		c.UpdateGauge("HeapIdle", float64(ms.HeapIdle))
		c.UpdateGauge("HeapInuse", float64(ms.HeapInuse))
		c.UpdateGauge("HeapObjects", float64(ms.HeapObjects))
		c.UpdateGauge("HeapReleased", float64(ms.HeapReleased))
		c.UpdateGauge("HeapSys", float64(ms.HeapSys))
		c.UpdateGauge("LastGC", float64(ms.LastGC))
		c.UpdateGauge("Lookups", float64(ms.Lookups))
		c.UpdateGauge("MCacheInuse", float64(ms.MCacheInuse))
		c.UpdateGauge("MCacheSys", float64(ms.MCacheSys))
		c.UpdateGauge("MSpanInuse", float64(ms.MSpanInuse))
		c.UpdateGauge("MSpanSys", float64(ms.MSpanSys))
		c.UpdateGauge("Mallocs", float64(ms.Mallocs))
		c.UpdateGauge("NextGC", float64(ms.NextGC))
		c.UpdateGauge("NumForcedGC", float64(ms.NumForcedGC))
		c.UpdateGauge("NumGC", float64(ms.NumGC))
		c.UpdateGauge("OtherSys", float64(ms.OtherSys))
		c.UpdateGauge("PauseTotalNs", float64(ms.PauseTotalNs))
		c.UpdateGauge("StackInuse", float64(ms.StackInuse))
		c.UpdateGauge("StackSys", float64(ms.StackSys))
		c.UpdateGauge("Sys", float64(ms.Sys))
		c.UpdateGauge("TotalAlloc", float64(ms.TotalAlloc))
		c.UpdateGauge("RandomValue", rand.Float64())
		c.UpdateCounter("PollCount", 1)
		c.LogUpdated()

		if reportTime.Compare(time.Now()) <= 0 {
			reportTime = time.Now().Add(cfg.ReportInterval)
			report(c.Metrics)
		}

		select {
		case <-ctx.Done():
			continue
		case <-ticker.C:
			continue
		}
	}

	logger.Debug("collect cancelled: %s\n", zap.Error(ctx.Err()))
}

func CollectAdvancedMetrics(
	ctx context.Context,
	cfg collector.Config,
	logger *zap.Logger,
	report Report,
) {
	c := collector.NewCollector(logger)
	reportTime := time.Now().Add(cfg.ReportInterval)

	ticker := time.NewTicker(cfg.PollInterval)
	defer ticker.Stop()

	for ctx.Err() == nil {
		v, err := mem.VirtualMemoryWithContext(ctx)
		if err != nil {
			log.Printf("error occured while collecting advanced metrics: %s\n", v)
		} else {
			c.UpdateGauge("TotalMemory", float64(v.Total))
			c.UpdateGauge("FreeMemory", float64(v.Free))
			c.UpdateGauge("CPUutilization1", v.UsedPercent)
			c.LogUpdated()
		}

		if reportTime.Compare(time.Now()) <= 0 {
			reportTime = time.Now().Add(cfg.ReportInterval)
			report(c.Metrics)
		}

		select {
		case <-ctx.Done():
			continue
		case <-ticker.C:
			continue
		}
	}

	logger.Debug("collect cancelled: %s\n", zap.Error(ctx.Err()))
}
