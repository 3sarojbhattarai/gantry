package tui

import (
	"context"
	"strings"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type category int

const (
	catContainers category = iota
	catImages
	catNetworks
	catVolumes
	categoryCount
)

func (c category) title() string {
	switch c {
	case catContainers:
		return "Containers"
	case catImages:
		return "Images"
	case catNetworks:
		return "Networks"
	case catVolumes:
		return "Volumes"
	}
	return ""
}

const maxLogLines = 1000

// Model is the root Bubbletea model for the read-only TUI.
type Model struct {
	client docker.Client
	ctx    context.Context

	width, height int
	ready         bool
	showHelp      bool
	err           error

	focus  category
	cursor [categoryCount]int

	containers []docker.Container
	images     []docker.ImageSummary
	networks   []docker.Network
	volumes    []docker.Volume

	// Selection-scoped state for the focused container: its logs, live stats,
	// and inspect detail all stream under selCancel and are tagged with selGen.
	selID     string
	selGen    int
	selCancel context.CancelFunc
	logCh     <-chan string
	statsCh   <-chan docker.Stats
	logLines  []string
	logVP     viewport.Model
	detail    string
	stats     docker.Stats
	hasStats  bool

	events <-chan docker.Event

	// destructive-op confirmation modal
	confirming    bool
	confirmPrompt string
	confirmAction func() tea.Cmd

	keys   keyMap
	help   help.Model
	styles styles

	// layout, computed on resize
	leftW, rightW int
	bodyH, logH   int
}

// New builds an initial model bound to ctx and client.
func New(ctx context.Context, client docker.Client) Model {
	return Model{
		client: client,
		ctx:    ctx,
		focus:  catContainers,
		keys:   defaultKeys(),
		help:   help.New(),
		styles: newStyles(),
		logVP:  viewport.New(0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		loadContainers(m.ctx, m.client),
		loadImages(m.ctx, m.client),
		loadNetworks(m.ctx, m.client),
		loadVolumes(m.ctx, m.client),
		startEvents(m.ctx, m.client),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.ready = true
		m.layout()
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case errMsg:
		m.err = msg.err
		return m, nil

	case containersMsg:
		m.containers = []docker.Container(msg)
		m.clampCursor(catContainers)
		return m, m.syncSelection()

	case imagesMsg:
		m.images = []docker.ImageSummary(msg)
		m.clampCursor(catImages)
		return m, nil

	case networksMsg:
		m.networks = []docker.Network(msg)
		m.clampCursor(catNetworks)
		return m, nil

	case volumesMsg:
		m.volumes = []docker.Volume(msg)
		m.clampCursor(catVolumes)
		return m, nil

	case eventsStartedMsg:
		m.events = msg.ch
		return m, recvEvent(m.events)

	case eventMsg:
		return m, tea.Batch(refreshFor(m.ctx, m.client, docker.Event(msg).Type), recvEvent(m.events))

	case eventsClosedMsg:
		return m, nil

	case logStreamMsg:
		if msg.gen != m.selGen {
			return m, nil
		}
		m.logCh = msg.ch
		return m, recvLog(m.logCh, msg.gen)

	case logLineMsg:
		if msg.gen != m.selGen {
			return m, nil
		}
		m.appendLog(msg.line)
		return m, recvLog(m.logCh, msg.gen)

	case logClosedMsg:
		return m, nil

	case statsStreamMsg:
		if msg.gen != m.selGen {
			return m, nil
		}
		m.statsCh = msg.ch
		return m, recvStats(m.statsCh, msg.gen)

	case statsMsg:
		if msg.gen != m.selGen {
			return m, nil
		}
		m.stats, m.hasStats = msg.stats, true
		return m, recvStats(m.statsCh, msg.gen)

	case statsClosedMsg:
		return m, nil

	case detailMsg:
		if msg.gen != m.selGen {
			return m, nil
		}
		m.detail = msg.text
		return m, nil

	case execFinishedMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		// Force a full repaint after the terminal was handed back.
		return m, tea.ClearScreen

	case mutationDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		} else {
			m.err = nil
		}
		// Refresh every list so the change shows immediately (the event stream
		// would also catch it, but this is snappier).
		return m, tea.Batch(
			loadContainers(m.ctx, m.client),
			loadImages(m.ctx, m.client),
			loadNetworks(m.ctx, m.client),
			loadVolumes(m.ctx, m.client),
		)
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// While a confirmation modal is open, keys answer it and nothing else.
	if m.confirming {
		switch msg.String() {
		case "y", "Y":
			action := m.confirmAction
			m.confirming = false
			m.confirmPrompt = ""
			m.confirmAction = nil
			if action != nil {
				return m, action()
			}
			return m, nil
		case "n", "N", "esc", "q":
			m.confirming = false
			m.confirmPrompt = ""
			m.confirmAction = nil
			return m, nil
		default:
			return m, nil
		}
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit
	case key.Matches(msg, m.keys.Help):
		m.showHelp = !m.showHelp
		m.layout()
		return m, nil
	case key.Matches(msg, m.keys.Up):
		m.move(-1)
		return m, m.syncSelection()
	case key.Matches(msg, m.keys.Down):
		m.move(1)
		return m, m.syncSelection()
	case key.Matches(msg, m.keys.Tab):
		m.focus = (m.focus + 1) % categoryCount
		return m, m.syncSelection()
	case key.Matches(msg, m.keys.ShiftTab):
		m.focus = (m.focus - 1 + categoryCount) % categoryCount
		return m, m.syncSelection()
	case key.Matches(msg, m.keys.Cat):
		switch msg.String() {
		case "1":
			m.focus = catContainers
		case "2":
			m.focus = catImages
		case "3":
			m.focus = catNetworks
		case "4":
			m.focus = catVolumes
		}
		return m, m.syncSelection()
	case key.Matches(msg, m.keys.Refresh):
		return m, tea.Batch(
			loadContainers(m.ctx, m.client),
			loadImages(m.ctx, m.client),
			loadNetworks(m.ctx, m.client),
			loadVolumes(m.ctx, m.client),
		)

	// Lifecycle (reversible; no confirmation) — containers only.
	case key.Matches(msg, m.keys.Start):
		return m, m.lifecycle(func(ctx context.Context, id string) error { return m.client.StartContainer(ctx, id) })
	case key.Matches(msg, m.keys.Stop):
		return m, m.lifecycle(func(ctx context.Context, id string) error {
			return m.client.StopContainer(ctx, id, docker.StopOptions{})
		})
	case key.Matches(msg, m.keys.Restart):
		return m, m.lifecycle(func(ctx context.Context, id string) error {
			return m.client.RestartContainer(ctx, id, docker.StopOptions{})
		})
	case key.Matches(msg, m.keys.Kill):
		return m, m.lifecycle(func(ctx context.Context, id string) error { return m.client.KillContainer(ctx, id, "KILL") })

	// Remove (destructive) — opens a confirmation modal for the focused item.
	case key.Matches(msg, m.keys.Remove):
		m.requestRemove()
		return m, nil

	// Exec — suspend the TUI and open a shell in the selected running container.
	case key.Matches(msg, m.keys.Exec):
		if m.focus == catContainers {
			if c, ok := m.selectedContainer(); ok && c.State == "running" {
				cmd := &execCommand{ctx: m.ctx, client: m.client, id: c.ID}
				return m, tea.Exec(cmd, func(err error) tea.Msg { return execFinishedMsg{err} })
			}
		}
		return m, nil
	}
	return m, nil
}

// lifecycle builds a mutation command for the selected container, or nil when
// the containers pane is not focused / nothing is selected.
func (m Model) lifecycle(fn func(context.Context, string) error) tea.Cmd {
	if m.focus != catContainers {
		return nil
	}
	c, ok := m.selectedContainer()
	if !ok {
		return nil
	}
	id, ctx := c.ID, m.ctx
	return mutate(func() error { return fn(ctx, id) })
}

// requestRemove opens a confirmation modal to remove the focused item. The
// modal's "yes" is the explicit consent the engine requires.
func (m *Model) requestRemove() {
	cl, ctx := m.client, m.ctx
	switch m.focus {
	case catContainers:
		if c, ok := m.selectedContainer(); ok {
			id := c.ID
			m.confirm("Remove container "+firstName(c.Names)+"?", func() tea.Cmd {
				return mutate(func() error {
					return cl.RemoveContainer(ctx, id, docker.RemoveContainerOptions{Consent: docker.Confirm(), Force: true})
				})
			})
		}
	case catImages:
		if im, ok := m.selectedImage(); ok {
			id := im.ID
			m.confirm("Remove image "+firstTag(im.RepoTags)+"?", func() tea.Cmd {
				return mutate(func() error {
					return cl.RemoveImage(ctx, id, docker.RemoveImageOptions{Consent: docker.Confirm(), PruneChildren: true})
				})
			})
		}
	case catNetworks:
		if n, ok := m.selectedNetwork(); ok {
			id := n.ID
			m.confirm("Remove network "+n.Name+"?", func() tea.Cmd {
				return mutate(func() error {
					return cl.RemoveNetwork(ctx, id, docker.RemoveNetworkOptions{Consent: docker.Confirm()})
				})
			})
		}
	case catVolumes:
		if v, ok := m.selectedVolume(); ok {
			name := v.Name
			m.confirm("Remove volume "+name+"?", func() tea.Cmd {
				return mutate(func() error {
					return cl.RemoveVolume(ctx, name, docker.RemoveVolumeOptions{Consent: docker.Confirm()})
				})
			})
		}
	}
}

func (m *Model) confirm(prompt string, action func() tea.Cmd) {
	m.confirming = true
	m.confirmPrompt = prompt
	m.confirmAction = action
}

// --- selection management ---------------------------------------------------

// syncSelection reconciles the selection-scoped streams with the currently
// selected container. When the selected container changes it cancels the old
// streams, bumps the generation, and starts fresh log/stats/detail streams. It
// returns nil when nothing changed.
func (m *Model) syncSelection() tea.Cmd {
	id := ""
	if m.focus == catContainers {
		if c, ok := m.selectedContainer(); ok {
			id = c.ID
		}
	}
	if id == m.selID {
		return nil
	}

	// Tear down the previous container's streams.
	if m.selCancel != nil {
		m.selCancel()
		m.selCancel = nil
	}
	m.selID = id
	m.logCh, m.statsCh = nil, nil
	m.logLines = nil
	m.detail = ""
	m.stats, m.hasStats = docker.Stats{}, false
	m.logVP.SetContent("")

	if id == "" {
		return nil
	}

	m.selGen++
	gen := m.selGen
	ctx, cancel := context.WithCancel(m.ctx)
	m.selCancel = cancel
	return tea.Batch(
		startLogs(ctx, m.client, id, gen),
		startStats(ctx, m.client, id, gen),
		loadDetail(ctx, m.client, id, gen),
	)
}

func (m Model) selectedContainer() (docker.Container, bool) {
	i := m.cursor[catContainers]
	if i < 0 || i >= len(m.containers) {
		return docker.Container{}, false
	}
	return m.containers[i], true
}

// --- small mutating helpers (pointer receiver; used within value Update) ----

func (m *Model) move(delta int) {
	n := m.lenFor(m.focus)
	if n == 0 {
		return
	}
	m.cursor[m.focus] = clamp(m.cursor[m.focus]+delta, 0, n-1)
}

func (m *Model) clampCursor(cat category) {
	n := m.lenFor(cat)
	if n == 0 {
		m.cursor[cat] = 0
		return
	}
	m.cursor[cat] = clamp(m.cursor[cat], 0, n-1)
}

func (m Model) lenFor(cat category) int {
	switch cat {
	case catContainers:
		return len(m.containers)
	case catImages:
		return len(m.images)
	case catNetworks:
		return len(m.networks)
	case catVolumes:
		return len(m.volumes)
	}
	return 0
}

func (m *Model) appendLog(line string) {
	m.logLines = append(m.logLines, line)
	if len(m.logLines) > maxLogLines {
		m.logLines = m.logLines[len(m.logLines)-maxLogLines:]
	}
	m.logVP.SetContent(strings.Join(m.logLines, "\n"))
	m.logVP.GotoBottom()
}

func (m *Model) layout() {
	if !m.ready {
		return
	}
	const headerH = 2 // title + tab bar
	const footerH = 1
	logH := m.height / 3
	logH = clamp(logH, 4, 14)
	bodyH := m.height - headerH - footerH - logH - 1 // -1 separator
	if bodyH < 3 {
		bodyH = 3
	}
	leftW := clamp(m.width*2/5, 20, 50)
	if leftW > m.width-10 {
		leftW = m.width - 10
	}
	m.leftW = leftW
	m.rightW = m.width - leftW - 1
	m.bodyH = bodyH
	m.logH = logH
	m.logVP.Width = m.width
	m.logVP.Height = logH
	m.logVP.SetContent(strings.Join(m.logLines, "\n"))
	m.logVP.GotoBottom()
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
