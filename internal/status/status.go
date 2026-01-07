package status

import (
	"docker-exporter/internal/docker"
	"docker-exporter/internal/log"
	"encoding/json"
	"net/http"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func HandleStatus(cli *docker.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := Response{
			Status: "healthy",
			Error:  "",
		}

		// Check if Docker daemon is responding
		err := cli.Ping(r.Context())
		if err != nil {
			log.ErrorWith("Docker daemon is not responding", "error", err)
			response.Status = "unhealthy"
			response.Error = err.Error()
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		err = json.NewEncoder(w).Encode(response)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})
}
