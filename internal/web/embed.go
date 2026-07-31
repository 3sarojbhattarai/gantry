// Package web serves gantry's React frontend. In Phase 4 this package embeds
// web/dist via embed.FS behind a build tag, so `go run` works without a
// frontend build while release builds ship the compiled assets.
//
// Phase 0 is a placeholder: Assets returns an empty filesystem. The signature
// is what the api layer will consume, so wiring can be written against it now.
package web

import (
	"io/fs"
	"testing/fstest"
)

// Assets returns the embedded frontend filesystem. Empty until Phase 4.
func Assets() fs.FS {
	return fstest.MapFS{}
}
