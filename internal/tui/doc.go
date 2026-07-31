// Package tui is gantry's Bubbletea terminal UI: a tabbed list of
// containers/images/networks/volumes on the left, a detail pane on the right,
// and a live log tail at the bottom. The daemon's event and stats streams,
// plus container log streaming, are wired as Bubbletea commands so the view
// updates itself rather than polling. Run takes over the terminal; the model,
// update logic, and rendering live in the sibling files.
package tui
