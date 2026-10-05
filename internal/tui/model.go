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
	desc := i.Description()
	if index == m.Index() {
		name = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Render("› " + name)
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  " + desc)
	} else {
		name = "  " + name
		desc = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("  " + desc)
	}
	fmt.Fprintf(w, "%s\n%s", name, desc)
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
	return Model{manager: manager, list: l, width: 100, height: 24}
}

func (m Model) selected() *theme.Theme {
	i, ok := m.list.SelectedItem().(item)
	if !ok {
		return nil
	}
	t := i.theme
	return &t
}

func (m Model) previewWidth() int {
	if m.width < 90 {
		return max(30, m.width-4)
	}
	return max(30, m.width-m.list.Width()-8)
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
		m.width, m.height = msg.Width, msg.Height
		if m.width >= 90 {
			m.list.SetSize(max(34, m.width*2/5), max(10, m.height-2))
		} else {
			m.list.SetSize(max(30, m.width-2), max(8, m.height*3/5))
		}
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

func (m Model) View() tea.View {
	if m.quitting {
		return tea.NewView("")
	}

	selected := m.selected()
	name := "Preview"
	if selected != nil {
		name = "Preview — " + selected.Name
	}
	preview := m.preview
	if m.previewErr != nil {
		preview = lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(m.previewErr.Error())
	} else if preview == "" || selected == nil || m.previewFor != selected.Name {
		preview = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Rendering Starship preview…")
	}
	if strings.TrimSpace(preview) == "" {
		preview = "(empty prompt)"
	}

	previewStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2)
	previewTitle := lipgloss.NewStyle().Bold(true).Render(name)
	help := lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter apply • / filter • ↑/↓ move • Esc/q quit")

	if m.width >= 90 {
		pw := max(30, m.previewWidth())
		panel := previewStyle.Width(pw).Height(max(8, m.height-6)).Render(previewTitle + "\n\n" + preview + "\n\n" + help)
		content := lipgloss.JoinHorizontal(lipgloss.Top, m.list.View(), "  ", panel)
		return tea.NewView(content)
	}

	panel := previewStyle.Width(max(26, m.width-6)).Render(previewTitle + "\n\n" + preview + "\n\n" + help)
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, m.list.View(), panel))
}

func (m Model) Choice() *theme.Theme {
	return m.choice
}
