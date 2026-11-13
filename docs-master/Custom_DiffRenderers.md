# Custom Diff Renderers

Custom diff renderers are useful for showing a better rendering of a diff than git's builtin raw diff, and using one is strongly recommended (I personally prefer delta myself, but that's a matter of personal preference). There are three types of diff renderers that lazygit supports:

- **stdin filters**, e.g. [delta](#delta) and [diff-so-fancy](#diff-so-fancy). They take git's raw output as stdin and produce something nicer on stdout, and they are hooked up using git's GIT_PAGER mechanism. (These used to be called "custom pagers" in earlier lazygit versions.)
- **external diff programs**, e.g. difftastic; these are called using git's `--ext-diff` flag, and they take over diff generation from git completely rather than post-processing git's output.
- **git's raw output using custom arguments**; mainly useful for `--color-words` (or `--word-diff` if you are color blind).

Diff renderers are configured with the `diffRenderers` array in the `git` section of lazygit's config file; it is an array because you can have multiple entries that you can cycle through with the `|` key. This can be useful if you usually prefer a particular diff renderer, but want to use a different one for certain kinds of diffs.

Fields that are shared by all renderer types:

- **type** The type of diff renderer; choices are `stdinFilter`, `extDiff`, or `rawGit`. `stdinFilter` is the default, because it's the most common one; so you can omit this if you use delta.
- **name** A name that is shown in the status bar toast when cycling renderers; defaults to the first word of the renderer command, but can be useful e.g. to distinguish "delta" from "delta side-by-side" if you have entries for both.

Fields only for `stdinFilter`:

- **command** The command line to use for `GIT_PAGER`.

- **colorArg** whether you want the `--color=always` arg in your `git diff` command. Some diff renderers want it set to `always`, others want it set to `never`. The default is `always`, since that's what most renderers need.

Fields only for `extDiff`:

- **command** The command line to use for the `diff.external` git config. If left empty, it uses the global value of git's `diff.external` config; this can be useful if you also want to use it for diffs on the command line, and it also has the advantage that you can configure it per file type in `.gitattributes`; see https://git-scm.com/docs/gitattributes#_defining_an_external_diff_driver.

Fields only for `rawGit`:

- **args** The additional arguments to use in the `git diff` or `git show` call (e.g. `--color-words`), as an array of strings.

The `command` of a `stdinFilter` or `extDiff` renderer is a [Go template](https://pkg.go.dev/text/template) with these variables:

- `{{width}}`: the width of the view that the diff is rendered into.
- `{{colorScheme}}`: `dark` or `light`, depending on whether the terminal has a dark or a light background. Lazygit asks the terminal about this; if yours doesn't tell, set `gui.colorScheme`.
- `{{columnWidth}}` (only for `stdinFilter`): the width of one side of a side-by-side rendering, e.g. for `ydiff -p cat -s -w {{columnWidth}}`.
- `{{diffContext}}` (only for `extDiff`): lazygit's current diff context size, the value controlled by the `{`/`}` keybindings.

A variable can also be written with a leading dot, as in `{{.width}}`. The command can use template expressions too; for example, `delta --paging=never {{if gt .width 160}}--side-by-side{{end}}` shows the diff side by side only when there is room for it, or `delta --syntax-theme={{if eq .colorScheme "light"}}Github{{else}}Dracula{{end}}` picks a different syntax theme based on the background.

Here's an example for a multi-renderer setup:

```yaml
git:
  diffRenderers:
    - command: delta --{{colorScheme}} --paging=never
    - command: ydiff -p cat
      colorArg: never
    - type: extDiff
      command: difft --color=always --background={{colorScheme}} --context={{diffContext}}
    - type: rawGit
      args: [--color-words]
      name: color-words
    - type: rawGit # git's default diff
      name: default
```

## Staging with another renderer

To stage, edit or copy lines from a diff in the focused main view, lazygit has to know which line of which file each row shows. A renderer that lays the diff out its own way, like difftastic, can only tell it that by speaking the OSC 1717 protocol; when the renderer you are browsing with doesn't, lazygit shows git's own diff while the main view is focused.

If you'd rather stage under a renderer of your own that does speak the protocol, name it in `focusedDiffRenderer`. Focusing the main view switches to it when the renderer you are browsing with can't be acted on, and leaving the main view switches back. It is only where the focused main view starts: `|` still cycles from there, and if you cycle while focused, the renderer you cycled to is kept when you leave. For example, to browse with difftastic and stage with delta:

```yaml
git:
  diffRenderers:
    - type: extDiff
      command: difft --color=always --background={{colorScheme}} --context={{diffContext}}
    - command: delta --{{colorScheme}} --paging=never
  focusedDiffRenderer: delta
```

An entry is named by its `name`, or failing that, by the first word of its command (or its first argument, for `rawGit`); if several have the same name, the first is used. If no entry has the name, or the named one doesn't speak the protocol either, git's own diff is shown as before.

## Staging entities

A renderer that understands the code it shows, such as semdiff, can state which entity (a function, a class) and which change unit of it each row belongs to, by adding an OSC 1717 `s` record to the row:

```
ESC ] 1717 ; 1;s;<unit>;<entity id>;<file> BEL
```

`<unit>` is the 1-based change unit, empty for a row outside every unit, such as the entity's header. `<entity id>` is percent-encoded. Such rows can be staged and unstaged as a whole in the focused main view of the files panel, with `stageEntity` (default `s`) for the entity of the selected row, and `stageChangeUnit` (default `u`) for its change unit: the units of every selected row, or the whole entity when the selected rows have no unit. On the staged side's diff they unstage instead. The two commands are listed only under a renderer that has an `entityPatchCommand`.

lazygit doesn't know how to cut a patch out of an entity, so the renderer has to be asked, with `entityPatchCommand`. It must print a patch that `git apply --cached` can apply as it is; unstaging is asked for by the flags below, so the command prints the reversed patch itself, and lazygit never passes `--reverse` to git. A failing command (non-zero exit) is reported to the user with its error output. On success, every line it writes to stderr that starts with `warning: ` (for example `warning: usesIntroduced: formatPrice uses currency, which is not staged`) is shown in a toast once the patch is staged; other stderr lines are ignored. When the patch holds every change the file has on that side, lazygit stages or unstages the whole file instead of applying it, as it does for line staging, so that a new, deleted or renamed file is staged as what it is. The command is run in the repo's root, and the entity staging commands are disabled, with that as the reason, for renderers that don't have one.

```yaml
git:
  diffRenderers:
    - command: semdiff --pager --color=always --width={{width}}
      colorArg: never
      entityPatchCommand: >-
        semdiff patch --entity {{entity}}
        {{if units}}--unit {{units}}{{end}}
        {{if staged}}--staged --reverse{{end}}
        -- {{path}} {{oldPath}}
```

The command is run by a shell, so the text values are quoted for you; don't put quotes around them. The values are:

| Value | Meaning |
| --- | --- |
| `{{entity}}` | The entity id, decoded |
| `{{path}}` | The file's path, relative to the repo root |
| `{{oldPath}}` | The path before a rename, only when unstaging a staged rename (the renderer needs both paths then); empty otherwise |
| `{{units}}` | The unit numbers, separated by spaces; empty to stage the whole entity |
| `{{staged}}`, `{{reverse}}` | True when unstaging (both are the same), false when staging |

## Showing the consumers of an entity

A renderer that says which entity each row belongs to (see above) can also be asked what depends on it, with `consumersCommand`. It must print JSON, described below, on stdout. A failing command (non-zero exit) is reported to the user with the first line of its error output. The command is run in the repo's root, and the consumers popup is disabled, with that as the reason, for renderers that don't have one.

```yaml
git:
  diffRenderers:
    - command: semdiff --pager --color=always --width={{width}}
      colorArg: never
      consumersCommand: semdiff impact --entity {{entity}} --depth {{depth}} --json
```

The command is run by a shell, so the text values are quoted for you; don't put quotes around them. The values are:

| Value | Meaning |
| --- | --- |
| `{{entity}}` | The entity id of the selected row, decoded |
| `{{path}}` | The file of the selected row, relative to the repo root |
| `{{depth}}` | How many steps of dependents to list: 1 when the popup opens, then 2, then 0 for all of them, as `d` cycles them |

The JSON has these fields; any others are ignored:

```
{ "schemaVersion": 1,
  "entity": { "name", "breadcrumb": [...] },
  "basis": "graph" | "textual",
  "complete": true | false,
  "incompleteReasons": [ { "detail" } ],
  "dependents": [ { "breadcrumb": [...], "file", "line", "depth" } ],
  "tests": [ { "breadcrumb": [...], "file", "line" } ],
  "possibleCallers": [ { "name", "file", "line", "sites": [ { "kind", "line" } ] } ] }
```

Each `breadcrumb` starts with the file, which the popup leaves out. A `basis` of `textual` says that the dependents were found by name rather than by what the code refers to, and the popup says it is approximate. When `complete` is false, the `possibleCallers` are listed under a heading of their own: they may or may not be consumers. `file` is relative to the repo root and `line` is 1-based.

The popup is opened with `showConsumers` (default `i`) in the focused main view, on a row that has an `s` record. `enter` opens the file of the selected row in the editor at its line, like `edit` does, and `d` lists the dependents of another depth.

## Delta:

```yaml
git:
  diffRenderers:
    - command: delta --{{colorScheme}} --paging=never
```

![](https://i.imgur.com/QJpQkF3.png)

`--{{colorScheme}}` passes `--dark` or `--light` to delta, so that it matches the background of your terminal.

A cool feature of delta is --hyperlinks, which renders clickable links for the line numbers in the left margin, and lazygit supports these. To use them, set the `command:` field to `delta --{{colorScheme}} --paging=never --line-numbers --hyperlinks --hyperlinks-file-link-format="lazygit-edit://{path}:{line}"`; this allows you to click on an underlined line number in the diff to jump right to that same line in your editor.

Note that delta's `--navigate` option doesn't work in lazygit, for technical reasons.

## Diff-so-fancy

```yaml
git:
  diffRenderers:
    - command: diff-so-fancy
```

![](https://i.imgur.com/rjH1TpT.png)

## ydiff

```yaml
gui:
  sidePanelWidth: 0.2 # gives you more space to show things side-by-side
git:
  diffRenderers:
    - colorArg: never
      command: ydiff -p cat
```

![](https://i.imgur.com/vaa8z0H.png)
