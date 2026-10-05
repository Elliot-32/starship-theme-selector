package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/Elliot-32/stheme/internal/theme"
)

const (
	wideBreakpoint         = 90
	previewPanelHeight     = 10
	previewPromptLines     = 2
	previewHorizontalFrame = 6 // 2 border cells + 2 cells of padding on each side
	previewFrameHeight     = previewPanelHeight
)

var (
	accent       = lipgloss.Color("212")
	muted        = lipgloss.Color("241")
	border       = lipgloss.Color("240")
	panelTitle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("252"))
	helpStyle    = lipgloss.NewStyle().Foreground(muted)
	loadingStyle = lipgloss.NewStyle().Foreground(muted).Italic(true)
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

type item struct {
	theme theme.Theme
}

func (i item) Title() string       { return i.theme.Name }
func (i item) Description() string { return string(i.theme.Source) }
func (i item) FilterValue() string { return i.theme.Name + " " + string(i.theme.Source) }

type previewMsg struct {
	name string
	text string
	err  error
}

type delegate struct{}

func (delegate) Height() int                             { return 2 }
func (delegate) Spacing() int                            { return 1 }
func (delegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (delegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	name := i.Title()
	source := i.Description()
	badge := lipgloss.NewStyle().Foreground(muted).Render(source)
	if i.theme.Source == theme.SourceCustom {
		badge = lipgloss.NewStyle().Foreground(lipgloss.Color("114")).Render(source)
	}

	if index == m.Index() {
		name = lipgloss.NewStyle().Bold(true).Foreground(accent).Render("› " + name)
		badge = "  " + badge
	} else {
		name = "  " + name
		badge = "  " + badge
	}
	fmt.Fprintf(w, "%s\n%s", name, badge)
}

type Model struct {
	manager    *theme.Manager
	list       list.Model
	choice     *theme.Theme
	preview    string
	previewErr error
	previewFor string
	width      int
	height     int
	quitting   bool
}

func New(manager *theme.Manager, themes []theme.Theme) Model {
	items := make([]list.Item, 0, len(themes))
	for _, t := range themes {
		items = append(items, item{theme: t})
	}
	l := list.New(items, delegate{}, 42, 18)
	l.Title = "stheme — Starship themes"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)
	l.AdditionalFullHelpKeys = nil

	m := Model{manager: manager, list: l, width: 100, height: 24}
	m.resize()
	return m
}

func (m Model) selected() *theme.Theme {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return nil
	}
	t := i.theme
	return &t
}

func (m *Model) resize() {
	if m.width >= wideBreakpoint {
		listWidth := max(34, min(48, m.width*2/5))
		m.list.SetSize(listWidth, max(8, m.height))
		return
	}

	listHeight := max(6, m.height-previewFrameHeight-1)
	m.list.SetSize(max(30, m.width), listHeight)
}

func (m Model) previewPanelWidth() int {
	if m.width < wideBreakpoint {
		return max(previewHorizontalFrame+16, m.width-2)
	}
	return max(previewHorizontalFrame+20, m.width-m.list.Width()-2)
}

func (m Model) previewContentWidth() int {
	return max(16, m.previewPanelWidth()-previewHorizontalFrame)
}

func (m Model) previewWidth() int {
	return max(20, m.previewContentWidth())
}

func (m Model) previewCmd(t theme.Theme) tea.Cmd {
	width := m.previewWidth()
	return func() tea.Msg {
		text, err := m.manager.Preview(context.Background(), t, width)
		return previewMsg{name: t.Name, text: text, err: err}
	}
}

func (m Model) Init() tea.Cmd {
	if selected := m.selected(); selected != nil {
		return m.previewCmd(*selected)
	}
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = max(1, msg.Width), max(1, msg.Height)
		m.resize()
		if selected := m.selected(); selected != nil {
			return m, m.previewCmd(*selected)
		}
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			if m.list.FilterState() == list.Filtering {
				break
			}
			m.quitting = true
			return m, tea.Quit
		case "q":
			if m.list.FilterState() == list.Filtering {
				break
			}
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if m.list.FilterState() == list.Filtering {
				break
			}
			if selected := m.selected(); selected != nil {
				m.choice = selected
				return m, tea.Quit
			}
		}

	case previewMsg:
		selected := m.selected()
		if selected != nil && selected.Name == msg.name {
			m.previewFor = msg.name
			m.preview = msg.text
			m.previewErr = msg.err
		}
		return m, nil
	}

	before := ""
	if selected := m.selected(); selected != nil {
		before = selected.Name
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	after := ""
	if selected := m.selected(); selected != nil {
		after = selected.Name
	}
	if after != "" && after != before {
		m.preview = ""
		m.previewErr = nil
		m.previewFor = ""
		selected := *m.selected()
		return m, tea.Batch(cmd, m.previewCmd(selected))
	}
	return m, cmd
}

func (m Model) previewPanel() string {
	selected := m.selected()
	name := "Preview"
	if selected != nil {
		name = "Preview — " + selected.Name
	}

	preview := m.preview
	if m.previewErr != nil {
		preview = errorStyle.Render(m.previewErr.Error())
	} else if preview == "" || selected == nil || m.previewFor != selected.Name {
		preview = loadingStyle.Render("Rendering Starship preview…")
	}
	if strings.TrimSpace(preview) == "" {
		preview = "(empty prompt)"
	}

	innerWidth := m.previewContentWidth()
	preview = lipgloss.Wrap(preview, innerWidth, " ")
	preview = lipgloss.NewStyle().MaxHeight(previewPromptLines).Render(preview)
	help := helpStyle.Render("Enter apply • / filter • ↑/↓ move • Esc/q quit")
	body := panelTitle.Render(name) + "\n\n" + preview + "\n\n" + help

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		Padding(1, 2).
		Width(m.previewPanelWidth()).
		MaxWidth(m.previewPanelWidth()).
		Height(previewPanelHeight).
		MaxHeight(previewPanelHeight).
		Render(body)
}

func (m Model) viewString() string {
	if m.quitting {
		return ""
	}

	panel := m.previewPanel()
	var content string
	if m.width >= wideBreakpoint {
		content = lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), "  ", panel)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Left, m.list.View(), panel)
	}

	// The preview panel itself has fixed outer dimensions, so async preview
	// updates cannot change the layout. Alt-screen rendering handles redraws;
	// forcing the entire view to terminal height would unnecessarily stretch
	// the list/panel composition.
	return content
}

func (m Model) View() tea.View {
	view := tea.NewView(m.viewString())
	view.AltScreen = true
	return view
}

func (m Model) Choice() *theme.Theme {
	return m.choice
}
