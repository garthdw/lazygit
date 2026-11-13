package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var NoEntityStagingWithoutAPatchCommand = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "The entity commands are disabled, with the reason, under a renderer that has no entityPatchCommand",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{entityStagingRenderer("")}
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
			Press(keys.Universal.OptionMenu)

		t.ExpectPopup().Menu().
			Title(Equals("Keybindings")).
			Tap(func() {
				t.Views().Menu().
					Content(DoesNotContain("Stage entity")).
					Content(DoesNotContain("Stage change unit"))
			}).
			Cancel()

		t.Views().Main().
			SelectedLines(Contains("-three")).
			Press(keys.Main.StageEntity)
		t.ExpectToast(Contains("The diff renderer has no entityPatchCommand"))

		t.Views().Main().Press(keys.Main.StageChangeUnit)
		t.ExpectToast(Contains("The diff renderer has no entityPatchCommand"))

		t.Git().IndexContent("file1", entityStagingFileBefore)
	},
})
