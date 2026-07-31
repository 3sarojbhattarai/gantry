//go:build embed

package web

import (
	"embed"
	"io/fs"
)

// distFS holds the built frontend. web/dist must exist at compile time when
// building with -tags embed (run `make web-build` first).
//
//go:embed all:dist
var distFS embed.FS

func assets() (fs.FS, bool) {
	sub, err := fs.Sub(distFS, "dist")
	if err != nil {
		return nil, false
	}
	return sub, true
}
