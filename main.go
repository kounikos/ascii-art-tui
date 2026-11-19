// Package main coordinates the Bubbletea terminal session and handles dynamic UI rendering.
package main

import (
	"log"
	"strings"
	"time"

	"ascii-art-tui/internal/ascii"
	"ascii-art-tui/internal/renderer"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	// Palettes
	primaryColor   = lipgloss.Color("#FF7CCB") // Neon Pink
	secondaryColor = lipgloss.Color("#B4BEFE") // Lavender
	textColor      = lipgloss.Color("#FFFDF5") // Off-white
	mutedColor     = lipgloss.Color("#383838") // Dark Gray
	lightMuted     = lipgloss.Color("#626262") // Lighter Gray
	darkBg         = lipgloss.Color("#1E1E2E") // Very dark

	titleStyle = lipgloss.NewStyle().Foreground(darkBg).Background(primaryColor).Padding(0, 1).Bold(true)

	activeTabStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(primaryColor).Foreground(primaryColor).Padding(0, 1).Bold(true)
	inactiveTabStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), true).BorderForeground(mutedColor).Foreground(lightMuted).Padding(0, 1)

	focusedBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(primaryColor).Padding(0, 1)
	blurredBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(mutedColor).Padding(0, 1)

	previewBoxStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(secondaryColor).Padding(0, 1)
	previewTitleStyle = lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
			
	centeredMessageStyle = lipgloss.NewStyle().Foreground(secondaryColor).Bold(true)
)

// appState defines the logical lifecycle phases of the core application router.
type appState int

const (
	stateLoading appState = iota
	stateDashboard
	stateExiting
)

// loadingTimeoutMsg signals the completion of the boot sequence animation.
type loadingTimeoutMsg struct{}

// exitingTimeoutMsg signals the completion of the graceful shutdown sleep.
type exitingTimeoutMsg struct{}

// tickLoading initiates a 1.5s delay command during instantiation.
func tickLoading() tea.Cmd {
	return tea.Tick(time.Millisecond*1500, func(t time.Time) tea.Msg {
		return loadingTimeoutMsg{}
	})
}

// tickExiting initiates a 1.0s delay command when a terminal exit is caught.
func tickExiting() tea.Cmd {
	return tea.Tick(time.Millisecond*1000, func(t time.Time) tea.Msg {
		return exitingTimeoutMsg{}
	})
}

// matchRow defines a singular paired row in the multi-rule Dashboard structure.
type matchRow struct {
	subInput textinput.Model
	colInput textinput.Model
}

// newMatchRow spawns a fresh Bubbletea form group instantiated natively with default char boundaries.
func newMatchRow() matchRow {
	su := textinput.New()
	su.Placeholder = "Match..."
	su.Prompt = "> "
	su.CharLimit = 100

	co := textinput.New()
	co.Placeholder = "Color"
	co.Prompt = " "
	co.CharLimit = 50
	return matchRow{subInput: su, colInput: co}
}

// model represents the root Bubbletea structural state tree mapping directly globally across the framework.
type model struct {
	state     appState
	spinner   spinner.Model
	textarea  textarea.Model
	viewport  viewport.Model
	rows      []matchRow
	focus     int // 0 = textarea, 1 = sub1, 2 = col1, 3 = sub2...
	
	styles    []string
	styleIdx  int
	ready     bool
	err       error
	output    string
	width     int
	height    int
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type something...\nSupports newlines!"
	ta.Prompt = "┃ "
	ta.CharLimit = 1000
	ta.ShowLineNumbers = false

	s := spinner.New()
	s.Spinner = spinner.MiniDot
	s.Style = lipgloss.NewStyle().Foreground(primaryColor)

	m := model{
		state:    stateLoading,
		spinner:  s,
		textarea: ta,
		rows:     []matchRow{newMatchRow()},
		focus:    0,
		styles:   []string{"standard", "shadow", "thinkertoy"},
		styleIdx: 0,
		ready:    false,
	}
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, tickLoading())
}

func (m *model) cycleFocusD() {
	maxFocus := len(m.rows) * 2
	m.focus = (m.focus + 1) % (maxFocus + 1)
}

func (m *model) cycleFocusU() {
	maxFocus := len(m.rows) * 2
	m.focus--
	if m.focus < 0 {
		m.focus = maxFocus
	}
}

