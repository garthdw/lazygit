package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var SelectEntities = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Where the renderer states entities, the select toggle steps through lines, change units and entities, which the options bar names, and staging in entity mode stages the whole entity",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
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
			SelectedLines(Contains("-three"))
		t.Views().Options().Content(Contains("Selecting lines"))

		t.Views().Main().
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(Contains("-three"), Contains("+THREE"))
		t.Views().Options().Content(Contains("Selecting change units"))

		t.Views().Main().
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(
				Contains("ENTITY early"),
				Contains("-three"),
				Contains("+THREE"),
				Contains("~ FOUR"),
			)
		t.Views().Options().Content(Contains("Selecting entities"))

		t.Views().Main().
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("ENTITY late"), Contains("-nine"), Contains("+NINE")).
			Press(keys.Universal.PrevItem).
			SelectedLines(Contains("ENTITY early"), Contains("-three"), Contains("+THREE"), Contains("~ FOUR")).
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(Contains("ENTITY early"))
		t.Views().Options().Content(Contains("Selecting lines"))

		t.Views().Main().
			Press(keys.Main.ToggleSelectHunk).
			Press(keys.Main.ToggleSelectHunk).
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("ENTITY late"), Contains("-nine"), Contains("+NINE")).
			Press(keys.Universal.Select)

		t.ExpectToast(Equals("usesIntroduced: NINE is used by nothing staged"))
		t.Git().IndexContent("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nNINE\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::late||false|file1|\n"))
	},
})
