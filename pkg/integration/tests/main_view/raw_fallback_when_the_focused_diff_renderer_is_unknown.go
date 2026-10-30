package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var RawFallbackWhenTheFocusedDiffRendererIsUnknown = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Focusing the main view under a diff renderer that says nothing about its rows brings git's own diff when the focused renderer names no configured one",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Name: "opaque", Type: "extDiff", Command: `sh -c 'seq -f "EXT-%g" 40'`},
		}
		cfg.GetUserConfig().Git.FocusedDiffRenderer = "nope"
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
