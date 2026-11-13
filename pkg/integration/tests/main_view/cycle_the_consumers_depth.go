package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var CycleTheConsumersDepth = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "d asks the consumersCommand again with the next depth: 1, 2, all, 1; a report found by name says it is approximate",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig:  setUpConsumersConfig,
	SetupRepo:    setUpConsumersRepo,
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			NavigateToLine(Contains("-three")).
			Press(keys.Main.ShowConsumers)

		t.ExpectPopup().Menu().
			Title(Equals("Consumers of Outer › early depth ≤ 1")).
			Tap(func() { t.Views().Menu().Press(config.Keybinding{"d"}) })

		t.ExpectPopup().Menu().
			Title(Equals("Consumers of Outer › early depth ≤ 2")).
			Lines(
				Contains("d1").Contains("onAdd"),
				Contains("d2").Contains("render"),
				Contains("Depth ≤ 2, change to ≤ ∞"),
			).
			Tap(func() { t.Views().Menu().Press(config.Keybinding{"d"}) })

		t.ExpectPopup().Menu().
			Title(Equals("Consumers of Outer › early depth ≤ ∞")).
			Tap(func() {
				t.Views().Menu().Content(Contains("Approximate: these were found by name, and still reference early"))
			}).
			ContainsLines(
				Contains("d1").Contains("legacy").Contains("src/old.ts:7"),
				Contains("Depth ≤ ∞, change to ≤ 1"),
			).
			Tap(func() { t.Views().Menu().Press(config.Keybinding{"d"}) })

		t.ExpectPopup().Menu().
			Title(Equals("Consumers of Outer › early depth ≤ 1")).
			Cancel()

		t.FileSystem().FileContent("../consumers-calls", Equals(
			"file1::fn::early;1|1|file1\n"+
				"file1::fn::early;1|2|file1\n"+
				"file1::fn::early;1|0|file1\n"+
				"file1::fn::early;1|1|file1\n"))
	},
})
