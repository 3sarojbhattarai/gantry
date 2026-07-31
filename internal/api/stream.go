package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// sseBegin sets Server-Sent Events headers and returns the flusher. The engine
// call must happen before this, so an engine error can still be reported as a
// normal JSON response.
func sseBegin(w http.ResponseWriter) (http.Flusher, bool) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return nil, false
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	f.Flush()
	return f, true
}

func sseJSON(w http.ResponseWriter, f http.Flusher, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", b)
	f.Flush()
}

func sseText(w http.ResponseWriter, f http.Flusher, line string) {
	fmt.Fprintf(w, "data: %s\n\n", line)
	f.Flush()
}

func (s *Server) streamEvents(w http.ResponseWriter, r *http.Request) {
	ch, err := s.engine.Events(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	f, ok := sseBegin(w)
	if !ok {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			sseJSON(w, f, ev)
		}
	}
}

func (s *Server) streamStats(w http.ResponseWriter, r *http.Request) {
	ch, err := s.engine.ContainerStats(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	f, ok := sseBegin(w)
	if !ok {
		return
	}
	for {
		select {
		case <-r.Context().Done():
			return
		case st, ok := <-ch:
			if !ok {
				return
			}
			sseJSON(w, f, st)
		}
	}
}

func (s *Server) streamLogs(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	tail := q.Get("tail")
	if tail == "" {
		tail = "200"
	}
	rc, err := s.engine.ContainerLogs(r.Context(), r.PathValue("id"), docker.LogOptions{
		Follow:     q.Get("follow") == "true",
		Timestamps: q.Get("timestamps") == "true",
		Tail:       tail,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	defer rc.Close()

	f, ok := sseBegin(w)
	if !ok {
		return
	}
	sc := bufio.NewScanner(rc)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		select {
		case <-r.Context().Done():
			return
		default:
		}
		sseText(w, f, sc.Text())
	}
}
