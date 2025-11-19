package components

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/ui"
	"github.com/yourusername/gitman/internal/ui/layout"
)

// ModalType defines the type of modal
type ModalType int

const (
	ModalInfo ModalType = iota
	ModalError
	ModalWarning
	ModalConfirm
	ModalForm
)

// ModalButton represents a button in the modal
type ModalButton struct {
	Label    string
	Primary  bool
	OnSelect func()
}

// ModalInput represents a text input in a form modal
type ModalInput struct {
	Label       string
	Placeholder string
	Input       textinput.Model
}

// Modal represents a reusable modal component
type Modal struct {
	Type        ModalType
	Title       string
	Message     string
	Content     string // For custom content
	Buttons     []ModalButton
	Inputs      []ModalInput
	Width       int
	Height      int

	// Internal state
	selectedButton int
	focusedInput   int
	inputMode      bool // true when focused on inputs, false when on buttons
}

// NewModal creates a new modal with default settings
func NewModal(modalType ModalType, title, message string) *Modal {
	width := layout.ModalWidthMD
	height := layout.ModalHeightMD

	// Adjust size based on type
	if modalType == ModalError || modalType == ModalInfo {
		height = layout.ModalHeightSM
	}

	return &Modal{
		Type:           modalType,
		Title:          title,
		Message:        message,
		Width:          width,
		Height:         height,
		selectedButton: 0,
		focusedInput:   0,
		inputMode:      false,
	}
}

// NewConfirmModal creates a Yes/No confirmation modal
func NewConfirmModal(title, message string, onYes, onNo func()) *Modal {
	m := NewModal(ModalConfirm, title, message)
	m.Buttons = []ModalButton{
		{Label: "Yes", Primary: true, OnSelect: onYes},
		{Label: "No", Primary: false, OnSelect: onNo},
	}
	return m
}

// NewErrorModal creates an error display modal
func NewErrorModal(message string) *Modal {
	m := NewModal(ModalError, "ERROR", message)
	m.Height = layout.ModalHeightSM
	return m
}

// NewFormModal creates a modal with inputs and buttons
func NewFormModal(title, message string, inputs []ModalInput, buttons []ModalButton) *Modal {
	m := NewModal(ModalForm, title, message)
	m.Inputs = inputs
	m.Buttons = buttons
	m.Height = layout.ModalHeightLG
	m.inputMode = len(inputs) > 0
	return m
}

// Update handles keyboard input
func (m *Modal) Update(msg string) {
	switch msg {
	case "tab", "shift+tab":
		if len(m.Inputs) > 0 && len(m.Buttons) > 0 {
			// Toggle between inputs and buttons
			m.inputMode = !m.inputMode
			if m.inputMode {
				m.focusedInput = 0
			} else {
				m.selectedButton = 0
			}
		}
	case "up", "left":
		if m.inputMode && len(m.Inputs) > 0 {
			m.focusedInput--
			if m.focusedInput < 0 {
				m.focusedInput = len(m.Inputs) - 1
			}
		} else if len(m.Buttons) > 0 {
			m.selectedButton--
			if m.selectedButton < 0 {
				m.selectedButton = len(m.Buttons) - 1
			}
		}
	case "down", "right":
		if m.inputMode && len(m.Inputs) > 0 {
			m.focusedInput++
			if m.focusedInput >= len(m.Inputs) {
				m.focusedInput = 0
			}
		} else if len(m.Buttons) > 0 {
			m.selectedButton++
			if m.selectedButton >= len(m.Buttons) {
				m.selectedButton = 0
			}
		}
	}
}

// UpdateInput forwards input to the focused text input
func (m *Modal) UpdateInput(msg interface{}) {
	if m.inputMode && len(m.Inputs) > 0 {
		m.Inputs[m.focusedInput].Input.Update(msg)
	}
}

// GetSelectedButton returns the currently selected button
func (m *Modal) GetSelectedButton() *ModalButton {
	if len(m.Buttons) > 0 && m.selectedButton >= 0 && m.selectedButton < len(m.Buttons) {
		return &m.Buttons[m.selectedButton]
	}
	return nil
}

// GetInputValues returns all input values as a map
func (m *Modal) GetInputValues() map[string]string {
	values := make(map[string]string)
	for _, input := range m.Inputs {
		values[input.Label] = input.Input.Value()
	}
	return values
}

