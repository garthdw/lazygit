package ui

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var MenuFitsSectionHeaders = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A menu with section headers is tall enough to show all its items without scrolling",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  func(config *config.AppConfig) {},
	SetupRepo:    func(shell *Shell) {},
	// Tall enough that the menu is not limited by the screen
	Width:  100,
	Height: 200,
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.OptionMenu)

		t.Views().Menu().
			IsFocused().
			Press(keys.Universal.GotoBottom).
			OriginY(0)
	},
})
