package output

import (
	"fmt"
	"strings"
	"sync"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"
)

// policyTreeFooterHint documents the policy tree view's navigation keys.
const policyTreeFooterHint = "↑/↓ navigate · enter/space expand · e/c expand/collapse all · PgUp/PgDn (b/f) page · q/Esc quit"

// policyTreeCursorStyle highlights the node under the cursor. It always
// replaces (never combines with) a kind's color below — nesting two
// lipgloss.Render calls on the same line would let the inner call's reset
// code cut the outer Reverse short partway through the line.
var policyTreeCursorStyle = lipgloss.NewStyle().Reverse(true)

// Per-kind colors, applied to a whole line so statement types are
// distinguishable at a glance without expanding every node. Colors are the
// shared adaptive palette (see interactive.go) so the tree reads correctly
// on both light and dark terminal backgrounds. Each is a function, not a
// resolved Style, so building it doesn't resolve the underlying adaptive
// color (and so query the terminal) until the tree actually renders — see
// hasDarkBackground's doc comment in interactive.go.
var (
	policyTreeBranchStyle = sync.OnceValue(func() lipgloss.Style { // nested policy branches
		return lipgloss.NewStyle().Bold(true).Foreground(adaptivePink())
	})
	policyTreeResourceStyle = sync.OnceValue(func() lipgloss.Style { // host/variable/group/layer/user/...
		return lipgloss.NewStyle().Foreground(adaptiveGreen())
	})
	policyTreePermitStyle = sync.OnceValue(func() lipgloss.Style { // permit/deny
		return lipgloss.NewStyle().Foreground(adaptiveAmber())
	})
	policyTreeLeafStyle = sync.OnceValue(func() lipgloss.Style { // flattened statement fields
		return lipgloss.NewStyle().Foreground(adaptiveGray())
	})
	policyTreeFallbackStyle = sync.OnceValue(func() lipgloss.Style {
		return lipgloss.NewStyle().Italic(true).Foreground(adaptiveGray())
	})
)

// policyTreeStyleFor picks the color for a node's kind; unrecognized kinds
// (the empty string never occurs today, but would mean "no special case")
// render in the terminal's default color.
func policyTreeStyleFor(kind string) lipgloss.Style {
	switch kind {
	case "policy":
		return policyTreeBranchStyle()
	case "resource":
		return policyTreeResourceStyle()
	case "permit":
		return policyTreePermitStyle()
	case "leaf":
		return policyTreeLeafStyle()
	case "fallback":
		return policyTreeFallbackStyle()
	default:
		return lipgloss.NewStyle()
	}
}

// policyNode is one statement (or one of its fields) in the policy tree.
// Nodes are shared pointers rather than values so expand/collapse state
// mutated by the tea.Model's Update survives being copied by value on every
// bubbletea update, the way the rest of this package's models are written.
// kind selects the node's color (see policyTreeStyleFor) — it is a display
// concern only, not used for tree logic.
type policyNode struct {
	label    string
	kind     string
	children []*policyNode
	expanded bool
}

// RunPolicyTree renders a fetched policy's top-level statements (the
// []any a policy fetch response unmarshals into) as a collapsible tree
// instead of the generic table/detail view: a flat array of deeply nested,
// differently-shaped statements (host/variable/permit/nested policy/...)
// reads as one unreadable blob per row through the generic flattener, so
// policy gets its own bespoke, recursive rendering.
func RunPolicyTree(cmd *cobra.Command, statements []any) error {
	if len(statements) == 0 {
		fmt.Fprintln(cmd.OutOrStdout(), "No policy statements found.")
		return nil
	}

	roots := buildPolicyTree(statements)
	expandTopLevel(roots)

	m := policyTreeModel{roots: roots, height: terminalHeight()}
	m.recomputeVisible()

	p := tea.NewProgram(m)
	_, err := p.Run()
	return err
}

// buildPolicyTree converts each top-level policy statement into a node.
func buildPolicyTree(statements []any) []*policyNode {
	nodes := make([]*policyNode, 0, len(statements))
	for _, s := range statements {
		nodes = append(nodes, buildStatementNode(s))
	}
	return nodes
}

