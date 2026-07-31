package tui

import (
	"fmt"
	"strings"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if !m.ready {
		return "Starting gantry…"
	}
	parts := []string{
		m.headerView(),
		m.bodyView(),
		m.logLabelView(),
		m.logVP.View(),
		m.footerView(),
	}
	return strings.Join(parts, "\n")
}

// --- header + tabs ----------------------------------------------------------

func (m Model) headerView() string {
	title := m.styles.title.Render("gantry")

	tabs := make([]string, 0, categoryCount)
	for c := category(0); c < categoryCount; c++ {
		label := fmt.Sprintf("%d %s (%d)", c+1, c.title(), m.lenFor(c))
		if c == m.focus {
			tabs = append(tabs, m.styles.activeTab.Render(label))
		} else {
			tabs = append(tabs, m.styles.tab.Render(label))
		}
	}
	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
	return title + "\n" + tabBar
}

// --- body: list | detail ----------------------------------------------------

func (m Model) bodyView() string {
	list := lipgloss.NewStyle().Width(m.leftW).Height(m.bodyH).MaxHeight(m.bodyH).Render(m.listView())
	gap := lipgloss.NewStyle().Width(1).Height(m.bodyH).Render(strings.Repeat("│\n", m.bodyH))
	detail := lipgloss.NewStyle().Width(m.rightW).Height(m.bodyH).MaxHeight(m.bodyH).Render(m.detailView())
	return lipgloss.JoinHorizontal(lipgloss.Top, list, gap, detail)
}

