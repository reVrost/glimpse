package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMarkdownRenderer(t *testing.T) {
	renderer, err := NewMarkdownRenderer()
	assert.NoError(t, err)
	assert.NotNil(t, renderer)
	
	// Test plain text (no markdown)
	plainText := "This is plain text"
	rendered := renderer.RenderResponse(plainText)
	assert.Equal(t, plainText, rendered)
	
	// Test markdown content
	markdown := "# Heading\n\nThis has **bold** text."
	rendered = renderer.RenderResponse(markdown)
	assert.NotEqual(t, markdown, rendered) // Should be different after rendering
	assert.Contains(t, rendered, "Heading") // Should contain heading
}

func TestSpinner(t *testing.T) {
	spinner := NewSpinner("test")
	firstTick := spinner.Tick()
	secondTick := spinner.Tick()
	
	assert.NotEmpty(t, firstTick)
	assert.NotEmpty(t, secondTick)
	assert.NotEqual(t, firstTick, secondTick) // Should advance frame
}