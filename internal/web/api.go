package web

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/log"
)

type infoResponse struct {
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

func HandleAPIInfo(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		hn, _ := os.Hostname()
		writeJSON(w, infoResponse{Hostname: hn, Version: version})
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
		usage, usageUser, usageSystem, err := readCPUPercent()
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read cpu", "error", err)
		}
		mem, err := readMemPercent()
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "failed to read mem", "error", err)
		}
		writeJSON(w, UsageResponse{CPUPercent: usage, CPUPercentUser: usageUser, CPUPercentSystem: usageSystem, MemPercent: mem})
	}
}

type containerItem struct {
	ID           string   `json:"id"`
	Names        []string `json:"names"`
	Created      int64    `json:"created"`
	State        string   `json:"state"`
	Exited       bool     `json:"exited"`
	ExitCode     int      `json:"exit_code"`
	RestartCount int      `json:"restart_count"`
	MemUsageKiB  uint64   `json:"mem_usage_kib"`
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
			// stats might fail for exited containers; ignore errors per item
			var memKiB uint64
			if st, err := cli.GetContainerStats(ctx, c.ID); err == nil {
				memKiB = st.MemoryUsageKiB
			}
			// inspect for exit code
			var exitCode int
			var restartCount int
			if insp, err := cli.InspectContainer(ctx, c.ID); err == nil {
				exitCode = insp.ExitCode
				restartCount = insp.RestartCount
			}
			stateStr := string(c.State)
			item := containerItem{
				ID:           c.ID,
				Names:        c.Names,
				Created:      c.Created,
				State:        stateStr,
				Exited:       strings.ToLower(stateStr) == "exited",
				ExitCode:     exitCode,
				RestartCount: restartCount,
				MemUsageKiB:  memKiB,
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
