package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageAnEntity = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Stage the entity of the selected row, which the renderer's command cuts out; the entity id reaches it decoded",
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
			Press(keys.Universal.OptionMenu)

		t.ExpectPopup().Menu().
			Title(Equals("Keybindings")).
			Tap(func() {
				t.Views().Menu().
					Content(Contains("Stage entity")).
					Content(Contains("Stage change unit"))
			}).
			Cancel()

		t.Views().Main().
			Press(keys.Universal.NextItem).
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("-nine"), Contains("+NINE")).
			Press(keys.Main.StageEntity)

		t.ExpectToast(Equals("usesIntroduced: NINE is used by nothing staged"))
		t.Git().IndexContent("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nNINE\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::late||false|file1|\n"))
	},
})
