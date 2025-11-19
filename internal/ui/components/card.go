package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/ui"
	"github.com/yourusername/gitman/internal/ui/layout"
)

// CardType defines the type of card
type CardType int

const (
	CardDefault CardType = iota
	CardPrimary
	CardSuccess
	CardWarning
	CardError
	CardInfo
)

// Card represents a reusable card component
type Card struct {
	Title    string
	Content  string
	Width    int
	Height   int
	Type     CardType
	Active   bool
	Border   bool
	Padding  int
}

// NewCard creates a new card with default settings
func NewCard(title, content string) *Card {
	return &Card{
		Title:   title,
		Content: content,
		Width:   0, // Auto width
		Height:  0, // Auto height
		Type:    CardDefault,
		Active:  false,
		Border:  true,
		Padding: layout.SpacingMD,
	}
}

// NewDashboardCard creates a card styled for the dashboard
func NewDashboardCard(title, content string, width, height int) *Card {
	return &Card{
		Title:   title,
		Content: content,
		Width:   width,
		Height:  height,
		Type:    CardDefault,
		Active:  false,
		Border:  true,
		Padding: layout.SpacingSM,
	}
}

// SetActive sets the active state of the card
func (c *Card) SetActive(active bool) *Card {
	c.Active = active
	return c
}

// SetType sets the card type
func (c *Card) SetType(cardType CardType) *Card {
	c.Type = cardType
	return c
}

// SetSize sets the card dimensions
func (c *Card) SetSize(width, height int) *Card {
	c.Width = width
	c.Height = height
	return c
}

// Render renders the card
func (c *Card) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	// Start with base style
	var cardStyle lipgloss.Style

	// Choose style based on active state for dashboard cards
	if c.Active {
		cardStyle = styles.DashboardCardActive
	} else {
		cardStyle = styles.DashboardCard
	}

	// Apply custom styling based on type
	switch c.Type {
	case CardPrimary:
		cardStyle = cardStyle.BorderForeground(styles.ColorPrimary)
	case CardSuccess:
		cardStyle = cardStyle.BorderForeground(styles.ColorSuccess)
	case CardWarning:
		cardStyle = cardStyle.BorderForeground(styles.ColorWarning)
	case CardError:
		cardStyle = cardStyle.BorderForeground(styles.ColorError)
	case CardInfo:
		cardStyle = cardStyle.BorderForeground(styles.ColorSecondary)
	}

	// Apply dimensions if set
	if c.Width > 0 {
		cardStyle = cardStyle.Width(c.Width - (c.Padding * 2) - 2) // Account for padding and border
	}
	if c.Height > 0 {
		cardStyle = cardStyle.Height(c.Height - (c.Padding * 2) - 2) // Account for padding and border
	}

	// Apply padding
	cardStyle = cardStyle.Padding(c.Padding)

	// Build content with title
	content := ""
	if c.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorText)
		if c.Active {
			titleStyle = titleStyle.Foreground(styles.ColorPrimary)
		}
		content = titleStyle.Render(c.Title) + "\n\n" + c.Content
	} else {
		content = c.Content
	}

	return cardStyle.Render(content)
}

// StatusCard creates a status card with an icon
type StatusCard struct {
	*Card
	Icon   string
	Status string
}

// NewStatusCard creates a new status card
func NewStatusCard(title, icon, status, details string) *StatusCard {
	card := NewCard(title, details)
	return &StatusCard{
		Card:   card,
		Icon:   icon,
		Status: status,
	}
}

// Render renders the status card
func (sc *StatusCard) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	// Build content with icon and status
	textStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorText)
	statusLine := sc.Icon + " " + textStyle.Render(sc.Status)
	content := statusLine

	if sc.Content != "" {
		content += "\n\n" + sc.Content
	}

	// Update card content and render
	sc.Content = content
	return sc.Card.Render()
}

// InfoCard creates an info card with details
type InfoCard struct {
	*Card
	Items []InfoItem
}

// InfoItem represents a key-value pair in an info card
type InfoItem struct {
	Label string
	Value string
	Icon  string
}

// NewInfoCard creates a new info card
func NewInfoCard(title string, items []InfoItem) *InfoCard {
	card := NewCard(title, "")
	return &InfoCard{
		Card:  card,
		Items: items,
	}
}

// Render renders the info card
func (ic *InfoCard) Render() string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	var content string
	for i, item := range ic.Items {
		if i > 0 {
			content += "\n"
		}

		line := ""
		if item.Icon != "" {
			line += item.Icon + " "
		}

		mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
		textStyle := lipgloss.NewStyle().Foreground(styles.ColorText)
		label := mutedStyle.Render(item.Label + ":")
		value := textStyle.Render(item.Value)

		line += label + " " + value
		content += line
	}

	ic.Content = content
	return ic.Card.Render()
}
