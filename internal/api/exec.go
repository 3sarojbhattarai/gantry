package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/coder/websocket"
)

// resizeMsg is the JSON control frame the browser sends to resize the TTY.
type resizeMsg struct {
	Type string `json:"type"` // "resize"
	Rows uint   `json:"rows"`
	Cols uint   `json:"cols"`
}

// execWS bridges a browser terminal to a container exec session over a
// WebSocket. Binary frames carry raw terminal I/O; text frames carry resize
// control messages.
func (s *Server) execWS(w http.ResponseWriter, r *http.Request) {
	shell := r.URL.Query().Get("cmd")
	if shell == "" {
		shell = "/bin/sh"
	}
	sess, err := s.engine.ContainerExec(r.Context(), r.PathValue("id"), docker.ExecOptions{
		Cmd: []string{shell},
		TTY: true,
	})
	if err != nil {
		writeError(w, err)
		return
	}
	defer sess.Close()

	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		// Same-origin only; the server binds loopback by default.
		OriginPatterns: []string{"*"},
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// Pump container output -> browser (binary frames).
	go func() {
		defer cancel()
		buf := make([]byte, 4096)
		for {
			n, err := sess.Read(buf)
			if n > 0 {
				if werr := conn.Write(ctx, websocket.MessageBinary, buf[:n]); werr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	// Pump browser -> container: binary = stdin, text = resize control.
	for {
		typ, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		switch typ {
		case websocket.MessageBinary:
			if _, werr := sess.Write(data); werr != nil {
				return
			}
		case websocket.MessageText:
			var msg resizeMsg
			if json.Unmarshal(data, &msg) == nil && msg.Type == "resize" {
				_ = sess.Resize(msg.Rows, msg.Cols)
			}
		}
	}
}

// createContainer creates a container from a JSON CreateSpec.
func (s *Server) createContainer(w http.ResponseWriter, r *http.Request) {
	var spec docker.CreateSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	if spec.Image == "" {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "image is required"})
		return
	}
	start := r.URL.Query().Get("start") == "true"
	id, err := s.engine.CreateContainer(r.Context(), spec, start)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": id})
}

// containerSpec returns an existing container's config as a CreateSpec (the
// --from prefill for the web create form).
func (s *Server) containerSpec(w http.ResponseWriter, r *http.Request) {
	spec, err := s.engine.SpecFromContainer(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, spec)
}

// exportSpec renders a posted CreateSpec as a `docker run` command or compose
// fragment, reusing the engine's formatting so the CLI and web agree.
func (s *Server) exportSpec(w http.ResponseWriter, r *http.Request) {
	var spec docker.CreateSpec
	if err := json.NewDecoder(r.Body).Decode(&spec); err != nil {
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "invalid JSON body"})
		return
	}
	var text string
	switch r.PathValue("format") {
	case "run":
		text = docker.SpecToDockerRun(spec)
	case "compose":
		text = docker.SpecToCompose(spec)
	default:
		writeJSON(w, http.StatusBadRequest, errorBody{Error: "unknown format"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"text": text})
}
