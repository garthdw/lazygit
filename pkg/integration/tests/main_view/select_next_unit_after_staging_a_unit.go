package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The renderer shows an entity whose unit 1 is a `~` row, a span with no addition or
// deletion of its own, followed by a `-`/`+` pair, and whose unit 2 is another `~` row,
// ahead of an entity of plain changes. It renders again from the diff it is given, so that once unit 1 is in the
// index the entity has only the `~` row left.
const unitThenSpanRenderer = `d=$(cat); case "$d" in *-three*) printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;;file1::fn::early%%3B1;file1\007\033]1717;1;e;3-4;3-4;file1\007ENTITY early\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;e;3-3;3-3;file1\007~ THREE\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;d;3;3;file1\007-three\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;a;3;;file1\007+THREE\n` +
	`\033]1717;1;s;2;file1::fn::early%%3B1;file1\007\033]1717;1;e;4-4;4-4;file1\007~ FOUR\n` +
	`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'` +
	`;; *) printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;;file1::fn::early%%3B1;file1\007\033]1717;1;e;4-4;4-4;file1\007ENTITY early\n` +
	`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;e;4-4;4-4;file1\007~ FOUR\n` +
	`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'` +
	`;; esac`

var SelectNextUnitAfterStagingAUnit = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "After staging a change unit the selection lands on the next unit, even a `~` row that is no addition or deletion, however many rows of the staged unit come before the selected one",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{{
			Command:            unitThenSpanRenderer,
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
			SelectedLines(Contains("-three")).
			Press(keys.Main.StageChangeUnit).
			SelectedLines(Contains("~ FOUR"))

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
	},
})
