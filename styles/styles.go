package styles

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color scheme for Glimpse
var (
	// Primary colors - brand colors
	PrimaryColor   = lipgloss.Color("99")  // Purple
	SecondaryColor = lipgloss.Color("205") // Pink
	AccentColor    = lipgloss.Color("86")  // Aqua

	// Semantic colors
	SuccessColor = lipgloss.Color("46")  // Green
	WarningColor = lipgloss.Color("208") // Orange
	ErrorColor   = lipgloss.Color("196") // Red
	InfoColor    = lipgloss.Color("39")  // Blue

	// Text colors
	TitleColor    = lipgloss.Color("231") // White
	SubtitleColor = lipgloss.Color("250") // Light gray
	TextColor     = lipgloss.Color("244") // Gray
	MutedColor    = lipgloss.Color("238") // Dark gray
	CodeColor     = lipgloss.Color("194") // Light cyan

	// Background colors
	PrimaryBg   = lipgloss.Color("99")  // Purple
	SecondaryBg = lipgloss.Color("205") // Pink
	SuccessBg   = lipgloss.Color("46")  // Green
	WarningBg   = lipgloss.Color("208") // Orange
	ErrorBg     = lipgloss.Color("196") // Red
	InfoBg      = lipgloss.Color("39")  // Blue
	HighlightBg = lipgloss.Color("236") // Dark gray
	BorderBg    = lipgloss.Color("238") // Dark gray
)

// Base styles
var (
	// Title style for headers and main titles
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)

	// Subtitle style for secondary headings
	Subtitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			MarginBottom(1)

	// Text style for regular text
	Text = lipgloss.NewStyle().
		Foreground(TextColor)

	// Muted text style for less important information
	Muted = lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	// Success text
	Success = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	// Warning text
	Warning = lipgloss.NewStyle().
		Foreground(WarningColor).
		Bold(true)

	// Error text
	Error = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	// Info text
	Info = lipgloss.NewStyle().
		Foreground(InfoColor).
		Bold(true)

	// Code style for inline code
	Code = lipgloss.NewStyle().
		Foreground(CodeColor).
		Background(HighlightBg).
		Padding(0, 1).
		SetString("`")

	// Border style
	Border = lipgloss.NewStyle().
		Foreground(BorderBg)

	// Highlight style for emphasis
	Highlight = lipgloss.NewStyle().
			Background(PrimaryBg).
			Foreground(TitleColor).
			Padding(0, 1).
			Bold(true)
)

// Component styles
var (
	// Header style for the app header
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Background(HighlightBg).
		Padding(0, 2).
		MarginBottom(1)

	// Footer style for the app footer
	Footer = lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	// Status style for status messages
	Status = lipgloss.NewStyle().
		Foreground(InfoColor).
		Padding(0, 1)

	// Loading style for loading indicators
	Loading = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Bold(true)

	// Spinner style for loading spinners
	Spinner = lipgloss.NewStyle().
		Foreground(AccentColor).
		Bold(true)

	// File path style
	FilePath = lipgloss.NewStyle().
			Foreground(InfoColor).
			Italic(true)

	// Batch header style
	BatchHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Background(HighlightBg).
			Padding(0, 2).
			MarginTop(1).
			MarginBottom(1)

	// Diff header style
	DiffHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(WarningColor).
			MarginTop(1).
			MarginBottom(0)

	// Provider info style
	ProviderInfo = lipgloss.NewStyle().
			Foreground(InfoColor).
			Background(HighlightBg).
			Padding(0, 1)

	// Error container style
	ErrorContainer = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Background(HighlightBg).
			Padding(1, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(ErrorColor).
			MarginTop(1)

	// Success container style
	SuccessContainer = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Background(HighlightBg).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(SuccessColor).
				MarginTop(1)

	// Warning container style
	WarningContainer = lipgloss.NewStyle().
				Foreground(WarningColor).
				Background(HighlightBg).
				Padding(1, 2).
				Border(lipgloss.NormalBorder()).
				BorderForeground(WarningColor).
				MarginTop(1)

	// Info container style
	InfoContainer = lipgloss.NewStyle().
			Foreground(InfoColor).
			Background(HighlightBg).
			Padding(1, 2).
			Border(lipgloss.NormalBorder()).
			BorderForeground(InfoColor).
			MarginTop(1)

	// Bold style for emphasis
	Bold = lipgloss.NewStyle().
		Bold(true)
)

// Border styles
var (
	// Normal border
	NormalBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(BorderBg)

	// Rounded border
	RoundedBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor)

	// Thick border
	ThickBorder = lipgloss.NewStyle().
			BorderStyle(lipgloss.ThickBorder()).
			BorderForeground(AccentColor)
)

// Utility functions for creating dynamic styles

// CreateErrorStyle creates a styled error message
func CreateErrorStyle(message string) string {
	return ErrorContainer.Render(message)
}

// CreateSuccessStyle creates a styled success message
func CreateSuccessStyle(message string) string {
	return SuccessContainer.Render(message)
}

// CreateInfoStyle creates a styled info message
func CreateInfoStyle(message string) string {
	return Info.Render(message)
}

// CreateWarningStyle creates a styled warning message
func CreateWarningStyle(message string) string {
	return Warning.Render(message)
}

// CreateBatchHeader creates a styled batch header
func CreateGlossMessage(message string) string {
	return BatchHeader.Render(message)
}

// CreateBatchHeader creates a styled batch header
func CreateBatchHeader(count int) string {
	return BatchHeader.Render(fmt.Sprintf("Processing batch of %d changes", count))
}

// CreateDiffHeader creates a styled diff header
func CreateDiffHeader(filename string) string {
	return DiffHeader.Render(fmt.Sprintf("File: %s", filename))
}

// CreateProviderInfo creates styled provider info
func CreateProviderInfo(provider, model string) string {
	return ProviderInfo.Render(fmt.Sprintf("Analyzing with %s (%s)", provider, model))
}

// CreateFileList creates a styled file list
func CreateFileList(files []string) string {
	var builder strings.Builder
	builder.WriteString(Subtitle.Render("Files changed:\n"))
	for i, file := range files {
		builder.WriteString(Text.Render(fmt.Sprintf("  %d. %s\n", i+1, FilePath.Render(file))))
	}
	return builder.String()
}

// CreateHeader creates a styled header with title
func CreateHeader(title string) string {
	return Header.Render(title)
}

// CreateFooter creates a styled footer with text
func CreateFooter(text string) string {
	return Footer.Render(text)
}
