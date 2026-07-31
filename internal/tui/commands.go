package tui

import (
	"bufio"
	"context"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	tea "github.com/charmbracelet/bubbletea"
)

// --- messages ---------------------------------------------------------------

type errMsg struct{ err error }

type (
	containersMsg []docker.Container
	imagesMsg     []docker.ImageSummary
	networksMsg   []docker.Network
	volumesMsg    []docker.Volume
)

// event-stream messages
type (
	eventsStartedMsg struct{ ch <-chan docker.Event }
	eventMsg         docker.Event
	eventsClosedMsg  struct{}
)

// Selection-scoped stream messages carry a generation number. When the user
// changes the selected container we bump the generation and cancel the old
// context; any late message from a previous stream arrives with a stale
// generation and is dropped.
type (
	logStreamMsg struct {
		gen int
		ch  <-chan string
	}
	logLineMsg struct {
		gen  int
		line string
	}
	logClosedMsg struct{ gen int }

	statsStreamMsg struct {
		gen int
		ch  <-chan docker.Stats
	}
	statsMsg struct {
		gen   int
		stats docker.Stats
	}
	statsClosedMsg struct{ gen int }

	detailMsg struct {
		gen  int
		text string
	}
)

// mutationDoneMsg reports the result of a mutation (start/stop/remove/…).
type mutationDoneMsg struct{ err error }

// mutate wraps a mutation call as a command. The closure captures whatever
// client/ctx/id it needs.
func mutate(fn func() error) tea.Cmd {
	return func() tea.Msg {
		return mutationDoneMsg{err: fn()}
	}
}

// --- list loaders -----------------------------------------------------------

func loadContainers(ctx context.Context, c docker.Client) tea.Cmd {
	return func() tea.Msg {
		cs, err := c.ListContainers(ctx, true)
		if err != nil {
			return errMsg{err}
		}
		return containersMsg(cs)
	}
}

func loadImages(ctx context.Context, c docker.Client) tea.Cmd {
	return func() tea.Msg {
		imgs, err := c.ListImages(ctx)
		if err != nil {
			return errMsg{err}
		}
		return imagesMsg(imgs)
	}
}

func loadNetworks(ctx context.Context, c docker.Client) tea.Cmd {
	return func() tea.Msg {
		ns, err := c.ListNetworks(ctx)
		if err != nil {
			return errMsg{err}
		}
		return networksMsg(ns)
	}
}

func loadVolumes(ctx context.Context, c docker.Client) tea.Cmd {
	return func() tea.Msg {
		vs, err := c.ListVolumes(ctx)
		if err != nil {
			return errMsg{err}
		}
		return volumesMsg(vs)
	}
}

// refreshFor returns the loader for the resource kind an event concerns.
func refreshFor(ctx context.Context, c docker.Client, eventType string) tea.Cmd {
	switch eventType {
	case "container":
		return loadContainers(ctx, c)
	case "image":
		return loadImages(ctx, c)
	case "network":
		return loadNetworks(ctx, c)
	case "volume":
		return loadVolumes(ctx, c)
	default:
		return nil
	}
}

// --- event stream -----------------------------------------------------------

func startEvents(ctx context.Context, c docker.Client) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.Events(ctx)
		if err != nil {
			return errMsg{err}
		}
		return eventsStartedMsg{ch: ch}
	}
}

func recvEvent(ch <-chan docker.Event) tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-ch
		if !ok {
			return eventsClosedMsg{}
		}
		return eventMsg(ev)
	}
}

// --- selection-scoped streams (logs, stats, detail) -------------------------

func startLogs(ctx context.Context, c docker.Client, id string, gen int) tea.Cmd {
	return func() tea.Msg {
		rc, err := c.ContainerLogs(ctx, id, docker.LogOptions{Follow: true, Tail: "200"})
		if err != nil {
			return logClosedMsg{gen: gen}
		}
		ch := make(chan string, 128)
		go func() {
			defer close(ch)
			defer rc.Close()
			sc := bufio.NewScanner(rc)
			sc.Buffer(make([]byte, 64*1024), 1024*1024)
			for sc.Scan() {
				select {
				case <-ctx.Done():
					return
				case ch <- sc.Text():
				}
			}
		}()
		return logStreamMsg{gen: gen, ch: ch}
	}
}

func recvLog(ch <-chan string, gen int) tea.Cmd {
	return func() tea.Msg {
		line, ok := <-ch
		if !ok {
			return logClosedMsg{gen: gen}
		}
		return logLineMsg{gen: gen, line: line}
	}
}

func startStats(ctx context.Context, c docker.Client, id string, gen int) tea.Cmd {
	return func() tea.Msg {
		ch, err := c.ContainerStats(ctx, id)
		if err != nil {
			return statsClosedMsg{gen: gen}
		}
		return statsStreamMsg{gen: gen, ch: ch}
	}
}

func recvStats(ch <-chan docker.Stats, gen int) tea.Cmd {
	return func() tea.Msg {
		s, ok := <-ch
		if !ok {
			return statsClosedMsg{gen: gen}
		}
		return statsMsg{gen: gen, stats: s}
	}
}

func loadDetail(ctx context.Context, c docker.Client, id string, gen int) tea.Cmd {
	return func() tea.Msg {
		d, err := c.InspectContainer(ctx, id)
		if err != nil {
			return detailMsg{gen: gen, text: ""}
		}
		return detailMsg{gen: gen, text: renderContainerDetail(d)}
	}
}
