package components

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/yourusername/gitman/internal/ui"
)

// LogoVariant defines the type of logo to display
type LogoVariant string

const (
	LogoDashboard LogoVariant = "dashboard"
	LogoCommit    LogoVariant = "commit"
	LogoMergePR   LogoVariant = "merge-pr"
	LogoSettings  LogoVariant = "settings"
	LogoOnboard   LogoVariant = "onboard"
)

// RenderLogo renders the appropriate logo based on the variant
func RenderLogo(variant LogoVariant, repoInfo string) string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	var ascii string
	switch variant {
	case LogoDashboard:
		ascii = `
   ██████╗ ███╗   ███╗
  ██╔════╝ ████╗ ████║
  ██║  ███╗██╔████╔██║
  ██║   ██║██║╚██╔╝██║
  ╚██████╔╝██║ ╚═╝ ██║
   ╚═════╝ ╚═╝     ╚═╝`

	case LogoCommit:
		ascii = `
   ██████╗ ██████╗ ███╗   ███╗███╗   ███╗██╗████████╗
  ██╔════╝██╔═══██╗████╗ ████║████╗ ████║██║╚══██╔══╝
  ██║     ██║   ██║██╔████╔██║██╔████╔██║██║   ██║
  ██║     ██║   ██║██║╚██╔╝██║██║╚██╔╝██║██║   ██║
  ╚██████╗╚██████╔╝██║ ╚═╝ ██║██║ ╚═╝ ██║██║   ██║
   ╚═════╝ ╚═════╝ ╚═╝     ╚═╝╚═╝     ╚═╝╚═╝   ╚═╝`

	case LogoMergePR:
		ascii = `
  ███╗   ███╗███████╗██████╗  ██████╗ ███████╗   ██████╗ ██████╗
  ████╗ ████║██╔════╝██╔══██╗██╔════╝ ██╔════╝   ██╔══██╗██╔══██╗
  ██╔████╔██║█████╗  ██████╔╝██║  ███╗█████╗     ██████╔╝██████╔╝
  ██║╚██╔╝██║██╔══╝  ██╔══██╗██║   ██║██╔══╝     ██╔═══╝ ██╔══██╗
  ██║ ╚═╝ ██║███████╗██║  ██║╚██████╔╝███████╗   ██║     ██║  ██║
  ╚═╝     ╚═╝╚══════╝╚═╝  ╚═╝ ╚═════╝ ╚══════╝   ╚═╝     ╚═╝  ╚═╝`

	case LogoSettings:
		ascii = `
   ██████╗ ███╗   ███╗
  ██╔════╝ ████╗ ████║
  ██║  ███╗██╔████╔██║
  ██║   ██║██║╚██╔╝██║
  ╚██████╔╝██║ ╚═╝ ██║
   ╚═════╝ ╚═╝     ╚═╝
       SETTINGS`

	case LogoOnboard:
		ascii = `
   ██████╗ ███╗   ███╗
  ██╔════╝ ████╗ ████║
  ██║  ███╗██╔████╔██║
  ██║   ██║██║╚██╔╝██║
  ╚██████╔╝██║ ╚═╝ ██║
   ╚═════╝ ╚═╝     ╚═╝
      ONBOARDING`

	default:
		ascii = `
   ██████╗ ███╗   ███╗
  ██╔════╝ ████╗ ████║
  ██║  ███╗██╔████╔██║
  ██║   ██║██║╚██╔╝██║
  ╚██████╔╝██║ ╚═╝ ██║
   ╚═════╝ ╚═╝     ╚═╝`
	}

	primaryStyle := lipgloss.NewStyle().Foreground(styles.ColorPrimary)
	logo := primaryStyle.Render(ascii)

	if repoInfo != "" {
		mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted).Italic(true)
		return logo + "\n" + mutedStyle.Render(repoInfo)
	}

	return logo
}

// RenderLogoCompact renders a compact version for limited space
func RenderLogoCompact(variant LogoVariant) string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	var text string
	switch variant {
	case LogoCommit:
		text = "[ COMMIT ]"
	case LogoMergePR:
		text = "[ MERGE/PR ]"
	case LogoSettings:
		text = "[ SETTINGS ]"
	case LogoOnboard:
		text = "[ ONBOARDING ]"
	default:
		text = "[ GM ]"
	}

	primaryStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary)
	return primaryStyle.Render(text)
}

// RenderBranding renders consistent branding text
func RenderBranding() string {
	styles := ui.GetGlobalThemeManager().GetStyles()
	primaryStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary)
	mutedStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
	title := primaryStyle.Render("GitMind")
	subtitle := mutedStyle.Render("AI-Powered Git Manager")
	return title + " - " + subtitle
}

// RenderHeader renders a consistent header with title and optional subtitle
func RenderHeader(title, subtitle string) string {
	styles := ui.GetGlobalThemeManager().GetStyles()

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(styles.ColorPrimary)
	header := titleStyle.Render(title)

	if subtitle != "" {
		subtitleStyle := lipgloss.NewStyle().Foreground(styles.ColorMuted)
		header += "\n" + subtitleStyle.Render(subtitle)
	}

	return header
}

// RenderDivider renders a horizontal divider
func RenderDivider(width int) string {
	styles := ui.GetGlobalThemeManager().GetStyles()
	dividerStyle := lipgloss.NewStyle().
		Foreground(styles.ColorBorder).
		Width(width)

	divider := ""
	for i := 0; i < width; i++ {
		divider += "─"
	}

	return dividerStyle.Render(divider)
}
