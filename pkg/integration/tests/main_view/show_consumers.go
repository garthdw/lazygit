package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

var ShowConsumers = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "The consumers of the selected row's entity are listed, with an incomplete section, and enter opens one in the editor at its line",
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
			Press(keys.Universal.OptionMenu)

		t.ExpectPopup().Menu().
			Title(Equals("Keybindings")).
			Tap(func() {
				t.Views().Menu().Content(Contains("Show consumers"))
			}).
			Cancel()

		t.Views().Main().
			NavigateToLine(Contains("-three")).
			Press(keys.Main.ShowConsumers)

		t.ExpectPopup().Menu().
			Title(Equals("Consumers of Outer › early depth ≤ 1")).
			Lines(
				Contains("d1").Contains("onAdd").Contains("src/checkout.ts:48"),
				Contains("test").Contains("cart › adds").Contains("src/cart.spec.ts:14"),
				Contains("INCOMPLETE · 1 possible callers (1 name used as a value)"),
				Contains("?").Contains("register").Contains("src/events.ts:33"),
				Contains("Depth ≤ 1, change to ≤ 2"),
			).
			Select(Contains("register")).
			Confirm()

		t.FileSystem().FileContent("edit-command", Contains("/repo/src/events.ts:33\n"))
		t.FileSystem().FileContent("../consumers-calls", Equals("file1::fn::early;1|1|file1\n"))

		t.Views().Main().IsFocused()
	},
})
