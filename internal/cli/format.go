package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/3sarojbhattarai/gantry/internal/docker"
)

// shortID trims the sha256: prefix and shortens an ID to the 12-character form
// Docker's own CLI displays.
func shortID(id string) string {
	id = strings.TrimPrefix(id, "sha256:")
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

// humanSize renders a byte count in SI units, matching `docker images` output
// (base 1000: kB, MB, GB…).
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

// truncate shortens s to at most n runes, marking elision with an ellipsis.
func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

// formatPorts renders a container's port mappings the way `docker ps` does,
// e.g. "0.0.0.0:8080->80/tcp, 53/udp", sorted and de-duplicated.
func formatPorts(ports []docker.Port) string {
	if len(ports) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(ports))
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		var s string
		switch {
		case p.Public != 0 && p.IP != "":
			s = fmt.Sprintf("%s:%d->%d/%s", p.IP, p.Public, p.Private, p.Type)
		case p.Public != 0:
			s = fmt.Sprintf("%d->%d/%s", p.Public, p.Private, p.Type)
		default:
			s = fmt.Sprintf("%d/%s", p.Private, p.Type)
		}
		if _, dup := seen[s]; dup {
			continue
		}
		seen[s] = struct{}{}
		parts = append(parts, s)
	}
	sort.Strings(parts)
	return strings.Join(parts, ", ")
}

// splitRepoTag splits a "repository:tag" reference on its final colon. A bare
// reference with no tag reports "<none>".
func splitRepoTag(ref string) (repo, tag string) {
	if i := strings.LastIndex(ref, ":"); i >= 0 && !strings.Contains(ref[i:], "/") {
		return ref[:i], ref[i+1:]
	}
	return ref, "<none>"
}
