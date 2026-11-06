package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageASpanOfLines = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A row the diff renderer states as a span stands for every change line within it, the ones it doesn't show included",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		// A renderer in the manner of an entity-level diff: a row for the entity
		// spanning lines 3-4, then only the change it thinks worth showing. The
		// whitespace change on line 4 has no row of its own.
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Command: `cat >/dev/null; printf '` +
				`\033]1717;1\007` +
				`\033]1717;1;f;;;file1\007FILE file1\n` +
				`\033]1717;1;e;3-4;3-4;file1\007ENTITY lines 3-4\n` +
				`\033]1717;1;d;3;3;file1\007-three\n` +
				`\033]1717;1;a;3;;file1\007+THREE\n` +
				`\033]1717;1;e;9-9;9-9;file1\007ENTITY line 9\n` +
				`\033]1717;1;d;9;9;file1\007-nine\n` +
				`\033]1717;1;a;9;;file1\007+NINE\n'`},
		}
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\ntwo\nTHREE\nfour \nfive\nsix\nseven\neight\nNINE\nten\n")
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
				Contains("ENTITY lines 3-4"),
			).
			PressPrimaryAction()

		// Staged: the shown change and the one the renderer left out, but not the
		// other entity.
		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour \nfive\nsix\nseven\neight\nnine\nten\n")
	},
})
