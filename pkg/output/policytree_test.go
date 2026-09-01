package output

import (
	"bytes"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/spf13/cobra"
)

func demoPolicyStatements() []any {
	return []any{
		map[string]any{
			"policy": map[string]any{
				"id": "demo",
				"body": []any{
					map[string]any{
						"host": map[string]any{
							"id":          "web-app",
							"annotations": map[string]any{"description": "Sample web app"},
						},
					},
					map[string]any{
						"variable": map[string]any{"id": "api-token"},
					},
					map[string]any{
						"permit": map[string]any{
							"role":       map[string]any{"host": "web-app"},
							"resource":   map[string]any{"variable": "api-token"},
							"privileges": []any{"execute", "read"},
						},
					},
				},
			},
		},
	}
}

func TestBuildPolicyTreeLabels(t *testing.T) {
	roots := buildPolicyTree(demoPolicyStatements())
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	policyNode := roots[0]
	if policyNode.label != "policy: demo" {
		t.Fatalf("expected policy branch label, got %q", policyNode.label)
	}
	if len(policyNode.children) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(policyNode.children))
	}

	host := policyNode.children[0]
	if host.label != "host: web-app" {
		t.Fatalf("expected host label, got %q", host.label)
	}

	variable := policyNode.children[1]
	if variable.label != "variable: api-token" {
		t.Fatalf("expected variable label, got %q", variable.label)
	}

	permit := policyNode.children[2]
	if permit.label != "permit: host:web-app → variable:api-token (execute, read)" {
		t.Fatalf("expected permit summary label, got %q", permit.label)
	}

	if got, want := policyNode.kind, "policy"; got != want {
		t.Fatalf("expected policy branch kind %q, got %q", want, got)
	}
	if got, want := host.kind, "resource"; got != want {
		t.Fatalf("expected host kind %q, got %q", want, got)
	}
	if got, want := permit.kind, "permit"; got != want {
		t.Fatalf("expected permit kind %q, got %q", want, got)
	}
	if len(host.children) == 0 || host.children[0].kind != "leaf" {
		t.Fatalf("expected host's flattened fields to have kind %q", "leaf")
	}
}

func TestBuildPolicyTreeFallbackForUnknownStatementType(t *testing.T) {
	roots := buildPolicyTree([]any{
		map[string]any{"grant": map[string]any{"role": map[string]any{"group": "admins"}}},
	})
	if len(roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(roots))
	}
	if roots[0].label != "grant" {
		t.Fatalf("expected fallback label 'grant', got %q", roots[0].label)
	}
	if roots[0].kind != "fallback" {
		t.Fatalf("expected fallback kind, got %q", roots[0].kind)
	}
	if len(roots[0].children) == 0 {
		t.Fatalf("expected fallback node to have flattened leaf children")
	}
}

func TestPolicyTreeModelDefaultsToCollapsed(t *testing.T) {
	// A policyTreeModel built directly (as opposed to via RunPolicyTree,
	// which calls expandTopLevel first) starts fully collapsed — that's
	// still the right default for a model that might be constructed for
	// reasons other than an initial render (e.g. after toggling back to
	// collapsed).
	m := policyTreeModel{roots: buildPolicyTree(demoPolicyStatements()), height: 24}
	m.recomputeVisible()

	if len(m.visible) != 1 {
		t.Fatalf("expected only the top-level policy node visible when collapsed, got %d", len(m.visible))
	}
}

func TestExpandTopLevelExpandsRootsButNotTheirChildren(t *testing.T) {
	roots := buildPolicyTree(demoPolicyStatements())
	expandTopLevel(roots)

	m := policyTreeModel{roots: roots, height: 24}
	m.recomputeVisible()

	// The demo policy is one root ("policy: demo") containing 3 direct
	// children (host/variable/permit) that each have their own leaf
	// fields — expandTopLevel should reveal exactly the 3 children, not
	// recurse into their fields too.
	if len(m.visible) != 4 {
		t.Fatalf("expected the root plus its 3 direct children visible, got %d", len(m.visible))
	}
	for _, vn := range m.visible[1:] {
		if vn.node.expanded {
			t.Fatalf("expected only the root to be pre-expanded, but a child was too: %q", vn.node.label)
		}
	}
}

func TestPolicyTreeToggleExpandsAndCollapses(t *testing.T) {
	m := policyTreeModel{roots: buildPolicyTree(demoPolicyStatements()), height: 24}
	m.recomputeVisible()

	m.toggleCursor()
	if len(m.visible) != 4 {
		t.Fatalf("expected policy node + 3 children visible after expand, got %d", len(m.visible))
	}

	m.toggleCursor()
	if len(m.visible) != 1 {
		t.Fatalf("expected collapse back to 1 visible node, got %d", len(m.visible))
	}
}

func TestPolicyTreeExpandAllAndCollapseAllShortcuts(t *testing.T) {
	m := policyTreeModel{roots: buildPolicyTree(demoPolicyStatements()), height: 24}
	m.recomputeVisible()

	updated, _ := m.Update(tea.KeyPressMsg{Code: 'e'})
	m = updated.(policyTreeModel)
	if len(m.visible) < 4 {
		t.Fatalf("expected 'e' to expand every node, got %d visible", len(m.visible))
	}

	updated, _ = m.Update(tea.KeyPressMsg{Code: 'c'})
	m = updated.(policyTreeModel)
	if len(m.visible) != 1 {
		t.Fatalf("expected 'c' to collapse back to 1 visible node, got %d", len(m.visible))
	}
	if m.cursor != 0 {
		t.Fatalf("expected cursor reset to 0 after collapse-all, got %d", m.cursor)
	}
}

func TestPolicyTreeArrowKeysNavigateAndClamp(t *testing.T) {
	m := policyTreeModel{roots: buildPolicyTree(demoPolicyStatements()), height: 24}
	m.recomputeVisible()
	m.toggleCursor() // expand to 4 visible nodes

	updated, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
	m = updated.(policyTreeModel)
	if m.cursor != 1 {
		t.Fatalf("expected cursor to advance to 1, got %d", m.cursor)
	}

	// Up past the top clamps at 0 rather than going negative.
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = updated.(policyTreeModel)
	updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyUp})
	m = updated.(policyTreeModel)
	if m.cursor != 0 {
		t.Fatalf("expected cursor clamped at 0, got %d", m.cursor)
	}

	// Down past the bottom clamps at len(visible)-1.
	for i := 0; i < 10; i++ {
		updated, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		m = updated.(policyTreeModel)
	}
	if m.cursor != len(m.visible)-1 {
		t.Fatalf("expected cursor clamped at %d, got %d", len(m.visible)-1, m.cursor)
	}
}

func TestPolicyTreeViewShowsFooterHint(t *testing.T) {
	m := policyTreeModel{roots: buildPolicyTree(demoPolicyStatements()), height: 24}
	m.recomputeVisible()

	view := m.View().Content
	if !strings.Contains(view, "expand/collapse all") {
		t.Fatalf("expected footer hint in view, got %q", view)
	}
	if !strings.Contains(view, "policy: demo") {
		t.Fatalf("expected the collapsed root label in view, got %q", view)
	}
}

func TestRunPolicyTreeReportsNoStatements(t *testing.T) {
	cmd := &cobra.Command{}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	if err := RunPolicyTree(cmd, []any{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := buf.String(); !strings.Contains(got, "No policy statements found.") {
		t.Fatalf("expected empty-statements message, got %q", got)
	}
}
