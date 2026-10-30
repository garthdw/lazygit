package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageWithTheFocusedDiffRenderer = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Focusing the main view under a diff renderer that says nothing about its rows brings the diff of the configured focused renderer, until the focus leaves it",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		// Browsing is with an external diff whose output has nothing to say about which
		// line of which file each row shows; staging is with a renderer that announces
		// the metadata protocol and passes the diff through.
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Name: "opaque", Type: "extDiff", Command: `sh -c 'seq -f "EXT-%g" 40'`},
			{Name: "conforming", Command: `printf '\033]1717;1\007RENDERED BY ME\n'; cat`},
		}
		cfg.GetUserConfig().Git.FocusedDiffRenderer = "conforming"
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\ntwo\nTHREE\nfour\nfive\nsix\nseven\neight\nNINE\nten\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().IsFocused()

		t.Views().Main().Content(Contains("EXT-1"))

		t.Views().Files().Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Content(Contains("RENDERED BY ME")).
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

		// The re-render after staging stays with the focused renderer, and the next hunk
		// can be selected in it.
		t.Views().Main().
			IsFocused().
			Content(Contains("RENDERED BY ME")).
			Content(DoesNotContain("EXT-1")).
			SelectedLines(
				Contains("-nine"),
				Contains("+NINE"),
			)

		// The first escape gives up the hunk selection, the second leaves.
		t.Views().Main().
			Press(keys.Universal.Return).
			Press(keys.Universal.Return)

		// Back in the files panel, the diff is the browsing renderer's again.
		t.Views().Files().IsFocused()
		t.Views().Main().Content(Contains("EXT-1"))
	},
})
