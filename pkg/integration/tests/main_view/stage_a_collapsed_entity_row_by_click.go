package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var StageACollapsedEntityRowByClick = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Clicking a row that is a span of a change unit selects the unit, as focusing by keyboard does",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  setUpCollapsedEntityConfig,
	SetupRepo:    setUpCollapsedEntityRepo,
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused()

		t.Views().Main().
			Click(0, 4).
			IsFocused().
			SelectedLines(Contains("formatting only")).
			PressPrimaryAction()

		expectCollapsedEntityStagedAsUnit(t)
	},
})
