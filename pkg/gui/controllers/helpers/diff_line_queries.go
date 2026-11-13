package helpers

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jesseduffield/lazygit/pkg/gocui"
	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/samber/lo"
)

// The questions a diff view can be asked about what it is showing — where the change
// lines are, which block or file a row belongs to — answered in the view-line terms a
// cursor and a click speak, except for what a selection covers (see
// DiffLinesInBufferRange). They are all built on the identities recovered in
// diff_line_helper.go, which is where the answering stops and the recovering starts.

// DiffLinesInBufferRange returns the identity of every diff line shown by the rows in
// the inclusive buffer-line range [first, last] of view's rendered diff, in display
// order. Rows whose identity can't be recovered are left out.
//
// The range is what a selection covers (see View.SelectedBufferLineRange). It is given
// in buffer lines because a buffer line stays on its line of the content however the
// view wraps it, so the selection can be read when a key is pressed and acted on after
// the view has changed.
//
// A row can show more than one diff line — a side-by-side rendering puts a deletion
// beside the addition replacing it — and all of them are reported: what the user
// pointed at is the row, so everything on it is selected.
func (self *DiffLineHelper) DiffLinesInBufferRange(view *gocui.View, first int, last int) []types.DiffLineInfo {
	identities := self.resolveDiffLineIdentities(diffLineContents(view))

	infos := []types.DiffLineInfo{}
	for bufferLine := first; bufferLine <= min(last, len(identities)-1); bufferLine++ {
		infos = append(infos, identities[bufferLine]...)
	}
	return self.inRepoTerms(view, infos)
}

// ChangeLineOrdinals says, for each of the given change lines, which of its file's
// changes it is in the given diff — its place among them, counted from the top of the
// file — keyed by file. Lines the diff doesn't have are left out.
//
// It is how a line is named in something built out of a diff rather than being that diff:
// the custom patch holds the lines it was given in the order the file has them, so a
// place among a file's changes is a line of the patch.
func (self *DiffLineHelper) ChangeLineOrdinals(
	diff string, infos []types.DiffLineInfo,
) map[string][]int {
	toDiffLineInfo := self.diffLineInfoConverter()
	ordinals := map[patchLine]int{}
	counts := map[string]int{}
	for _, parsed := range parseAllDiffLinesFromBuffer(strings.Split(diff, "\n")) {
		if !parsed.ok {
			continue
		}
		info := toDiffLineInfo(parsed.parsed)
		if !info.IsChange() {
			continue
		}
		ordinals[patchLineOf(info)] = counts[info.Path]
		counts[info.Path]++
	}

	ordinalsByPath := map[string][]int{}
	for _, info := range infos {
		if ordinal, ok := ordinals[patchLineOf(info)]; ok {
			ordinalsByPath[info.Path] = append(ordinalsByPath[info.Path], ordinal)
		}
	}
	return ordinalsByPath
}

// inRepoTerms brings the paths of lines recovered from a view into the repo's terms.
//
// They are in them already for a diff of the repo's own files. The pane previewing the
// custom patch, though, shows a diff of the two trees the patch was materialized into: a
// diff renderer states the path it was handed there, which is under the tree's own name,
// while the diff's text names the trees where an ordinary diff has git's a/ and b/
// prefixes and so needs nothing.
func (self *DiffLineHelper) inRepoTerms(view *gocui.View, infos []types.DiffLineInfo) []types.DiffLineInfo {
	toRepoTerms := self.repoTermsMapper(view)
	return lo.Map(infos, func(info types.DiffLineInfo, _ int) types.DiffLineInfo {
		info.Path = toRepoTerms(info.Path)
		return info
	})
}

// repoTermsMapper returns how a path recovered from view is brought into the repo's
// terms, for the callers that have a path rather than a whole identity to bring. The
// mapping is worked out once, per view rather than per path.
func (self *DiffLineHelper) repoTermsMapper(view *gocui.View) func(string) string {
	if !self.ShowsCustomPatch(view) {
		return func(path string) string { return path }
	}

	worktreePath := self.c.Git().RepoPaths.WorktreePath()
	treesDir := self.c.Git().Patch.PatchBuilder.TempDir()
	return func(path string) string {
		return repoPathOfTreePath(path, treesDir, worktreePath)
	}
}

