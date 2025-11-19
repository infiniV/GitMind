package layout

// Spacing constants for consistent padding and margins
const (
	SpacingXS = 1
	SpacingSM = 2
	SpacingMD = 3
	SpacingLG = 4
	SpacingXL = 6
)

// Modal width constants
const (
	ModalWidthSM = 50
	ModalWidthMD = 60
	ModalWidthLG = 70
	ModalWidthXL = 80
)

// Modal height constants
const (
	ModalHeightSM = 10
	ModalHeightMD = 15
	ModalHeightLG = 20
)

// Layout ratios for split views
const (
	SplitRatio35_65 = 0.35
	SplitRatio40_60 = 0.40
	SplitRatio50_50 = 0.50
)

// Standard UI element heights
const (
	HeaderHeight       = 8
	FooterHeight       = 2
	TabBarHeight       = 3
	StatusBarHeight    = 1
	LogoHeight         = 6
	ButtonHeight       = 3
	InputHeight        = 3
)

// Card dimensions
const (
	DashboardCardMinWidth  = 30
	DashboardCardMinHeight = 8
)

// Helper functions

// CalculateContentHeight calculates available height for content after headers/footers
func CalculateContentHeight(windowHeight int) int {
	return windowHeight - HeaderHeight - FooterHeight
}

// CalculateContentHeightWithTabs calculates height with tab bar
func CalculateContentHeightWithTabs(windowHeight int) int {
	return windowHeight - HeaderHeight - FooterHeight - TabBarHeight
}

// CalculateSplitWidths calculates left and right widths for split view
func CalculateSplitWidths(totalWidth int, ratio float64) (int, int) {
	leftWidth := int(float64(totalWidth) * ratio)
	rightWidth := totalWidth - leftWidth - SpacingMD
	return leftWidth, rightWidth
}

// CalculateViewportHeight calculates viewport height for scrollable content
// Subtracts logo, header, and footer space
func CalculateViewportHeight(windowHeight int) int {
	return windowHeight - LogoHeight - HeaderHeight - FooterHeight - SpacingMD
}

// CenterHorizontal calculates x position to center content
func CenterHorizontal(windowWidth, contentWidth int) int {
	if windowWidth <= contentWidth {
		return 0
	}
	return (windowWidth - contentWidth) / 2
}

// CenterVertical calculates y position to center content
func CenterVertical(windowHeight, contentHeight int) int {
	if windowHeight <= contentHeight {
		return 0
	}
	return (windowHeight - contentHeight) / 2
}