// expandTopLevel marks each root statement expanded (but not their
// children), so a freshly-launched tree opens showing what's actually in
// the policy — a fully collapsed tree's only visible line per root is
// "policy: <id>" (or "host: ...", "permit: ...", etc.), which tells the
// user nothing until they press Enter.
func expandTopLevel(roots []*policyNode) {
	for _, root := range roots {
		root.expanded = true
	}
}

// buildStatementNode expects a policy statement's usual shape: a
// single-key map naming the statement type (host, variable, permit, a
// nested policy branch, ...). Anything else (an unexpected server
// response shape) still renders, just without a type-specific label.
func buildStatementNode(stmt any) *policyNode {
	m, ok := stmt.(map[string]any)
	if !ok || len(m) != 1 {
		return &policyNode{label: fmt.Sprintf("%v", stmt), kind: "fallback"}
	}
	for kind, value := range m {
		return buildTypedNode(kind, value)
	}
	return &policyNode{}
}

func buildTypedNode(kind string, value any) *policyNode {
	switch kind {
	case "policy":
		return buildPolicyBranchNode(value)
	case "permit", "deny":
		return buildPermitNode(kind, value)
	case "host", "variable", "group", "layer", "user", "hostfactory", "webservice":
		return buildResourceNode(kind, value)
	default:
		return &policyNode{label: kind, kind: "fallback", children: leafChildrenFor(value)}
	}
}

// buildPolicyBranchNode recurses into a nested policy's body, so a
// multi-branch policy document renders as a genuine tree rather than a
// single, deeply flattened node.
func buildPolicyBranchNode(value any) *policyNode {
	m, _ := value.(map[string]any)
	label := "policy"
	if id, ok := m["id"].(string); ok && id != "" {
		label = "policy: " + id
	}

	body, ok := m["body"].([]any)
	if !ok {
		return &policyNode{label: label, kind: "policy", children: leafChildrenFor(value)}
	}
	return &policyNode{label: label, kind: "policy", children: buildPolicyTree(body)}
}

// buildResourceNode handles statement types identified by an "id" field
// (host, variable, group, layer, user, ...): the label is "<kind>: <id>",
// and every other field of the statement becomes a leaf child.
func buildResourceNode(kind string, value any) *policyNode {
	label := kind
	if m, ok := value.(map[string]any); ok {
		if id, ok := m["id"].(string); ok && id != "" {
			label = kind + ": " + id
		}
	}
	return &policyNode{label: label, kind: "resource", children: leafChildrenFor(value)}
}

// buildPermitNode summarizes a permit/deny statement's role, resource, and
// privileges on one line instead of requiring the user to expand three
// separate fields to see who was granted what.
func buildPermitNode(kind string, value any) *policyNode {
	m, ok := value.(map[string]any)
	if !ok {
		return &policyNode{label: kind, kind: "permit", children: leafChildrenFor(value)}
	}
	label := fmt.Sprintf("%s: %s → %s (%s)",
		kind,
		describeRoleOrResource(m["role"]),
		describeRoleOrResource(m["resource"]),
		describePrivileges(m["privileges"]),
	)
	return &policyNode{label: label, kind: "permit", children: leafChildrenFor(value)}
}

// describeRoleOrResource renders a role/resource reference (e.g.
// {"host": "web-app"}) as "host:web-app".
func describeRoleOrResource(v any) string {
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	for kind, id := range m {
		return fmt.Sprintf("%s:%v", kind, id)
	}
	return "?"
}

func describePrivileges(v any) string {
	items, ok := v.([]any)
	if !ok {
		return fmt.Sprintf("%v", v)
	}
	parts := make([]string, len(items))
	for i, item := range items {
		parts[i] = fmt.Sprintf("%v", item)
	}
	return strings.Join(parts, ", ")
}

// leafChildrenFor flattens value the same way the generic detail view
// does (dotted paths, array count headers) and wraps each resulting row as
// a leaf, so fallback rendering for less common or unrecognized fields
// isn't reinvented.
func leafChildrenFor(value any) []*policyNode {
	rows := flattenDetailRows(value)
	children := make([]*policyNode, 0, len(rows))
	for _, r := range rows {
		field := strings.TrimSpace(r.field)
		label := field + ": " + r.value
		if r.header {
			label = field + " " + r.value
		}
		children = append(children, &policyNode{label: label, kind: "leaf"})
	}
	return children
}

