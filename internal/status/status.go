package status

import (
	"encoding/json"
	"net/http"

	"github.com/h3rmt/docker-exporter/internal/docker"
	"github.com/h3rmt/docker-exporter/internal/log"
)

type Response struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Version string `json:"version"`
}

func HandleStatus(cli *docker.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		w.Header().Set("Content-Type", "application/json")

		response := Response{
			Status:  "healthy",
			Error:   "",
			Version: "",
		}

		// Check if Docker daemon is responding
		ver, err := cli.Ping(ctx)
		if err != nil {
			log.GetLogger().ErrorContext(ctx, "Docker daemon is not responding", "error", err)
			response.Status = "unhealthy"
			response.Error = err.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
			response.Version = ver
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
