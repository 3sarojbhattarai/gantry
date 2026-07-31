// Package api exposes the docker engine over HTTP: REST for lists and inspects,
// Server-Sent Events for the event/stats/log streams, and mutation endpoints.
// It depends only on docker.Client, so it is tested against fakedocker with no
// daemon. The built React frontend is served from internal/web.
package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/web"
)

// Server routes HTTP requests to the engine.
type Server struct {
	engine docker.Client
	mux    *http.ServeMux
}

// NewServer builds a Server over the given engine, wiring the API routes and
// the embedded frontend.
func NewServer(engine docker.Client) *Server {
	s := &Server{engine: engine, mux: http.NewServeMux()}
	s.routes()
	return s
}

func (s *Server) routes() {
	// Reads.
	s.mux.HandleFunc("GET /api/containers", s.listContainers)
	s.mux.HandleFunc("GET /api/containers/{id}", s.inspectContainer)
	s.mux.HandleFunc("GET /api/images", s.listImages)
	s.mux.HandleFunc("GET /api/images/{id}", s.inspectImage)
	s.mux.HandleFunc("GET /api/networks", s.listNetworks)
	s.mux.HandleFunc("GET /api/volumes", s.listVolumes)

	// Streams (SSE / chunked).
	s.mux.HandleFunc("GET /api/events", s.streamEvents)
	s.mux.HandleFunc("GET /api/containers/{id}/stats", s.streamStats)
	s.mux.HandleFunc("GET /api/containers/{id}/logs", s.streamLogs)

	// Mutations.
	s.mux.HandleFunc("POST /api/containers/{id}/start", s.startContainer)
	s.mux.HandleFunc("POST /api/containers/{id}/stop", s.stopContainer)
	s.mux.HandleFunc("POST /api/containers/{id}/restart", s.restartContainer)
	s.mux.HandleFunc("POST /api/containers/{id}/kill", s.killContainer)
	s.mux.HandleFunc("DELETE /api/containers/{id}", s.removeContainer)
	s.mux.HandleFunc("DELETE /api/images/{id}", s.removeImage)
	s.mux.HandleFunc("DELETE /api/networks/{id}", s.removeNetwork)
	s.mux.HandleFunc("DELETE /api/volumes/{name}", s.removeVolume)
	s.mux.HandleFunc("POST /api/prune/{kind}", s.prune)
	s.mux.HandleFunc("POST /api/networks", s.createNetwork)
	s.mux.HandleFunc("POST /api/networks/{id}/connect", s.connectNetwork)
	s.mux.HandleFunc("POST /api/networks/{id}/disconnect", s.disconnectNetwork)

	// Exec (WebSocket) + container create (Phases 6-7).
	s.mux.HandleFunc("GET /api/containers/{id}/exec", s.execWS)
	s.mux.HandleFunc("GET /api/containers/{id}/spec", s.containerSpec)
	s.mux.HandleFunc("POST /api/containers", s.createContainer)
	s.mux.HandleFunc("POST /api/export/{format}", s.exportSpec)

	// Health + frontend.
	s.mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})
	s.mux.Handle("/", web.Handler())
}

// Handler returns the HTTP handler (with middleware applied).
func (s *Server) Handler() http.Handler {
	return noCacheAPI(s.mux)
}

// Run serves until ctx is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context, addr string) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           s.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	errCh := make(chan error, 1)
	go func() {
		err := srv.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}

// noCacheAPI disables caching for API responses so the live UI never sees stale
// data from an intermediary.
func noCacheAPI(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) >= 4 && r.URL.Path[:4] == "/api" {
			w.Header().Set("Cache-Control", "no-store")
		}
		next.ServeHTTP(w, r)
	})
}
