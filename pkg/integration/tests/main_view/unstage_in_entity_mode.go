package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var UnstageInEntityMode = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Space in entity mode takes the whole entity of the selection back out of the index through the entityPatchCommand",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		setUpEntityStagingConfig(cfg)
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = true
	},
	SetupRepo: func(shell *Shell) {
		setUpEntityStagingRepo(shell, true)
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			Press(keys.Universal.TogglePanel)

		t.Views().Secondary().
			IsFocused().
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(Contains("ENTITY early"), Contains("-three"), Contains("+THREE"), Contains("-four"), Contains("+FOUR")).
			Press(keys.Universal.Select)

		t.Git().IndexContent("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::early;1||true|file1|\n"))
	},
})
