package tui

import (
	"strings"
	"testing"

	"github.com/3sarojbhattarai/gantry/internal/docker"
	"github.com/3sarojbhattarai/gantry/internal/docker/fakedocker"
	tea "github.com/charmbracelet/bubbletea"
)

func keyRune(r rune) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}}
}

func TestLifecycleKeyProducesCommand(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", Names: []string{"a"}, State: "exited"}}}
	m := newTestModel(f)
	m = send(t, m, containersMsg(f.Containers))

	// 's' = start, on the focused container.
	_, cmd := m.Update(keyRune('s'))
	if cmd == nil {
		t.Fatal("start key produced no command")
	}
	// Executing the command should start the container in the fake.
	cmd()
	if f.Containers[0].State != "running" {
		t.Fatalf("state = %q, want running", f.Containers[0].State)
	}
}

func TestRemoveOpensConfirmModalThenExecutes(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", Names: []string{"a"}, State: "exited"}}}
	m := newTestModel(f)
	m = send(t, m, containersMsg(f.Containers))

	// 'd' opens the confirmation modal, does not remove yet.
	m = send(t, m, keyRune('d'))
	if !m.confirming {
		t.Fatal("remove key did not open confirmation modal")
	}
	if !strings.Contains(m.View(), "Remove container a?") {
		t.Fatalf("modal prompt not shown:\n%s", m.View())
	}
	if len(f.Containers) != 1 {
		t.Fatal("container removed before confirmation")
	}

	// 'y' confirms: the returned command performs the removal.
	next, cmd := m.Update(keyRune('y'))
	m = next.(Model)
	if m.confirming {
		t.Fatal("still confirming after y")
	}
	if cmd == nil {
		t.Fatal("confirm produced no command")
	}
	cmd()
	if len(f.Containers) != 0 {
		t.Fatal("container not removed after confirming")
	}
}

func TestRemoveModalCancelled(t *testing.T) {
	f := &fakedocker.Fake{Containers: []docker.Container{{ID: "a", Names: []string{"a"}, State: "exited"}}}
	m := newTestModel(f)
	m = send(t, m, containersMsg(f.Containers))

	m = send(t, m, keyRune('d'))
	m = send(t, m, keyRune('n')) // cancel
	if m.confirming {
		t.Fatal("modal still open after 'n'")
	}
	if len(f.Containers) != 1 {
		t.Fatal("container removed despite cancel")
	}
}

func TestLifecycleIgnoredOffContainersPane(t *testing.T) {
	f := &fakedocker.Fake{
		Containers: []docker.Container{{ID: "a", State: "exited"}},
		Images:     []docker.ImageSummary{{ID: "img"}},
	}
	m := newTestModel(f)
	m = send(t, m, containersMsg(f.Containers))
	m = send(t, m, imagesMsg(f.Images))
	m = send(t, m, keyRune('2')) // focus images

	if _, cmd := m.Update(keyRune('s')); cmd != nil {
		t.Fatal("start should be a no-op when the containers pane isn't focused")
	}
}
