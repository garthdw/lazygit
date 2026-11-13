package helpers

import (
	"fmt"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/commands/git_commands"
	"github.com/jesseduffield/lazygit/pkg/config"
	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/jesseduffield/lazygit/pkg/tasks"
)

// Falling back to another diff when the configured one can't be acted on.
//
// A diff renderer is free to lay a diff out however it likes, and once it has, we can
// only tell which line of which file a row shows if the renderer says so. A renderer
// that doesn't produces a diff that can be read but not staged, edited or copied from,
// so when the user focuses the main view to act on it, we show another diff instead:
// the one git.focusedDiffRenderer renders, if that renderer does say so, and git's own
// otherwise. Browsing keeps the renderer's version; only acting on it needs one we can
// follow.

// SwitchToFocusedDiffRenderer makes git.focusedDiffRenderer the current renderer, as the
// main view takes focus, when the renderer being browsed with can't be acted on and that
// one can. It is only the renderer the focused main view starts with: the user can cycle
// on from it, and leaving the main view goes back to the one they were browsing with
// unless they did. It reports whether it switched.
func (self *DiffLineHelper) SwitchToFocusedDiffRenderer() bool {
	manager := self.c.State().GetDiffRendererConfigManager()
	if !self.diffNeedsMetadata() || self.diffRendererEmitsMetadata(manager) {
		return false
	}
	focused, ok := manager.FocusedDiffRenderer()
	if !ok || !self.diffRendererEmitsMetadata(focused) {
		return false
	}
	manager.SwitchToFocusedDiffRenderer(focused)
	return true
}

// MainViewDiffMode says how a side panel should produce the diff it renders into the
// main view: as the user configured it, or as git's own — while the main view holds
// focus and what the renderer would produce couldn't be acted on.
//
// Every panel that renders a diff into the main view asks, so that a re-render while
// focused — after staging a hunk, say — stays with git's own diff rather than flipping
// back to the renderer's.
func (self *DiffLineHelper) MainViewDiffMode() git_commands.DiffMode {
	if self.mainViewIsFocused() && self.diffNeedsMetadata() && !self.diffRendererEmitsMetadata(self.c.State().GetDiffRendererConfigManager()) {
		return git_commands.DiffModeRaw
	}
	return git_commands.DiffModeRendered
}

// DiffRowsCanBePlaced reports whether the rows of the diff the main view is about to be
// given can be placed in the file they show. git's own diff describes itself, whether it
// is what the user configured or what MainViewDiffMode is about to substitute for a
// rendering that can't be acted on; any other rendering says where its rows belong only
// if it states records for them.
//
// It is what anything that means to go from a row back to the file it shows has to ask
// first: with neither records nor a diff that describes itself, there is nothing to go
// on, and offering the user the way there would be offering nothing.
func (self *DiffLineHelper) DiffRowsCanBePlaced() bool {
	return !self.diffNeedsMetadata() ||
		self.MainViewDiffMode() == git_commands.DiffModeRaw ||
		self.diffRendererEmitsMetadata(self.c.State().GetDiffRendererConfigManager())
}

// RenderFocusedMainViewAgain has the panel beneath the focused main view render its
// diff again — which, the main view now holding focus, is the focused renderer's or
// git's own rather than the browsing renderer's — and calls place once that is on
// screen.
//
// The whole diff is read before it is shown, rather than the first screenful: place
// looks at what is there to decide where to put the selection, and a change line
// further down would otherwise be missed.
//
// It is the same diff of the same files, so the view keeps the scroll position it has
// rather than starting from the top: the new rendering lays the changes out its own way
// and the line the user was on is somewhere else now, but the offset still puts them
// among the same part of the file — and the selection is then established from what
// that leaves on screen.
func (self *DiffLineHelper) RenderFocusedMainViewAgain(view *gocui.View, sidePanel types.Context, place func()) {
	manager := self.c.GetOrCreateViewBufferManagerForView(view)
	if manager == nil {
		return
	}

	manager.SetKeepScrollPositionForNextTask()
	manager.SetRestoreForNextTask(&tasks.RenderRestore{
		FirstPaintReady: func() bool { return false },
		Apply: func(swapIn func()) {
			swapIn()
			place()
		},
	})

	sidePanel.HandleRenderToMain()
}

func (self *DiffLineHelper) mainViewIsFocused() bool {
	current := self.c.Context().CurrentStatic().GetKey()
	return current == self.c.Contexts().Normal.GetKey() ||
		current == self.c.Contexts().NormalSecondary.GetKey()
}

// diffNeedsMetadata reports whether the diff we would show is one whose rows can only
// be placed in the file by the records the renderer states. Any custom renderer may
// restructure the diff; so may git itself, once the renderer's arguments ask for a word
// diff, whose markup is inline. Plain git output describes itself, and needs no records.
func (self *DiffLineHelper) diffNeedsMetadata() bool {
	manager := self.c.State().GetDiffRendererConfigManager()
	if manager.GetDiffRendererType() != config.DiffRendererType_RawGit {
		return true
	}
	return len(manager.GetRawGitArgs()) > 0
}

// diffRendererEmitsMetadata is the probed verdict about the given diff renderer, asked
// once and remembered for as long as the renderer stays the same — until the user cycles
// to another one, or a changed config is reloaded.
func (self *DiffLineHelper) diffRendererEmitsMetadata(manager *config.DiffRendererConfigManager) bool {
	signature := diffRendererSignature(manager)
	verdict, ok := self.rendererEmitsMetadata[signature]
	if !ok {
		verdict = self.c.Git().Diff.ProbeDiffRendererEmitsMetadata(manager)
		self.rendererEmitsMetadata[signature] = verdict
	}
	return verdict
}

// diffRendererSignature identifies a diff renderer, so that a remembered verdict is only
// used for the renderer it describes. The values a command is resolved with are no part
// of its identity, so fixed ones are used.
func diffRendererSignature(manager *config.DiffRendererConfigManager) string {
	index, _ := manager.CurrentDiffRendererIndex()
	values := config.DiffRendererValues{}
	externalDiff, _ := manager.GetExternalDiffCommand(values)
	stdinFilter, _ := manager.GetStdinFilterCommand(values)
	return fmt.Sprintf("%d\x00%s\x00%s\x00%s",
		index,
		externalDiff,
		stdinFilter,
		strings.Join(manager.GetRawGitArgs(), "\x00"))
}
