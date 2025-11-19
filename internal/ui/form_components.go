package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TextInput represents a text input field
type TextInput struct {
	Label       string
	Value       string
	Placeholder string
	Password    bool
	Focused     bool
	Width       int
}

// NewTextInput creates a new text input
func NewTextInput(label, placeholder string) TextInput {
	return TextInput{
		Label:       label,
		Placeholder: placeholder,
		Value:       "",
		Password:    false,
		Focused:     false,
		Width:       40,
	}
}

// Update handles key input for text input
func (t *TextInput) Update(msg tea.KeyMsg) {
	if !t.Focused {
		return
	}

	switch msg.String() {
	case "backspace":
		if len(t.Value) > 0 {
			t.Value = t.Value[:len(t.Value)-1]
		}
	case "space":
		t.Value += " "
	default:
		// Only add printable characters
		if len(msg.String()) == 1 {
			t.Value += msg.String()
		}
	}
}

// View renders the text input
func (t TextInput) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	// Label
	label := styles.FormLabel.Render(t.Label + ":")

	// Input value or placeholder
	displayValue := t.Value
	if displayValue == "" && !t.Focused {
		displayValue = t.Placeholder
	} else if displayValue == "" && t.Focused {
		displayValue = "" // Show empty with cursor
	} else if t.Password {
		displayValue = strings.Repeat("*", len(t.Value))
	}

	// Add cursor when focused
	if t.Focused {
		displayValue += "█" // Block cursor
	}

	// Input field
	var inputStyle lipgloss.Style
	if t.Focused {
		inputStyle = styles.FormInputFocused.Width(t.Width)
		// Apply different text color for the value vs placeholder
		if t.Value == "" {
			displayValue = lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(t.Placeholder) +
				lipgloss.NewStyle().Foreground(styles.ColorPrimary).Render("█")
		}
	} else {
		inputStyle = styles.FormInput.Width(t.Width)
		if t.Value == "" {
			displayValue = lipgloss.NewStyle().Foreground(styles.ColorMuted).Render(t.Placeholder)
		}
	}

	input := inputStyle.Render(displayValue)

	return lipgloss.JoinHorizontal(lipgloss.Top, label, " ", input)
}

// Checkbox represents a single checkbox
type Checkbox struct {
	Label   string
	Checked bool
	Focused bool
}

// NewCheckbox creates a new checkbox
func NewCheckbox(label string, checked bool) Checkbox {
	return Checkbox{
		Label:   label,
		Checked: checked,
		Focused: false,
	}
}

// Toggle toggles the checkbox state
func (c *Checkbox) Toggle() {
	c.Checked = !c.Checked
}

// View renders the checkbox
func (c Checkbox) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	checkbox := "☐"
	if c.Checked {
		checkbox = "☑"
	}

	var style lipgloss.Style
	prefix := "  "
	if c.Focused {
		style = styles.OptionCursor
		prefix = "> " // Arrow indicator for focused
	} else {
		style = styles.OptionNormal
	}

	return prefix + style.Render(checkbox+" "+c.Label)
}

// RadioGroup represents a group of radio buttons
type RadioGroup struct {
	Label    string
	Options  []string
	Selected int
	Focused  bool
}

// NewRadioGroup creates a new radio group
func NewRadioGroup(label string, options []string, defaultIndex int) RadioGroup {
	return RadioGroup{
		Label:    label,
		Options:  options,
		Selected: defaultIndex,
		Focused:  false,
	}
}

// Next moves to the next option
func (r *RadioGroup) Next() {
	r.Selected = (r.Selected + 1) % len(r.Options)
}

// Previous moves to the previous option
func (r *RadioGroup) Previous() {
	r.Selected = (r.Selected - 1 + len(r.Options)) % len(r.Options)
}

// View renders the radio group
func (r RadioGroup) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Label
	if r.Label != "" {
		lines = append(lines, styles.FormLabel.Render(r.Label+":"))
	}

	// Options
	for i, option := range r.Options {
		radio := "☐"
		if i == r.Selected {
			radio = "☑"
		}

		var style lipgloss.Style
		prefix := "  "
		suffix := ""

		if r.Focused && i == r.Selected {
			style = styles.OptionCursor
			prefix = "> " // Arrow for focused
			// Show navigation hint on focused option
			suffix = " " + lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("(←/→)")
		} else if i == r.Selected {
			style = styles.OptionSelected
		} else {
			style = styles.OptionNormal
		}

		lines = append(lines, prefix+style.Render(radio+" "+option)+suffix)
	}

	return strings.Join(lines, "\n")
}

// GetSelected returns the currently selected option
func (r RadioGroup) GetSelected() string {
	if r.Selected >= 0 && r.Selected < len(r.Options) {
		return r.Options[r.Selected]
	}
	return ""
}

// CheckboxGroup represents multiple checkboxes
type CheckboxGroup struct {
	Label      string
	Items      []Checkbox
	FocusedIdx int
}

