package main_view

import (
	"github.com/jesseduffield/lazygit/pkg/config"
	. "github.com/jesseduffield/lazygit/pkg/integration/components"
)

// The tests of staging by entity share a renderer in the manner of an entity-aware one,
// and an entityPatchCommand that stands in for it. The repo has file1 with `three` and
// `four` changed, which the renderer calls the units 1 and 2 of one entity, and `nine`
// changed, which is another entity. The entity id has a `;` in it, written percent-encoded
// in the records, to show that it reaches the command decoded, and as one argument.
const (
	entityStagingFileBefore = "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\n"
	entityStagingFileAfter  = "one\ntwo\nTHREE\nFOUR\nfive\nsix\nseven\neight\nNINE\nten\n"
)

// The patch the stand-in prints depends on what it is asked for, which it also records
// in ../entity-calls as entity|units|unstage|path|oldPath. The patches are written for an
// index that has none of the changes, or all of the entity's, as the tests have it.
const fakeEntityPatchScript = `#!/bin/sh
printf '%s|%s|%s|%s|%s\n' "$1" "$2" "$3" "$4" "$5" >> ../entity-calls
case "$1" in
'file2::fn::all')
printf '%s\n' 'diff --git a/file2 b/file2' 'new file mode 100644' '--- /dev/null' '+++ b/file2' '@@ -0,0 +1,2 @@' '+a' '+WRONG'
exit 0 ;;
esac
echo "diff --git a/file1 b/file1"
echo "--- a/file1"
echo "+++ b/file1"
case "$1|$2|$3" in
'file1::fn::early;1||false' | 'file1::fn::early;1|1 2|false') cat <<'PATCH'
@@ -2,4 +2,4 @@
 two
-three
-four
+THREE
+FOUR
 five
PATCH
;;
'file1::fn::early;1|1|false') cat <<'PATCH'
@@ -2,3 +2,3 @@
 two
-three
+THREE
 four
PATCH
;;
'file1::fn::early;1|2|false') cat <<'PATCH'
@@ -3,3 +3,3 @@
 three
-four
+FOUR
 five
PATCH
;;
'file1::fn::early;1||true') cat <<'PATCH'
@@ -2,4 +2,4 @@
 two
-THREE
-FOUR
+three
+four
 five
PATCH
;;
'file1::fn::early;1|1|true') cat <<'PATCH'
@@ -2,3 +2,3 @@
 two
-THREE
+three
 FOUR
PATCH
;;
'file1::fn::late||false') cat <<'PATCH'
@@ -8,3 +8,3 @@
 eight
-nine
+NINE
 ten
PATCH
echo "warning: usesIntroduced: NINE is used by nothing staged" >&2
echo "not a warning" >&2
;;
*) echo "no patch for $1|$2|$3" >&2; exit 1 ;;
esac
`

// entityStagingRenderer is the diff renderer of the tests: it ignores its input and
// shows the rows of file1 with the records an entity-aware renderer would give them.
func entityStagingRenderer(entityPatchCommand string) config.DiffRendererConfig {
	return config.DiffRendererConfig{
		Command: `cat >/dev/null; printf '` +
			`\033]1717;1\007` +
			`\033]1717;1;f;;;file1\007FILE file1\n` +
			`\033]1717;1;s;;file1::fn::early%%3B1;file1\007\033]1717;1;e;3-4;3-4;file1\007ENTITY early\n` +
			`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;d;3;3;file1\007-three\n` +
			`\033]1717;1;s;1;file1::fn::early%%3B1;file1\007\033]1717;1;a;3;;file1\007+THREE\n` +
			`\033]1717;1;s;2;file1::fn::early%%3B1;file1\007\033]1717;1;d;4;4;file1\007-four\n` +
			`\033]1717;1;s;2;file1::fn::early%%3B1;file1\007\033]1717;1;a;4;;file1\007+FOUR\n` +
			`\033]1717;1;s;;file1::fn::late;file1\007\033]1717;1;e;9-9;9-9;file1\007ENTITY late\n` +
			`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;d;9;9;file1\007-nine\n` +
			`\033]1717;1;s;1;file1::fn::late;file1\007\033]1717;1;a;9;;file1\007+NINE\n'`,
		EntityPatchCommand: entityPatchCommand,
	}
}

const entityStagingCommand = `sh ../entity-patch.sh {{entity}} "{{units}}" {{staged}} {{path}} {{oldPath}}`

func setUpEntityStagingConfig(cfg *config.AppConfig) {
	cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
	cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{entityStagingRenderer(entityStagingCommand)}
}

