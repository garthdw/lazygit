package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The renderer shows an entity whose unit 1 is a plain `-`/`+` pair and whose unit 2 is a
// `~` row alone, followed by an entity of plain changes. It renders again from the diff
// it is given, so that once the `~` row is in the index only the other units are left.
const unitsWithASpanOnlyUnitRenderer = `d=$(cat); case "$d" in *four*) printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;;file1::fn::early%%3B1;file1\007\033]1717;1;e;3-4;3-4;file1\007ENTITY early\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;d;3;3;file1\007-three\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;a;3;;file1\007+THREE\n` +
	`\033]1717;1;s;2;file1::fn::early%%3B1;file1\007\033]1717;1;e;4-4;4-4;file1\007~ FOUR\n` +
	`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'` +
	`;; *) printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;;file1::fn::early%%3B1;file1\007\033]1717;1;e;3-3;3-3;file1\007ENTITY early\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;d;3;3;file1\007-three\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;a;3;;file1\007+THREE\n` +
	`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'` +
	`;; esac`

var SelectChangeUnits = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Where the renderer states change units, hunk mode steps from unit to unit, a unit that is a `~` row alone included, and staging stages the unit",
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
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("~ FOUR")).
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("-nine"), Contains("+NINE")).
			Press(keys.Universal.PrevItem).
			SelectedLines(Contains("~ FOUR")).
			Press(keys.Universal.PrevItem).
			SelectedLines(Contains("-three"), Contains("+THREE")).
			Press(keys.Universal.NextItem).
			Press(keys.Universal.Select).
			SelectedLines(Contains("-nine"), Contains("+NINE"))

		t.Git().IndexContent("file1", "one\ntwo\nthree\nFOUR\nfive\nsix\nseven\neight\nnine\nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::fn::early;1|2|false|file1|\n"))
	},
})
