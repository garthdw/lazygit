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

	assert.Equal(t, []bool{false, true, false, true, false, true, false, true}, stepRows(rows, true))
	assert.Equal(t, []bool{false, true, false, false, false, false, false, true}, stepRows(rows, false))
}
