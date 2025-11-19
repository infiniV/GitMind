package components

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/lipgloss"
)

// WrapText wraps text to a specified width while preserving words
func WrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	// Use lipgloss for wrapping with word boundaries
	return lipgloss.NewStyle().Width(width).Render(text)
}

// WrapTextManual wraps text manually with custom word breaking logic
func WrapTextManual(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		// If word is too long for the width, break it
		if len(word) > width {
			// Flush current line if it has content
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}

			// Break long word into chunks
			for len(word) > width {
				lines = append(lines, word[:width])
				word = word[width:]
			}

			if len(word) > 0 {
				currentLine.WriteString(word)
			}
			continue
		}

		// Check if adding this word would exceed width
		testLine := currentLine.String()
		if testLine != "" {
			testLine += " "
		}
		testLine += word

		if len(testLine) > width {
			// Start a new line
			lines = append(lines, currentLine.String())
			currentLine.Reset()
			currentLine.WriteString(word)
		} else {
			if currentLine.Len() > 0 {
				currentLine.WriteString(" ")
			}
			currentLine.WriteString(word)
		}
	}

	// Add the last line if it has content
	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}

// TruncateText truncates text to a maximum length with ellipsis
func TruncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	if maxLen < 3 {
		return text[:maxLen]
	}

	return text[:maxLen-3] + "..."
}

// TruncateMiddle truncates text in the middle with ellipsis
func TruncateMiddle(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}

	if maxLen < 3 {
		return text[:maxLen]
	}

	// Keep equal parts from start and end
	halfLen := (maxLen - 3) / 2
	return text[:halfLen] + "..." + text[len(text)-halfLen:]
}

// PadRight pads text to the right with spaces
func PadRight(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return text + strings.Repeat(" ", width-len(text))
}

// PadLeft pads text to the left with spaces
func PadLeft(text string, width int) string {
	if len(text) >= width {
		return text
	}
	return strings.Repeat(" ", width-len(text)) + text
}

// Center centers text within a given width
func Center(text string, width int) string {
	textLen := len(text)
	if textLen >= width {
		return text
	}

	leftPad := (width - textLen) / 2
	rightPad := width - textLen - leftPad

	return strings.Repeat(" ", leftPad) + text + strings.Repeat(" ", rightPad)
}

// IndentText indents all lines of text by a specified amount
func IndentText(text string, spaces int) string {
	indent := strings.Repeat(" ", spaces)
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if line != "" {
			lines[i] = indent + line
		}
	}

	return strings.Join(lines, "\n")
}

// StripANSI removes ANSI color codes from text
func StripANSI(text string) string {
	// Simple ANSI stripper - removes escape sequences
	var result strings.Builder
	inEscape := false

	for _, r := range text {
		if r == '\x1b' {
			inEscape = true
			continue
		}

		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		result.WriteRune(r)
	}

	return result.String()
}

// CountVisibleChars counts visible characters (excluding ANSI codes)
func CountVisibleChars(text string) int {
	return len(StripANSI(text))
}

// Pluralize returns singular or plural form based on count
func Pluralize(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}

// FormatCount formats a count with singular/plural noun
func FormatCount(count int, singular, plural string) string {
	countStr := lipgloss.NewStyle().Bold(true).Render(string(rune('0' + count)))
	word := Pluralize(count, singular, plural)
	return countStr + " " + word
}

// SplitIntoLines splits text into lines respecting existing line breaks
func SplitIntoLines(text string, maxWidth int) []string {
	var result []string

	// First split by existing newlines
	paragraphs := strings.Split(text, "\n")

	for _, paragraph := range paragraphs {
		if paragraph == "" {
			result = append(result, "")
			continue
		}

		// Wrap each paragraph
		wrapped := WrapTextManual(paragraph, maxWidth)
		lines := strings.Split(wrapped, "\n")
		result = append(result, lines...)
	}

	return result
}

// TitleCase converts text to title case
func TitleCase(text string) string {
	words := strings.Fields(text)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = string(unicode.ToUpper(rune(word[0]))) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

// HighlightText highlights a substring within text
func HighlightText(text, highlight string, highlightStyle lipgloss.Style) string {
	if highlight == "" {
		return text
	}

	// Case-insensitive search
	lowerText := strings.ToLower(text)
	lowerHighlight := strings.ToLower(highlight)

	index := strings.Index(lowerText, lowerHighlight)
	if index == -1 {
		return text
	}

	before := text[:index]
	match := text[index : index+len(highlight)]
	after := text[index+len(highlight):]

	return before + highlightStyle.Render(match) + after
}

// FormatList formats a list of items with bullets
func FormatList(items []string, bulletChar string) string {
	if bulletChar == "" {
		bulletChar = "•"
	}

	var lines []string
	for _, item := range items {
		lines = append(lines, bulletChar+" "+item)
	}

	return strings.Join(lines, "\n")
}

// FormatNumberedList formats a list with numbers
func FormatNumberedList(items []string) string {
	var lines []string
	for i, item := range items {
		lines = append(lines, lipgloss.NewStyle().Render(string(rune('1'+i))+". "+item))
	}

	return strings.Join(lines, "\n")
}

// JoinWithSeparator joins strings with a separator and styling
func JoinWithSeparator(items []string, separator string) string {
	return strings.Join(items, separator)
}

// EllipsizeMiddle shortens a path or long string by removing middle content
func EllipsizeMiddle(text string, maxLen int, separator string) string {
	if len(text) <= maxLen {
		return text
	}

	if separator == "" {
		separator = "..."
	}

	sepLen := len(separator)
	if maxLen < sepLen {
		return text[:maxLen]
	}

	// Calculate how much to keep on each side
	sideLen := (maxLen - sepLen) / 2

	return text[:sideLen] + separator + text[len(text)-sideLen:]
}
