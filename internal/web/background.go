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

type DataPoint struct {
	Time time.Time
	Data UsageResponse
}

func CollectInBg() {
	ctx := context.Background()
	data = make([]DataPoint, DataPoints)
	usage, usageUser, usageSystem, err := readCPUInfo(ctx, 1000*time.Millisecond)
	if err != nil {
		log.GetLogger().Error("failed to read cpu", "error", err)
	}
	mem, err := readMemPercent(ctx)
	if err != nil {
		log.GetLogger().Error("failed to read ram", "error", err)
	}
	now := time.Now()
	for i := 0; i < len(data); i++ {
		data[i] = DataPoint{Time: now.Add(-time.Duration(len(data)-i) * 10 * time.Second), Data: UsageResponse{CPUPercent: usage, CPUPercentUser: usageUser, CPUPercentSystem: usageSystem, MemPercent: mem}}
	}

	if ticker != nil {
		ticker.Stop()
	}
	ticker = time.NewTicker(Delay)

	index := 0
	for range ticker.C {
		//log.GetLogger().Debug("collecting data")
		usage, usageUser, usageSystem, err := readCPUInfo(ctx, Delay/2)
		if err != nil {
			log.GetLogger().Error("failed to read cpu", "error", err)
		}
		mem, err := readMemPercent(ctx)
		if err != nil {
			log.GetLogger().Error("failed to read ram", "error", err)
		}
		now := time.Now()
		data[index] = DataPoint{Time: now, Data: UsageResponse{CPUPercent: usage, CPUPercentUser: usageUser, CPUPercentSystem: usageSystem, MemPercent: mem}}
		index = (index + 1) % DataPoints
	}
}

func StopCollect() {
	ticker.Stop()
}

func GetData() []DataPoint {
	return data
}
