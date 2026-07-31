//go:build !embed

package web

import "io/fs"

// assets reports no embedded frontend in the default build; Handler falls back
// to the placeholder page.
func assets() (fs.FS, bool) { return nil, false }
