package ping

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Pong(log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Debug("Received request", "method", r.Method, "url", r.URL.Path)

		mess := message{Message: "Pong!"}

		w.Header().Set("Content-type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mess)
	}
}