// Render renders the modal
func (m *Modal) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	// Build content
	var content strings.Builder

	// Title with icon
	titleStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
	titleIcon := ""

	switch m.Type {
	case ModalError:
		titleStyle = styles.StatusError
		titleIcon = "✗ "
	case ModalWarning:
		titleStyle = styles.StatusWarning
		titleIcon = "⚠ "
	case ModalInfo:
		titleStyle = styles.StatusInfo
		titleIcon = "ℹ "
	case ModalConfirm:
		titleStyle = lipgloss.NewStyle().Foreground(styles.ColorPrimary)
		titleIcon = "? "
	case ModalForm:
		titleStyle = lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	}

	title := titleStyle.Bold(true).Render(titleIcon + m.Title)
	content.WriteString(title + "\n\n")

	// Message or custom content
	if m.Content != "" {
		content.WriteString(m.Content)
	} else if m.Message != "" {
		messageStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
		if m.Type == ModalError {
			messageStyle = styles.StatusError
		}
		content.WriteString(messageStyle.Render(m.Message))
	}

	content.WriteString("\n\n")

	// Inputs (for form modals)
	if len(m.Inputs) > 0 {
		textStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
		for i, input := range m.Inputs {
			focused := m.inputMode && i == m.focusedInput

			label := textStyle.Render(input.Label + ":")
			content.WriteString(label + "\n")

			inputStyle := styles.FormInput
			if focused {
				inputStyle = styles.FormInputFocused
			}

			content.WriteString(inputStyle.Render(input.Input.View()) + "\n\n")
		}
	}

	// Buttons
	if len(m.Buttons) > 0 {
		content.WriteString("\n")
		var buttons []string
		for i, btn := range m.Buttons {
			selected := !m.inputMode && i == m.selectedButton
			buttons = append(buttons, m.renderButton(btn, selected))
		}
		content.WriteString(lipgloss.JoinHorizontal(lipgloss.Center, buttons...))
	}

	// Help text
	mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	if m.Type == ModalError {
		content.WriteString("\n\n")
		help := mutedStyle.Render("Press any key to dismiss")
		content.WriteString(help)
	} else {
		content.WriteString("\n\n")
		helpParts := []string{}

		if len(m.Inputs) > 0 && len(m.Buttons) > 0 {
			helpParts = append(helpParts, "Tab to navigate")
		}
		if len(m.Buttons) > 1 {
			helpParts = append(helpParts, "←/→ to switch")
		}
		if len(m.Buttons) > 0 {
			helpParts = append(helpParts, "Enter to confirm")
		}
		helpParts = append(helpParts, "Esc to cancel")

		help := mutedStyle.Render(strings.Join(helpParts, " • "))
		content.WriteString(help)
	}

	// Modal container
	theme := ui.GetGlobalThemeManager().GetCurrentTheme()
	modalStyle := lipgloss.NewStyle().
		Width(m.Width).
		Padding(layout.SpacingMD).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder)

	// Set background based on modal type
	switch m.Type {
	case ModalError:
		modalStyle = modalStyle.Background(lipgloss.Color(theme.Backgrounds.ErrorModal))
	case ModalConfirm, ModalForm:
		modalStyle = modalStyle.Background(lipgloss.Color(theme.Backgrounds.Confirmation))
	default:
		modalStyle = modalStyle.Background(lipgloss.Color(theme.Backgrounds.Modal))
	}

	return modalStyle.Render(content.String())
}

// renderButton renders a single button
func (m *Modal) renderButton(btn ModalButton, selected bool) string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	buttonStyle := lipgloss.NewStyle().
		Padding(0, layout.SpacingMD).
		MarginRight(layout.SpacingSM).
		Border(lipgloss.RoundedBorder())

	if selected {
		if btn.Primary {
			buttonStyle = buttonStyle.
				Background(styles.ColorPrimary).
				Foreground(lipgloss.Color("#000000")).
				BorderForeground(styles.ColorPrimary).
				Bold(true)
		} else {
			buttonStyle = buttonStyle.
				BorderForeground(styles.ColorPrimary).
				Foreground(styles.ColorPrimary).
				Bold(true)
		}
	} else {
		buttonStyle = buttonStyle.
			BorderForeground(styles.ColorBorder).
			Foreground(styles.ColorMuted)
	}

	return buttonStyle.Render(btn.Label)
}

// RenderCentered renders the modal centered on screen
func (m *Modal) RenderCentered(windowWidth, windowHeight int) string {
	modalContent := m.Render()

	// Calculate position
	x := layout.CenterHorizontal(windowWidth, m.Width)
	y := layout.CenterVertical(windowHeight, m.Height)

	// Create overlay with modal positioned
	overlay := lipgloss.NewStyle().
		Width(windowWidth).
		Height(windowHeight).
		Padding(y, 0, 0, x)

	return overlay.Render(modalContent)
}

// Helper function to create a text input for modals
func NewModalInput(label, placeholder, initialValue string) ModalInput {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetValue(initialValue)
	ti.CharLimit = 200
	ti.Width = layout.ModalWidthMD - 10

	return ModalInput{
		Label:       label,
		Placeholder: placeholder,
		Input:       ti,
	}
}
