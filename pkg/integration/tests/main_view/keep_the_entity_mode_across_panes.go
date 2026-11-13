package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var KeepTheEntityModeAcrossPanes = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Switching between the unstaged and staged panes keeps the entity selection mode when the renderer states entities in both",
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
			Press(keys.Main.ToggleSelectHunk).
			Press(keys.Universal.TogglePanel)

		t.Views().Secondary().
			IsFocused().
			SelectedLines(Contains("ENTITY early"), Contains("-three"), Contains("+THREE"), Contains("-four"), Contains("+FOUR"))
		t.Views().Options().Content(Contains("Selecting entities"))
	},
})