// repoPathOfTreePath maps a path under one of the trees the custom patch was materialized
// into to the file of the repo it stands for: the path is the tree's name followed by the
// file's own, stated either against the directory holding the trees or against the repo,
// depending on how the renderer that stated it was given it.
func repoPathOfTreePath(path string, treesDir string, worktreePath string) string {
	root := worktreePath
	if treesDir != "" && strings.HasPrefix(path, treesDir+string(filepath.Separator)) {
		root = treesDir
	}
	relativePath, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}

	segments := strings.Split(filepath.ToSlash(relativePath), "/")
	if len(segments) > 1 && (segments[0] == "a" || segments[0] == "b") {
		relativePath = filepath.Join(segments[1:]...)
	}
	return filepath.Join(worktreePath, relativePath)
}

// ChangeLinesInBufferRange returns the change lines — the additions and deletions —
// among the diff lines shown by the rows in the inclusive buffer-line range, plus the
// spans standing for groups of them (see types.DiffLineSpan). Those are
// the lines a patch is built from: a patch carries whatever context it needs around
// them by itself, so a selection contributes only its changes.
func (self *DiffLineHelper) ChangeLinesInBufferRange(view *gocui.View, first int, last int) []types.DiffLineInfo {
	return lo.Filter(self.DiffLinesInBufferRange(view, first, last),
		func(info types.DiffLineInfo, _ int) bool { return info.IsChange() || info.Type == types.DiffLineSpan })
}

// changeLines resolves view's rendered diff to one flag per buffer line: whether
// that row is a change line (an addition or a deletion), as opposed to context, a
// header, or a row whose identity couldn't be recovered. Those are the rows a
// selection is anchored on and navigation moves between.
//
// A rendering may show every change as a span, such as a `~` row standing for a line
// changed in place, and no addition or deletion of its own. Then the spans are the
// change lines, or there would be nothing to select in it.
func (self *DiffLineHelper) changeLines(view *gocui.View) []bool {
	resolved := self.resolveDiffLines(diffLineContents(view))
	isChange := make([]bool, len(resolved))
	for i, r := range resolved {
		isChange[i] = r.ok && r.info.IsChange()
	}
	if lo.Contains(isChange, true) {
		return isChange
	}
	for i, r := range resolved {
		isChange[i] = r.ok && r.info.Type == types.DiffLineSpan
	}
	return isChange
}

// FirstChangeLineInView returns the view line of the first change line on screen. It
// is where the selection goes when the main view is focused by keyboard: focusing a
// diff you are reading points at something in it without moving it, so the search
// stops at the bottom of the viewport rather than going after a change further down.
// ok is false when the viewport holds no change line — scrolled into a long stretch
// of context, or past the last change.
func (self *DiffLineHelper) FirstChangeLineInView(view *gocui.View) (int, bool) {
	top, bottom, ok := visibleBufferLines(view)
	if !ok {
		return 0, false
	}

	isChange := self.changeLines(view)
	for i := top; i <= min(bottom, len(isChange)-1); i++ {
		if isChange[i] {
			return view.ViewLineForBufferLine(i)
		}
	}
	return 0, false
}

// FirstChangeBlockInView returns the view line of the first change block on screen:
// the first one that *begins* in the viewport, and failing that the one that reaches
// into the viewport from above, whose start is off screen. Hunk mode wants that order
// for the block it offers up on focus: preferably a block whose beginning the user can
// see, rather than the tail of one they have scrolled past the start of. The block
// bleeding in from above is kept as the answer for a change too long to fit on screen,
// where there is no other. ok is false when the viewport shows no change line.
func (self *DiffLineHelper) FirstChangeBlockInView(view *gocui.View) (int, bool) {
	top, bottom, ok := visibleBufferLines(view)
	if !ok {
		return 0, false
	}

	blocks := self.changeBlocks(view, false)
	for i := top; i <= min(bottom, len(blocks)-1); i++ {
		if blocks[i] >= 0 && (i == 0 || blocks[i-1] != blocks[i]) {
			return view.ViewLineForBufferLine(i)
		}
	}
	// A block covering the top line is one that began above it: nothing else can put a
	// change there once no block starts on screen.
	if top < len(blocks) && blocks[top] >= 0 {
		return view.ViewLineForBufferLine(top)
	}
	return 0, false
}

