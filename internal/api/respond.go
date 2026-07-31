package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// writeJSON encodes v as JSON with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// errorBody is the JSON shape returned for errors.
type errorBody struct {
	Error string `json:"error"`
}

// writeError maps a domain error to an HTTP status and writes a JSON body.
func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, docker.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, docker.ErrConsentRequired):
		status = http.StatusConflict
	case errors.Is(err, docker.ErrDryRunUnsupported):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, errorBody{Error: err.Error()})
}
