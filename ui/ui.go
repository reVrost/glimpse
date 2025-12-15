package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/revrost/glimpse/styles"
)

// LoadingModel represents a loading animation model
type LoadingModel struct {
	frames []string
	index  int
	text   string
	quit   bool
}

// LoadingMsg is a message to update the loading animation
type LoadingMsg struct{}

// NewLoading creates a new loading animation model
func NewLoading(text string) *LoadingModel {
	return &LoadingModel{
		frames: []string{
			"[    ]",
			"[=   ]",
			"[==  ]",
			"[=== ]",
			"[====]",
			"[ ===]",
			"[  ==]",
			"[   =]",
		},
		index: 0,
		text:  text,
		quit:  false,
	}
}

// Init implements tea.Model
func (m *LoadingModel) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle("Glimpse"),
		tickCmd(),
	)
}

// Update implements tea.Model
func (m *LoadingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.quit = true
			return m, tea.Quit
		}
	case LoadingMsg:
		m.index = (m.index + 1) % len(m.frames)
		return m, tickCmd()
	}
	return m, nil
}

// View implements tea.Model
func (m *LoadingModel) View() string {
	frame := m.frames[m.index]
	text := m.text

	return styles.Loading.Render(frame) + " " + styles.Text.Render(text)
}

// tickCmd sends a message to update the loading animation
func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(time.Time) tea.Msg {
		return LoadingMsg{}
	})
}

// RunLoadingAnimation starts a loading animation and returns a function to stop it
func RunLoadingAnimation(text string) (*tea.Program, func()) {
	p := tea.NewProgram(NewLoading(text), tea.WithOutput(os.Stdout))

	go func() {
		if _, err := p.Run(); err != nil {
			fmt.Fprintln(os.Stderr, styles.CreateErrorStyle(fmt.Sprintf("Error running loading animation: %v", err)))
		}
	}()

	stop := func() {
		p.Quit()
	}

	return p, stop
}

// SpinnerModel represents a compact spinner model
type SpinnerModel struct {
	frames []string
	index  int
	text   string
}

// NewSpinner creates a new spinner model
func NewSpinner(text string) *SpinnerModel {
	return &SpinnerModel{
		frames: []string{
			"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
		},
		index: 0,
		text:  text,
	}
}

// Tick updates the spinner
func (m *SpinnerModel) Tick() string {
	frame := m.frames[m.index%len(m.frames)]
	m.index++
	return styles.Spinner.Render(frame + " " + m.text)
}

// MarkdownRenderer handles markdown rendering with glamour
type MarkdownRenderer struct {
	renderer *glamour.TermRenderer
}

// NewMarkdownRenderer creates a new markdown renderer
func NewMarkdownRenderer() (*MarkdownRenderer, error) {
	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create glamour renderer: %w", err)
	}

	return &MarkdownRenderer{
		renderer: renderer,
	}, nil
}

// Render renders markdown text
func (r *MarkdownRenderer) Render(markdown string) (string, error) {
	return r.renderer.Render(markdown)
}

// RenderResponse renders an LLM response with markdown
func (r *MarkdownRenderer) RenderResponse(content string) string {
	// Check if content contains markdown syntax
	hasMarkdown := strings.Contains(content, "#") ||
		strings.Contains(content, "**") ||
		strings.Contains(content, "*") ||
		strings.Contains(content, "`") ||
		strings.Contains(content, "```") ||
		strings.Contains(content, "- ") ||
		strings.Contains(content, "1. ")

	if !hasMarkdown {
		// If no markdown detected, return as-is
		return content
	}

	rendered, err := r.renderer.Render(content)
	if err != nil {
		// Fallback to plain text if rendering fails
		return content
	}

	return rendered
}