// visibleBufferLines returns the first and last line of view's content that the
// viewport shows any part of, for the queries that only care about what the user can
// see. The last line is the one at the bottom edge, or the content's last when the
// content ends above it. ok is false for a view showing no content at all.
func visibleBufferLines(view *gocui.View) (int, int, bool) {
	top, ok := view.BufferLineForViewLine(view.OriginY())
	if !ok {
		return 0, 0, false
	}

	lastVisible := min(view.OriginY()+view.InnerHeight(), view.ViewLinesHeight()) - 1
	bottom, ok := view.BufferLineForViewLine(lastVisible)
	if !ok {
		return top, top, true
	}
	return top, bottom, true
}

// ViewHasChangeLines reports whether view's rendered diff holds any change line at
// all, i.e. whether there is anything to select. It is false over a non-diff
// placeholder, and over a diff with nothing in it — an empty commit, a binary file —
// which are the cases where the focused main view shows no selection.
func (self *DiffLineHelper) ViewHasChangeLines(view *gocui.View) bool {
	return lo.Contains(self.changeLines(view), true)
}

// IsChangeLine reports whether the given view line of view's rendered diff is a
// change line rather than context, a header, or an unresolvable row — i.e. whether
// pointing at it points at something a patch could be built from.
func (self *DiffLineHelper) IsChangeLine(view *gocui.View, viewLineIdx int) bool {
	info, ok := self.GetDiffLineInfo(view, viewLineIdx)
	return ok && info.IsChange()
}

// IsSingleHunkForWholeFile reports whether the file the given change line belongs to
// is shown as one solid block of changes — every row of its diff a change of the same
// kind, no context — which is what a newly added or deleted file looks like. That is
// the case where widening the selection to the change block would select the file
// entire, so hunk mode drops to a single line there instead.
//
// It says false while the diff is still being read in, since the rows that would
// answer otherwise — a context line, a change of the other kind — may not have
// arrived yet. That errs towards hunk mode, which is what the user asked for.
func (self *DiffLineHelper) IsSingleHunkForWholeFile(view *gocui.View, changeViewLine int) bool {
	if manager := self.c.GetViewBufferManagerForView(view); manager != nil && manager.IsLoading() {
		return false
	}

	anchor, ok := view.BufferLineForViewLine(changeViewLine)
	if !ok {
		return false
	}
	resolved := self.resolveDiffLines(diffLineContents(view))
	if anchor >= len(resolved) || !resolved[anchor].ok {
		return false
	}

	// The question is per file: a commit's diff may hold a newly added file next to an
	// edited one.
	path := resolved[anchor].info.Path
	kind := resolved[anchor].info.Type
	for _, row := range resolved {
		if !row.ok || row.info.Path != path {
			continue
		}
		if row.info.Type == types.DiffLineContext {
			return false
		}
		if row.info.IsChange() && row.info.Type != kind {
			return false
		}
	}
	return true
}

// ChangeBlockBounds returns the inclusive view-line range of the change block to
// select in hunk mode around anchorViewLine. A change block is lazygit's notion of a
// hunk — a run of consecutive added or deleted lines bounded by context, of which a
// single git @@ hunk may hold several. When the anchor is context, the block used is
// the first at or below it, or — with nothing below, the cursor sitting past the last
// change — the nearest above, so that hunk mode always has a block to select. ok is
// false only when the diff holds no change line at all.
func (self *DiffLineHelper) ChangeBlockBounds(view *gocui.View, anchorViewLine int, byEntity bool) (int, int, bool) {
	anchor, ok := view.BufferLineForViewLine(anchorViewLine)
	if !ok {
		return 0, 0, false
	}

	blocks := self.changeBlocks(view, byEntity)
	start := anchor
	for start < len(blocks) && blocks[start] < 0 {
		start++
	}
	if start >= len(blocks) {
		for start = min(anchor, len(blocks)-1); start >= 0 && blocks[start] < 0; start-- {
		}
		if start < 0 {
			return 0, 0, false
		}
	}
	end := start
	for start > 0 && blocks[start-1] == blocks[start] {
		start--
	}
	for end < len(blocks)-1 && blocks[end+1] == blocks[end] {
		end++
	}

	startView, startOk := view.ViewLineForBufferLine(start)
	// The block's last line goes to its last view line, so that a line the view
	// wrapped is highlighted to its end rather than only where it begins.
	endView, endOk := view.LastViewLineForBufferLine(end)
	if !startOk || !endOk {
		return 0, 0, false
	}
	return startView, endView, true
}

