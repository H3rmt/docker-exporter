package status

import (
	"encoding/json"
	"github.com/docker/docker/client"
	"net/http"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error"`
}

func HandleStatus(cli *client.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		response := Response{
			Status: "healthy",
			Error:  "",
		}

		// Check if Docker daemon is responding
		_, err := cli.Ping(r.Context())
		if err != nil {
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