// NewCheckboxGroup creates a new checkbox group
func NewCheckboxGroup(label string, options []string, checked []bool) CheckboxGroup {
	items := make([]Checkbox, len(options))
	for i, opt := range options {
		isChecked := false
		if i < len(checked) {
			isChecked = checked[i]
		}
		items[i] = NewCheckbox(opt, isChecked)
	}

	return CheckboxGroup{
		Label:      label,
		Items:      items,
		FocusedIdx: 0,
	}
}

// Next moves focus to next item
func (c *CheckboxGroup) Next() {
	c.FocusedIdx = (c.FocusedIdx + 1) % len(c.Items)
}

// Previous moves focus to previous item
func (c *CheckboxGroup) Previous() {
	c.FocusedIdx = (c.FocusedIdx - 1 + len(c.Items)) % len(c.Items)
}

// Toggle toggles the focused item
func (c *CheckboxGroup) Toggle() {
	if c.FocusedIdx >= 0 && c.FocusedIdx < len(c.Items) {
		c.Items[c.FocusedIdx].Toggle()
	}
}

// View renders the checkbox group
func (c CheckboxGroup) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var lines []string

	// Label
	if c.Label != "" {
		lines = append(lines, styles.FormLabel.Render(c.Label+":"))
	}

	// Checkboxes
	for i := range c.Items {
		c.Items[i].Focused = (i == c.FocusedIdx)
		// Don't add extra prefix - Checkbox.View() already handles it ("> " or "  ")
		lines = append(lines, c.Items[i].View())
	}

	return strings.Join(lines, "\n")
}

// GetChecked returns a slice of checked values
func (c CheckboxGroup) GetChecked() []string {
	var checked []string
	for _, item := range c.Items {
		if item.Checked {
			checked = append(checked, item.Label)
		}
	}
	return checked
}

// Button represents a clickable button
type Button struct {
	Label   string
	Active  bool
	Focused bool
}

// NewButton creates a new button
func NewButton(label string) Button {
	return Button{
		Label:   label,
		Active:  true,
		Focused: false,
	}
}

// View renders the button
func (b Button) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	var style lipgloss.Style
	if !b.Active {
		style = styles.FormButtonInactive
	} else if b.Focused {
		style = styles.FormButton.Border(lipgloss.RoundedBorder()).BorderForeground(styles.ColorPrimary)
	} else {
		style = styles.FormButton
	}

	return style.Render(b.Label)
}

// Dropdown represents a dropdown/select component
type Dropdown struct {
	Label    string
	Options  []string
	Selected int
	Focused  bool
	Open     bool
}

// NewDropdown creates a new dropdown
func NewDropdown(label string, options []string, defaultIndex int) Dropdown {
	return Dropdown{
		Label:    label,
		Options:  options,
		Selected: defaultIndex,
		Focused:  false,
		Open:     false,
	}
}

// Next moves to next option
func (d *Dropdown) Next() {
	d.Selected = (d.Selected + 1) % len(d.Options)
}

// Previous moves to previous option
func (d *Dropdown) Previous() {
	d.Selected = (d.Selected - 1 + len(d.Options)) % len(d.Options)
}

// Toggle toggles the dropdown open/closed
func (d *Dropdown) Toggle() {
	d.Open = !d.Open
}

// View renders the dropdown
func (d Dropdown) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	label := styles.FormLabel.Render(d.Label + ":")

	selectedValue := d.Options[d.Selected]
	arrow := "v"
	navHint := ""

	if d.Open {
		arrow = "^"
		if d.Focused {
			// Show navigation hint when dropdown is open and focused
			navHint = " " + lipgloss.NewStyle().Foreground(styles.ColorMuted).Render("(←/→)")
		}
	}

	var style lipgloss.Style
	if d.Focused {
		style = styles.FormInputFocused.Width(38)
	} else {
		style = styles.FormInput.Width(38)
	}

	dropdown := style.Render(selectedValue + " " + arrow)

	result := lipgloss.JoinHorizontal(lipgloss.Top, label, " ", dropdown) + navHint

	// If open, show options below
	if d.Open {
		var options []string
		for i, opt := range d.Options {
			if i == d.Selected {
				options = append(options, styles.OptionSelected.Render("  > "+opt))
			} else {
				options = append(options, styles.OptionNormal.Render("    "+opt))
			}
		}
		result += "\n" + strings.Join(options, "\n")
	}

	return result
}

// GetSelected returns the selected option
func (d Dropdown) GetSelected() string {
	if d.Selected >= 0 && d.Selected < len(d.Options) {
		return d.Options[d.Selected]
	}
	return ""
}

// HelpText represents helper text for a field
type HelpText struct {
	Text string
}

// View renders the help text
func (h HelpText) View() string {
	styles := GetGlobalThemeManager().GetStyles()
	return styles.FormHelp.Render(h.Text)
}