// AdjacentChangeBlock returns the view line to move to for next/previous change-block
// navigation in view's rendered diff, starting from anchorViewLine. A change block is
// lazygit's notion of a hunk (see ChangeBlockBounds). forward=true targets the start
// of the next block, forward=false the start of the previous one — from mid-block that
// means the previous block, rather than the one we are in. ok is false when there's no
// further block, so the caller leaves the view where it is.
func (self *DiffLineHelper) AdjacentChangeBlock(
	view *gocui.View, anchorViewLine int, forward bool, byEntity bool,
) (int, bool) {
	anchor, ok := view.BufferLineForViewLine(anchorViewLine)
	if !ok {
		return 0, false
	}

	target, ok := changeBlockStart(self.changeBlocks(view, byEntity), anchor, forward)
	if !ok {
		return 0, false
	}
	return view.ViewLineForBufferLine(target)
}

// AdjacentFile returns the view line to move to for next/previous file navigation in
// view's (possibly multi-file) rendered diff, starting from anchorViewLine: the first
// located row of the neighbouring file, found where the rows' file changes. ok is
// false at the first or last file.
func (self *DiffLineHelper) AdjacentFile(view *gocui.View, anchorViewLine int, forward bool) (int, bool) {
	anchor, ok := view.BufferLineForViewLine(anchorViewLine)
	if !ok {
		return 0, false
	}

	target, ok := fileStart(self.filePaths(view), anchor, forward)
	if !ok {
		return 0, false
	}
	return view.ViewLineForBufferLine(target)
}

// FilesInDiff lists the files of view's (possibly multi-file) rendered diff, in the
// order it shows them, by the paths of the repo's files. It is what a menu offering to
// jump between them is built from; jumping to one of them goes to StartOfFileInDiff.
func (self *DiffLineHelper) FilesInDiff(view *gocui.View) []string {
	return lo.Map(fileStarts(self.filePaths(view)),
		func(start diffFileStart, _ int) string { return start.path })
}

// StartOfFileInDiff returns the view line the given file's section of view's rendered
// diff begins at. That is the row file navigation lands on, so jumping to a file from a
// menu and stepping to it with next-file land in the same place. ok is false for a file
// the diff doesn't show, e.g. because it was re-rendered since the file was listed.
func (self *DiffLineHelper) StartOfFileInDiff(view *gocui.View, path string) (int, bool) {
	start, ok := lo.Find(fileStarts(self.filePaths(view)), func(start diffFileStart) bool {
		return start.path == path
	})
	if !ok {
		return 0, false
	}
	return view.ViewLineForBufferLine(start.row)
}

// filePaths resolves view's rendered diff to the path each buffer line belongs to, in
// the repo's terms, and empty for a row whose identity couldn't be recovered. Naming the
// files the way the rest of the queries name them means a row of the custom patch's
// preview belongs to the repo's file rather than to the copy of it in the tree the patch
// was materialized into, so that both halves of a change belong to the same file however
// the diff renderer states them.
func (self *DiffLineHelper) filePaths(view *gocui.View) []string {
	resolved := self.resolveDiffLines(diffLineContents(view))
	toRepoTerms := self.repoTermsMapper(view)
	paths := make([]string, len(resolved))
	for i, row := range resolved {
		if row.ok {
			paths[i] = toRepoTerms(row.info.Path)
		}
	}
	return paths
}

// diffFileStart is where one file of a diff begins: the path of the file, and the row
// of the diff its section starts at.
type diffFileStart struct {
	path string
	row  int
}

