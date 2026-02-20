package status

import (
	"encoding/json"
	"net/http"
	"sync/atomic"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/log"
)

var ready atomic.Bool

type Response struct {
	Status        string `json:"status"`
	DockerError   string `json:"docker_error,omitempty"`
	DockerVersion string `json:"docker_version,omitempty"`
	Version       string `json:"version"`
}

// SetReady marks the exporter as ready to serve metrics
func SetReady() {
	ready.Store(true)
}

// IsReady returns true if the exporter is ready to serve metrics
func IsReady() bool {
	return ready.Load()
}

func HandleStatus(cli *docker.Client, version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Content-Type", "application/json")

		response := Response{
			Status:        "healthy",
			DockerError:   "",
			DockerVersion: "",
			Version:       version,
		}

		// Check if ready
		if !IsReady() {
			response.Status = "starting"
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(response)
			return
		}

		// Check if Docker daemon is responding
		ver, err := cli.Ping(ctx)
		log.GetLogger().Log(ctx, log.LevelTrace, "Checking Docker daemon health", "version", ver, "error", err)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "Docker daemon is not responding", "error", err)
			response.Status = "unhealthy"
			response.DockerError = err.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			response.DockerVersion = ver
			w.WriteHeader(http.StatusOK)
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