// setUpEntityStagingRepo leaves the changes unstaged, or, with stagedEarly, the first
// entity's staged.
func setUpEntityStagingRepo(shell *Shell, stagedEarly bool) {
	shell.CreateFile("../entity-patch.sh", fakeEntityPatchScript)
	shell.CreateFileAndAdd("file1", entityStagingFileBefore)
	shell.Commit("one")

	if stagedEarly {
		shell.UpdateFileAndAdd("file1", "one\ntwo\nTHREE\nFOUR\nfive\nsix\nseven\neight\nnine\nten\n")
	}
	shell.UpdateFile("file1", entityStagingFileAfter)
}

// entityStagingWholeFileRenderer shows file2, a new file that the one entity of it, which
// the stand-in's patch is for, makes up all of. The patch is wrong on purpose, so that
// applying it, rather than staging the file, shows.
func entityStagingWholeFileRenderer() config.DiffRendererConfig {
	return config.DiffRendererConfig{
		Command: `cat >/dev/null; printf '` +
			`\033]1717;1\007` +
			`\033]1717;1;f;;;file2\007FILE file2\n` +
			`\033]1717;1;s;;file2::fn::all;file2\007\033]1717;1;e;1-2;;file2\007ENTITY all\n` +
			`\033]1717;1;s;1;file2::fn::all;file2\007\033]1717;1;a;1;;file2\007+a\n` +
			`\033]1717;1;s;1;file2::fn::all;file2\007\033]1717;1;a;2;;file2\007+b\n'`,
		EntityPatchCommand: entityStagingCommand,
	}
}

// The tests of the consumers popup use the renderer of the staging tests, whose rows
// say which entity they belong to, and a consumersCommand that stands in for one: it
// prints a fixed report for each depth, and records what it was asked for in
// ../consumers-calls as entity|depth|path.
//
// The reports are not what a renderer would print for one repo; they are what shows
// each part of the popup: at depth 1 an incomplete report with a possible caller, at
// depth 2 a complete one with more rows, and for all depths (0) one found by name.
const fakeConsumersScript = `#!/bin/sh
printf '%s|%s|%s\n' "$1" "$2" "$3" >> ../consumers-calls
case "$2" in
1) cat <<'JSON'
{"schemaVersion": 1, "entity": {"name": "early", "breadcrumb": ["file1", "Outer", "early"]},
 "basis": "graph", "depth": 1, "complete": false,
 "incompleteReasons": [{"code": "value_refs", "detail": "1 name used as a value"}],
 "dependents": [{"name": "onAdd", "breadcrumb": ["src/checkout.ts", "onAdd"], "file": "src/checkout.ts", "line": 48, "depth": 1}],
 "tests": [{"name": "adds", "breadcrumb": ["src/cart.spec.ts", "cart", "adds"], "file": "src/cart.spec.ts", "line": 14}],
 "possibleCallers": [{"name": "register", "file": "src/events.ts", "line": 30, "sites": [{"kind": "value_ref", "line": 33}]}],
 "futureField": true}
JSON
;;
2) cat <<'JSON'
{"schemaVersion": 1, "entity": {"name": "early", "breadcrumb": ["file1", "Outer", "early"]},
 "basis": "graph", "depth": 2, "complete": true, "incompleteReasons": [],
 "dependents": [
  {"name": "render", "breadcrumb": ["src/checkout.ts", "render"], "file": "src/checkout.ts", "line": 12, "depth": 2},
  {"name": "onAdd", "breadcrumb": ["src/checkout.ts", "onAdd"], "file": "src/checkout.ts", "line": 48, "depth": 1}],
 "tests": [], "possibleCallers": []}
JSON
;;
0) cat <<'JSON'
{"schemaVersion": 1, "entity": {"name": "early", "breadcrumb": ["file1", "Outer", "early"]},
 "basis": "textual", "depth": 0, "complete": false,
 "incompleteReasons": [{"code": "textual", "detail": "matched by name"}],
 "dependents": [{"name": "legacy", "breadcrumb": ["src/old.ts", "legacy"], "file": "src/old.ts", "line": 7, "depth": 1}],
 "tests": [], "possibleCallers": []}
JSON
;;
*) echo "no report for depth $2
second line" >&2; exit 1 ;;
esac
`

const consumersCommand = `sh ../consumers.sh {{entity}} {{depth}} {{path}}`

func consumersRenderer(consumersCommand string) config.DiffRendererConfig {
	renderer := entityStagingRenderer("")
	renderer.ConsumersCommand = consumersCommand
	return renderer
}

func setUpConsumersConfig(cfg *config.AppConfig) {
	cfg.GetUserConfig().Gui.UseHunkModeInDiffView = false
	cfg.GetUserConfig().OS.EditAtLine = "echo {{filename}}:{{line}} > edit-command"
	cfg.GetUserConfig().Git.DiffRenderers = []config.DiffRendererConfig{consumersRenderer(consumersCommand)}
}

func setUpConsumersRepo(shell *Shell) {
	shell.CreateFile("../consumers.sh", fakeConsumersScript)
	setUpEntityStagingRepo(shell, false)
}