// fileStarts finds, in a diff whose lines carry the file path they belong to (empty for
// a row no backend could place), where each file of it begins, in the order the diff
// shows them.
//
// A file is identified by its path, and the rows showing it are consecutive, so a path
// differing from the one before it begins a file. Rows carrying no path are passed over:
// those are the blank separator rows between files, or the header rows of a diff
// renderer that doesn't state which file its headers belong to. So a file begins at its
// header wherever the source says so (a parseable buffer, or a renderer that tags its
// headers), and at its first content line otherwise, which is an accepted degradation.
func fileStarts(paths []string) []diffFileStart {
	starts := []diffFileStart{}
	previousPath := ""
	for row, path := range paths {
		if path == "" || path == previousPath {
			continue
		}
		previousPath = path
		starts = append(starts, diffFileStart{path: path, row: row})
	}
	return starts
}

// fileStart returns where the file adjacent to `from` in the given direction begins —
// the row file navigation lands on. It is the pure index arithmetic behind AdjacentFile.
// ok is false at the first or last file of the diff.
func fileStart(paths []string, from int, forward bool) (int, bool) {
	anchorPath, ok := anchorFilePath(paths, from)
	if !ok {
		return 0, false
	}

	starts := fileStarts(paths)
	_, anchor, ok := lo.FindIndexOf(starts, func(start diffFileStart) bool {
		return start.path == anchorPath
	})
	if !ok {
		return 0, false
	}

	target := anchor + lo.Ternary(forward, 1, -1)
	if target < 0 || target >= len(starts) {
		return 0, false
	}
	return starts[target].row, true
}

// anchorFilePath returns the path of the file the anchor sits in: the first row at or
// below it that carries a path — the file whose content is at or below the top of the
// view — falling back to the nearest above when there is nothing below. Scanning down
// first matters because the anchor is often a file-header row that carries no path of
// its own, whose nearest tagged row above is the *previous* file's content; taking
// that would make next-file navigation jump back into the file just left, so a second
// press wouldn't advance. ok is false when no row carries a path.
func anchorFilePath(paths []string, from int) (string, bool) {
	if from < 0 {
		return "", false
	}
	for i := from; i < len(paths); i++ {
		if paths[i] != "" {
			return paths[i], true
		}
	}
	for i := min(from, len(paths)) - 1; i >= 0; i-- {
		if paths[i] != "" {
			return paths[i], true
		}
	}
	return "", false
}

// changeBlockStart finds, in a diff whose lines are flagged by the change block they are
// in (see changeBlocks), the first line of the change block adjacent to `from` in the
// given direction. It is the pure index arithmetic behind AdjacentChangeBlock.
func changeBlockStart(blocks []int, from int, forward bool) (int, bool) {
	if from < 0 || from >= len(blocks) {
		return 0, false
	}

	current := blocks[from]
	if forward {
		i := from
		for i < len(blocks) && current >= 0 && blocks[i] == current { // leave the current block
			i++
		}
		for i < len(blocks) && blocks[i] < 0 { // skip the separating context
			i++
		}
		if i == len(blocks) {
			return 0, false
		}
		return i, true
	}

	i := from
	for i >= 0 && current >= 0 && blocks[i] == current { // leave the current block
		i--
	}
	for i >= 0 && blocks[i] < 0 { // skip context, landing on the previous block's last line
		i--
	}
	if i < 0 {
		return 0, false
	}
	for i > 0 && blocks[i-1] == blocks[i] { // walk back to that block's first line
		i--
	}
	return i, true
}

// hasEntityRecords reports whether the renderer states which entity rows belong to,
// which is what lets a change block be a change unit or an entity.
func hasEntityRecords(rows []resolvedDiffLine) bool {
	return lo.SomeBy(rows, func(row resolvedDiffLine) bool { return row.ok && row.info.EntityID != "" })
}

