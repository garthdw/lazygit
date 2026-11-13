package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageAChangeUnit = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Stage only the change unit of the selected row",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  setUpEntityStagingConfig,
	SetupRepo: func(shell *Shell) {
		setUpEntityStagingRepo(shell, false)
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(Contains("-three")).
			Press(keys.Main.StageChangeUnit)

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::early;1|1|false|file1|\n"))
	},
})
