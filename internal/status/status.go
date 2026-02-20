package status

import (
	"encoding/json"
	"net/http"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/glob"
	"github.com/h3rmt/docker-exporter/internal/log"
)

type Response struct {
	Status        string            `json:"status"`
	Errors        map[string]string `json:"errors,omitempty"`
	DockerError   string            `json:"dockerError,omitempty"`
	DockerVersion string            `json:"dockerVersion,omitempty"`
	Version       string            `json:"version"`
}

func HandleStatus(cli *docker.Client, version string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Content-Type", "application/json")

		response := Response{
			Status:        "healthy",
			Errors:        glob.GetErrorDescriptions(),
			DockerError:   "",
			DockerVersion: "",
			Version:       version,
		}

		// Check if Docker daemon is responding
		ver, err := cli.Ping(ctx)
		log.GetLogger().Log(ctx, log.LevelTrace, "Checking Docker daemon health", "version", ver, "error", err)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "Docker daemon is not responding", "error", err)
			response.Status = "unhealthy"
			response.DockerError = err.Error()
		} else {
			response.DockerVersion = ver
		}

		// Check if healthy
		if !glob.IsOk() {
			response.Status = "unhealthy"
		}

		// Check if ready
		if !glob.IsReady() {
			response.Status = "starting"
		}

		switch response.Status {
		case "healthy":
			w.WriteHeader(http.StatusOK)
			break
		case "starting":
			w.WriteHeader(http.StatusServiceUnavailable)
			break
		case "unhealthy":
			w.WriteHeader(http.StatusInternalServerError)
			break

		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
