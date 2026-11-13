package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The file's first change is a unit that is a `~` row alone, and its first plain
// `-`/`+` change comes after it.
const firstUnitIsASpanRenderer = `cat >/dev/null; printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;;file1::module::imports;file1\007\033]1717;1;e;3-3;3-3;file1\007ENTITY imports\n` +
	`\033]1717;1;s;;file1::module::imports;file1\007\033]1717;1;c;2;2;file1\007 two\n` +
	`\033]1717;1;s;1;file1::module::imports;file1\007\033]1717;1;e;3-3;3-3;file1\007~ THREE\n` +
	`\033]1717;1;s;;file1::module::imports;file1\007\033]1717;1;c;4;4;file1\007 four\n` +
	`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'`

var SelectFirstUnitOnFocusingMainView = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Where the renderer states change units, focusing the main view selects the first unit, a `~` row alone included, and steps by unit, whatever useHunkModeInDiffView says",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{{
			Command:            firstUnitIsASpanRenderer,
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
			SelectedLines(Contains("~ THREE")).
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("-nine"), Contains("+NINE"))
	},
})