func (m Model) listView() string {
	lines := m.itemLines(m.focus)
	if len(lines) == 0 {
		return m.styles.dim.Render("(none)")
	}
	cursor := m.cursor[m.focus]
	start := 0
	if len(lines) > m.bodyH {
		if cursor >= m.bodyH {
			start = cursor - m.bodyH + 1
		}
		if start+m.bodyH > len(lines) {
			start = len(lines) - m.bodyH
		}
		if start < 0 {
			start = 0
		}
	}
	var b strings.Builder
	for i := start; i < len(lines) && i < start+m.bodyH; i++ {
		text := fitLine(lines[i], m.leftW)
		if i == cursor {
			text = m.styles.selected.Render(text)
		}
		b.WriteString(text)
		if i < len(lines)-1 && i < start+m.bodyH-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}

func (m Model) itemLines(cat category) []string {
	switch cat {
	case catContainers:
		out := make([]string, len(m.containers))
		for i, c := range m.containers {
			mark := "○"
			if c.State == "running" {
				mark = "●"
			}
			out[i] = fmt.Sprintf("%s %s  %s", mark, firstName(c.Names), c.Image)
		}
		return out
	case catImages:
		out := make([]string, len(m.images))
		for i, im := range m.images {
			out[i] = fmt.Sprintf("%s  %s", firstTag(im.RepoTags), humanSize(im.Size))
		}
		return out
	case catNetworks:
		out := make([]string, len(m.networks))
		for i, n := range m.networks {
			out[i] = fmt.Sprintf("%s  %s", n.Name, n.Driver)
		}
		return out
	case catVolumes:
		out := make([]string, len(m.volumes))
		for i, v := range m.volumes {
			out[i] = fmt.Sprintf("%s  %s", v.Name, v.Driver)
		}
		return out
	}
	return nil
}

// --- detail -----------------------------------------------------------------

func (m Model) detailView() string {
	switch m.focus {
	case catContainers:
		return m.containerDetailView()
	case catImages:
		if im, ok := m.selectedImage(); ok {
			return m.styles.paneTitle.Render(firstTag(im.RepoTags)) + "\n\n" +
				kv("ID", shortID(im.ID)) +
				kv("Size", humanSize(im.Size)) +
				kv("Tags", strings.Join(im.RepoTags, ", ")) +
				kv("Created", im.Created.Format("2006-01-02 15:04:05"))
		}
	case catNetworks:
		if n, ok := m.selectedNetwork(); ok {
			return m.styles.paneTitle.Render(n.Name) + "\n\n" +
				kv("ID", shortID(n.ID)) +
				kv("Driver", n.Driver) +
				kv("Scope", n.Scope) +
				kv("Internal", fmt.Sprintf("%t", n.Internal))
		}
	case catVolumes:
		if v, ok := m.selectedVolume(); ok {
			return m.styles.paneTitle.Render(v.Name) + "\n\n" +
				kv("Driver", v.Driver) +
				kv("Mountpoint", v.Mountpoint)
		}
	}
	return m.styles.dim.Render("(nothing selected)")
}

func (m Model) containerDetailView() string {
	c, ok := m.selectedContainer()
	if !ok {
		return m.styles.dim.Render("(no containers)")
	}
	var b strings.Builder
	b.WriteString(m.styles.paneTitle.Render(firstName(c.Names)))
	b.WriteString("\n\n")
	b.WriteString(kv("ID", shortID(c.ID)))
	b.WriteString(kv("Image", c.Image))
	b.WriteString(kv("State", c.State))
	b.WriteString(kv("Status", c.Status))
	b.WriteString(kv("Command", c.Command))
	if p := portsLine(c.Ports); p != "" {
		b.WriteString(kv("Ports", p))
	}
	if m.detail != "" {
		b.WriteString("\n")
		b.WriteString(m.detail)
	}
	if m.hasStats {
		b.WriteString("\n")
		b.WriteString(m.styles.paneTitle.Render("Stats"))
		b.WriteString("\n")
		b.WriteString(kv("CPU", fmt.Sprintf("%.2f%%", m.stats.CPUPercent)))
		b.WriteString(kv("Mem", fmt.Sprintf("%s / %s (%.1f%%)",
			humanSize(int64(m.stats.MemUsage)), humanSize(int64(m.stats.MemLimit)), m.stats.MemPercent)))
		b.WriteString(kv("Net", fmt.Sprintf("↓ %s  ↑ %s", humanSize(int64(m.stats.NetRx)), humanSize(int64(m.stats.NetTx)))))
		b.WriteString(kv("PIDs", fmt.Sprintf("%d", m.stats.PIDs)))
	}
	return b.String()
}

// renderContainerDetail formats the inspect-only extras (used by loadDetail).
func renderContainerDetail(d *docker.ContainerDetails) string {
	var b strings.Builder
	if d.Platform != "" {
		b.WriteString(kv("Platform", d.Platform))
	}
	b.WriteString(kv("Restarts", fmt.Sprintf("%d", d.RestartCount)))
	if len(d.Env) > 0 {
		b.WriteString("Env:\n")
		for _, e := range d.Env {
			b.WriteString("  " + e + "\n")
		}
	}
	return b.String()
}

// --- logs + footer ----------------------------------------------------------

func (m Model) logLabelView() string {
	label := "Logs"
	if m.focus == catContainers {
		if c, ok := m.selectedContainer(); ok {
			label = "Logs: " + firstName(c.Names)
		}
	} else {
		label = "Logs (select a container)"
	}
	prefix := "── " + label + " "
	dashes := m.width - lipgloss.Width(prefix)
	if dashes < 0 {
		dashes = 0
	}
	return m.styles.sep.Render(prefix + strings.Repeat("─", dashes))
}

func (m Model) footerView() string {
	if m.confirming {
		return m.styles.errText.Render(m.confirmPrompt + "  [y/n]")
	}
	if m.err != nil {
		return m.styles.errText.Render("error: " + m.err.Error())
	}
	m.help.ShowAll = m.showHelp
	return m.help.View(m.keys)
}

// --- selection accessors for non-container categories -----------------------

func (m Model) selectedImage() (docker.ImageSummary, bool) {
	i := m.cursor[catImages]
	if i < 0 || i >= len(m.images) {
		return docker.ImageSummary{}, false
	}
	return m.images[i], true
}

func (m Model) selectedNetwork() (docker.Network, bool) {
	i := m.cursor[catNetworks]
	if i < 0 || i >= len(m.networks) {
		return docker.Network{}, false
	}
	return m.networks[i], true
}

func (m Model) selectedVolume() (docker.Volume, bool) {
	i := m.cursor[catVolumes]
	if i < 0 || i >= len(m.volumes) {
		return docker.Volume{}, false
	}
	return m.volumes[i], true
}

// --- formatting helpers -----------------------------------------------------

func kv(key, val string) string {
	return fmt.Sprintf("%-10s %s\n", key+":", val)
}

func fitLine(s string, w int) string {
	if w <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) > w {
		if w > 1 {
			return string(r[:w-1]) + "…"
		}
		return string(r[:w])
	}
	return s + strings.Repeat(" ", w-len(r))
}

func firstName(names []string) string {
	if len(names) == 0 {
		return "<none>"
	}
	return names[0]
}

func firstTag(tags []string) string {
	if len(tags) == 0 {
		return "<none>"
	}
	return tags[0]
}

func shortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func humanSize(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%dB", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func portsLine(ports []docker.Port) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		if p.Public != 0 {
			parts = append(parts, fmt.Sprintf("%d->%d/%s", p.Public, p.Private, p.Type))
		} else {
			parts = append(parts, fmt.Sprintf("%d/%s", p.Private, p.Type))
		}
	}
	return strings.Join(parts, ", ")
}
