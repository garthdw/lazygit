package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectADiffOfSpansOnly = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A diff whose rows are all spans, as when every change is one in place, still has something to select and act on",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{
			{Command: `cat >/dev/null; printf '` +
				`\033]1717;1\007` +
				`\033]1717;1;f;;;file1\007FILE file1\n` +
				`\033]1717;1;e;3-3;3-3;file1\007~ three\n'`},
		}
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\n")
		shell.Commit("one")

		shell.UpdateFile("file1", "one\ntwo\nTHREE\nfour\nfive\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(Contains("~ three")).
			PressPrimaryAction()

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour\nfive\n")
	},
})
