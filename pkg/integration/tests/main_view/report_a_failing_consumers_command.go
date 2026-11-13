package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ReportAFailingConsumersCommand = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A consumersCommand that fails, or prints something other than the consumers, says so with the first line of what it printed to stderr",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		setUpConsumersConfig(cfg)
		cfg.GetUserConfig().Git.DiffRenderers[0].ConsumersCommand = `sh ../consumers.sh {{entity}} 7 {{path}}`
	},
	SetupRepo: setUpConsumersRepo,
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			NavigateToLine(Contains("-three")).
			Press(keys.Main.ShowConsumers)

		t.ExpectPopup().Alert().
			Title(Equals("Error")).
			Content(Equals("The consumersCommand failed: no report for depth 7")).
			Confirm()

		t.Views().Main().IsFocused()
	},
})
