package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/revrost/glimpse/styles"
)

// ProgressModel represents a progress bar
type ProgressModel struct {
	width      int
	total      int
	current    int
	percent    float64
	text       string
	characters []string
}

// NewProgress creates a new progress bar
func NewProgress(width, total int, text string) *ProgressModel {
	return &ProgressModel{
		width:      width,
		total:      total,
		current:    0,
		percent:    0.0,
		text:       text,
		characters: []string{" ", "▏", "▎", "▍", "▌", "▋", "▊", "▉", "█"},
	}
}

// Update updates the progress bar
func (p *ProgressModel) Update(current int) {
	if current > p.total {
		current = p.total
	}
	p.current = current
	p.percent = float64(current) / float64(p.total)
}

// View renders the progress bar
func (p *ProgressModel) View() string {
	filled := int(p.percent * float64(p.width))
	remaining := p.width - filled
	
	// Build progress bar
	var bar strings.Builder
	bar.WriteString(styles.Success.Render(strings.Repeat("█", filled)))
	
	if remaining > 0 {
		bar.WriteString(styles.Muted.Render(strings.Repeat("░", remaining)))
	}
	
	// Build full progress view
	progressText := fmt.Sprintf("%s [%d/%d] %.1f%%", 
		p.text, p.current, p.total, p.percent*100)
	
	return styles.Text.Render(progressText) + " " + bar.String()
}

// AnimatedProgress represents an animated progress indicator
type AnimatedProgress struct {
	frames []string
	index  int
	text   string
	active bool
}

// NewAnimatedProgress creates a new animated progress
func NewAnimatedProgress(text string) *AnimatedProgress {
	return &AnimatedProgress{
		frames: []string{
			"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏",
		},
		index:  0,
		text:   text,
		active: true,
	}
}

// Stop stops the animation
func (a *AnimatedProgress) Stop() {
	a.active = false
}

// Tick advances the animation
func (a *AnimatedProgress) Tick() string {
	if !a.active {
		return ""
	}
	
	frame := a.frames[a.index%len(a.frames)]
	a.index++
	
	return styles.Spinner.Render(frame + " " + a.text)
}

// FileTable represents a styled file table
type FileTable struct {
	table *table.Table
}

// NewFileTable creates a new file table
func NewFileTable() *FileTable {
	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(styles.RoundedBorder).
		Headers("File", "Status", "Type").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return styles.Subtitle.Padding(0, 1)
			}
			
			switch col {
			case 0: // File column
				return styles.FilePath.Padding(0, 1)
			case 1: // Status column
				return styles.Text.Padding(0, 1)
			case 2: // Type column
				return styles.Code.Padding(0, 1)
			default:
				return styles.Text.Padding(0, 1)
			}
		})
	
	return &FileTable{table: t}
}

// AddRow adds a row to the table
func (ft *FileTable) AddRow(file, status, fileType string) {
	ft.table.Row(file, status, fileType)
}

// String returns the table as a string
func (ft *FileTable) String() string {
	return ft.table.String()
}

// StatusBar represents a status bar
type StatusBar struct {
	left  string
	right string
}

// NewStatusBar creates a new status bar
func NewStatusBar(left, right string) *StatusBar {
	return &StatusBar{
		left:  left,
		right: right,
	}
}

// View renders the status bar
func (s *StatusBar) View() string {
	// Get terminal width (using a reasonable default if unknown)
	width := 80
	
	// Calculate padding needed
	leftLen := lipgloss.Width(s.left)
	rightLen := lipgloss.Width(s.right)
	padding := width - leftLen - rightLen
	if padding < 0 {
		padding = 0
	}
	
	leftStyle := styles.Status.Background(styles.HighlightBg)
	rightStyle := styles.Status.Background(styles.HighlightBg)
	
	return leftStyle.Render(s.left) + 
		strings.Repeat(" ", padding) + 
		rightStyle.Render(s.right)
}

// BorderedBox creates a bordered box around content
func BorderedBox(title, content string) string {
	titleStyle := styles.Title.Padding(0, 2)
	contentStyle := styles.Text.Padding(1, 2)
	
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.PrimaryColor).
		BorderTop(true).
		BorderBottom(true).
		BorderLeft(true).
		BorderRight(true).
		Padding(0).
		Width(80)
	
	// Create title and content
	result := titleStyle.Render(title) + "\n"
	result += contentStyle.Render(content)
	
	return box.Render(result)
}

// InfoBox creates an info box with styled content
func InfoBox(title, message string) string {
	boxStyle := styles.InfoContainer.
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.InfoColor)
	
	content := styles.Bold.Render(title) + "\n" + styles.Text.Render(message)
	return boxStyle.Render(content)
}

// WarningBox creates a warning box with styled content
func WarningBox(title, message string) string {
	boxStyle := styles.WarningContainer.
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.WarningColor)
	
	content := styles.Bold.Render(title) + "\n" + styles.Text.Render(message)
	return boxStyle.Render(content)
}

// ErrorBox creates an error box with styled content
func ErrorBox(title, message string) string {
	boxStyle := styles.ErrorContainer.
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.ErrorColor)
	
	content := styles.Bold.Render(title) + "\n" + styles.Text.Render(message)
	return boxStyle.Render(content)
}

// SuccessBox creates a success box with styled content
func SuccessBox(title, message string) string {
	boxStyle := styles.SuccessContainer.
		Border(lipgloss.NormalBorder()).
		BorderForeground(styles.SuccessColor)
	
	content := styles.Bold.Render(title) + "\n" + styles.Text.Render(message)
	return boxStyle.Render(content)
}

// Key bindings display
type KeyBindings struct {
	bindings map[string]string
}

// NewKeyBindings creates a new key bindings display
func NewKeyBindings() *KeyBindings {
	return &KeyBindings{
		bindings: make(map[string]string),
	}
}

// Add adds a key binding
func (kb *KeyBindings) Add(key, description string) {
	kb.bindings[key] = description
}

// View renders the key bindings
func (kb *KeyBindings) View() string {
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(styles.Border).
		Headers("Key", "Action").
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return styles.Subtitle.Padding(0, 1)
			}
			
			if col == 0 { // Key column
				return styles.Code.Padding(0, 1)
			}
			
			return styles.Text.Padding(0, 1)
		})
	
	for key, desc := range kb.bindings {
		t.Row(key, desc)
	}
	
	return t.String()
}

// LoadingWithText creates a loading indicator with custom text
func LoadingWithText(text string) string {
	return styles.Loading.Render("⠋") + " " + styles.Text.Render(text)
}

// Separator creates a visual separator
func Separator(width int) string {
	return styles.Border.Render(strings.Repeat("─", width))
}

// Timestamp formats a timestamp with style
func Timestamp() string {
	now := time.Now().Format("15:04:05")
	return styles.Muted.Render("[" + now + "]")
}