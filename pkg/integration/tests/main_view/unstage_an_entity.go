package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var UnstageAnEntity = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Take the entity of the selected row back out of the index from the staged half of the main view",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  setUpEntityStagingConfig,
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
			Press(keys.Universal.OptionMenu)

		t.ExpectPopup().Menu().
			Title(Equals("Keybindings")).
			Tap(func() {
				t.Views().Menu().
					Content(Contains("Unstage entity")).
					Content(Contains("Unstage change unit"))
			}).
			Cancel()

		t.Views().Secondary().
			SelectedLines(Contains("-three"), Contains("+THREE")).
			Press(keys.Main.StageEntity)

		t.Git().IndexContent("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::early;1||true|file1|\n"))
	},
})
