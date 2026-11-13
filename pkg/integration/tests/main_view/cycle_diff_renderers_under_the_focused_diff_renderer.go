package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CycleDiffRenderersUnderTheFocusedDiffRenderer = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "The focused diff renderer is only where cycling starts from in the focused main view, and leaving the main view keeps the renderer cycled to",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Name: "opaque", Type: "extDiff", Command: `sh -c 'seq -f "EXT-%g" 40'`},
			{Name: "conforming", Command: `printf '\033]1717;1\007RENDERED BY ME\n'; cat`},
			{Name: "git", Type: "rawGit"},
		}
		cfg.GetUserConfig().Git.FocusedDiffRenderer = "conforming"
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\ntwo\nTHREE\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().IsFocused()
		t.Views().Main().Content(Contains("EXT-1"))

		t.Views().Files().Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Content(Contains("RENDERED BY ME")).
			Press(keys.Universal.CycleDiffRenderers)
		t.ExpectToast(Equals("Diff renderer: git (3 of 3)"))
		t.Views().Main().
			Content(DoesNotContain("RENDERED BY ME")).
			Content(Contains("+THREE"))

		// Cycling on to a renderer that can't be acted on shows git's own diff rather
		// than going back to the focused renderer.
		t.Views().Main().
			Press(keys.Universal.CycleDiffRenderers)
		t.ExpectToast(Equals("Diff renderer: opaque (1 of 3)"))
		t.Views().Main().
			Content(DoesNotContain("RENDERED BY ME")).
			Content(DoesNotContain("EXT-1")).
			Content(Contains("+THREE")).
			Press(keys.Universal.CycleDiffRenderersReverse)
		t.ExpectToast(Equals("Diff renderer: git (3 of 3)"))
		t.Views().Main().Press(keys.Universal.Return)

		// The user chose git while focused, so it stays.
		t.Views().Files().IsFocused()
		t.Views().Main().
			Content(DoesNotContain("EXT-1")).
			Content(DoesNotContain("RENDERED BY ME")).
			Content(Contains("+THREE"))
	},
})
