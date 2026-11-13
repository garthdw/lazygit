package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var RawFallbackWhenTheFocusedDiffRendererIsUnsupported = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Focusing the main view under a diff renderer that says nothing about its rows brings git's own diff when the focused renderer says nothing about them either",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		// The focused renderer numbers the lines of the diff, so its rows no longer read
		// as a diff, and it doesn't announce the metadata protocol.
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Name: "opaque", Type: "extDiff", Command: `sh -c 'seq -f "EXT-%g" 40'`},
			{Name: "numbered", Command: "cat -n"},
		}
		cfg.GetUserConfig().Git.FocusedDiffRenderer = "numbered"
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\ntwo\nTHREE\nfour\nfive\nsix\nseven\neight\nNINE\nten\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Main().Content(Contains("EXT-1"))

		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Content(Contains("+THREE")).
			Content(DoesNotContain("EXT-1")).
			Content(DoesNotContain("1  diff --git")).
			SelectedLines(
				Contains("-three"),
			).
			Press(keys.Main.ToggleSelectHunk).
			PressPrimaryAction()

		t.Views().Secondary().ContainsLines(
			Contains("-three"),
			Contains("+THREE"),
		)
	},
})
