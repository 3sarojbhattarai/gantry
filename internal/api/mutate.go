package api

import (
	"encoding/json"
	"net/http"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// confirmed reports whether the request carries explicit consent for a
// destructive operation. The web dialog sets ?confirm=true after the user
// agrees; the handler translates that into docker.Confirm().
func confirmed(r *http.Request) docker.Consent {
	if r.URL.Query().Get("confirm") == "true" {
		return docker.Confirm()
	}
	return docker.Consent{}
}

func boolParam(r *http.Request, key string) bool {
	return r.URL.Query().Get(key) == "true"
}

// okOrError writes 204 on success or the mapped error.
func okOrError(w http.ResponseWriter, err error) {
	if err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) startContainer(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.StartContainer(r.Context(), r.PathValue("id")))
}

func (s *Server) stopContainer(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.StopContainer(r.Context(), r.PathValue("id"), docker.StopOptions{}))
}

func (s *Server) restartContainer(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.RestartContainer(r.Context(), r.PathValue("id"), docker.StopOptions{}))
}

func (s *Server) killContainer(w http.ResponseWriter, r *http.Request) {
	signal := r.URL.Query().Get("signal")
	if signal == "" {
		signal = "KILL"
	}
	okOrError(w, s.engine.KillContainer(r.Context(), r.PathValue("id"), signal))
}

func (s *Server) removeContainer(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.RemoveContainer(r.Context(), r.PathValue("id"), docker.RemoveContainerOptions{
		Consent:       confirmed(r),
		Force:         boolParam(r, "force"),
		RemoveVolumes: boolParam(r, "volumes"),
	}))
}

func (s *Server) removeImage(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.RemoveImage(r.Context(), r.PathValue("id"), docker.RemoveImageOptions{
		Consent:       confirmed(r),
		Force:         boolParam(r, "force"),
		PruneChildren: true,
	}))
}

func (s *Server) removeNetwork(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.RemoveNetwork(r.Context(), r.PathValue("id"), docker.RemoveNetworkOptions{
		Consent: confirmed(r),
	}))
}

func (s *Server) removeVolume(w http.ResponseWriter, r *http.Request) {
	okOrError(w, s.engine.RemoveVolume(r.Context(), r.PathValue("name"), docker.RemoveVolumeOptions{
		Consent: confirmed(r),
		Force:   boolParam(r, "force"),
	}))
}

func (s *Server) prune(w http.ResponseWriter, r *http.Request) {
	opts := docker.PruneOptions{Consent: confirmed(r), DryRun: boolParam(r, "dryRun")}
	var (
		report docker.PruneReport
		err    error
	)
	switch r.PathValue("kind") {
	case "containers":
		report, err = s.engine.PruneContainers(r.Context(), opts)
	case "images":
		report, err = s.engine.PruneImages(r.Context(), opts)
	case "volumes":
		report, err = s.engine.PruneVolumes(r.Context(), opts)
	case "networks":
		report, err = s.engine.PruneNetworks(r.Context(), opts)
	default:
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "unknown prune kind"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, report)
}

func (s *Server) createNetwork(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Driver   string `json:"driver"`
		Internal bool   `json:"internal"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if body.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "name is required"})
		return
	}
	id, err := s.engine.CreateNetwork(r.Context(), docker.CreateNetworkOptions{
		Name:     body.Name,
		Driver:   body.Driver,
		Internal: body.Internal,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

func (s *Server) connectNetwork(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Container string `json:"container"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	okOrError(w, s.engine.ConnectNetwork(r.Context(), r.PathValue("id"), body.Container))
}

func (s *Server) disconnectNetwork(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Container string `json:"container"`
		Force     bool   `json:"force"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	okOrError(w, s.engine.DisconnectNetwork(r.Context(), r.PathValue("id"), body.Container, body.Force))
}
