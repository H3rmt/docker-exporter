package web

import (
	"context"
	"time"

	linuxproc "github.com/c9s/goprocinfo/linux"
	"github.com/h3rmt/docker-exporter/internal/log"
)

func readMemPercent(ctx context.Context) (float64, error) {
	total, avail, err := readMemInfo(ctx)
	if err != nil {
		return 0, err
	}
	if total == 0 {
		return 0, nil
	}
	used := total - avail
	return (float64(used) / float64(total)) * 100.0, nil
}

func readMemInfo(ctx context.Context) (uint64, uint64, error) {
	mem, err := linuxproc.ReadMemInfo("/proc/meminfo")
	log.GetLogger().Log(ctx, log.LevelTrace, "readMemInfo", "mem", mem, "err", err)
	if err != nil {
		return 0, 0, err
	}
	return mem.MemTotal * 1024, mem.MemAvailable * 1024, nil
}

// readCPUInfo computes a short-sampled CPU usage percent using /proc/stat
func readCPUInfo(ctx context.Context, measureDuration time.Duration) (float64, float64, float64, error) {
	user0, system0, idle0, total0, _, err := readProcStat(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	time.Sleep(measureDuration)
	user1, system1, idle1, total1, _, err := readProcStat(ctx)
	if err != nil {
		return 0, 0, 0, err
	}
	user := float64(user1 - user0)
	system := float64(system1 - system0)
	idle := float64(idle1 - idle0)
	total := float64(total1 - total0)
	if total <= 0 {
		return 0, 0, 0, nil
	}

	usage := (1.0 - idle/total) * 100.0
	if usage < 0 {
		usage = 0
	}
	if usage > 100 {
		usage = 100
	}

	usageUser := (user / total) * 100.0
	if usageUser < 0 {
		usageUser = 0
	}
	if usageUser > 100 {
		usageUser = 100
	}

	usageSystem := (system / total) * 100.0
	if usageSystem < 0 {
		usageSystem = 0
	}
	if usageSystem > 100 {
		usageSystem = 100
	}
	return usage, usageUser, usageSystem, nil
}

func readProcStat(ctx context.Context) (uint64, uint64, uint64, uint64, uint64, error) {
	stat, err := linuxproc.ReadStat("/proc/stat")
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	s := stat.CPUStatAll
	c := len(stat.CPUStats)

	log.GetLogger().Log(ctx, log.LevelTrace, "readProcStat", "stat", s, "cpus", c)
	return s.User, s.System, s.Idle + s.IOWait, s.User + s.Nice + s.System + s.Idle + s.IOWait + s.IRQ + s.SoftIRQ + s.Steal + s.Guest + s.GuestNice, uint64(c), nil
}
