package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectARangeOfChangeUnits = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "In hunk mode where the renderer states change units, growing the range takes in a unit at a time, and staging stages them all",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = true
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{{
			Command:            unitsWithASpanOnlyUnitRenderer,
			EntityPatchCommand: entityStagingCommand,
		}}
	},
	SetupRepo: func(shell *Shell) {
		setUpEntityStagingRepo(shell, false)
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(Contains("-three"), Contains("+THREE")).
			Press(keys.Universal.RangeSelectDown).
			SelectedLines(Contains("-three"), Contains("+THREE"), Contains("~ FOUR")).
			Press(keys.Universal.Select)

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nFOUR\nfive\nsix\nseven\neight\nnine\nten\n")
	},
})
