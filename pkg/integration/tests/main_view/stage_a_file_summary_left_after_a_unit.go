package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The renderer shows a unit of an entity, then a file summary: a row with no entity
// that is a span over each of two whitespace changes. Once the unit is in the index it
// shows the summary alone.
const unitThenFileSummaryRenderer = `d=$(cat); case "$d" in *three*) printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\033]1717;1;s;1;file1::module::imports;file1\007\033]1717;1;d;3;3;file1\007-three\n` +
	`\033]1717;1;s;1;file1::module::imports;file1\007\033]1717;1;a;3;;file1\007+THREE\n` +
	`\n` +
	`\033]1717;1;e;6-6;6-6;file1\007\033]1717;1;e;9-9;9-9;file1\007SUMMARY formatting only\n'` +
	`;; *) printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;file1\007FILE file1\n` +
	`\n` +
	`\033]1717;1;e;6-6;6-6;file1\007\033]1717;1;e;9-9;9-9;file1\007SUMMARY formatting only\n'` +
	`;; esac`

const unitThenFileSummaryPatchScript = `#!/bin/sh
printf '%s|%s|%s|%s|%s\n' "$1" "$2" "$3" "$4" "$5" >> ../entity-calls
cat <<'PATCH'
diff --git a/file1 b/file1
--- a/file1
+++ b/file1
@@ -2,3 +2,3 @@
 two
-three
+THREE
 four
PATCH
`

var StageAFileSummaryLeftAfterAUnit = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "After staging a unit, the file summary that is all that is left, a block with no entity, is staged by its lines in the unit mode carried over",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{{
			Command:            unitThenFileSummaryRenderer,
			EntityPatchCommand: entityStagingCommand,
		}}
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("../entity-patch.sh", unitThenFileSummaryPatchScript)
		shell.CreateFileAndAdd("file1", "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n")
		shell.Commit("one")
		shell.UpdateFile("file1", "one\ntwo\nTHREE\nfour\nfive\nsix \nseven\neight\nnine \nten\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(Contains("-three"), Contains("+THREE")).
			PressPrimaryAction()

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour\nfive\nsix\nseven\neight\nnine\nten\n")

		t.Views().Main().
			SelectedLines(Contains("SUMMARY formatting only")).
			PressPrimaryAction()

		t.Git().IndexContent("file1", "one\ntwo\nTHREE\nfour\nfive\nsix \nseven\neight\nnine \nten\n")
		t.FileSystem().FileContent("../entity-calls", Equals("file1::module::imports|1|false|file1|\n"))
	},
})
