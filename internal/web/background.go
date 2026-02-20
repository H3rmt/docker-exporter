package web

import (
	"context"
	"time"

	"github.com/h3rmt/docker-exporter/internal/log"
)

const Delay = 20 * time.Second
const DataPoints = 30

var ticker *time.Ticker

var data []DataPoint

// fill with dummy data
func init() {
	data = make([]DataPoint, DataPoints)
	now := time.Now()
	for i := 0; i < len(data); i++ {
		data[i] = DataPoint{Time: now.Add(-time.Duration(len(data)-i) * 10 * time.Second), Data: UsageResponse{CPUPercent: 0, CPUPercentUser: 0, CPUPercentSystem: 0, MemPercent: 0}}
	}
}

type DataPoint struct {
	Time time.Time
	Data UsageResponse
}

func CollectInBg() {
	ctx := context.Background()
	if ticker != nil {
		ticker.Stop()
	}
	ticker = time.NewTicker(Delay)

	for range ticker.C {
		// calculate cpu usage over long period to be more accurate
		usage, usageUser, usageSystem, err := readCPUInfo(ctx, Delay/2)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read cpu", "error", err)
		}
		mem, err := readMemPercent(ctx)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read ram", "error", err)
		}
		now := time.Now()
		data = append(data, DataPoint{Time: now, Data: UsageResponse{CPUPercent: usage, CPUPercentUser: usageUser, CPUPercentSystem: usageSystem, MemPercent: mem}})
		if len(data) > DataPoints {
			data = data[1:]
		}
	}
}

func StopCollect() {
	if ticker == nil {
		return
	}
	ticker.Stop()
}

func GetData() []DataPoint {
	return data
}
