package ui

// LayoutMode represents the layout display mode based on terminal size.
type LayoutMode int

const (
	// LayoutSinglePane shows only the left menu pane (fallback for small terminals).
	LayoutSinglePane LayoutMode = iota
	// LayoutDualPaneCompact shows both panes with minimal left panel width (80-119 chars).
	LayoutDualPaneCompact
	// LayoutDualPaneFull shows both panes with comfortable widths (120+ chars).
	LayoutDualPaneFull
)

// LayoutConfig contains calculated dimensions for the UI layout.
type LayoutConfig struct {
	Mode            LayoutMode // Current layout mode
	TermWidth       int        // Terminal width
	TermHeight      int        // Terminal height
	LeftPaneWidth   int        // Width of left pane (menu)
	RightPaneWidth  int        // Width of right pane (git tree)
	SeparatorWidth  int        // Width of separator (1 char or 0 if hidden)
	ContentHeight   int        // Available height for content (excluding header/footer)
	ShowSeparator   bool       // Whether to show the vertical separator
	ShowGitTree     bool       // Whether to show the git tree pane
}

// Responsive breakpoints (in characters)
const (
	MinWidthDualPane     = 80  // Minimum width to show dual-pane layout
	MinWidthComfortable  = 120 // Width for comfortable dual-pane with full features
	MinLeftPaneWidth     = 25  // Minimum left pane width (for menu items)
	MaxLeftPaneWidth     = 40  // Maximum left pane width
	ReservedHeaderHeight = 3   // Header lines (title + separator)
	ReservedFooterHeight = 2   // Footer lines (help text)
)

// CalculateLayout determines the optimal layout configuration based on terminal size.
func CalculateLayout(termWidth, termHeight int) LayoutConfig {
	config := LayoutConfig{
		TermWidth:      termWidth,
		TermHeight:     termHeight,
		SeparatorWidth: 1,
		ShowSeparator:  true,
	}

	// Calculate content height (total - header - footer)
	config.ContentHeight = termHeight - ReservedHeaderHeight - ReservedFooterHeight
	if config.ContentHeight < 10 {
		config.ContentHeight = 10 // Minimum content height
	}

	// Determine layout mode based on terminal width
	if termWidth < MinWidthDualPane {
		// Single pane fallback for very small terminals
		config.Mode = LayoutSinglePane
		config.LeftPaneWidth = termWidth
		config.RightPaneWidth = 0
		config.SeparatorWidth = 0
		config.ShowSeparator = false
		config.ShowGitTree = false
	} else if termWidth < MinWidthComfortable {
		// Compact dual-pane mode (80-119 chars)
		config.Mode = LayoutDualPaneCompact
		config.LeftPaneWidth = calculateLeftPaneWidth(termWidth, 0.35) // 35% for left pane
		config.RightPaneWidth = termWidth - config.LeftPaneWidth - config.SeparatorWidth
		config.ShowGitTree = true
	} else {
		// Full dual-pane mode (120+ chars)
		config.Mode = LayoutDualPaneFull
		config.LeftPaneWidth = calculateLeftPaneWidth(termWidth, 0.30) // 30% for left pane
		config.RightPaneWidth = termWidth - config.LeftPaneWidth - config.SeparatorWidth
		config.ShowGitTree = true
	}

	return config
}

// calculateLeftPaneWidth calculates the left pane width based on percentage, with min/max constraints.
func calculateLeftPaneWidth(termWidth int, percentage float64) int {
	width := int(float64(termWidth) * percentage)

	// Apply min/max constraints
	if width < MinLeftPaneWidth {
		width = MinLeftPaneWidth
	}
	if width > MaxLeftPaneWidth {
		width = MaxLeftPaneWidth
	}

	return width
}

// String returns a human-readable description of the layout mode.
func (lm LayoutMode) String() string {
	switch lm {
	case LayoutSinglePane:
		return "single-pane"
	case LayoutDualPaneCompact:
		return "dual-pane-compact"
	case LayoutDualPaneFull:
		return "dual-pane-full"
	default:
		return "unknown"
	}
}

// IsShowingGitTree returns true if the git tree pane should be visible.
func (config LayoutConfig) IsShowingGitTree() bool {
	return config.ShowGitTree && config.RightPaneWidth > 0
}

// GetAvailableTreeHeight returns the height available for the git tree viewport.
func (config LayoutConfig) GetAvailableTreeHeight() int {
	// Reserve space for branch structure overlay (8 lines) and commit details (10 lines)
	const (
		branchOverlayHeight = 8
		commitDetailsHeight = 10
		padding             = 2
	)

	availableHeight := config.ContentHeight - branchOverlayHeight - commitDetailsHeight - padding
	if availableHeight < 5 {
		availableHeight = 5 // Minimum viewport height
	}

	return availableHeight
}
