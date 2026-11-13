package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageEveryChangeOfANewFileAsTheFile = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "An entity that is every change of a file stages the file itself, rather than applying the renderer's patch",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{entityStagingWholeFileRenderer()}
	},
	SetupRepo: func(shell *Shell) {
		setUpEntityStagingRepo(shell, false)
		shell.UpdateFile("file1", entityStagingFileBefore)
		shell.CreateFile("file2", "a\nb\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Lines(Contains("file2").IsSelected()).
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			NavigateToLine(Contains("ENTITY all")).
			Press(keys.Main.StageEntity)

		// The patch the stand-in prints says WRONG; it was never applied.
		t.Git().IndexContent("file2", "a\nb\n")
		t.Views().Files().Lines(Contains("A").Contains("file2"))
	},
})
