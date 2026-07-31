package api

import (
	"net/http"
)

func (s *Server) listContainers(w http.ResponseWriter, r *http.Request) {
	all := r.URL.Query().Get("all") == "true"
	cs, err := s.engine.ListContainers(r.Context(), all)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

func (s *Server) inspectContainer(w http.ResponseWriter, r *http.Request) {
	d, err := s.engine.InspectContainer(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) listImages(w http.ResponseWriter, r *http.Request) {
	imgs, err := s.engine.ListImages(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, imgs)
}

func (s *Server) inspectImage(w http.ResponseWriter, r *http.Request) {
	d, err := s.engine.InspectImage(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, d)
}

func (s *Server) listNetworks(w http.ResponseWriter, r *http.Request) {
	ns, err := s.engine.ListNetworks(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ns)
}

func (s *Server) listVolumes(w http.ResponseWriter, r *http.Request) {
	vs, err := s.engine.ListVolumes(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, vs)
}
