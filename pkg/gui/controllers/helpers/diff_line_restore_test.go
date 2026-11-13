package helpers

import (
	"testing"

	"github.com/jesseduffield/lazygit/pkg/gui/types"
	"github.com/stretchr/testify/assert"
)

func TestStepRows(t *testing.T) {
	row := func(lineType types.DiffLineType, entityID string, unit int) resolvedDiffLine {
		return resolvedDiffLine{info: types.DiffLineInfo{Type: lineType, Path: "f", EntityID: entityID, Unit: unit}, ok: true}
	}
	rows := []resolvedDiffLine{
		row(types.DiffLineSpan, "a", 0),    // header of an entity with rows under it
		row(types.DiffLineDeleted, "a", 1), // change
		row(types.DiffLineContext, "a", 0), // context
		row(types.DiffLineSpan, "a", 2),    // a `~` unit
		{},                                 // a row without identity
		row(types.DiffLineSpan, "b", 0),    // collapsed entity: nothing under it
		row(types.DiffLineSpan, "c", 0),    // header of c...
		row(types.DiffLineAdded, "c", 0),   // ...whose body is an added line
	}

	assert.Equal(t, []bool{false, true, false, true, false, true, false, true}, stepRows(rows, stepLinesAndSpans))
	assert.Equal(t, []bool{false, true, false, false, false, false, false, true}, stepRows(rows, stepChangeLines))
}

func TestBlocksOfRows(t *testing.T) {
	row := func(lineType types.DiffLineType, entityID string, unit int) resolvedDiffLine {
		return resolvedDiffLine{info: types.DiffLineInfo{Type: lineType, Path: "f", EntityID: entityID, Unit: unit}, ok: true}
	}
	rows := []resolvedDiffLine{
		row(types.DiffLineSpan, "a", 0),    // 0 header with units under it
		row(types.DiffLineSpan, "a", 1),    // 1 `~` unit
		row(types.DiffLineDeleted, "a", 2), // 2 unit 2
		row(types.DiffLineContext, "a", 2), // 3 context in unit 2
		{},                                 // 4 row without identity
		row(types.DiffLineSpan, "b", 0),    // 5 collapsed entity
		row(types.DiffLineSpan, "c", 0),    // 6 header of c...
		row(types.DiffLineAdded, "c", 0),   // 7 ...whose body is an added line
		row(types.DiffLineSpan, "", 0),     // 8 file summary
		row(types.DiffLineContext, "c", 0), // 9 context stands for nothing
	}

	assert.Equal(t, []int{-1, 0, 1, 1, -1, 2, 3, 3, 4, -1}, blocksOfRows(rows, false))
	assert.Equal(t, []int{0, 0, 0, 0, -1, 1, 2, 2, 3, 2}, blocksOfRows(rows, true))
}

func TestBlocksOfRowsHeaderBeforeUnchangedLines(t *testing.T) {
	rows := []resolvedDiffLine{
		{info: types.DiffLineInfo{Type: types.DiffLineSpan, Path: "f", EntityID: "a"}, ok: true},
		{}, // a line counting unchanged lines
		{info: types.DiffLineInfo{Type: types.DiffLineDeleted, Path: "f", EntityID: "a", Unit: 1}, ok: true},
	}

	assert.Equal(t, []int{-1, -1, 0}, blocksOfRows(rows, false))
}
