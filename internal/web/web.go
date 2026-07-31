// Package web serves gantry's React frontend. The built assets are embedded via
// embed.FS behind the `embed` build tag (see embed.go), so release builds ship
// a single binary. Without that tag — plain `go build` / `go run` — no frontend
// is compiled in, and Handler serves a small placeholder pointing at the Vite
// dev server. This keeps the Go build working without a Node toolchain.
package web

import (
	"io/fs"
	"net/http"
)

// Handler serves the frontend: the embedded SPA when built with -tags embed,
// otherwise a placeholder. API routes are registered separately and take
// precedence, so this only handles non-/api paths.
func Handler() http.Handler {
	sub, ok := assets()
	if !ok {
		return http.HandlerFunc(placeholder)
	}
	fileServer := http.FileServer(http.FS(sub))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SPA fallback: serve index.html for paths that aren't real files.
		if _, err := fs.Stat(sub, trimLeadingSlash(r.URL.Path)); err != nil {
			r = cloneToIndex(r)
		}
		fileServer.ServeHTTP(w, r)
	})
}

func trimLeadingSlash(p string) string {
	if len(p) > 0 && p[0] == '/' {
		p = p[1:]
	}
	if p == "" {
		return "index.html"
	}
	return p
}

func cloneToIndex(r *http.Request) *http.Request {
	r2 := r.Clone(r.Context())
	r2.URL.Path = "/"
	return r2
}

func placeholder(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(placeholderHTML))
}

const placeholderHTML = `<!doctype html>
<html><head><meta charset="utf-8"><title>gantry</title>
<style>body{font-family:ui-monospace,monospace;background:#0b0f14;color:#cbd5e1;
display:flex;min-height:100vh;align-items:center;justify-content:center;margin:0}
.card{max-width:38rem;padding:2rem;line-height:1.6}code{color:#7dd3fc}</style></head>
<body><div class="card">
<h1>gantry</h1>
<p>The API is running, but the frontend was not embedded in this build.</p>
<p>For development, run the Vite dev server:</p>
<pre><code>make web-install
make web-dev</code></pre>
<p>For a single self-contained binary, build with the frontend embedded:</p>
<pre><code>make web-build
go build -tags embed ./cmd/gantry</code></pre>
</div></body></html>`
