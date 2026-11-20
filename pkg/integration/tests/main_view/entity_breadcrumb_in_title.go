package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The first entity's name contains the id's separator, as a Rust trait impl's does, so
// splitting the id would name it wrongly.
const breadcrumbRenderer = `cat >/dev/null; printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;;file1::impl::fmt::Display for Cart::early;file1\007` +
	`\033]1717;1;b;file1::impl::fmt::Display for Cart::early;fmt::Display for Cart %%E2%%80%%BA early;file1\007` +
	`\033]1717;1;e;3-3;3-3;file1\007ENTITY early\n` +
	`\033]1717;1;s;1;file1::impl::fmt::Display for Cart::early;file1\007\033]1717;1;d;3;3;file1\007-three\n` +
	`\033]1717;1;s;1;file1::impl::fmt::Display for Cart::early;file1\007\033]1717;1;a;3;;file1\007+THREE\n` +
	`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;b;file1::fn::late;late;file1\007` +
	`\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
	`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'`

var EntityBreadcrumbInTitle = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "Where the renderer states entities' breadcrumbs, the focused main view's title names the entity of the selected change unit or entity, but not of a selected line",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{{Command: breadcrumbRenderer}}
	},
	SetupRepo: func(shell *Shell) {
		setUpEntityStagingRepo(shell, false)
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Main().Title(Equals("Unstaged changes"))

		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(Contains("-three"), Contains("+THREE")).
			Title(Equals("Unstaged changes · fmt::Display for Cart › early")).
			Press(keys.Universal.NextItem).
			SelectedLines(Contains("-nine"), Contains("+NINE")).
			Title(Equals("Unstaged changes · late")).
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(Contains("ENTITY late"), Contains("-nine"), Contains("+NINE")).
			Title(Equals("Unstaged changes · late")).
			Press(keys.Main.ToggleSelectHunk).
			SelectedLines(Contains("ENTITY late")).
			Title(Equals("Unstaged changes")).
			Press(keys.Universal.Return)

		t.Views().Files().IsFocused()
		t.Views().Main().Title(Equals("Unstaged changes"))
	},
})
