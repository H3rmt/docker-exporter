package web

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/log"
	"github.com/moby/moby/api/types/container"
)

type infoResponse struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
	HostIP   string `json:"host_ip"`
}

func HandleAPIInfo(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hn, _ := os.ReadFile("/etc/hostname")
		hostname := strings.TrimSpace(string(hn))
		hostIP := os.Getenv("IP")
		if hostIP == "" {
			hostIP = "???"
		}
		writeJSON(w, infoResponse{Hostname: hostname, Version: version, HostIP: hostIP})
	}
}

type UsageResponse struct {
	CPUPercent       float64 `json:"cpu_percent"`
	CPUPercentUser   float64 `json:"cpu_percent_user"`
	CPUPercentSystem float64 `json:"cpu_percent_system"`
	MemPercent       float64 `json:"mem_percent"`
}

func HandleAPIUsage() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		usage, usageUser, usageSystem, err := readCPUInfo(ctx, 500*time.Millisecond)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read cpu", "error", err)
		}
		mem, err := readMemPercent(ctx)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read mem", "error", err)
		}
		writeJSON(w, UsageResponse{CPUPercent: usage, CPUPercentUser: usageUser, CPUPercentSystem: usageSystem, MemPercent: mem})
	}
}

type containerItem struct {
	ID              string   `json:"id"`
	Names           []string `json:"names"`
	Created         int64    `json:"created"`
	State           string   `json:"state"`
	Exited          bool     `json:"exited"`
	ExitCode        int      `json:"exit_code"`
	RestartCount    int      `json:"restart_count"`
	MemUsageKiB     uint64   `json:"mem_usage_kib"`
	MemLimitKiB     uint64   `json:"mem_limit_kib"`
	CpuUsage        uint64   `json:"cpu_usage"`
	MaxCpus         float64  `json:"max_cpus"`
	CpuLimitedUsage uint64   `json:"cpu_limited_usage"`
	MaxLimitedCpus  float64  `json:"max_limited_cpus"`
}

func HandleAPIContainers(cli *docker.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		var items []containerItem
		containers, err := cli.ListAllRunningContainers(ctx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, c := range containers {
			// inspect for exit code
			var exitCode int
			var restartCount int
			var nanoCpus int64
			if insp, err := cli.InspectContainer(ctx, c.ID, false); err == nil {
				log.GetLogger().Log(ctx, log.LevelTrace, "inspect container", "inspect", insp)
				exitCode = insp.ExitCode
				restartCount = insp.RestartCount
				nanoCpus = insp.NanoCpus
			}

			// stats might fail for exited containers; ignore errors per item
			var memKiB uint64
			var memLimitKiB uint64
			var cpuPercentOfSystem uint64
			var maxCPUs float64
			var maxLimitedCpus float64
			if c.State == container.StateRunning {
				if st, err := cli.GetContainerStats(ctx, c.ID); err == nil {
					memKiB = st.MemoryUsageKiB
					memLimitKiB = st.MemoryLimitKiB

					// Compute CPU% of system (docker stats style)
					cpuDelta := st.Cpu.UsageNS - st.Cpu.PreUsageNS
					sysDelta := st.Cpu.SystemUsageNS - st.Cpu.PreSystemUsageNS
					maxCPUs = float64(st.Cpu.OnlineCpus)
					if nanoCpus > 0 {
						maxLimitedCpus = float64(nanoCpus) / 1000000000.0
					}
					if sysDelta > 0 {
						cpuPercentOfSystem = uint64(float64(cpuDelta) / float64(sysDelta) * 100.0)
					}
				}
			}
			stateStr := string(c.State)
			item := containerItem{
				ID:              c.ID,
				Names:           c.Names,
				Created:         c.Created,
				State:           stateStr,
				Exited:          strings.ToLower(stateStr) == "exited",
				ExitCode:        exitCode,
				RestartCount:    restartCount,
				MemUsageKiB:     memKiB,
				MemLimitKiB:     memLimitKiB,
				CpuUsage:        cpuPercentOfSystem,
				MaxCpus:         maxCPUs,
				CpuLimitedUsage: uint64((float64(cpuPercentOfSystem) / maxLimitedCpus) * maxCPUs),
				MaxLimitedCpus:  maxLimitedCpus,
			}
			items = append(items, item)
		}
		writeJSON(w, items)
	})
}

// Helpers
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(v)
}