func (m *model) updateFocusCmds() []tea.Cmd {
	var cmds []tea.Cmd
	m.textarea.Blur()
	for i := range m.rows {
		m.rows[i].subInput.Blur()
		m.rows[i].colInput.Blur()
	}

	if m.focus == 0 {
		cmds = append(cmds, m.textarea.Focus())
	} else {
		idx := (m.focus - 1) / 2
		isCol := (m.focus - 1) % 2 == 1
		if isCol {
			cmds = append(cmds, m.rows[idx].colInput.Focus())
		} else {
			cmds = append(cmds, m.rows[idx].subInput.Focus())
		}
	}
	// blink cursors natively
	cmds = append(cmds, textinput.Blink, textarea.Blink)
	return cmds
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd  tea.Cmd

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		m.width = msg.Width
		m.height = msg.Height

		// Layout logic: dynamically expand the dashboard bounds
		// Control panel takes up max 20 lines, header 4 lines.
		headerHeight := 20
		innerWidth := msg.Width - 4 

		if !m.ready {
			m.viewport = viewport.New(innerWidth-4, m.height-headerHeight)
			m.ready = true
		} else {
			m.viewport.Width = innerWidth - 4
			m.viewport.Height = max(10, m.height-headerHeight)
		}

		m.textarea.SetHeight(5)
		
		// 40% left, 60% right
		leftWidth := (innerWidth * 2) / 5   
		rightWidth := innerWidth - leftWidth - 2 

		m.textarea.SetWidth(leftWidth - 4) 
		
		for i := range m.rows {
			m.rows[i].subInput.Width = (rightWidth / 2) - 6
			m.rows[i].colInput.Width = (rightWidth / 2) - 6
		}

		m.updateAscii()
	}

	switch m.state {
	case stateLoading:
		switch msg := msg.(type) {
		case loadingTimeoutMsg:
			m.state = stateDashboard
			cmds = append(cmds, m.updateFocusCmds()...)
			return m, tea.Batch(cmds...)
		case spinner.TickMsg:
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		case tea.KeyMsg:
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
				return m, tea.Quit
			}
		}

	case stateExiting:
		switch msg := msg.(type) {
		case exitingTimeoutMsg:
			return m, tea.Quit
		case tea.KeyMsg:
			if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEsc {
				return m, tea.Quit 
			}
		}

	case stateDashboard:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyEsc:
				m.state = stateExiting
				return m, tickExiting()
			case tea.KeyCtrlT:
				m.styleIdx = (m.styleIdx + 1) % len(m.styles)
				m.updateAscii()
				return m, nil
			case tea.KeyCtrlA:
				if len(m.rows) < 4 {
					m.rows = append(m.rows, newMatchRow())
					if m.width > 0 {
						innerWidth := m.width - 4
						leftWidth := (innerWidth * 2) / 5
						rightWidth := innerWidth - leftWidth - 2
						idx := len(m.rows) - 1
						m.rows[idx].subInput.Width = (rightWidth / 2) - 6
						m.rows[idx].colInput.Width = (rightWidth / 2) - 6
					}
				}
				return m, nil
			case tea.KeyCtrlX:
				if len(m.rows) > 1 {
					m.rows = m.rows[:len(m.rows)-1]
					maxFocus := len(m.rows) * 2
					if m.focus > maxFocus {
						m.focus = maxFocus // clamp focus seamlessly
					}
					cmds = append(cmds, m.updateFocusCmds()...)
					m.updateAscii()
				}
				return m, tea.Batch(cmds...)
			case tea.KeyTab, tea.KeyShiftTab:
				if msg.Type == tea.KeyShiftTab {
					m.cycleFocusU()
				} else {
					m.cycleFocusD()
				}
				cmds = append(cmds, m.updateFocusCmds()...)
				return m, tea.Batch(cmds...)
			}
		}

		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)

		if m.focus == 0 {
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
		} else {
			idx := (m.focus - 1) / 2
			isCol := (m.focus - 1) % 2 == 1
			if isCol {
				m.rows[idx].colInput, cmd = m.rows[idx].colInput.Update(msg)
			} else {
				m.rows[idx].subInput, cmd = m.rows[idx].subInput.Update(msg)
			}
			cmds = append(cmds, cmd)
		}

		if _, ok := msg.(tea.KeyMsg); ok {
			m.updateAscii()
		}
	}

	// Always sync widths actively just in case a row was added without a WS msg
	if len(m.rows) > 0 && m.width > 0 {
		innerWidth := m.width - 4
		leftWidth := (innerWidth * 2) / 5   
		rightWidth := innerWidth - leftWidth - 2
		for i := range m.rows {
			m.rows[i].subInput.Width = (rightWidth / 2) - 6
			m.rows[i].colInput.Width = (rightWidth / 2) - 6
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *model) updateAscii() {
	if !m.ready {
		return
	}
	text := m.textarea.Value()
	if text == "" {
		m.output = ""
		m.err = nil
		m.viewport.SetContent("")
		return
	}

	style := m.styles[m.styleIdx]
	
	colorMap := map[string]string{
		"red":     "#FF5555",
		"green":   "#50FA7B",
		"yellow":  "#F1FA8C",
		"blue":    "#8BE9FD",
		"magenta": "#FF79C6",
		"cyan":    "#8BE9FD",
		"white":   "#F8F8F2",
		"black":   "#000000",
		"pink":    "#FF7CCB",
		"orange":  "#FFB86C",
		"purple":  "#BD93F9",
	}

	var parsedRules []renderer.ColorRule
	for _, row := range m.rows {
		hex := row.colInput.Value()
		if parsed, ok := colorMap[strings.ToLower(strings.TrimSpace(hex))]; ok {
			hex = parsed
		}
		parsedRules = append(parsedRules, renderer.ColorRule{
			Substring: row.subInput.Value(),
			ColorHex:  hex,
		})
	}

	cfg := &renderer.ColorConfig{
		Rules: parsedRules,
	}

	out, err := ascii.GenerateAscii(text, style, cfg)
	if err != nil {
		m.err = err
		m.viewport.SetContent(lipgloss.NewStyle().Foreground(primaryColor).Render(err.Error()))
	} else {
		m.err = nil
		m.output = out
		m.viewport.SetContent(m.output)
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m model) View() string {
	if m.width == 0 {
		return "\n  Booting system..."
	}

	switch m.state {
	case stateLoading:
		msg := lipgloss.JoinHorizontal(lipgloss.Center, m.spinner.View(), " ", centeredMessageStyle.Render("Booting ASCII Engine..."))
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	
	case stateExiting:
		msg := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render("✨"),
			"\n"+centeredMessageStyle.Render("Goodbye! Shutting down..."),
		)
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	
	case stateDashboard:
		return m.dashboardView()
	}

	return ""
}

func (m model) dashboardView() string {
	if !m.ready {
		return "\n  Initializing Dashboard..."
	}

	var b strings.Builder

	// Header
	title := titleStyle.Render(" ASCII STUDIO ")
	b.WriteString(title)
	b.WriteString("\n\n")

	// Tabs
	var tabs []string
	for i, style := range m.styles {
		display := strings.ToUpper(style[:1]) + style[1:]
		if i == m.styleIdx {
			tabs = append(tabs, activeTabStyle.Render(display))
		} else {
			tabs = append(tabs, inactiveTabStyle.Render(display))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...))
	b.WriteString("\n")

	// Split panel layout
	innerWidth := m.width - 4
	leftWidth := (innerWidth * 2) / 5
	rightWidth := innerWidth - leftWidth - 2

	// Left: Text Input
	var leftPane string
	leftTitle := lipgloss.NewStyle().Foreground(lightMuted).Render(" Input Text ")
	if m.focus == 0 {
		leftTitle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(" Input Text ")
		leftPane = focusedBoxStyle.Width(leftWidth).Render(leftTitle + "\n" + m.textarea.View())
	} else {
		leftPane = blurredBoxStyle.Width(leftWidth).Render(leftTitle + "\n" + m.textarea.View())
	}

	// Right: Dynamic Rules Manager
	var rulesBuilder strings.Builder
	for i, r := range m.rows {
		subPrefix := lipgloss.NewStyle().Foreground(lightMuted).Render(" Match: ")
		colPrefix := lipgloss.NewStyle().Foreground(lightMuted).Render(" Color: ")
		if m.focus == 1+(i*2) {
			subPrefix = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(" Match: ")
		}
		if m.focus == 2+(i*2) {
			colPrefix = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(" Color: ")
		}

		rowContent := lipgloss.JoinHorizontal(lipgloss.Top, 
			subPrefix+r.subInput.View(), 
			"   ", 
			colPrefix+r.colInput.View(),
		)
		rulesBuilder.WriteString(rowContent + "\n")
	}

	rightTitle := lipgloss.NewStyle().Foreground(lightMuted).Render(" Highlighting Rules ([CTRL+A] Add • [CTRL+X] Delete) ")
	var rightPane string
	if m.focus > 0 {
		rightTitle = lipgloss.NewStyle().Foreground(primaryColor).Bold(true).Render(" Highlighting Rules ([CTRL+A] Add • [CTRL+X] Delete) ")
		rightPane = focusedBoxStyle.Width(rightWidth).Render(rightTitle + "\n" + rulesBuilder.String())
	} else {
		rightPane = blurredBoxStyle.Width(rightWidth).Render(rightTitle + "\n" + rulesBuilder.String())
	}

	controlsRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "  ", rightPane)
	b.WriteString(controlsRow)
	b.WriteString("\n")

	// Viewport
	previewTitle := previewTitleStyle.Render(" Output Viewport ")
	
	// Dynamically squeeze the vertical remaining height robustly
	calcH := m.height - lipgloss.Height(b.String()) - 3
	if calcH < 5 {
		calcH = 5
	}
	m.viewport.Height = calcH

	previewBox := previewBoxStyle.Width(m.viewport.Width + 2).Render(previewTitle + "\n" + m.viewport.View())
	b.WriteString(previewBox)

	body := lipgloss.NewStyle().MarginLeft(2).Render(b.String())

	// Highly Visible Status Footer Bar
	mode := lipgloss.NewStyle().Foreground(darkBg).Background(primaryColor).Padding(0, 1).Bold(true).Render(" READY ")
	helpStr := " [TAB] Switch Input • [CTRL+A] New Rule • [CTRL+T] Theme • [ESC] Quit "
	helpView := lipgloss.NewStyle().Foreground(darkBg).Background(secondaryColor).Padding(0, 1).Bold(true).Render(helpStr)

	gap := m.width - lipgloss.Width(mode) - lipgloss.Width(helpView)
	var filler string
	if gap > 0 {
		filler = lipgloss.NewStyle().Background(mutedColor).Width(gap).Render("")
	}

	statusBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		mode,
		filler,
		helpView,
	)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		body,
		"\n",
		statusBar,
	)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
