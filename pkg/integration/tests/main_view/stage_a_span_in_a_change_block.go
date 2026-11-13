package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageASpanInAChangeBlock = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A span standing for one deletion and addition in the middle of a change block is staged in place, as a unit",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		// A renderer that shows the second pair of a block of three deletions and
		// four additions as a row of its own.
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Command: `cat >/dev/null; printf '` +
				`\033]1717;1\007` +
				`\033]1717;1;f;;;file1\007FILE file1\n` +
				`\033]1717;1;e;3-3;3-3;file1\007PAIR line 3\n` +
				`\033]1717;1;d;3;3;file1\007-three\n` +
				`\033]1717;1;a;3;;file1\007+THREE\n'`},
		}
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\nsix\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\nTWO\nTHREE\nFOUR\nFOUR AND A HALF\nfive\nsix\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(
				Contains("-three"),
			).
			SelectPreviousItem().
			SelectedLines(
				Contains("PAIR line 3"),
			).
			PressPrimaryAction()

		// The unselected deletions stay, ahead of and after the staged pair.
		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour\nfive\nsix\n")
	},
})