// visiblePolicyNode is one line of the currently-rendered tree: a node
// plus how deep it is, used only for indentation.
type visiblePolicyNode struct {
	node  *policyNode
	depth int
}

type policyTreeModel struct {
	roots   []*policyNode
	visible []visiblePolicyNode
	cursor  int
	top     int
	height  int
}

func (m policyTreeModel) Init() tea.Cmd {
	return nil
}

// recomputeVisible flattens the tree into the currently-visible line list,
// descending only into expanded nodes' children — this is what lets ↑/↓
// and paging operate on a plain index rather than tree position.
func (m *policyTreeModel) recomputeVisible() {
	m.visible = nil
	for _, root := range m.roots {
		m.appendVisible(root, 0)
	}
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *policyTreeModel) appendVisible(n *policyNode, depth int) {
	m.visible = append(m.visible, visiblePolicyNode{node: n, depth: depth})
	if n.expanded {
		for _, c := range n.children {
			m.appendVisible(c, depth+1)
		}
	}
}

// policyTreeVisibleHeight reserves one line for policyTreeFooterHint,
// mirroring detailViewportHeight's reservation for the detail view.
func policyTreeVisibleHeight(height int) int {
	h := height - 1
	if h < 1 {
		h = 1
	}
	return h
}

func (m *policyTreeModel) ensureCursorVisible() {
	vh := policyTreeVisibleHeight(m.height)
	if m.cursor < m.top {
		m.top = m.cursor
	}
	if m.cursor >= m.top+vh {
		m.top = m.cursor - vh + 1
	}
}

func (m *policyTreeModel) pageDown() {
	vh := policyTreeVisibleHeight(m.height)
	m.cursor += vh
	if m.cursor >= len(m.visible) {
		m.cursor = len(m.visible) - 1
	}
	m.ensureCursorVisible()
}

func (m *policyTreeModel) pageUp() {
	vh := policyTreeVisibleHeight(m.height)
	m.cursor -= vh
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.ensureCursorVisible()
}

// toggleCursor expands or collapses the node under the cursor; leaf nodes
// (no children) ignore the toggle.
func (m *policyTreeModel) toggleCursor() {
	if m.cursor < 0 || m.cursor >= len(m.visible) {
		return
	}
	node := m.visible[m.cursor].node
	if len(node.children) == 0 {
		return
	}
	node.expanded = !node.expanded
	m.recomputeVisible()
}

func setExpandedAll(nodes []*policyNode, expanded bool) {
	for _, n := range nodes {
		n.expanded = expanded
		setExpandedAll(n.children, expanded)
	}
}

func (m policyTreeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.ensureCursorVisible()
		return m, nil
	case tea.KeyPressMsg:
		if msg.Mod == tea.ModCtrl && msg.Code == 'c' {
			return m, tea.Quit
		}
		switch msg.Code {
		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
				m.ensureCursorVisible()
			}
		case tea.KeyDown:
			if m.cursor < len(m.visible)-1 {
				m.cursor++
				m.ensureCursorVisible()
			}
		case tea.KeyPgUp, 'b':
			m.pageUp()
		case tea.KeyPgDown, 'f':
			m.pageDown()
		case tea.KeyEnter, ' ':
			m.toggleCursor()
		case 'e':
			setExpandedAll(m.roots, true)
			m.recomputeVisible()
		case 'c':
			setExpandedAll(m.roots, false)
			m.recomputeVisible()
			m.cursor = 0
			m.top = 0
		case tea.KeyEsc, 'q':
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m policyTreeModel) View() tea.View {
	var b strings.Builder

	vh := policyTreeVisibleHeight(m.height)
	end := m.top + vh
	if end > len(m.visible) {
		end = len(m.visible)
	}

	for i := m.top; i < end; i++ {
		vn := m.visible[i]
		marker := "  "
		if len(vn.node.children) > 0 {
			if vn.node.expanded {
				marker = "▾ "
			} else {
				marker = "▸ "
			}
		}
		line := strings.Repeat("  ", vn.depth) + marker + vn.node.label
		if i == m.cursor {
			line = policyTreeCursorStyle.Render(line)
		} else {
			line = policyTreeStyleFor(vn.node.kind).Render(line)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	view := b.String() + policyTreeFooterHint
	v := tea.NewView(view)
	v.AltScreen = true
	return v
}
