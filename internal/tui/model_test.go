package tui

import (
	"context"
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
	tea "github.com/charmbracelet/bubbletea"
)

func newTestModel(f *fakedocker.Fake) Model {
	m := New(context.Background(), f)
	// Give it a size so layout is computed and View is safe to call.
	next, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return next.(Model)
}

func send(t *testing.T, m Model, msg tea.Msg) Model {
	t.Helper()
	next, _ := m.Update(msg)
	return next.(Model)
}

func TestWindowSizeMakesReady(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	if !m.ready {
		t.Fatal("model not ready after WindowSizeMsg")
	}
	if m.leftW <= 0 || m.bodyH <= 0 || m.logH <= 0 {
		t.Fatalf("layout not computed: leftW=%d bodyH=%d logH=%d", m.leftW, m.bodyH, m.logH)
	}
}

func TestNavigationMovesCursorAndClamps(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	m = send(t, m, containersMsg([]docker.Container{
		{ID: "a", Names: []string{"a"}, State: "running"},
		{ID: "b", Names: []string{"b"}, State: "exited"},
	}))

	m = send(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor[catContainers] != 1 {
		t.Fatalf("cursor = %d, want 1", m.cursor[catContainers])
	}
	// Clamp at the bottom.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.cursor[catContainers] != 1 {
		t.Fatalf("cursor overran to %d, want clamped at 1", m.cursor[catContainers])
	}
	// And back up, clamped at the top.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyUp})
	m = send(t, m, tea.KeyMsg{Type: tea.KeyUp})
	if m.cursor[catContainers] != 0 {
		t.Fatalf("cursor = %d, want clamped at 0", m.cursor[catContainers])
	}
}

func TestTabSwitchesFocus(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	if m.focus != catContainers {
		t.Fatalf("initial focus = %v", m.focus)
	}
	m = send(t, m, tea.KeyMsg{Type: tea.KeyTab})
	if m.focus != catImages {
		t.Fatalf("after tab, focus = %v, want images", m.focus)
	}
	// Number keys jump directly.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("4")})
	if m.focus != catVolumes {
		t.Fatalf("after '4', focus = %v, want volumes", m.focus)
	}
}

func TestSelectionSyncBumpsGenerationOnChange(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	m = send(t, m, containersMsg([]docker.Container{
		{ID: "a", Names: []string{"a"}, State: "running"},
		{ID: "b", Names: []string{"b"}, State: "running"},
	}))
	// Loading containers selects index 0 -> generation 1, selID "a".
	if m.selID != "a" || m.selGen != 1 {
		t.Fatalf("after load: selID=%q selGen=%d, want a/1", m.selID, m.selGen)
	}
	// Move to container b -> new generation, new selID.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyDown})
	if m.selID != "b" || m.selGen != 2 {
		t.Fatalf("after down: selID=%q selGen=%d, want b/2", m.selID, m.selGen)
	}
	// Switching away from containers clears the selection.
	m = send(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("2")})
	if m.selID != "" {
		t.Fatalf("after switching to images, selID=%q, want empty", m.selID)
	}
}

func TestStaleStreamMessagesIgnored(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	m = send(t, m, containersMsg([]docker.Container{{ID: "a", Names: []string{"a"}, State: "running"}}))
	// Current generation is 1. A stats message from an older generation must be
	// dropped, not applied.
	m = send(t, m, statsMsg{gen: 0, stats: docker.Stats{CPUPercent: 99}})
	if m.hasStats {
		t.Fatal("stale stats applied")
	}
	// A current-generation message is applied.
	m = send(t, m, statsMsg{gen: m.selGen, stats: docker.Stats{CPUPercent: 12.5}})
	if !m.hasStats || m.stats.CPUPercent != 12.5 {
		t.Fatalf("current stats not applied: hasStats=%v cpu=%v", m.hasStats, m.stats.CPUPercent)
	}
}

func TestEventTriggersRefreshCommand(t *testing.T) {
	f := &fakedocker.Fake{}
	m := newTestModel(f)
	m.events = make(chan docker.Event) // so recvEvent has a channel to re-listen on
	_, cmd := m.Update(eventMsg(docker.Event{Type: "container", Action: "start"}))
	if cmd == nil {
		t.Fatal("container event produced no refresh command")
	}
}

func TestViewDoesNotPanicWhenEmpty(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	if got := m.View(); got == "" {
		t.Fatal("empty view")
	}
}

func TestViewRendersContainer(t *testing.T) {
	m := newTestModel(&fakedocker.Fake{})
	m = send(t, m, containersMsg([]docker.Container{
		{ID: "abcdef0123456789", Names: []string{"web"}, Image: "nginx", State: "running"},
	}))
	view := m.View()
	if !strings.Contains(view, "web") || !strings.Contains(view, "Containers") {
		t.Fatalf("view missing expected content:\n%s", view)
	}
}