// changeBlocks flags every buffer line of view's rendered diff with the change block it
// is in, or -1 for none. Where the renderer states entities, a block is a change unit
// (or with byEntity an entity); otherwise it is a run of change lines.
func (self *DiffLineHelper) changeBlocks(view *gocui.View, byEntity bool) []int {
	rows := self.resolveDiffLines(diffLineContents(view))
	if hasEntityRecords(rows) {
		return blocksOfRows(rows, byEntity)
	}
	blocks := make([]int, len(rows))
	next, inRun := 0, false
	for i, isChange := range self.changeLines(view) {
		switch {
		case !isChange:
			blocks[i] = -1
			inRun = false
		default:
			if !inRun {
				next++
			}
			blocks[i] = next - 1
			inRun = true
		}
	}
	return blocks
}

// blocksOfRows is changeBlocks for rows that state their entities.
//
// A unit is the rows that share an entity and a unit number. Rows of an entity that are
// in no unit — a collapsed entity, an added or deleted one — are one block per run of
// them, and so is any other change the renderer shows without an entity, except that a
// span, such as a file's summary, is a block by itself. A header that stands in for the
// units under it is no block of its own: staging it is staging the entity. With
// byEntity every row of an entity, header and context included, is one block.
//
// Rows between the first and last of a block, such as a line of unchanged-line count,
// belong to it, so that stepping from block to block never lands inside one.
func blocksOfRows(rows []resolvedDiffLine, byEntity bool) []int {
	keys := make([]string, len(rows))
	runs := 0
	for i, row := range rows {
		info := row.info
		isChange := row.ok && (info.IsChange() || info.Type == types.DiffLineSpan)
		switch {
		case !row.ok:
		case info.EntityID != "" && byEntity:
			keys[i] = fmt.Sprintf("e\x00%s\x00%s", info.Path, info.EntityID)
		case info.EntityID != "" && info.Unit > 0:
			keys[i] = fmt.Sprintf("u\x00%s\x00%s\x00%d", info.Path, info.EntityID, info.Unit)
		case !isChange:
		case info.EntityID == "" && info.Type == types.DiffLineSpan:
			runs++
			keys[i] = fmt.Sprintf("s%d", runs)
		default:
			previous := rows[max(i-1, 0)].info
			if i == 0 || keys[i-1] == "" || keys[i-1][0] != 'r' ||
				previous.Path != info.Path || previous.EntityID != info.EntityID {
				runs++
			}
			keys[i] = fmt.Sprintf("r%d", runs)
		}
	}
	if !byEntity {
		dropStandInHeaders(rows, keys)
	}

	ids := map[string]int{}
	first, last := map[string]int{}, map[string]int{}
	for i, key := range keys {
		if key == "" {
			continue
		}
		if _, seen := ids[key]; !seen {
			ids[key] = len(ids)
			first[key] = i
		}
		last[key] = i
	}
	blocks := make([]int, len(rows))
	for i, key := range keys {
		blocks[i] = -1
		if key != "" {
			blocks[i] = ids[key]
		}
	}
	filled := slices.Clone(blocks)
	for key, id := range ids {
		for i := first[key]; i <= last[key]; i++ {
			if blocks[i] == -1 && (filled[i] == -1 || id < filled[i]) {
				filled[i] = id
			}
		}
	}
	return filled
}

// dropStandInHeaders clears the keys of runs that are nothing but spans of an entity
// whose next row is in one of its units: that is an entity's header, a stand-in for the
// units beneath it.
func dropStandInHeaders(rows []resolvedDiffLine, keys []string) {
	var end int
	for start := 0; start < len(rows); start = end + 1 {
		end = start
		for end+1 < len(rows) && keys[end+1] == keys[start] {
			end++
		}
		onlySpans := keys[start] != "" && keys[start][0] == 'r'
		for i := start; i <= end; i++ {
			onlySpans = onlySpans && rows[i].info.Type == types.DiffLineSpan && rows[i].info.EntityID != ""
		}
		// Lines of context, and one counting unchanged lines, may stand between.
		next := end + 1
		for next < len(rows) && (!rows[next].ok || rows[next].info.Type == types.DiffLineContext && rows[next].info.Unit == 0) {
			next++
		}
		if onlySpans && next < len(rows) {
			next, current := rows[next], rows[end].info
			if next.ok && next.info.Path == current.Path && next.info.EntityID == current.EntityID && next.info.Unit > 0 {
				for i := start; i <= end; i++ {
					keys[i] = ""
				}
			}
		}
	}
}
