package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The renderer shows the file's only change, a reformatted function, as one collapsed
// row: the row is the entity's only unit, and a span over the line it changed.
const collapsedEntityRenderer = `cat >/dev/null; printf '` +
	`\033]1717;1\007` +
	`\033]1717;1;f;;;a.ts\007a.ts · TypeScript · 1 entity\n` +
	`\n` +
	`  ≈ t\n` +
	`\n` +
	`\033]1717;1;s;1;a.ts::function::t;a.ts\007\033]1717;1;e;2-2;2-2;a.ts\007── a.ts › t L1–3 formatting only\n'`

const collapsedEntityPatchScript = `#!/bin/sh
printf '%s|%s|%s|%s|%s\n' "$1" "$2" "$3" "$4" "$5" >> ../entity-calls
cat <<'PATCH'
diff --git a/a.ts b/a.ts
--- a/a.ts
+++ b/a.ts
@@ -1,3 +1,3 @@
 export function t(a: number) {
-  return a + 1
+    return a +  1
 }
PATCH
`

var StageACollapsedEntityRow = NewIntegrationTest(NewIntegrationTestArgs{
	Description:  "A file whose only change is an entity the renderer collapses to one row of a unit can be selected and staged there",
	ExtraCmdArgs: []string{},
	Skip:         false,
	SetupConfig: func(cfg *config.AppConfig) {
		cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{{
			Command:            collapsedEntityRenderer,
			EntityPatchCommand: `sh ../entity-patch.sh {{entity}} "{{units}}" {{staged}} {{path}} {{oldPath}}`,
		}}
	},
	SetupRepo: func(shell *Shell) {
		shell.CreateFile("../entity-patch.sh", collapsedEntityPatchScript)
		shell.CreateFileAndAdd("a.ts", "export function t(a: number) {\n  return a + 1\n}\n")
		shell.Commit("one")
		shell.UpdateFile("a.ts", "export function t(a: number) {\n    return a +  1\n}\n")
	},
	Run: func(t *TestDriver, keys config.KeybindingConfig) {
		t.Views().Files().
			IsFocused().
			Press(keys.Universal.FocusMainView)

		t.Views().Main().
			IsFocused().
			SelectedLines(Contains("formatting only")).
			PressPrimaryAction()

		t.Git().IndexContent("a.ts", "export function t(a: number) {\n    return a +  1\n}\n")
		t.FileSystem().FileContent("../entity-calls", Equals("a.ts::function::t|1|false|a.ts|\n"))
	},
})
