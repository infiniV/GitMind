package ui

import "github.com/yourusername/gitman/internal/domain"

// Available theme presets for the TUI.
var (
	// ThemeClaudeWarm is the default theme with warm orange-rust tones.
	ThemeClaudeWarm = domain.Theme{
		Name:        "claude-warm",
		Description: "Professional warm theme with orange-rust accents (default)",
		Colors: domain.ThemeColors{
			Primary:          "#C15F3C",
			Secondary:        "#A14A2F",
			Success:          "#7A9A6E",
			Warning:          "#D4945A",
			Error:            "#C16B6B",
			Muted:            "#B1ADA1",
			Border:           "#3A3631",
			Selected:         "#C15F3C",
			Text:             "#E8E6E3",
			HighConfidence:   "#7A9A6E",
			MediumConfidence: "#D4945A",
			LowConfidence:    "#C16B6B",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F3A2C",
			BadgeMedium:  "#3A2F1F",
			BadgeLow:     "#3A1F1F",
			FormInput:    "#2F2A1F",
			FormFocused:  "#3A2F1F",
			Modal:        "#1F2937",
			Submenu:      "#1A1A1A",
			Dashboard:    "#1A1A1A",
			Confirmation: "#1A1A1A",
			ErrorModal:   "#1A1A1A",
		},
	}

	// ThemeOceanBlue is a calm blue theme for focused coding.
	ThemeOceanBlue = domain.Theme{
		Name:        "ocean-blue",
		Description: "Cool blue theme for focus and reduced eye strain",
		Colors: domain.ThemeColors{
			Primary:          "#4A90E2",
			Secondary:        "#357ABD",
			Success:          "#6EA06E",
			Warning:          "#E2A04A",
			Error:            "#E24A4A",
			Muted:            "#A1B1C1",
			Border:           "#2A3641",
			Selected:         "#4A90E2",
			Text:             "#E3E8ED",
			HighConfidence:   "#6EA06E",
			MediumConfidence: "#E2A04A",
			LowConfidence:    "#E24A4A",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F3A33",
			BadgeMedium:  "#3A321F",
			BadgeLow:     "#3A1F1F",
			FormInput:    "#1F2A37",
			FormFocused:  "#2A3641",
			Modal:        "#1A2532",
			Submenu:      "#151F2A",
			Dashboard:    "#151F2A",
			Confirmation: "#1A2532",
			ErrorModal:   "#1A2532",
		},
	}

	// ThemeForestGreen is a natural green theme for balanced work.
	ThemeForestGreen = domain.Theme{
		Name:        "forest-green",
		Description: "Natural green theme for balanced, calming coding",
		Colors: domain.ThemeColors{
			Primary:          "#6B9A6B",
			Secondary:        "#557A55",
			Success:          "#7AAA7A",
			Warning:          "#D4A45A",
			Error:            "#C17B6B",
			Muted:            "#A1B1A1",
			Border:           "#2A3A2A",
			Selected:         "#6B9A6B",
			Text:             "#E3EDE3",
			HighConfidence:   "#7AAA7A",
			MediumConfidence: "#D4A45A",
			LowConfidence:    "#C17B6B",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F3A2C",
			BadgeMedium:  "#3A341F",
			BadgeLow:     "#3A1F1F",
			FormInput:    "#1F2A1F",
			FormFocused:  "#2A3A2A",
			Modal:        "#1A251A",
			Submenu:      "#15201A",
			Dashboard:    "#15201A",
			Confirmation: "#1A251A",
			ErrorModal:   "#1A251A",
		},
	}

	// ThemeMonochrome is a minimalist grayscale theme.
	ThemeMonochrome = domain.Theme{
		Name:        "monochrome",
		Description: "Minimalist grayscale theme for distraction-free coding",
		Colors: domain.ThemeColors{
			Primary:          "#888888",
			Secondary:        "#666666",
			Success:          "#999999",
			Warning:          "#AAAAAA",
			Error:            "#777777",
			Muted:            "#666666",
			Border:           "#333333",
			Selected:         "#888888",
			Text:             "#EEEEEE",
			HighConfidence:   "#999999",
			MediumConfidence: "#AAAAAA",
			LowConfidence:    "#777777",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#2A2A2A",
			BadgeMedium:  "#323232",
			BadgeLow:     "#282828",
			FormInput:    "#252525",
			FormFocused:  "#2A2A2A",
			Modal:        "#1F1F1F",
			Submenu:      "#1A1A1A",
			Dashboard:    "#1A1A1A",
			Confirmation: "#1F1F1F",
			ErrorModal:   "#1F1F1F",
		},
	}

	// ThemeMagma is a scientific colormap with purple-orange tones.
	ThemeMagma = domain.Theme{
		Name:        "magma",
		Description: "Scientific colormap with vibrant purple-orange gradient",
		Colors: domain.ThemeColors{
			Primary:          "#D8456C",
			Secondary:        "#B12A5B",
			Success:          "#7AAA7A",
			Warning:          "#FCAD52",
			Error:            "#FB8861",
			Muted:            "#B39DB5",
			Border:           "#3B1F40",
			Selected:         "#D8456C",
			Text:             "#F0E4F1",
			HighConfidence:   "#7AAA7A",
			MediumConfidence: "#FCAD52",
			LowConfidence:    "#FB8861",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F3A2C",
			BadgeMedium:  "#3A2A1F",
			BadgeLow:     "#3A1F2A",
			FormInput:    "#2B1A33",
			FormFocused:  "#3B1F40",
			Modal:        "#1F1A2A",
			Submenu:      "#1A1525",
			Dashboard:    "#1A1525",
			Confirmation: "#1F1A2A",
			ErrorModal:   "#1F1A2A",
		},
	}

	// ThemeViridis is a scientific colormap with blue-green-yellow tones.
	ThemeViridis = domain.Theme{
		Name:        "viridis",
		Description: "Scientific colormap with perceptually uniform blue-green gradient",
		Colors: domain.ThemeColors{
			Primary:          "#4AC29A",
			Secondary:        "#2FA785",
			Success:          "#68D391",
			Warning:          "#FDE047",
			Error:            "#FB7185",
			Muted:            "#A1C9BA",
			Border:           "#1F3A38",
			Selected:         "#4AC29A",
			Text:             "#E8F4F1",
			HighConfidence:   "#68D391",
			MediumConfidence: "#FDE047",
			LowConfidence:    "#FB7185",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F3A33",
			BadgeMedium:  "#3A381F",
			BadgeLow:     "#3A1F2A",
			FormInput:    "#1A2F2D",
			FormFocused:  "#1F3A38",
			Modal:        "#1A2528",
			Submenu:      "#15201E",
			Dashboard:    "#15201E",
			Confirmation: "#1A2528",
			ErrorModal:   "#1A2528",
		},
	}

	// ThemePlasma is a scientific colormap with pink-purple-yellow tones.
	ThemePlasma = domain.Theme{
		Name:        "plasma",
		Description: "Scientific colormap with vibrant pink-purple-yellow gradient",
		Colors: domain.ThemeColors{
			Primary:          "#CC4778",
			Secondary:        "#A63367",
			Success:          "#7AAA7A",
			Warning:          "#F89441",
			Error:            "#EB5E8D",
			Muted:            "#C5A3BA",
			Border:           "#3A1F35",
			Selected:         "#CC4778",
			Text:             "#F5E8F0",
			HighConfidence:   "#7AAA7A",
			MediumConfidence: "#F89441",
			LowConfidence:    "#EB5E8D",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F3A2C",
			BadgeMedium:  "#3A2A1F",
			BadgeLow:     "#3A1F2F",
			FormInput:    "#2A1A30",
			FormFocused:  "#3A1F35",
			Modal:        "#1F1A28",
			Submenu:      "#1A1523",
			Dashboard:    "#1A1523",
			Confirmation: "#1F1A28",
			ErrorModal:   "#1F1A28",
		},
	}

	// ThemeTwilight is a purple-blue theme for evening coding.
	ThemeTwilight = domain.Theme{
		Name:        "twilight",
		Description: "Purple-blue theme optimized for evening coding sessions",
		Colors: domain.ThemeColors{
			Primary:          "#8B7EC8",
			Secondary:        "#6B5FA8",
			Success:          "#7AAA88",
			Warning:          "#D9A85A",
			Error:            "#C17B8B",
			Muted:            "#ADA1C1",
			Border:           "#2A2541",
			Selected:         "#8B7EC8",
			Text:             "#EDE8F5",
			HighConfidence:   "#7AAA88",
			MediumConfidence: "#D9A85A",
			LowConfidence:    "#C17B8B",
		},
		Backgrounds: domain.ThemeBackgrounds{
			BadgeHigh:    "#1F2A3A",
			BadgeMedium:  "#3A2F3A",
			BadgeLow:     "#3A1F2F",
			FormInput:    "#1F1A2F",
			FormFocused:  "#2A2541",
			Modal:        "#1A1628",
			Submenu:      "#15111E",
			Dashboard:    "#15111E",
			Confirmation: "#1A1628",
			ErrorModal:   "#1A1628",
		},
	}
)

// AllThemes returns a slice of all available themes.
func AllThemes() []domain.Theme {
	return []domain.Theme{
		ThemeClaudeWarm,
		ThemeOceanBlue,
		ThemeForestGreen,
		ThemeMonochrome,
		ThemeMagma,
		ThemeViridis,
		ThemePlasma,
		ThemeTwilight,
	}
}

// GetThemeByName returns a theme by its name, or the default theme if not found.
func GetThemeByName(name string) domain.Theme {
	for _, theme := range AllThemes() {
		if theme.Name == name {
			return theme
		}
	}
	// Return default theme if not found
	return ThemeClaudeWarm
}

// GetThemeNames returns a slice of all theme names.
func GetThemeNames() []string {
	themes := AllThemes()
	names := make([]string, len(themes))
	for i, theme := range themes {
		names[i] = theme.Name
	}
	return names
}
