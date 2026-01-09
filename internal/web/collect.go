package web

import (
	"context"
	"io"
	"net"
	"os"
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
	// handle special case there meminfo is exposed via a socket because /proc doesnt work with docker + lxc
	path := "/proc/meminfo"
	if CopyDataFromSocket(ctx, "/meminfo.sock", "/meminfo") {
		path = "/meminfo"
	}
	mem, err := linuxproc.ReadMemInfo(path)
	log.GetLogger().Log(ctx, log.LevelTrace, "readMemInfo", "mem", mem, "err", err)
	if err != nil {
		return 0, 0, err
	}
	return mem.MemTotal * 1024, mem.MemAvailable * 1024, nil
}

func CopyDataFromSocket(ctx context.Context, from string, to string) bool {
	// check if the socket exists
	if _, err := os.Stat(from); os.IsNotExist(err) {
		return false
	}

	// Connect to the Unix socket
	conn, err := net.Dial("unix", from)
	if err != nil {
		log.GetLogger().ErrorContext(ctx, "Failed to connect to socket", "error", err, "socket", from)
		return false
	}
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "Failed to close socket connection", "error", err)
		}
	}(conn)

	// Read everything from the socket
	data, err := io.ReadAll(conn)
	if err != nil {
		log.GetLogger().ErrorContext(ctx, "Failed to read data from socket", "error", err)
		return false
	}

	// Write the data to the file using ioutil
	err = os.WriteFile(to, data, 0644)
	if err != nil {
		log.GetLogger().ErrorContext(ctx, "Failed to write data to file", "error", err)
		return false
	}

	return true
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
