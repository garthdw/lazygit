package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ReportAFailingEntityPatchCommand = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A command that cannot cut the patch says why, and nothing is staged",
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
			NavigateToLine(Contains("-nine")).
			Press(keys.Main.StageChangeUnit)

		t.ExpectPopup().Alert().
			Title(Equals("Error")).
			Content(Equals("The entityPatchCommand failed: no patch for file1::fn::late|1|false")).
			Confirm()

		t.Git().IndexContent("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
	},
})
