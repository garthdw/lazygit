package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageTheUnitsOfARange = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A range of rows stages every change unit it covers, once each",
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
			Press(keys.Universal.ToggleRangeSelect).
			SelectNextItem().
			SelectNextItem().
			SelectNextItem().
			SelectedLines(
				Contains("-three"),
				Contains("+THREE"),
				Contains("-four"),
				Contains("+FOUR"),
			).
			Press(keys.Main.StageChangeUnit)

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nFOUR\nfive\nsix\nseven\neight\nnine\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::early;1|1 2|false|file1|\n"))
	},
})
