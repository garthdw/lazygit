package git_commands

import (
	"github.com/jesseduffield/lazygit/pkg/config"
)

// DiffMode says what a diff command's output is for. This decides whether the
// configured diff renderer produces it, and whether it is coloured.
type DiffMode int

const (
	// DiffModeRendered is the diff as the user has arranged for it to look: through the
	// diff renderer, with the renderer's own arguments and its preference about colour.
	DiffModeRendered DiffMode = iota
	// DiffModeRaw is git's own coloured diff, for showing a diff whose rendered form
	// couldn't be acted on.
	DiffModeRaw
	// DiffModeFocused is the diff as git.focusedDiffRenderer renders it, for a focused
	// main view whose diff, as the user's current renderer would produce it, couldn't be
	// acted on.
	DiffModeFocused
	// DiffModePlain is git's own uncoloured diff, for building patches from and copying
	// text out of rather than for looking at.
	DiffModePlain
)

// DiffRenderer returns the diff renderer that produces a diff in this mode, of those the
// given manager knows about. Only DiffModeFocused has one of its own; for any other mode
// it is the given manager's current one, which the modes that don't render ignore. A
// config reload can take the focused renderer away after the mode was chosen, and the
// current one stands in for it then.
func (self DiffMode) DiffRenderer(diffRendererConfigManager *config.DiffRendererConfigManager) *config.DiffRendererConfigManager {
	if self == DiffModeFocused {
		if focused, ok := diffRendererConfigManager.FocusedDiffRenderer(); ok {
			return focused
		}
	}
	return diffRendererConfigManager
}

// colorArg returns the value to pass to git's --color for this mode. Rendered output is
// coloured however the renderer wants its input; a raw diff gets git's own colour, which
// is the point of it; a plain one is for reading as text, not for looking at.
func (self DiffMode) colorArg(diffRendererConfigManager *config.DiffRendererConfigManager) string {
	switch self {
	case DiffModeRendered, DiffModeFocused:
		return self.DiffRenderer(diffRendererConfigManager).GetColorArg()
	case DiffModeRaw:
		return "always"
	default:
		return "never"
	}
}
